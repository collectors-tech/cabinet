param(
  [string]$PackagePath = "",
  [string]$Version = "",
  [int]$Port = 19833,
  [int]$MockOpenAIPort = 19834,
  [string]$LogRoot = ".work-agent\logs\issue-1933-chat-agent-planner-packaged-smoke"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $repoRoot "scripts\lib\cabinet-console.ps1")

function Invoke-CabinetJson {
  param(
    [string]$Method,
    [string]$Url,
    [object]$Body = $null
  )
  $headers = @{ "Content-Type" = "application/json" }
  if ($null -eq $Body) {
    return Invoke-RestMethod -Method $Method -Uri $Url -Headers $headers
  }
  $json = if ($Body -is [string]) { $Body } else { $Body | ConvertTo-Json -Depth 20 -Compress }
  return Invoke-RestMethod -Method $Method -Uri $Url -Headers $headers -Body $json
}

function Wait-CabinetEndpoint {
  param([string]$Url, [int]$Seconds = 45)
  $deadline = (Get-Date).AddSeconds($Seconds)
  do {
    try {
      Invoke-RestMethod -Method Get -Uri $Url | Out-Null
      return
    } catch {
      Start-Sleep -Milliseconds 500
    }
  } while ((Get-Date) -lt $deadline)
  throw "Timed out waiting for $Url"
}

Write-CabinetBanner -Command "chat-agent-planner-packaged-smoke" -Summary "Run #1933 packaged Windows planner smoke."

$resolvedLogRoot = Resolve-Path -LiteralPath (New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot $LogRoot)).FullName
$runID = Get-Date -Format "yyyyMMdd-HHmmss"
$runDir = Join-Path $resolvedLogRoot $runID
New-Item -ItemType Directory -Force -Path $runDir | Out-Null

if ([string]::IsNullOrWhiteSpace($PackagePath)) {
  & (Join-Path $repoRoot "scripts\package-installers.ps1") -Version $Version
  if ($LASTEXITCODE -ne 0) {
    throw "package-installers failed with exit code $LASTEXITCODE"
  }
  if ([string]::IsNullOrWhiteSpace($Version)) {
    $versionPayload = Get-Content -LiteralPath (Join-Path $repoRoot "release\cabinet-beta-version.json") -Raw | ConvertFrom-Json
    $Version = [string]$versionPayload.version
  }
  $PackagePath = Join-Path $repoRoot "dist\cabinet-$Version-windows-amd64-portable.zip"
}

$packageFullPath = (Resolve-Path -LiteralPath $PackagePath).Path
$extractDir = Join-Path $runDir "package"
$dataDir = Join-Path $runDir "data"
$runtimeLog = Join-Path $runDir "cabinet-runtime.log"
$runtimeErrorLog = Join-Path $runDir "cabinet-runtime.err.log"
$mockLog = Join-Path $runDir "mock-openai.log"
$summaryPath = Join-Path $runDir "summary.json"
Expand-Archive -LiteralPath $packageFullPath -DestinationPath $extractDir -Force
$cabinetExe = Join-Path $extractDir "cabinet.exe"
if (-not (Test-Path -LiteralPath $cabinetExe)) {
  throw "Package did not contain cabinet.exe at $cabinetExe"
}

$mockJob = Start-Job -ArgumentList $MockOpenAIPort, $mockLog -ScriptBlock {
  param($Port, $LogPath)
  $ErrorActionPreference = "Stop"
  $listener = [System.Net.HttpListener]::new()
  $listener.Prefixes.Add("http://127.0.0.1:$Port/")
  $listener.Start()
  try {
    while ($listener.IsListening) {
      $ctx = $listener.GetContext()
      $path = $ctx.Request.Url.AbsolutePath
      $body = ""
      if ($ctx.Request.HasEntityBody) {
        $reader = [System.IO.StreamReader]::new($ctx.Request.InputStream, $ctx.Request.ContentEncoding)
        $body = $reader.ReadToEnd()
        $reader.Close()
      }
      Add-Content -LiteralPath $LogPath -Value "$($ctx.Request.HttpMethod) $path $body"
      $content = '{"choices":[{"message":{"content":"{\"decision\":\"select_skill\",\"skill_id\":\"cabinet.inventory.search_items\",\"parameters\":{\"query\":\"SMOKE-READ-1933\"},\"message\":\"Searching inventory for SMOKE-READ-1933.\"}"}}]}'
      if ($path -eq "/v1/models") {
        $content = '{"data":[{"id":"gpt-4o-mini"}]}'
      } elseif ($body -like "*SMOKE-WRITE-1933*") {
        $content = '{"choices":[{"message":{"content":"{\"decision\":\"select_skill\",\"skill_id\":\"cabinet.inventory.create_item\",\"parameters\":{\"part_number\":\"SMOKE-WRITE-1933\",\"title\":\"Packaged Planner Write Item\",\"brand\":\"AFX\",\"category\":\"Slot\"},\"message\":\"I prepared a preview for SMOKE-WRITE-1933.\"}"}}]}'
      }
      $bytes = [System.Text.Encoding]::UTF8.GetBytes($content)
      $ctx.Response.ContentType = "application/json"
      $ctx.Response.StatusCode = 200
      $ctx.Response.OutputStream.Write($bytes, 0, $bytes.Length)
      $ctx.Response.Close()
    }
  } finally {
    $listener.Close()
  }
}

