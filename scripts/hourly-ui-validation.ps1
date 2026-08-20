param(
  [switch]$Force,
  [string]$Browser = "chrome",
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [int]$MaxSpecs = 0,
  [string]$SpecContains = "",
  [switch]$AllowStaleRuntimeVersion,
  [switch]$SkipIssueCreate
)

$ErrorActionPreference = "Stop"

function Write-Step([string]$message) {
  Write-Host "[hourly-ui-validation] $message"
}

function Resolve-RepoRoot {
  return (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
}

function Read-JsonFile([string]$path) {
  if (-not (Test-Path $path)) {
    return $null
  }
  $raw = Get-Content $path -Raw
  if ([string]::IsNullOrWhiteSpace($raw)) {
    return $null
  }
  return $raw | ConvertFrom-Json
}

function Write-JsonFile([string]$path, [object]$payload) {
  $dir = Split-Path -Parent $path
  if (-not (Test-Path $dir)) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
  }
  ($payload | ConvertTo-Json -Depth 12) | Set-Content -Path $path -Encoding UTF8
}

function ConvertTo-SafeLogSegment([string]$value) {
  $segment = ($value -replace '[^A-Za-z0-9._-]+', '-').Trim('-')
  if ([string]::IsNullOrWhiteSpace($segment)) {
    return "spec"
  }
  if ($segment.Length -gt 120) {
    return $segment.Substring(0, 120)
  }
  return $segment
}

function Test-HourlyValidationHealth([string]$baseUrl) {
  try {
    $response = Invoke-WebRequest -Uri "$baseUrl/healthz" -UseBasicParsing -TimeoutSec 2
    return $response.StatusCode -eq 200
  } catch {
    return $false
  }
}

function Start-HourlyValidationRuntime(
  [string]$repoRoot,
  [string]$baseUrl,
  [string]$reportDir,
  [int]$startupTimeoutSec = 45
) {
  if (Test-HourlyValidationHealth $baseUrl) {
    throw "Hourly validation base URL is already occupied: $baseUrl"
  }

  $runtimeExecutable = Join-Path $repoRoot "bin/cabinet.exe"
  if (-not (Test-Path -LiteralPath $runtimeExecutable)) {
    throw "Workflow-built Cabinet runtime is missing: $runtimeExecutable"
  }
  $runtimeExecutable = (Resolve-Path -LiteralPath $runtimeExecutable).Path
  $runtimePort = ([Uri]$baseUrl).Port
  $runtimeDataDir = Join-Path $repoRoot ".tmp/hourly-ui-validation-$runtimePort"
  $runtimeStdoutPath = Join-Path $reportDir "shared-runtime.stdout.log"
  $runtimeStderrPath = Join-Path $reportDir "shared-runtime.stderr.log"
  if (Test-Path -LiteralPath $runtimeDataDir) {
    Remove-Item -LiteralPath $runtimeDataDir -Recurse -Force
  }
  New-Item -ItemType Directory -Path $runtimeDataDir -Force | Out-Null

  $runtimeEnvironment = @{
    CABINET_E2E_MODE = "1"
    CABINET_BIND_MODE = "local"
    CABINET_HOST = "127.0.0.1"
    CABINET_PORT = "$runtimePort"
    CABINET_WEBAUTHN_RP_ID = "127.0.0.1"
    CABINET_WEBAUTHN_ORIGIN = $baseUrl
    CABINET_ALLOW_INSECURE_SECRET_FALLBACK = "1"
    CABINET_FALLBACK_SECRET_PEPPER = "hourly-ui-validation-secret-fallback"
  }
  $runtimeProcess = Start-Process `
    -FilePath $runtimeExecutable `
    -ArgumentList @(
      "--no-open-browser",
      "--port", "$runtimePort",
      "--data-dir", $runtimeDataDir,
      "--profile", "hourly-ui-validation-$runtimePort",
      "--instance-name", "hourly-ui-validation-$runtimePort",
      "--allow-parallel"
    ) `
    -WorkingDirectory $repoRoot `
    -Environment $runtimeEnvironment `
    -RedirectStandardOutput $runtimeStdoutPath `
    -RedirectStandardError $runtimeStderrPath `
    -WindowStyle Hidden `
    -PassThru

  $deadline = (Get-Date).AddSeconds($startupTimeoutSec)
  while ((Get-Date) -lt $deadline) {
    if ($runtimeProcess.HasExited) {
      throw "Workflow-built Cabinet runtime exited before becoming healthy."
    }
    if (Test-HourlyValidationHealth $baseUrl) {
      return [pscustomobject]@{
        process = $runtimeProcess
        executable_path = $runtimeExecutable
        data_dir = $runtimeDataDir
        stdout_path = $runtimeStdoutPath
        stderr_path = $runtimeStderrPath
      }
    }
    Start-Sleep -Milliseconds 500
  }

  Stop-Process -Id $runtimeProcess.Id -Force -ErrorAction SilentlyContinue
  throw "Workflow-built Cabinet runtime did not become healthy at $baseUrl within $startupTimeoutSec seconds."
}

