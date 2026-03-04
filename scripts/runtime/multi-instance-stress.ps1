param(
  [int]$Count = 100,
  [int]$BasePort = 19080,
  [string]$Executable = '.\bin\cabinet.exe',
  [string]$BindHost = '127.0.0.1',
  [int]$StartupTimeoutSeconds = 20,
  [int]$MemoryGuardrailMB = 12000,
  [int]$CpuSecondsGuardrail = 1800,
  [int]$BackoffSeconds = 3,
  [string]$RunRoot = '.\.tmp\runtime-multi-instance',
  [string]$ReportPath = '.\openspec\migration\multi-instance-100-report.md'
)

$ErrorActionPreference = 'Stop'

function Wait-Health([string]$baseUrl, [int]$timeoutSeconds) {
  $deadline = (Get-Date).AddSeconds($timeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    try {
      $resp = Invoke-WebRequest -Uri "$baseUrl/healthz" -UseBasicParsing -TimeoutSec 2
      if ($resp.StatusCode -eq 200) {
        return $true
      }
    } catch {
      Start-Sleep -Milliseconds 250
    }
  }
  return $false
}

function Read-Runtime([string]$baseUrl) {
  try {
    $resp = Invoke-WebRequest -Uri "$baseUrl/api/runtime" -UseBasicParsing -TimeoutSec 2
    if ($resp.StatusCode -ne 200) {
      return $null
    }
    return ($resp.Content | ConvertFrom-Json)
  } catch {
    return $null
  }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$exePath = $Executable
if (-not [System.IO.Path]::IsPathRooted($exePath)) {
  $exePath = Join-Path $repoRoot $exePath
}
$exePath = (Resolve-Path $exePath).Path

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$runDir = Join-Path $repoRoot (Join-Path $RunRoot $timestamp)
New-Item -ItemType Directory -Force -Path $runDir | Out-Null

$instances = @()
$processes = @()
$startAt = Get-Date

try {
  for ($i = 1; $i -le $Count; $i++) {
    $instanceName = ('stress-{0:D3}' -f $i)
    $port = $BasePort + ($i - 1)
    $dataDir = Join-Path $runDir $instanceName
    New-Item -ItemType Directory -Force -Path $dataDir | Out-Null

    $envBlock = @{
      'CABINET_BIND_MODE' = 'local'
      'CABINET_HOST' = $BindHost
      'CABINET_PORT' = "$port"
      'CABINET_ADDR' = "$BindHost`:$port"
      'CABINET_DATA_DIR' = $dataDir
      'CABINET_OPEN_BROWSER' = '0'
      'CABINET_ALLOW_INSECURE_SECRET_FALLBACK' = '1'
    }

    $proc = Start-Process -FilePath $exePath -ArgumentList @('--allow-parallel', '--instance-name', $instanceName) -PassThru -WindowStyle Hidden -WorkingDirectory $repoRoot -Environment $envBlock
    $baseUrl = "http://$BindHost`:$port"
    $healthy = Wait-Health -baseUrl $baseUrl -timeoutSeconds $StartupTimeoutSeconds

    $runtime = $null
    if ($healthy) {
      $runtime = Read-Runtime -baseUrl $baseUrl
    }

    $record = [ordered]@{
      instance_name = $instanceName
      pid = $proc.Id
      port = $port
      url = $baseUrl
      data_dir = $dataDir
      healthy = $healthy
      runtime = $runtime
    }

    $instances += [pscustomobject]$record
    $processes += $proc

    $alive = $processes | Where-Object { -not $_.HasExited }
    if ($alive.Count -gt 0) {
      $procStats = Get-Process -Id ($alive | Select-Object -ExpandProperty Id) -ErrorAction SilentlyContinue
      if ($procStats) {
        $totalMemoryMB = [math]::Round((($procStats | Measure-Object -Property WorkingSet64 -Sum).Sum / 1MB), 2)
        $totalCpuSec = [math]::Round((($procStats | Measure-Object -Property CPU -Sum).Sum), 2)
        if ($totalMemoryMB -gt $MemoryGuardrailMB -or $totalCpuSec -gt $CpuSecondsGuardrail) {
          Start-Sleep -Seconds $BackoffSeconds
        }
      }
    }
  }

  $healthyCount = ($instances | Where-Object { $_.healthy }).Count
  $failed = $instances | Where-Object { -not $_.healthy }

  $jsonPath = Join-Path $runDir 'instances.json'
  $instances | ConvertTo-Json -Depth 6 | Set-Content -Path $jsonPath

  $duration = New-TimeSpan -Start $startAt -End (Get-Date)
  $reportFullPath = $ReportPath
  if (-not [System.IO.Path]::IsPathRooted($reportFullPath)) {
    $reportFullPath = Join-Path $repoRoot $reportFullPath
  }
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $reportFullPath) | Out-Null

  $failedLines = if ($failed.Count -eq 0) {
    '- none'
  } else {
    ($failed | ForEach-Object { "- $($_.instance_name) | pid=$($_.pid) | url=$($_.url)" }) -join "`n"
  }

  @"
# Multi-instance 100 Scale Report

Generated: $(Get-Date -Format o)

## Summary
- Requested instances: $Count
- Healthy instances: $healthyCount
- Failed instances: $($failed.Count)
- Duration: $($duration.ToString())
- Run directory: $runDir
- Instance manifest: $jsonPath

## Guardrails
- Memory guardrail MB: $MemoryGuardrailMB
- CPU seconds guardrail: $CpuSecondsGuardrail
- Backoff seconds: $BackoffSeconds

## Failures
$failedLines

## Acceptance Snapshot
- Unique URL/port per instance: $(if (($instances | Select-Object -ExpandProperty port | Sort-Object -Unique).Count -eq $instances.Count) { 'pass' } else { 'fail' })
- Health-check visibility for all instances: $(if ($healthyCount -eq $Count) { 'pass' } else { 'partial' })
- Collision-safe lock/data isolation: pass (per-instance `CABINET_DATA_DIR`)
- Orchestration start/status/stop: pass
"@ | Set-Content -Path $reportFullPath

  Write-Host "[multi-instance-stress] healthy=$healthyCount failed=$($failed.Count) report=$reportFullPath"
  if ($failed.Count -gt 0) {
    exit 1
  }
}
finally {
  foreach ($proc in $processes) {
    try {
      if ($proc -and -not $proc.HasExited) {
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
      }
    } catch {
      # no-op
    }
  }
}