$previousEnv = @{
  CABINET_DATA_DIR = $env:CABINET_DATA_DIR
  CABINET_PORT = $env:CABINET_PORT
  CABINET_OPEN_BROWSER = $env:CABINET_OPEN_BROWSER
  CABINET_ALLOW_PARALLEL = $env:CABINET_ALLOW_PARALLEL
  CABINET_ALLOW_INSECURE_SECRET_FALLBACK = $env:CABINET_ALLOW_INSECURE_SECRET_FALLBACK
}
$cabinetProcess = $null
try {
  Wait-CabinetEndpoint -Url "http://127.0.0.1:$MockOpenAIPort/v1/models" -Seconds 15

  $env:CABINET_DATA_DIR = $dataDir
  $env:CABINET_PORT = [string]$Port
  $env:CABINET_OPEN_BROWSER = "false"
  $env:CABINET_ALLOW_PARALLEL = "true"
  $env:CABINET_ALLOW_INSECURE_SECRET_FALLBACK = "1"
  $cabinetProcess = Start-Process -FilePath $cabinetExe -ArgumentList "--no-open-browser" -PassThru -NoNewWindow -RedirectStandardOutput $runtimeLog -RedirectStandardError $runtimeErrorLog
  $baseUrl = "http://127.0.0.1:$Port"
  Wait-CabinetEndpoint -Url "$baseUrl/healthz" -Seconds 60

  $runtime = Invoke-CabinetJson -Method Get -Url "$baseUrl/api/runtime"
  $profile = Invoke-CabinetJson -Method Post -Url "$baseUrl/api/profiles" -Body @{ name = "Issue 1933 packaged smoke" }
  $profileID = $profile.id
  Invoke-CabinetJson -Method Put -Url "$baseUrl/api/profiles/active" -Body @{ profile_id = $profileID } | Out-Null
  $thread = Invoke-CabinetJson -Method Post -Url "$baseUrl/api/chat/threads" -Body @{ profile_id = $profileID; title = "Packaged planner smoke" }
  $threadID = $thread.id

  Invoke-CabinetJson -Method Post -Url "$baseUrl/api/profiles/$profileID/integration-instances" -Body @{
    provider_id = "openai"
    display_name = "OpenAI packaged smoke"
    enabled = $true
    config = @{
      "openai.active_auth_method" = "api_key"
      "assistant_default_model" = "gpt-4o-mini"
      "base_url" = "http://127.0.0.1:$MockOpenAIPort"
    }
    secrets = @{ "openai_api_key" = "sk-packaged-smoke" }
    auth_state = "configured"
    health_state = "ready"
  } | Out-Null

  Invoke-CabinetJson -Method Post -Url "$baseUrl/api/items" -Body @{
    part_number = "SMOKE-READ-1933"
    title = "Packaged Planner Read Item"
    brand = "AFX"
    category = "Slot"
  } | Out-Null

  $readResp = Invoke-CabinetJson -Method Post -Url "$baseUrl/api/chat/messages" -Body @{
    profile_id = $profileID
    thread_id = $threadID
    role = "user"
    content = "Find the item with part number SMOKE-READ-1933"
    context = @{
      route = @{ pathname = "/chats" }
      assistant = @{ provider = "openai"; model = "gpt-4o-mini" }
    }
  }
  $readJson = $readResp | ConvertTo-Json -Depth 30 -Compress
  if ($readJson -notlike "*agent_planner*" -or $readJson -notlike "*Packaged Planner Read Item*" -or $readJson -notlike "*execution_result*") {
    throw "Provider-backed conversational read did not return grounded planner execution evidence: $readJson"
  }

  $writeResp = Invoke-CabinetJson -Method Post -Url "$baseUrl/api/chat/messages" -Body @{
    profile_id = $profileID
    thread_id = $threadID
    role = "user"
    content = "Add catalog record for part number SMOKE-WRITE-1933 titled Packaged Planner Write Item"
    context = @{
      route = @{ pathname = "/chats" }
      assistant = @{ provider = "openai"; model = "gpt-4o-mini" }
    }
  }
  $preview = $writeResp.agent_planner.preview_result
  if ($null -eq $preview -or [string]::IsNullOrWhiteSpace([string]$preview.preview_id) -or $preview.mutation_applied -ne $false) {
    throw "Local-write planner request did not create a preview-only result: $($writeResp | ConvertTo-Json -Depth 30 -Compress)"
  }

  $beforeApply = Invoke-CabinetJson -Method Get -Url "$baseUrl/api/items?profile_id=$profileID"
  if (($beforeApply | ConvertTo-Json -Depth 20 -Compress) -like "*SMOKE-WRITE-1933*") {
    throw "Planner preview mutated inventory before confirmation"
  }

  $applyResp = Invoke-CabinetJson -Method Post -Url "$baseUrl/api/chat/actions/apply" -Body @{
    profile_id = $profileID
    thread_id = $threadID
    preview_id = $preview.preview_id
    confirm = $true
  }
  if ($applyResp.applied -ne $true) {
    throw "Confirmed local write did not apply: $($applyResp | ConvertTo-Json -Depth 20 -Compress)"
  }
  try {
    Invoke-CabinetJson -Method Post -Url "$baseUrl/api/chat/actions/apply" -Body @{
      profile_id = $profileID
      thread_id = $threadID
      preview_id = $preview.preview_id
      confirm = $true
    } | Out-Null
    throw "Replay apply unexpectedly succeeded"
  } catch {
    if ($_.Exception.Message -like "*Replay apply unexpectedly succeeded*") {
      throw
    }
  }

  $afterApply = Invoke-CabinetJson -Method Get -Url "$baseUrl/api/items?profile_id=$profileID"
  $afterJson = $afterApply | ConvertTo-Json -Depth 20 -Compress
  if ($afterJson -notlike "*SMOKE-WRITE-1933*" -or $afterJson -notlike "*Packaged Planner Write Item*") {
    throw "Confirmed local write item was not visible after apply: $afterJson"
  }

  $summary = [ordered]@{
    status = "passed"
    issue = 1933
    package_path = $packageFullPath
    package_sha256 = (Get-FileHash -LiteralPath $packageFullPath -Algorithm SHA256).Hash.ToLowerInvariant()
    runtime_app_version = $runtime.app_version
    runtime_build_date = $runtime.build_date
    runtime_port = $runtime.runtime_port
    data_dir = $dataDir
    profile_id = $profileID
    thread_id = $threadID
    provider_base_url = "http://127.0.0.1:$MockOpenAIPort"
    provider_backed_read = "passed"
    confirmed_local_write = "passed"
    replay_idempotency = "passed"
    preview_id = $preview.preview_id
    runtime_log = $runtimeLog
    runtime_error_log = $runtimeErrorLog
    mock_openai_log = $mockLog
  }
  $summary | ConvertTo-Json -Depth 20 | Set-Content -LiteralPath $summaryPath -Encoding utf8
  Write-CabinetStatus -State "ok" -Message "Packaged Chat planner smoke passed."
  Write-CabinetKeyValue -Key "Summary" -Value $summaryPath
} finally {
  try {
    Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:$Port/api/runtime/shutdown" -ContentType "application/json" -Body '{"reason":"issue-1933-packaged-smoke-complete"}' | Out-Null
  } catch {}
  if ($cabinetProcess -and -not $cabinetProcess.HasExited) {
    if (-not $cabinetProcess.WaitForExit(5000)) {
      Stop-Process -Id $cabinetProcess.Id -Force -ErrorAction SilentlyContinue
    }
  }
  Stop-Job -Job $mockJob -ErrorAction SilentlyContinue | Out-Null
  Remove-Job -Job $mockJob -Force -ErrorAction SilentlyContinue | Out-Null
  foreach ($key in $previousEnv.Keys) {
    Set-Item -Path "Env:$key" -Value $previousEnv[$key] -ErrorAction SilentlyContinue
  }
}
