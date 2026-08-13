param(
  [string]$Spec = "cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts",
  [string]$Browser = "electron",
  [switch]$Headed,
  [switch]$NoServer,
  [switch]$ReuseServer,
  [switch]$RequireE2EHooks,
  [switch]$ApiContractSmoke,
  [int[]]$ApiContractSmokeAllowedRuntimePorts = @(),
  [switch]$AllowStaleRuntimeVersion,
  [string]$RuntimeExecutablePath = "",
  [switch]$AllowTempRuntimePath,
  [switch]$SkipDependencyPrep,
  [switch]$SkipRuntimeBuild,
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [int]$StartupTimeoutSec = 45,
  [int]$ExecutionTimeoutSec = 900,
  [string]$LogDir = ".work-agent\logs\cypress",
  [string]$LogName = ""
)

$ErrorActionPreference = "Stop"
$script:CypressStepEvents = @()
$script:LastApiContractSmokeSummaryPath = ""
$script:CypressRunnerPhase = "initializing"
$script:CypressChildPids = @()
$script:CypressProcessTree = @()
$script:CypressElapsedMs = $null
$script:CypressLastOutput = @()
$script:CypressCleanupResult = ""

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

function Get-AppRuntimeMetadata([string]$url) {
  try {
    $res = Invoke-WebRequest -Uri "$url/api/runtime" -UseBasicParsing -TimeoutSec 5
    if ($res.StatusCode -lt 200 -or $res.StatusCode -ge 300) {
      throw "status $($res.StatusCode)"
    }
    return ([string]$res.Content | ConvertFrom-Json)
  }
  catch {
    throw "Unable to read runtime metadata at $url/api/runtime: $($_.Exception.Message)"
  }
}

function Test-AppVersionMatchesSourceCommit([string]$appVersion, [string]$sourceCommit) {
  if ([string]::IsNullOrWhiteSpace($sourceCommit)) {
    return $true
  }
  if ([string]::IsNullOrWhiteSpace($appVersion)) {
    return $false
  }
  $shortCommit = $sourceCommit.Trim()
  if ($shortCommit.Length -gt 12) {
    $shortCommit = $shortCommit.Substring(0, 12)
  }
  return $appVersion -eq "rev-$shortCommit"
}

