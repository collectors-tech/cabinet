param(
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [string]$RunId = "",
  [string]$LogRoot = ".work-agent\logs\api-contract-smoke",
  [int]$TimeoutSec = 5,
  [switch]$RequireE2EHooks
)

$ErrorActionPreference = "Stop"

function ConvertTo-SafeSegment([string]$value) {
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

function Join-UrlPath([string]$baseUrl, [string]$path) {
  return $baseUrl.TrimEnd("/") + "/" + $path.TrimStart("/")
}

function Invoke-ApiSmokeCheck([string]$baseUrl, [hashtable]$check, [int]$timeoutSec) {
  $checkStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
  $method = if ($check.ContainsKey("Method")) { $check.Method } else { "GET" }
  $uri = Join-UrlPath $baseUrl $check.Path
  $result = [ordered]@{
    name = $check.Name
    method = $method
    path = $check.Path
    url = $uri
    expected_status = $check.Status
    status = 0
    passed = $false
    duration_ms = 0
    error = ""
  }

  try {
    $body = if ($check.ContainsKey("Body")) { $check.Body } else { $null }
    $contentType = if ($body) { "application/json" } else { $null }
    $response = Invoke-WebRequest -Uri $uri -Method $method -Body $body -ContentType $contentType -UseBasicParsing -TimeoutSec $timeoutSec
    $result.status = [int]$response.StatusCode
    if ([int]$response.StatusCode -ne [int]$check.Status) {
      $result.error = "expected status $($check.Status), got $($response.StatusCode)"
      $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
      return [pscustomobject]$result
    }

    if ($check.ContainsKey("ContentType")) {
      $actualContentType = [string]$response.Headers["Content-Type"]
      if (-not $actualContentType.ToLowerInvariant().Contains($check.ContentType.ToLowerInvariant())) {
        $result.error = "expected content type containing $($check.ContentType), got $actualContentType"
        $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
        return [pscustomobject]$result
      }
    }

    $text = if ($response.Content -is [byte[]]) {
      [System.Text.Encoding]::UTF8.GetString($response.Content)
    } else {
      [string]$response.Content
    }
    if ($check.ContainsKey("Contains") -and -not $text.Contains($check.Contains)) {
      $result.error = "response body did not contain required marker $($check.Contains)"
      $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
      return [pscustomobject]$result
    }

    if ($check.ContainsKey("Json") -and $check.Json) {
      try {
        $null = $text | ConvertFrom-Json
      }
      catch {
        $result.error = "response body was not valid JSON: $($_.Exception.Message)"
        $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
        return [pscustomobject]$result
      }
    }

    $result.passed = $true
    $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
    return [pscustomobject]$result
  }
  catch {
    $statusCode = 0
    if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
      $statusCode = [int]$_.Exception.Response.StatusCode
    }
    $result.status = $statusCode
    $result.error = $_.Exception.Message
    $result.duration_ms = [int]$checkStopwatch.ElapsedMilliseconds
    return [pscustomobject]$result
  }
}

function Get-SourceCommit([string]$repoRoot) {
  try {
    $commit = (& git -C $repoRoot rev-parse HEAD 2>$null)
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($commit)) {
      return [string]$commit
    }
  }
  catch {
  }
  return ""
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$runStopwatch = [System.Diagnostics.Stopwatch]::StartNew()
$runStamp = if ([string]::IsNullOrWhiteSpace($RunId)) { Get-Date -Format "yyyyMMdd-HHmmss" } else { ConvertTo-SafeSegment $RunId }
$resolvedLogRoot = if ([System.IO.Path]::IsPathRooted($LogRoot)) { $LogRoot } else { Join-Path $repoRoot $LogRoot }
$runLogDir = Join-Path $resolvedLogRoot $runStamp
$summaryPath = Join-Path $runLogDir "api-contract-smoke.summary.json"
$sourceCommit = Get-SourceCommit $repoRoot

New-Item -ItemType Directory -Force -Path $runLogDir | Out-Null

$checks = @(
  @{
    Name = "healthz"
    Method = "GET"
    Path = "/healthz"
    Status = 200
  },
  @{
    Name = "runtime API"
    Method = "GET"
    Path = "/api/runtime"
    Status = 200
    ContentType = "json"
    Json = $true
  },
  @{
    Name = "OpenAPI YAML"
    Method = "GET"
    Path = "/api/openapi.yaml"
    Status = 200
    ContentType = "yaml"
    Contains = "openapi:"
  },
  @{
    Name = "sign-in route"
    Method = "GET"
    Path = "/sign-in"
    Status = 200
    Contains = "<"
  }
)

if ($RequireE2EHooks) {
  $checks += @{
    Name = "E2E reset hook"
    Method = "POST"
    Path = "/api/test/reset"
    Status = 200
    Body = "{}"
    ContentType = "json"
  }
}

Write-Host "[api-contract-smoke] Base URL: $BaseUrl"
Write-Host "[api-contract-smoke] Summary: $summaryPath"

$results = @()
foreach ($check in $checks) {
  $result = Invoke-ApiSmokeCheck $BaseUrl $check $TimeoutSec
  $results += $result
  $state = if ($result.passed) { "pass" } else { "fail" }
  Write-Host "[api-contract-smoke] $state $($result.method) $($result.path) status=$($result.status)"
}

$failed = @($results | Where-Object { -not $_.passed })
$exitCode = if ($failed.Count -gt 0) { 1 } else { 0 }
$summary = [ordered]@{
  timestamp = (Get-Date).ToString("o")
  exit_code = $exitCode
  base_url = $BaseUrl
  source_commit = $sourceCommit
  elapsed_ms = [int]$runStopwatch.ElapsedMilliseconds
  check_count = $results.Count
  failed_count = $failed.Count
  require_e2e_hooks = $RequireE2EHooks.IsPresent
  log_dir = $runLogDir
  checks = $results
}
$summary | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $summaryPath -Encoding UTF8

if ($exitCode -ne 0) {
  Write-Error "API contract smoke failed; summary written to $summaryPath"
  exit $exitCode
}

Write-Host "[api-contract-smoke] All checks passed; summary written to $summaryPath"
exit 0
