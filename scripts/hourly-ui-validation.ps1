param(
  [switch]$Force,
  [string]$Browser = "chrome",
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [int]$MaxSpecs = 0,
  [string]$SpecContains = "",
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
$hadFailures = $false

foreach ($spec in $specFiles) {
  $relativeSpec = $spec.FullName.Substring((Join-Path $repoRoot "ui.web").Length).TrimStart('\').Replace('\', '/')
  Write-Step "Running spec: $relativeSpec"
  $command = "pwsh -File ./cypress.ps1 -Spec $relativeSpec -Browser $Browser"
  & pwsh -NoLogo -NoProfile -File (Join-Path $repoRoot "cypress.ps1") -Spec $relativeSpec -Browser $Browser
  $exitCode = $LASTEXITCODE

  $entry = [ordered]@{
    spec = $relativeSpec
    command = $command
    exit_code = $exitCode
    status = if ($exitCode -eq 0) { "pass" } else { "fail" }
  }
  $results += $entry

  if ($exitCode -ne 0) {
    $hadFailures = $true
    $failures += [ordered]@{
      screen = $relativeSpec
      action = "managed-cypress-spec-run"
      expected = "spec exits with code 0"
      actual = "spec exited with code $exitCode"
      error_text = "managed cypress run failed for $relativeSpec"
      evidence_path = $reportPath
    }
  }
}

$reportPayload = [ordered]@{
  generated_at = (Get-Date).ToString("o")
  status = if ($hadFailures) { "failed" } else { "passed" }
  base_url = $BaseUrl
  current_version = $currentVersion
  current_commit = $currentCommit
  last_validated_version = $lastVersion
  last_validated_commit = $lastCommit
  specs_run = $results
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
    gh issue create --title $title --body $body --label "bug,high-priority,general" | Out-Null
  }
  Write-Step "Validation finished with failures."
  exit 1
}

Write-Step "Validation finished successfully."
exit 0