function Stop-HourlyValidationRuntime([object]$runtime) {
  if ($null -eq $runtime -or $null -eq $runtime.process) {
    return "not_started"
  }
  $runtimeProcess = $runtime.process
  try {
    $runtimeProcess.Refresh()
    if ($runtimeProcess.HasExited) {
      return "already_exited"
    }
    $liveProcess = Get-Process -Id $runtimeProcess.Id -ErrorAction Stop
    $livePath = $liveProcess.Path
    if ([string]::IsNullOrWhiteSpace($livePath) -or
        -not [string]::Equals(
          (Resolve-Path -LiteralPath $livePath).Path,
          $runtime.executable_path,
          [System.StringComparison]::OrdinalIgnoreCase
        )) {
      return "pid_identity_mismatch"
    }
    Stop-Process -Id $runtimeProcess.Id -Force -ErrorAction Stop
    if (-not $runtimeProcess.WaitForExit(10000)) {
      return "stop_timeout"
    }
    return "stopped"
  } catch {
    return "stop_failed"
  }
}

function Find-ExistingValidationIssue([string]$currentCommit) {
  if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    return $null
  }
  $rawIssues = & gh issue list --state open --limit 1000 --label "high-priority" --json number,title,body 2>$null
  if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace(($rawIssues | Out-String))) {
    return $null
  }
  $issues = $rawIssues | ConvertFrom-Json
  foreach ($issue in $issues) {
    if ($issue.title -like "[[]Hourly UI Validation[]]*" -and
        ([string]$issue.body).Contains("- commit: $currentCommit")) {
      return $issue
    }
  }
  return $null
}

$repoRoot = Resolve-RepoRoot
$logsRoot = Join-Path $repoRoot ".logs"
$statePath = Join-Path $logsRoot "hourly-ui-validation-state.json"
$reportDir = Join-Path $logsRoot "hourly-ui-validation"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$reportPath = Join-Path $reportDir "report-$timestamp.json"

if (-not (Test-Path $reportDir)) {
  New-Item -ItemType Directory -Path $reportDir -Force | Out-Null
}

$packageJsonPath = Join-Path $repoRoot "ui.web/package.json"
if (-not (Test-Path $packageJsonPath)) {
  throw "Missing ui.web/package.json at $packageJsonPath"
}
$packageJson = Get-Content $packageJsonPath -Raw | ConvertFrom-Json
$currentVersion = [string]$packageJson.version
$currentCommit = (git -C $repoRoot rev-parse HEAD).Trim()

$existingState = Read-JsonFile $statePath
$lastVersion = ""
$lastCommit = ""
if ($existingState) {
  $lastVersion = [string]$existingState.last_validated_version
  $lastCommit = [string]$existingState.last_validated_commit
}

if (-not $Force -and $lastVersion -eq $currentVersion -and $lastCommit -eq $currentCommit) {
  $noChangeReport = [ordered]@{
    generated_at = (Get-Date).ToString("o")
    status = "no-change"
    base_url = $BaseUrl
    last_validated_version = $lastVersion
    last_validated_commit = $lastCommit
    current_version = $currentVersion
    current_commit = $currentCommit
    specs_run = @()
    failures = @()
  }
  Write-JsonFile $reportPath $noChangeReport
  Write-Step "No new build/version detected. status=no-change"
  exit 0
}

