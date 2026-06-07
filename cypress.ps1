param(
  [string]$Spec = "cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts",
  [string]$Browser = "electron",
  [switch]$Headed,
  [switch]$NoServer,
  [switch]$ReuseServer,
  [switch]$RequireE2EHooks,
  [switch]$ApiContractSmoke,
  [string]$RuntimeExecutablePath = "",
  [switch]$AllowTempRuntimePath,
  [switch]$SkipDependencyPrep,
  [switch]$SkipRuntimeBuild,
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [int]$StartupTimeoutSec = 45,
  [string]$LogDir = ".work-agent\logs\cypress",
  [string]$LogName = ""
)

$ErrorActionPreference = "Stop"
$script:CypressStepEvents = @()

function Write-Step([string]$msg) {
  $timestamp = (Get-Date).ToString("o")
  $script:CypressStepEvents += [pscustomobject]@{
    timestamp = $timestamp
    message = $msg
  }
  Write-Host "[$timestamp] [cypress.ps1] $msg"
}

function ConvertTo-SafeLogSegment([string]$value) {
  if ([string]::IsNullOrWhiteSpace($value)) {
    return "run"
  }

  $safe = $value -replace '[^A-Za-z0-9._-]+', '-'
  $safe = $safe.Trim('-')
  if ([string]::IsNullOrWhiteSpace($safe)) {
    return "run"
  }
  if ($safe.Length -gt 80) {
    return $safe.Substring(0, 80)
  }
  return $safe
}

function Test-Health([string]$url) {
  try {
    $res = Invoke-WebRequest -Uri "$url/healthz" -UseBasicParsing -TimeoutSec 2
    return $res.StatusCode -eq 200
  }
  catch {
    return $false
  }
}

function Test-E2EHooks([string]$url) {
  try {
    $res = Invoke-WebRequest -Uri "$url/api/test/reset" -Method Post -Body "{}" -ContentType "application/json" -UseBasicParsing -TimeoutSec 2
    return $res.StatusCode -ge 200 -and $res.StatusCode -lt 300
  }
  catch {
    return $false
  }
}

function Test-AppEndpoint([string]$url, [string]$path) {
  try {
    $res = Invoke-WebRequest -Uri "$url$path" -UseBasicParsing -TimeoutSec 5
    return $res.StatusCode -ge 200 -and $res.StatusCode -lt 300
  }
  catch {
    return $false
  }
}

function Assert-AppPreflight([string]$url) {
  $checks = @(
    @{ Name = "healthz"; Path = "/healthz" },
    @{ Name = "runtime API"; Path = "/api/runtime" },
    @{ Name = "sign-in route"; Path = "/sign-in" }
  )

  foreach ($check in $checks) {
    if (-not (Test-AppEndpoint $url $check.Path)) {
      throw "Server preflight failed: $($check.Name) did not return 2xx at $url$($check.Path)."
    }
    Write-Step "Server preflight passed: $($check.Name) at $($check.Path)."
  }
}

function Get-SourceCommit([string]$repoRoot) {
  try {
    $commit = & git -C $repoRoot rev-parse HEAD
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($commit)) {
      return $commit.Trim()
    }
  }
  catch {
    return ""
  }
  return ""
}

function Invoke-ApiContractSmoke([string]$repoRoot, [string]$url, [string]$logDir, [bool]$requireE2EHooks) {
  $args = @(
    "-NoLogo",
    "-NoProfile",
    "-File", (Join-Path $repoRoot "scripts\run-api-contract-smoke.ps1"),
    "-BaseUrl", $url,
    "-LogRoot", $logDir,
    "-RunId", "cypress-preflight-$runStamp"
  )
  if ($requireE2EHooks) {
    $args += "-RequireE2EHooks"
  }

  Write-Step "Running API contract smoke preflight."
  & pwsh @args | ForEach-Object {
    Write-Host $_
  }
  if ($LASTEXITCODE -ne 0) {
    throw "API contract smoke preflight failed with exit code $LASTEXITCODE."
  }
}