function Assert-RuntimeAppVersionMatchesSourceCommit([string]$url, [string]$sourceCommit, [bool]$allowStaleRuntimeVersion) {
  if ([string]::IsNullOrWhiteSpace($sourceCommit)) {
    Write-Step "Runtime app version preflight skipped; source commit is unavailable."
    return
  }

  $runtimeMetadata = Get-AppRuntimeMetadata $url
  $appVersion = [string]$runtimeMetadata.app_version
  $expectedVersion = "rev-$($sourceCommit.Substring(0, [Math]::Min(12, $sourceCommit.Length)))"
  if (Test-AppVersionMatchesSourceCommit $appVersion $sourceCommit) {
    Write-Step "Runtime app version preflight passed: app_version=$appVersion matches source_commit=$sourceCommit"
    return
  }

  $message = "Runtime app version mismatch: /api/runtime app_version=$appVersion did not match expected $expectedVersion for source_commit=$sourceCommit. Rebuild the runtime or pass -AllowStaleRuntimeVersion for explicit stale-runtime baseline testing."
  if ($allowStaleRuntimeVersion) {
    Write-Step "Runtime app version preflight allowed stale runtime: $message"
    return
  }
  throw $message
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

function Invoke-ApiContractSmoke([string]$repoRoot, [string]$url, [string]$logDir, [bool]$requireE2EHooks, [int[]]$allowedRuntimePorts) {
  $runId = "cypress-preflight-$runStamp"
  $resolvedLogRoot = if ([System.IO.Path]::IsPathRooted($logDir)) { $logDir } else { Join-Path $repoRoot $logDir }
  $summaryPath = Join-Path (Join-Path $resolvedLogRoot $runId) "api-contract-smoke.summary.json"
  $script:LastApiContractSmokeSummaryPath = $summaryPath
  $args = @(
    "-NoLogo",
    "-NoProfile",
    "-File", (Join-Path $repoRoot "scripts\run-api-contract-smoke.ps1"),
    "-BaseUrl", $url,
    "-LogRoot", $logDir,
    "-RunId", $runId
  )
  if ($requireE2EHooks) {
    $args += "-RequireE2EHooks"
  }
  if ($allowedRuntimePorts -and $allowedRuntimePorts.Count -gt 0) {
    $args += "-AllowedRuntimePorts"
    $args += @($allowedRuntimePorts | ForEach-Object { "$_" })
  }

  Write-Step "Running API contract smoke preflight."
  & pwsh @args | ForEach-Object {
    Write-Host $_
  }
  if ($LASTEXITCODE -ne 0) {
    throw "API contract smoke preflight failed with exit code $LASTEXITCODE."
  }
  Write-Step "API contract smoke summary: $summaryPath"
  return $summaryPath
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

function Reset-CypressRuntimeDataDir([string]$runtimeDataDir) {
  if ([string]::IsNullOrWhiteSpace($runtimeDataDir)) {
    return
  }
  if (-not (Test-Path $runtimeDataDir)) {
    return
  }
  Write-Step "Clearing managed Cypress runtime data dir: $runtimeDataDir"
  Remove-Item -LiteralPath $runtimeDataDir -Recurse -Force
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

function Resolve-CypressSpecArgument([string]$uiRoot, [string]$specValue) {
  $resolvedSpecs = @()
  foreach ($entry in ($specValue -split ",")) {
    $trimmed = $entry.Trim()
    if ([string]::IsNullOrWhiteSpace($trimmed)) {
      continue
    }
    $candidate = if ([System.IO.Path]::IsPathRooted($trimmed)) { $trimmed } else { Join-Path $uiRoot $trimmed }
    if (-not (Test-Path $candidate)) {
      throw "Missing Cypress spec: $candidate"
    }
    $resolvedSpecs += (Resolve-Path $candidate).Path
  }
  if ($resolvedSpecs.Count -eq 0) {
    throw "Missing Cypress spec: no spec paths were provided."
  }
  return ($resolvedSpecs -join ",")
}

function Start-ProcessWithEnvironment([string]$filePath, [string[]]$argumentList, [string]$workingDirectory, [hashtable]$environment) {
  $originalValues = @{}
  foreach ($key in $environment.Keys) {
    $originalValues[$key] = [Environment]::GetEnvironmentVariable($key, "Process")
    [Environment]::SetEnvironmentVariable($key, [string]$environment[$key], "Process")
  }

  try {
    return Start-Process -FilePath $filePath -ArgumentList $argumentList -WorkingDirectory $workingDirectory -PassThru
  }
  finally {
    foreach ($key in $environment.Keys) {
      [Environment]::SetEnvironmentVariable($key, $originalValues[$key], "Process")
    }
  }
}

function Get-ChildProcessIds([int]$parentProcessId) {
  $descendants = @()
  $children = @(Get-CimInstance Win32_Process -Filter "ParentProcessId = $parentProcessId" -ErrorAction SilentlyContinue)
  foreach ($child in $children) {
    $childId = [int]$child.ProcessId
    $descendants += $childId
    $descendants += @(Get-ChildProcessIds $childId)
  }
  return $descendants
}

function ConvertTo-RedactedCommandLine([string]$commandLine) {
  if ([string]::IsNullOrWhiteSpace($commandLine)) {
    return ""
  }

  $redacted = $commandLine
  $redacted = $redacted -replace '(?i)(token|secret|password|passwd|api[_-]?key)(=|\s+)[^\s"]+', '$1$2[REDACTED]'
  $redacted = $redacted -replace '(?i)(--(?:token|secret|password|passwd|api[_-]?key)(?:=|\s+))[^\s"]+', '$1[REDACTED]'
  if ($redacted.Length -gt 500) {
    return $redacted.Substring(0, 500)
  }
  return $redacted
}

function Get-ProcessTreeSnapshot([int]$rootProcessId) {
  $processIds = @()
  $processIds += $rootProcessId
  $processIds += @(Get-ChildProcessIds $rootProcessId)
  $snapshot = @()

  foreach ($processId in ($processIds | Select-Object -Unique)) {
    try {
      $process = Get-CimInstance Win32_Process -Filter "ProcessId = $processId" -ErrorAction SilentlyContinue
      if ($process) {
        $snapshot += [pscustomobject]@{
          pid = [int]$process.ProcessId
          parent_pid = [int]$process.ParentProcessId
          name = [string]$process.Name
          command_line = ConvertTo-RedactedCommandLine ([string]$process.CommandLine)
        }
      }
    }
    catch {
      $snapshot += [pscustomobject]@{
        pid = $processId
        parent_pid = $null
        name = ""
        command_line = "snapshot_error=$($_.Exception.Message)"
      }
    }
  }

  return $snapshot
}

function Stop-OwnedProcessTree([int]$processId) {
  $stopped = @()
  $errors = @()
  $processIds = @()
  $processIds += @(Get-ChildProcessIds $processId)
  $processIds += $processId
  foreach ($ownedProcessId in ($processIds | Select-Object -Unique | Sort-Object -Descending)) {
    try {
      $process = Get-Process -Id $ownedProcessId -ErrorAction SilentlyContinue
      if ($process) {
        Stop-Process -Id $ownedProcessId -Force -ErrorAction Stop
        $stopped += $ownedProcessId
      }
    }
    catch {
      $errors += "pid ${ownedProcessId}: $($_.Exception.Message)"
    }
  }
  if ($errors.Count -gt 0) {
    return "stopped_pids=$($stopped -join ','); errors=$($errors -join '; ')"
  }
  return "stopped_pids=$($stopped -join ',')"
}

function Get-LogTail([string[]]$paths, [int]$lineCount) {
  $lines = @()
  foreach ($path in $paths) {
    if (-not [string]::IsNullOrWhiteSpace($path) -and (Test-Path $path)) {
      $lines += @(Get-Content -LiteralPath $path -Tail $lineCount -ErrorAction SilentlyContinue)
    }
  }
  if ($lines.Count -gt $lineCount) {
    return @($lines | Select-Object -Last $lineCount)
  }
  return $lines
}

function Resolve-CypressRunnerCommand {
  if ($IsWindows) {
    return "npx.cmd"
  }

  return "npx"
}

function Invoke-CypressProcessWithTimeout([string]$uiRoot, [string[]]$arguments, [int]$timeoutSeconds) {
  if ($timeoutSeconds -le 0) {
    throw "ExecutionTimeoutSec must be greater than zero."
  }

  $stdoutPath = Join-Path $env:TEMP ("cabinet-cypress-{0}-stdout.log" -f ([guid]::NewGuid().ToString("N")))
  $stderrPath = Join-Path $env:TEMP ("cabinet-cypress-{0}-stderr.log" -f ([guid]::NewGuid().ToString("N")))
  $startedAt = Get-Date
  $script:CypressRunnerPhase = "cypress_process_started"
  $cypressCommand = Resolve-CypressRunnerCommand

  $process = Start-Process -FilePath $cypressCommand -ArgumentList $arguments -WorkingDirectory $uiRoot -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -NoNewWindow -PassThru
  $script:CypressChildPids = @($process.Id)
  Write-Step "Cypress child process started: command=$cypressCommand pid=$($process.Id) timeout_sec=$timeoutSeconds"

  $completed = $process.WaitForExit($timeoutSeconds * 1000)
  $script:CypressElapsedMs = [int64][Math]::Max(0, [Math]::Round(((Get-Date) - $startedAt).TotalMilliseconds))
  $script:CypressChildPids = @($process.Id) + @(Get-ChildProcessIds $process.Id)
  $script:CypressProcessTree = @(Get-ProcessTreeSnapshot $process.Id)
  $script:CypressLastOutput = @(Get-LogTail @($stdoutPath, $stderrPath) 80)

  foreach ($line in $script:CypressLastOutput) {
    Write-Host $line
  }

  if (-not $completed) {
    $script:CypressRunnerPhase = "execution_timeout"
    $script:CypressCleanupResult = Stop-OwnedProcessTree $process.Id
    Write-Step "Cypress timed out after $($script:CypressElapsedMs) ms; $script:CypressCleanupResult"
    return 124
  }

  $process.Refresh()
  if ($process.ExitCode -eq 0) {
    $script:CypressRunnerPhase = "completed"
  } else {
    $script:CypressRunnerPhase = "cypress_failed"
  }
  $script:CypressCleanupResult = "not_required"
  return [int]$process.ExitCode
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
  [datetime]$runStartedAt,
  [string[]]$runnerCommand,
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
  [string]$apiContractSmokeSummaryPath,
  [bool]$allowStaleRuntimeVersion,
  [bool]$startedServer,
  [string]$runnerPhase,
  [int[]]$cypressChildPids,
  [object[]]$cypressProcessTree,
  [Nullable[int64]]$cypressElapsedMs,
  [string[]]$cypressLastOutput,
  [string]$cypressCleanupResult,
  [int]$executionTimeoutSec
) {
  $runFinishedAt = Get-Date
  $apiContractSmokeStatus = ""
  $apiContractSmokeCheckCount = $null
  $apiContractSmokeFailedCount = $null
  $apiContractSmokeElapsedMs = $null
  $apiContractSmokeFailedChecks = @()
  if (-not [string]::IsNullOrWhiteSpace($apiContractSmokeSummaryPath) -and (Test-Path $apiContractSmokeSummaryPath)) {
    try {
      $apiContractSmokeSummary = Get-Content -Raw -LiteralPath $apiContractSmokeSummaryPath | ConvertFrom-Json
      if ($apiContractSmokeSummary.PSObject.Properties.Name -contains "status") {
        $apiContractSmokeStatus = [string]$apiContractSmokeSummary.status
      }
      if ($apiContractSmokeSummary.PSObject.Properties.Name -contains "check_count") {
        $apiContractSmokeCheckCount = [int]$apiContractSmokeSummary.check_count
      }
      if ($apiContractSmokeSummary.PSObject.Properties.Name -contains "failed_count") {
        $apiContractSmokeFailedCount = [int]$apiContractSmokeSummary.failed_count
      }
      if ($apiContractSmokeSummary.PSObject.Properties.Name -contains "elapsed_ms") {
        $apiContractSmokeElapsedMs = [int]$apiContractSmokeSummary.elapsed_ms
      }
      if ($apiContractSmokeSummary.PSObject.Properties.Name -contains "failed_checks") {
        $apiContractSmokeFailedChecks = @($apiContractSmokeSummary.failed_checks)
      }
    }
    catch {
      Write-Step "Unable to read API contract smoke summary metadata: $($_.Exception.Message)"
    }
  }

  $summary = [ordered]@{
    timestamp = (Get-Date).ToString("o")
    started_at = $runStartedAt.ToString("o")
    finished_at = $runFinishedAt.ToString("o")
    duration_ms = [int64][Math]::Max(0, [Math]::Round(($runFinishedAt - $runStartedAt).TotalMilliseconds))
    exit_code = $exitCode
    error = $errorMessage
    runner_command = $runnerCommand
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
    allow_stale_runtime_version = $allowStaleRuntimeVersion
    started_server = $startedServer
    runner_phase = $runnerPhase
    cypress_child_pids = $cypressChildPids
    cypress_process_tree = $cypressProcessTree
    cypress_elapsed_ms = $cypressElapsedMs
    cypress_last_output = $cypressLastOutput
    cypress_cleanup_result = $cypressCleanupResult
    execution_timeout_sec = $executionTimeoutSec
    log_path = $logPath
    api_contract_smoke_summary_path = $apiContractSmokeSummaryPath
    api_contract_smoke_status = $apiContractSmokeStatus
    api_contract_smoke_check_count = $apiContractSmokeCheckCount
    api_contract_smoke_failed_count = $apiContractSmokeFailedCount
    api_contract_smoke_elapsed_ms = $apiContractSmokeElapsedMs
    api_contract_smoke_failed_checks = $apiContractSmokeFailedChecks
    steps = $script:CypressStepEvents
  }

  $summaryJson = $summary | ConvertTo-Json -Depth 6
  Set-Content -LiteralPath $summaryPath -Value $summaryJson -Encoding UTF8
}

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$uiRoot = Join-Path $repoRoot "ui.web"
$defaultRuntimeExecutable = Join-Path $repoRoot "bin/cabinet.exe"
$configPath = Join-Path $uiRoot "cypress.config.runtime.cjs"
$specPath = Resolve-CypressSpecArgument $uiRoot $Spec
$baseUri = [Uri]$BaseUrl
$runtimePort = $baseUri.Port
$e2eDataDir = Join-Path $repoRoot ".tmp\cypress-runtime-$runtimePort"
$e2eProfile = "e2e-cypress-$runtimePort"
$e2eInstanceName = "cypress-$runtimePort"
$sourceCommit = Get-SourceCommit $repoRoot
$runStartedAt = Get-Date
$resolvedLogDir = if ([System.IO.Path]::IsPathRooted($LogDir)) { $LogDir } else { Join-Path $repoRoot $LogDir }
$runStamp = Get-Date -Format "yyyyMMdd-HHmmss"
$logSegment = if ([string]::IsNullOrWhiteSpace($LogName)) {
  ConvertTo-SafeLogSegment ([System.IO.Path]::GetFileNameWithoutExtension($Spec))
} else {
  ConvertTo-SafeLogSegment $LogName
}
$logPath = Join-Path $resolvedLogDir "$runStamp-$logSegment.log"
$summaryPath = Join-Path $resolvedLogDir "$runStamp-$logSegment.summary.json"
$scriptRelativePath = (Resolve-Path -Relative $PSCommandPath).TrimStart(".\").Replace("\", "/")
$runnerCommand = @(
  "pwsh",
  "-NoLogo",
  "-NoProfile",
  "-File",
  $scriptRelativePath,
  "-Spec",
  $Spec,
  "-Browser",
  $Browser,
  "-BaseUrl",
  $BaseUrl,
  "-StartupTimeoutSec",
  "$StartupTimeoutSec",
  "-ExecutionTimeoutSec",
  "$ExecutionTimeoutSec",
  "-LogDir",
  $LogDir
)
if ($Headed) {
  $runnerCommand += "-Headed"
}
if ($NoServer) {
  $runnerCommand += "-NoServer"
}
if ($ReuseServer) {
  $runnerCommand += "-ReuseServer"
}
if ($RequireE2EHooks) {
  $runnerCommand += "-RequireE2EHooks"
}
if ($ApiContractSmoke) {
  $runnerCommand += "-ApiContractSmoke"
}
if ($AllowStaleRuntimeVersion) {
  $runnerCommand += "-AllowStaleRuntimeVersion"
}
if (-not [string]::IsNullOrWhiteSpace($RuntimeExecutablePath)) {
  $runnerCommand += @("-RuntimeExecutablePath", $RuntimeExecutablePath)
}
if ($AllowTempRuntimePath) {
  $runnerCommand += "-AllowTempRuntimePath"
}
if ($SkipDependencyPrep) {
  $runnerCommand += "-SkipDependencyPrep"
}
if ($SkipRuntimeBuild) {
  $runnerCommand += "-SkipRuntimeBuild"
}
if (-not [string]::IsNullOrWhiteSpace($LogName)) {
  $runnerCommand += @("-LogName", $LogName)
}
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
$apiContractSmokeSummaryPath = ""
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
      Assert-RuntimeAppVersionMatchesSourceCommit $BaseUrl $sourceCommit $AllowStaleRuntimeVersion.IsPresent
    }
    else {
      Write-Step "Starting Cabinet server..."
      Reset-CypressRuntimeDataDir $e2eDataDir
      if (-not [string]::IsNullOrWhiteSpace($resolvedRuntimeExecutablePath)) {
        Write-Step "Runtime executable resolved: $resolvedRuntimeExecutablePath"
        $serverEnv = @{
          CABINET_E2E_MODE = "1"
          CABINET_ALLOW_INSECURE_SECRET_FALLBACK = "1"
          CABINET_FALLBACK_SECRET_PEPPER = "cypress-e2e-secret-fallback"
        }
        $serverProc = Start-ProcessWithEnvironment -FilePath $resolvedRuntimeExecutablePath -ArgumentList @(
          "--no-open-browser",
          "--port", "$runtimePort",
          "--data-dir", "$e2eDataDir",
          "--profile", "$e2eProfile",
          "--instance-name", "$e2eInstanceName",
          "--allow-parallel"
        ) -WorkingDirectory $repoRoot -Environment $serverEnv
      } else {
        Write-Step "Runtime executable resolved: go run ./cmd/cabinet (project-local bin executable missing)"
        $serverEnv = @{
          CABINET_E2E_MODE = "1"
          CABINET_ALLOW_INSECURE_SECRET_FALLBACK = "1"
          CABINET_FALLBACK_SECRET_PEPPER = "cypress-e2e-secret-fallback"
        }
        $serverProc = Start-ProcessWithEnvironment -FilePath "go" -ArgumentList @(
          "run",
          "./cmd/cabinet",
          "--no-open-browser",
          "--port", "$runtimePort",
          "--data-dir", "$e2eDataDir",
          "--profile", "$e2eProfile",
          "--instance-name", "$e2eInstanceName",
          "--allow-parallel"
        ) -WorkingDirectory $repoRoot -Environment $serverEnv
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
      Assert-RuntimeAppVersionMatchesSourceCommit $BaseUrl $sourceCommit $AllowStaleRuntimeVersion.IsPresent
    }
  }

  if ($ApiContractSmoke) {
    $apiContractSmokeSummaryPath = Invoke-ApiContractSmoke $repoRoot $BaseUrl ".work-agent\logs\api-contract-smoke" $RequireE2EHooks.IsPresent $ApiContractSmokeAllowedRuntimePorts
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
    $exitCode = Invoke-CypressProcessWithTimeout $uiRoot $args $ExecutionTimeoutSec
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
  if ([string]::IsNullOrWhiteSpace($apiContractSmokeSummaryPath)) {
    $apiContractSmokeSummaryPath = $script:LastApiContractSmokeSummaryPath
  }
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
    -runStartedAt $runStartedAt `
    -runnerCommand $runnerCommand `
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
    -apiContractSmokeSummaryPath $apiContractSmokeSummaryPath `
    -allowStaleRuntimeVersion $AllowStaleRuntimeVersion.IsPresent `
    -startedServer $startedServer `
    -runnerPhase $script:CypressRunnerPhase `
    -cypressChildPids $script:CypressChildPids `
    -cypressProcessTree $script:CypressProcessTree `
    -cypressElapsedMs $script:CypressElapsedMs `
    -cypressLastOutput $script:CypressLastOutput `
    -cypressCleanupResult $script:CypressCleanupResult `
    -executionTimeoutSec $ExecutionTimeoutSec
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