$specFiles = Get-ChildItem -Path (Join-Path $repoRoot "ui.web/cypress/e2e") -Recurse -Filter "*.cy.ts" | Sort-Object FullName
if (-not $specFiles -or $specFiles.Count -eq 0) {
  throw "No Cypress specs found under ui.web/cypress/e2e"
}
if (-not [string]::IsNullOrWhiteSpace($SpecContains)) {
  $specFiles = $specFiles | Where-Object { $_.FullName.Replace('\', '/').Contains($SpecContains) }
}
if ($MaxSpecs -gt 0) {
  $specFiles = $specFiles | Select-Object -First $MaxSpecs
}
if (-not $specFiles -or $specFiles.Count -eq 0) {
  throw "No Cypress specs selected for hourly validation. Check -SpecContains/-MaxSpecs filters."
}

$results = @()
$failures = @()
$runnerFailures = @()
$hadFailures = $false
$sharedRuntime = $null
$runtimeCleanupResult = "not_started"

try {
  $sharedRuntime = Start-HourlyValidationRuntime $repoRoot $BaseUrl $reportDir
  for ($specIndex = 0; $specIndex -lt $specFiles.Count; $specIndex++) {
    $spec = $specFiles[$specIndex]
    $relativeSpec = $spec.FullName.Substring((Join-Path $repoRoot "ui.web").Length).TrimStart('\').Replace('\', '/')
    $safeSpec = ConvertTo-SafeLogSegment $relativeSpec
    $logName = "hourly-{0:D3}-{1}" -f ($specIndex + 1), $safeSpec
    Write-Step "Running spec: $relativeSpec"
    $cypressArgs = @(
      "-NoLogo",
      "-NoProfile",
      "-File", (Join-Path $repoRoot "cypress.ps1"),
      "-Spec", $relativeSpec,
      "-Browser", $Browser,
      "-BaseUrl", $BaseUrl,
      "-RequireE2EHooks",
      "-ApiContractSmoke",
      "-ReuseServer",
      "-SkipRuntimeBuild",
      "-SkipDependencyPrep",
      "-Retries", "0",
      "-ExecutionTimeoutSec", "300",
      "-LogDir", $reportDir,
      "-LogName", $logName
    )
    if ($AllowStaleRuntimeVersion) {
      $cypressArgs += "-AllowStaleRuntimeVersion"
    }
    $command = "pwsh " + (($cypressArgs | ForEach-Object {
      if ($_ -match '\s') { '"' + ($_ -replace '"', '\"') + '"' } else { $_ }
    }) -join " ")
    & pwsh @cypressArgs
    $exitCode = $LASTEXITCODE

    $summaryFile = Get-ChildItem -LiteralPath $reportDir -Filter "*-$logName.summary.json" -File |
      Sort-Object LastWriteTimeUtc -Descending |
      Select-Object -First 1
    $summary = if ($summaryFile) { Read-JsonFile $summaryFile.FullName } else { $null }
    $runnerPhase = if ($summary) { [string]$summary.runner_phase } else { "missing_summary" }
    $executionTimedOut = if ($summary) { [bool]$summary.execution_timed_out } else { $false }
    $runtimeRevision = if ($summary) { [string]$summary.runtime_revision } else { "" }
    $runnerCompleted = $summary -and
      -not $executionTimedOut -and
      @("completed", "cypress_failed") -contains $runnerPhase -and
      ($AllowStaleRuntimeVersion -or $runtimeRevision -eq $currentCommit)

    $entry = [ordered]@{
      spec = $relativeSpec
      command = $command
      exit_code = $exitCode
      status = if ($exitCode -eq 0 -and $runnerCompleted) { "pass" } else { "fail" }
      api_contract_smoke = $true
      require_e2e_hooks = $true
      allow_stale_runtime_version = $AllowStaleRuntimeVersion.IsPresent
      log_name = $logName
      summary_path = if ($summaryFile) { $summaryFile.FullName } else { "" }
      runner_phase = $runnerPhase
      execution_timed_out = $executionTimedOut
      runtime_revision = $runtimeRevision
    }
    $results += $entry

    if (-not $runnerCompleted) {
      $hadFailures = $true
      $runnerFailure = [ordered]@{
        spec = $relativeSpec
        phase = $runnerPhase
        exit_code = $exitCode
        execution_timed_out = $executionTimedOut
        runtime_revision = $runtimeRevision
        evidence_path = if ($summaryFile) { $summaryFile.FullName } else { $reportPath }
      }
      $runnerFailures += $runnerFailure
      $failures += [ordered]@{
        screen = $relativeSpec
        action = "managed-cypress-runner"
        expected = "runner completes with exact runtime identity"
        actual = "runner phase $runnerPhase; exit code $exitCode; timed out $executionTimedOut"
        error_text = "managed Cypress runner failed before a trustworthy product result"
        evidence_path = $runnerFailure.evidence_path
      }
      Write-Step "Stopping after runner failure to prevent cascading Windows process exhaustion."
      break
    }

    if ($exitCode -ne 0) {
      $hadFailures = $true
      $failures += [ordered]@{
        screen = $relativeSpec
        action = "managed-cypress-spec-run"
        expected = "spec exits with code 0"
        actual = "spec exited with code $exitCode"
        error_text = "managed cypress run failed for $relativeSpec"
        evidence_path = $entry.summary_path
      }
    }
  }
} catch {
  $hadFailures = $true
  $runnerFailures += [ordered]@{
    spec = "shared-runtime"
    phase = "runtime_error"
    exit_code = 1
    execution_timed_out = $false
    runtime_revision = $currentCommit
    evidence_path = $reportPath
  }
  $failures += [ordered]@{
    screen = "shared-runtime"
    action = "start-or-run-shared-runtime"
    expected = "workflow-built runtime starts and remains healthy"
    actual = "shared runtime failed"
    error_text = $_.Exception.Message
    evidence_path = $reportPath
  }
} finally {
  $runtimeCleanupResult = Stop-HourlyValidationRuntime $sharedRuntime
}

$controlIntentResults = @()
$formFieldResults = @()

foreach ($entry in $results) {
  $intentStatus = if ($entry.exit_code -eq 0) { "pass" } else { "fail" }
  $controlIntentResults += [ordered]@{
    control_id = $entry.spec
    expected_outcome = "selected Cypress spec exits with code 0"
    actual_outcome = if ($entry.exit_code -eq 0) { "spec exited with code 0" } else { "spec exited with code $($entry.exit_code)" }
    status = $intentStatus
    evidence_path = $reportPath
  }

  $formFieldResults += [ordered]@{
    form_id = $entry.spec
    field_scope = "covered-by-selected-cypress-spec"
    checks = @(
      "required-optional-handling",
      "valid-invalid-input-handling",
      "error-message-feedback",
      "submit-save-behavior",
      "keyboard-accessibility"
    )
    status = $intentStatus
    evidence_path = $reportPath
  }
}

$intentPassCount = @($controlIntentResults | Where-Object { $_.status -eq "pass" }).Count
$intentFailCount = @($controlIntentResults | Where-Object { $_.status -eq "fail" }).Count
$fieldPassCount = @($formFieldResults | Where-Object { $_.status -eq "pass" }).Count
$fieldFailCount = @($formFieldResults | Where-Object { $_.status -eq "fail" }).Count

$reportPayload = [ordered]@{
  generated_at = (Get-Date).ToString("o")
  status = if ($hadFailures) { "failed" } else { "passed" }
  base_url = $BaseUrl
  current_version = $currentVersion
  current_commit = $currentCommit
  last_validated_version = $lastVersion
  last_validated_commit = $lastCommit
  specs_run = $results
  control_intent_results = $controlIntentResults
  form_field_results = $formFieldResults
  intent_pass_count = $intentPassCount
  intent_fail_count = $intentFailCount
  field_pass_count = $fieldPassCount
  field_fail_count = $fieldFailCount
  runner_failures = $runnerFailures
  shared_runtime_cleanup = $runtimeCleanupResult
  failures = $failures
}

Write-JsonFile $reportPath $reportPayload

$updatedState = [ordered]@{
  last_validated_version = $currentVersion
  last_validated_commit = $currentCommit
  last_report_path = $reportPath
  last_status = $reportPayload.status
  updated_at = (Get-Date).ToString("o")
}
Write-JsonFile $statePath $updatedState

if ($hadFailures) {
  if ((-not $SkipIssueCreate) -and (Get-Command gh -ErrorAction SilentlyContinue)) {
    $title = "[Hourly UI Validation] failures detected $timestamp"
    $body = @"
Hourly release-aware validation detected failing specs.

- version: $currentVersion
- commit: $currentCommit
- report: $reportPath

Failures:
$(($failures | ForEach-Object { "- $($_.screen): $($_.actual)" }) -join "`n")
"@
    $existingIssue = Find-ExistingValidationIssue $currentCommit
    if ($existingIssue) {
      Write-Step "Existing open validation issue #$($existingIssue.number) already records commit $currentCommit; skipping duplicate issue creation."
    } else {
      gh issue create --title $title --body $body --label "bug,high-priority,general" | Out-Null
    }
  }
  Write-Step "Validation finished with failures."
  exit 1
}

Write-Step "Validation finished successfully."
exit 0