function Stop-PortListener([string]$url) {
  $uri = [Uri]$url
  $port = $uri.Port
  try {
    $listeners = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction Stop
    foreach ($listener in $listeners) {
      if ($listener.OwningProcess -gt 0) {
        Write-Step "Stopping process $($listener.OwningProcess) on port $port"
        Stop-Process -Id $listener.OwningProcess -Force -ErrorAction SilentlyContinue
      }
    }
  }
  catch {
    Write-Step "Unable to inspect port listeners for $port; continuing."
  }
}

function Test-IsEphemeralRuntimePath([string]$path) {
  if ([string]::IsNullOrWhiteSpace($path)) {
    return $false
  }
  $normalized = $path.ToLowerInvariant().Replace('/', '\')
  $markers = @(
    '\appdata\local\temp\',
    '\temp\',
    '\tmp\',
    '\.tmp\',
    '\template\',
    '\templates\'
  )
  foreach ($marker in $markers) {
    if ($normalized.Contains($marker)) {
      return $true
    }
  }
  return $false
}

function Test-CypressDependencyReady([string]$uiRoot) {
  $nodeModules = Join-Path $uiRoot "node_modules"
  if (-not (Test-Path $nodeModules)) {
    return $false
  }

  $cypressPackage = Join-Path $nodeModules "cypress\package.json"
  $cypressCmd = Join-Path $nodeModules ".bin\cypress.cmd"
  $cypressBin = Join-Path $nodeModules ".bin\cypress"
  return (Test-Path $cypressPackage) -and ((Test-Path $cypressCmd) -or (Test-Path $cypressBin))
}

function Get-ReusableNodeModulesPath([string]$repoRoot, [string]$uiRoot) {
  $candidates = @()

  if (-not [string]::IsNullOrWhiteSpace($env:CABINET_UI_NODE_MODULES_PATH)) {
    $candidates += $env:CABINET_UI_NODE_MODULES_PATH
  }

  $repoParent = Split-Path -Parent $repoRoot
  if ($repoParent) {
    $candidates += (Join-Path $repoParent "cabinet\ui.web\node_modules")

    $repoGrandParent = Split-Path -Parent $repoParent
    if ($repoGrandParent) {
      $candidates += (Join-Path $repoGrandParent "cabinet\ui.web\node_modules")
    }
  }

  $currentNodeModules = Join-Path $uiRoot "node_modules"
  foreach ($candidate in $candidates) {
    if ([string]::IsNullOrWhiteSpace($candidate)) {
      continue
    }
    if ($candidate -eq $currentNodeModules) {
      continue
    }
    if (Test-CypressDependencyReady (Split-Path -Parent $candidate)) {
      return (Resolve-Path $candidate).Path
    }
  }

  return ""
}

function Invoke-NpmCi([string]$uiRoot) {
  $lockPath = Join-Path $uiRoot "package-lock.json"
  if (-not (Test-Path $lockPath)) {
    throw "Cannot install UI dependencies: missing $lockPath"
  }

  Write-Step "Installing UI dependencies with npm ci."
  Push-Location $uiRoot
  try {
    & npm ci
    if ($LASTEXITCODE -ne 0) {
      throw "npm ci failed with exit code $LASTEXITCODE."
    }
  }
  finally {
    Pop-Location
  }
}

function Ensure-UiDependencies([string]$repoRoot, [string]$uiRoot) {
  if (Test-CypressDependencyReady $uiRoot) {
    Write-Step "UI dependencies already prepared."
    return
  }

  $nodeModules = Join-Path $uiRoot "node_modules"
  if (Test-Path $nodeModules) {
    Write-Step "UI node_modules exists but Cypress is not ready; reinstalling dependencies."
    Remove-Item -LiteralPath $nodeModules -Recurse -Force
    Invoke-NpmCi $uiRoot
  }
  else {
    $reusableNodeModules = Get-ReusableNodeModulesPath $repoRoot $uiRoot
    if (-not [string]::IsNullOrWhiteSpace($reusableNodeModules)) {
      Write-Step "Linking UI dependencies from $reusableNodeModules"
      New-Item -ItemType Junction -Path $nodeModules -Target $reusableNodeModules | Out-Null
    }
    else {
      Invoke-NpmCi $uiRoot
    }
  }

  if (-not (Test-CypressDependencyReady $uiRoot)) {
    throw "UI dependency prep completed, but Cypress is still not ready under $nodeModules."
  }
}

function Ensure-RuntimeExecutable([string]$repoRoot, [string]$runtimeExecutablePath, [bool]$skipRuntimeBuild) {
  if ($skipRuntimeBuild -and (Test-Path $runtimeExecutablePath)) {
    Write-Step "Runtime build skipped; using existing $runtimeExecutablePath"
    return (Resolve-Path $runtimeExecutablePath).Path
  }

  Write-Step "Building static UI and project-local runtime executable at $runtimeExecutablePath"
  Push-Location $repoRoot
  try {
    & (Join-Path $repoRoot "scripts\build-cabinet.ps1") | ForEach-Object {
      Write-Host $_
    }
    if ($LASTEXITCODE -ne 0) {
      throw "build-cabinet.ps1 failed with exit code $LASTEXITCODE."
    }
  }
  finally {
    Pop-Location
  }

  if (-not (Test-Path $runtimeExecutablePath)) {
    throw "Runtime build completed without producing $runtimeExecutablePath"
  }
  return (Resolve-Path $runtimeExecutablePath).Path
}

function Write-RunSummary(
  [string]$summaryPath,
  [int]$exitCode,
  [string]$errorMessage,
  [string]$repoRoot,
  [string]$uiRoot,
  [string]$specPath,
  [string]$browser,
  [string]$baseUrl,
  [int]$runtimePort,
  [string]$runtimeDataDir,
  [string]$runtimeProfile,
  [string]$runtimeInstanceName,
  [string]$runtimeExecutablePath,
  [string]$sourceCommit,
  [string]$logPath,
  [bool]$startedServer
) {
  $summary = [ordered]@{
    timestamp = (Get-Date).ToString("o")
    exit_code = $exitCode
    error = $errorMessage
    repo_root = $repoRoot
    ui_root = $uiRoot
    spec = $specPath
    browser = $browser
    base_url = $baseUrl
    runtime_port = $runtimePort
    runtime_data_dir = $runtimeDataDir
    runtime_profile = $runtimeProfile
    runtime_instance_name = $runtimeInstanceName
    runtime_executable_path = $runtimeExecutablePath
    source_commit = $sourceCommit
    started_server = $startedServer
    log_path = $logPath
    steps = $script:CypressStepEvents
  }

  $summaryJson = $summary | ConvertTo-Json -Depth 6
  Set-Content -LiteralPath $summaryPath -Value $summaryJson -Encoding UTF8
}

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$uiRoot = Join-Path $repoRoot "ui.web"
$defaultRuntimeExecutable = Join-Path $repoRoot "bin/cabinet.exe"
$configPath = Join-Path $uiRoot "cypress.config.runtime.cjs"
$specPath = if ([System.IO.Path]::IsPathRooted($Spec)) { $Spec } else { Join-Path $uiRoot $Spec }
$baseUri = [Uri]$BaseUrl
$runtimePort = $baseUri.Port
$e2eDataDir = Join-Path $repoRoot ".tmp\cypress-runtime-$runtimePort"
$e2eProfile = "e2e-cypress-$runtimePort"
$e2eInstanceName = "cypress-$runtimePort"
$sourceCommit = Get-SourceCommit $repoRoot
$resolvedLogDir = if ([System.IO.Path]::IsPathRooted($LogDir)) { $LogDir } else { Join-Path $repoRoot $LogDir }
$runStamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logSegment = if ([string]::IsNullOrWhiteSpace($LogName)) {
  ConvertTo-SafeLogSegment ([System.IO.Path]::GetFileNameWithoutExtension($Spec))
} else {
  ConvertTo-SafeLogSegment $LogName
}
$logPath = Join-Path $resolvedLogDir "$runStamp-$logSegment.log"
$summaryPath = Join-Path $resolvedLogDir "$runStamp-$logSegment.summary.json"
$transcriptStarted = $false

New-Item -ItemType Directory -Force -Path $resolvedLogDir | Out-Null
try {
  Start-Transcript -Path $logPath -Force | Out-Null
  $transcriptStarted = $true
}
catch {
  Write-Step "Unable to start transcript log at ${logPath}: $($_.Exception.Message)"
}
Write-Step "Run log: $logPath"
Write-Step "Run summary: $summaryPath"
Write-Step "Lane isolation: port=$runtimePort data_dir=$e2eDataDir profile=$e2eProfile instance=$e2eInstanceName commit=$sourceCommit"

if (-not (Test-Path $configPath)) {
  throw "Missing Cypress config: $configPath"
}
if (-not (Test-Path $specPath)) {
  throw "Missing Cypress spec: $specPath"
}
if (-not $SkipDependencyPrep) {
  Ensure-UiDependencies $repoRoot $uiRoot
}

$resolvedRuntimeExecutablePath = ""
if ([string]::IsNullOrWhiteSpace($RuntimeExecutablePath)) {
  $resolvedRuntimeExecutablePath = Ensure-RuntimeExecutable $repoRoot $defaultRuntimeExecutable $SkipRuntimeBuild
} else {
  $candidate = $RuntimeExecutablePath
  if (-not [System.IO.Path]::IsPathRooted($candidate)) {
    $candidate = Join-Path $repoRoot $candidate
  }
  if (-not (Test-Path $candidate)) {
    throw "Configured runtime executable path does not exist: $candidate"
  }
  $resolvedRuntimeExecutablePath = (Resolve-Path $candidate).Path
}

if ((Test-IsEphemeralRuntimePath $resolvedRuntimeExecutablePath) -and -not $AllowTempRuntimePath) {
  throw "ephemeral temp/template runtime path was rejected: $resolvedRuntimeExecutablePath. Pass -AllowTempRuntimePath to override explicitly."
}

$serverProc = $null
$startedServer = $false
$exitCode = 1
$runError = $null

try {
  if (-not $NoServer) {
    $alreadyRunning = Test-Health $BaseUrl
    $canReuse = $alreadyRunning -and $ReuseServer
    if ($alreadyRunning -and -not $ReuseServer) {
      Write-Step "Existing server found at $BaseUrl; recycling to ensure current worktree runtime and UI."
      Stop-PortListener $BaseUrl
      Start-Sleep -Seconds 1
      $alreadyRunning = Test-Health $BaseUrl
      if ($alreadyRunning) {
        Write-Step "Existing server still responds after recycle attempt; continuing with startup and health validation."
      }
    }
    elseif ($alreadyRunning -and $RequireE2EHooks) {
      $canReuse = Test-E2EHooks $BaseUrl
      if (-not $canReuse) {
        Write-Step "Existing server lacks E2E hooks; recycling runtime with CABINET_E2E_MODE=1."
        Stop-PortListener $BaseUrl
      }
    }
    if ($canReuse) {
      Write-Step "Reusing existing server at $BaseUrl"
      Assert-AppPreflight $BaseUrl
    }
    else {
      Write-Step "Starting Cabinet server..."
      if (-not [string]::IsNullOrWhiteSpace($resolvedRuntimeExecutablePath)) {
        Write-Step "Runtime executable resolved: $resolvedRuntimeExecutablePath"
        $serverProc = Start-Process -FilePath $resolvedRuntimeExecutablePath -ArgumentList @(
          "--no-open-browser",
          "--port", "$runtimePort",
          "--data-dir", "$e2eDataDir",
          "--profile", "$e2eProfile",
          "--instance-name", "$e2eInstanceName",
          "--allow-parallel"
        ) -WorkingDirectory $repoRoot -Environment @{ CABINET_E2E_MODE = "1" } -PassThru
      } else {
        Write-Step "Runtime executable resolved: go run ./cmd/cabinet (project-local bin executable missing)"
        $serverProc = Start-Process -FilePath "go" -ArgumentList @(
          "run",
          "./cmd/cabinet",
          "--no-open-browser",
          "--port", "$runtimePort",
          "--data-dir", "$e2eDataDir",
          "--profile", "$e2eProfile",
          "--instance-name", "$e2eInstanceName",
          "--allow-parallel"
        ) -WorkingDirectory $repoRoot -Environment @{ CABINET_E2E_MODE = "1" } -PassThru
      }
      $startedServer = $true

      $deadline = (Get-Date).AddSeconds($StartupTimeoutSec)
      while ((Get-Date) -lt $deadline) {
        if (Test-Health $BaseUrl) {
          break
        }
        Start-Sleep -Seconds 1
      }
      if (-not (Test-Health $BaseUrl)) {
        throw "Server did not become healthy at $BaseUrl within $StartupTimeoutSec seconds."
      }
      if ($RequireE2EHooks -and -not (Test-E2EHooks $BaseUrl)) {
        throw "Server is healthy but E2E hooks are unavailable at $BaseUrl/api/test/reset."
      }
      Write-Step "Server is healthy."
      Assert-AppPreflight $BaseUrl
    }
  }

  if ($ApiContractSmoke) {
    Invoke-ApiContractSmoke $repoRoot $BaseUrl ".work-agent\logs\api-contract-smoke" $RequireE2EHooks.IsPresent
  }

  Write-Step "Running Cypress spec: $specPath"
  $args = @(
    "cypress", "run",
    "--browser", $Browser,
    "--config-file", $configPath,
    "--config", "baseUrl=$BaseUrl",
    "--spec", $specPath
  )
  if ($Headed) {
    $args += "--headed"
  }

  Push-Location $uiRoot
  try {
    $hadElectronRunAsNode = Test-Path Env:ELECTRON_RUN_AS_NODE
    $originalElectronRunAsNode = $env:ELECTRON_RUN_AS_NODE
    if ($hadElectronRunAsNode) {
      Write-Step "Temporarily clearing ELECTRON_RUN_AS_NODE for Cypress runtime compatibility."
    }
    Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue
    & npx @args
    $exitCode = $LASTEXITCODE
    Write-Step "Cypress exited with code $exitCode."
  }
  finally {
    if ($hadElectronRunAsNode) {
      $env:ELECTRON_RUN_AS_NODE = $originalElectronRunAsNode
    } else {
      Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue
    }
    Pop-Location
  }
}
catch {
  $runError = $_
  $exitCode = 1
  Write-Step "Run failed: $($_.Exception.Message)"
}
finally {
  if ($startedServer -and $serverProc) {
    Write-Step "Stopping server process $($serverProc.Id)"
    Stop-Process -Id $serverProc.Id -Force -ErrorAction SilentlyContinue
  }
  $errorMessage = if ($runError) { $runError.Exception.Message } else { "" }
  Write-RunSummary `
    -summaryPath $summaryPath `
    -exitCode $exitCode `
    -errorMessage $errorMessage `
    -repoRoot $repoRoot `
    -uiRoot $uiRoot `
    -specPath $specPath `
    -browser $Browser `
    -baseUrl $BaseUrl `
    -runtimePort $runtimePort `
    -runtimeDataDir $e2eDataDir `
    -runtimeProfile $e2eProfile `
    -runtimeInstanceName $e2eInstanceName `
    -runtimeExecutablePath $resolvedRuntimeExecutablePath `
    -sourceCommit $sourceCommit `
    -logPath $logPath `
    -startedServer $startedServer
  Write-Step "Run summary written: $summaryPath"
  if ($transcriptStarted) {
    Stop-Transcript | Out-Null
  }
}

if ($runError) {
  Write-Error -Message $runError.Exception.Message
  exit 1
}
exit $exitCode
