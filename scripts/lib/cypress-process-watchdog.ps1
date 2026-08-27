function Get-CypressProcessInventory {
  # Windows PowerShell's first CIM query can spend most of the watchdog guard
  # loading the CIM stack. Prefer its in-box WMI cmdlet there; PowerShell 7
  # does not expose Get-WmiObject, so retain CIM as the compatible fallback.
  $processes = if (Get-Command Get-WmiObject -ErrorAction SilentlyContinue) {
    @(Get-WmiObject Win32_Process -Property ProcessId, ParentProcessId, Name, CommandLine -ErrorAction SilentlyContinue)
  } else {
    @(Get-CimInstance Win32_Process -Property ProcessId, ParentProcessId, Name, CommandLine -ErrorAction SilentlyContinue)
  }
  return @($processes |
    Select-Object ProcessId, ParentProcessId, Name, CommandLine)
}

function Get-CypressOwnedProcessIds([int]$RootProcessId, $ProcessInventory = $null) {
  $processes = if ($null -eq $ProcessInventory) { @(Get-CypressProcessInventory) } else { @($ProcessInventory) }
  $childrenByParent = @{}
  foreach ($process in $processes) {
    $parentId = [int]$process.ParentProcessId
    if (-not $childrenByParent.ContainsKey($parentId)) { $childrenByParent[$parentId] = [System.Collections.Generic.List[int]]::new() }
    $childrenByParent[$parentId].Add([int]$process.ProcessId)
  }
  $owned = [System.Collections.Generic.List[int]]::new()
  $pending = [System.Collections.Generic.Queue[int]]::new()
  $pending.Enqueue($RootProcessId)
  while ($pending.Count -gt 0) {
    $parentId = $pending.Dequeue()
    if (-not $childrenByParent.ContainsKey($parentId)) { continue }
    foreach ($childId in $childrenByParent[$parentId]) {
      if ($childId -le 0 -or $owned.Contains($childId)) { continue }
      $owned.Add($childId)
      $pending.Enqueue($childId)
    }
  }
  return @($owned)
}

function Protect-CypressSummaryOutput([string[]]$Lines) {
  $safe = @()
  foreach ($line in @($Lines)) {
    if ($null -eq $line) { continue }
    $redacted = [string]$line
    $redacted = $redacted -replace '(?i)\b(Bearer\s+)[A-Za-z0-9._~+/=-]+', '$1[REDACTED]'
    $redacted = $redacted -replace '(?i)\b(token|api[_-]?key|authorization|password|secret)(\s*[:=]\s*)[^\s,;]+', '$1$2[REDACTED]'
    $redacted = $redacted -replace '\b\d{6,12}:[A-Za-z0-9_-]{20,}\b', '[REDACTED_TELEGRAM_TOKEN]'
    $safe += $redacted
  }
  return @($safe)
}

function ConvertTo-CypressRedactedCommandLine([string]$CommandLine) {
  if ([string]::IsNullOrWhiteSpace($CommandLine)) { return "" }
  $redacted = @(Protect-CypressSummaryOutput @($CommandLine))[0]
  if ($redacted.Length -gt 500) { return $redacted.Substring(0, 500) }
  return $redacted
}

function Get-CypressProcessTreeSnapshot([int]$RootProcessId, $ProcessInventory = $null) {
  $processes = if ($null -eq $ProcessInventory) { @(Get-CypressProcessInventory) } else { @($ProcessInventory) }
  $processIds = @($RootProcessId) + @(Get-CypressOwnedProcessIds -RootProcessId $RootProcessId -ProcessInventory $processes)
  $processesById = @{}
  foreach ($process in $processes) { $processesById[[int]$process.ProcessId] = $process }
  $snapshot = @()
  foreach ($processId in @($processIds | Where-Object { $_ -gt 0 } | Select-Object -Unique)) {
    if ($processesById.ContainsKey([int]$processId)) {
      $process = $processesById[[int]$processId]
      $snapshot += [pscustomobject][ordered]@{
        pid = [int]$process.ProcessId
        parent_pid = [int]$process.ParentProcessId
        name = [string]$process.Name
        command_line = ConvertTo-CypressRedactedCommandLine ([string]$process.CommandLine)
      }
    } else {
      $snapshot += [pscustomobject][ordered]@{
        pid = [int]$processId
        parent_pid = $null
        name = "unavailable"
        command_line = ""
      }
    }
  }
  return @($snapshot)
}

function Get-CypressOutputTail([string]$StandardOutputPath, [string]$StandardErrorPath, [int]$LineCount) {
  $lines = @()
  foreach ($entry in @(@{ Label = "stdout"; Path = $StandardOutputPath }, @{ Label = "stderr"; Path = $StandardErrorPath })) {
    if ([string]::IsNullOrWhiteSpace($entry.Path) -or -not (Test-Path -LiteralPath $entry.Path)) { continue }
    try {
      foreach ($line in @(Get-Content -LiteralPath $entry.Path -Tail $LineCount -ErrorAction Stop)) { $lines += "[$($entry.Label)] $line" }
    } catch { $lines += "[$($entry.Label)] unable to read captured output" }
  }
  if ($lines.Count -gt $LineCount) { $lines = @($lines | Select-Object -Last $LineCount) }
  return @(Protect-CypressSummaryOutput $lines)
}

function Test-CypressObservedCleanupCandidate([int]$ProcessId, [int[]]$CurrentOwnedProcessIds, $ProcessInventory = $null) {
  if ($ProcessId -le 0 -or $ProcessId -eq $PID) { return $false }
  if ($CurrentOwnedProcessIds -contains $ProcessId) { return $true }
  $processes = if ($null -eq $ProcessInventory) { @(Get-CypressProcessInventory) } else { @($ProcessInventory) }
  $process = $processes | Where-Object { [int]$_.ProcessId -eq $ProcessId } | Select-Object -First 1
  if (-not $process) { return $false }
  $commandLine = [string]$process.CommandLine
  if ([string]::IsNullOrWhiteSpace($commandLine)) { return $false }
  return $commandLine -match '(?i)(\\cypress\\|/cypress/|cypress-runtime-|Cypress\\cy\\production\\browsers|cypress\.config\.runtime\.cjs|run-cypress\.mjs)'
}

function Stop-CypressOwnedProcessTree(
  [int]$RootProcessId,
  [int[]]$ObservedChildProcessIds = @(),
  $ProcessInventory = $null
) {
  # Observed PIDs are guarded by current command-line evidence before cleanup to avoid killing reused PIDs.
  $processes = if ($null -eq $ProcessInventory) { @(Get-CypressProcessInventory) } else { @($ProcessInventory) }
  $ownedIds = @(Get-CypressOwnedProcessIds -RootProcessId $RootProcessId -ProcessInventory $processes | Where-Object { $_ -gt 0 -and $_ -ne $PID } | Select-Object -Unique)
  $observedLiveIds = @($ObservedChildProcessIds | Where-Object {
    Test-CypressObservedCleanupCandidate -ProcessId $_ -CurrentOwnedProcessIds $ownedIds -ProcessInventory $processes
  } | Select-Object -Unique)
  $ownedIds = @($ownedIds + $observedLiveIds | Select-Object -Unique)
  [array]::Reverse($ownedIds)
  $targets = @($ownedIds) + @($RootProcessId)
  $stopped = @()
  $stoppedProcesses = @()
  foreach ($processId in $targets) {
    if ($processId -le 0 -or $processId -eq $PID) { continue }
    $targetProcess = Get-Process -Id $processId -ErrorAction SilentlyContinue
    if (-not $targetProcess) { continue }
    Stop-Process -InputObject $targetProcess -Force -ErrorAction SilentlyContinue
    $stoppedProcesses += $targetProcess
    $stopped += $processId
  }
  $cleanupDeadline = (Get-Date).AddSeconds(2)
  foreach ($targetProcess in $stoppedProcesses) {
    try {
      $remainingMs = [int][Math]::Max(0, [Math]::Floor(($cleanupDeadline - (Get-Date)).TotalMilliseconds))
      if ($remainingMs -gt 0) { $targetProcess.WaitForExit($remainingMs) | Out-Null }
    } catch {}
  }
  $remaining = @($targets | Select-Object -Unique | Where-Object { Get-Process -Id $_ -ErrorAction SilentlyContinue })
  foreach ($targetProcess in $stoppedProcesses) { try { $targetProcess.Dispose() } catch {} }
  if ($remaining.Count -gt 0) { return "owned_process_tree_cleanup_incomplete stopped=$($stopped -join ',') remaining=$($remaining -join ',')" }
  $verifiedStopped = @($targets | Select-Object -Unique)
  return "owned_process_tree_stopped pids=$($verifiedStopped -join ',')"
}

function Invoke-CypressOwnedProcess(
  [string]$FilePath, [string[]]$ArgumentList, [string]$WorkingDirectory,
  [int]$TimeoutSec, [string]$StandardOutputPath, [string]$StandardErrorPath,
  [int]$OutputTailLineCount = 40
) {
  if ($TimeoutSec -le 0) { throw "Cypress execution timeout must be greater than zero seconds." }
  if ($StandardOutputPath -eq $StandardErrorPath) { throw "Cypress stdout and stderr capture paths must be different." }
  foreach ($capturePath in @($StandardOutputPath, $StandardErrorPath)) {
    $parent = Split-Path -Parent $capturePath
    if (-not [string]::IsNullOrWhiteSpace($parent)) { New-Item -ItemType Directory -Force -Path $parent | Out-Null }
  }
  $process = Start-Process -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory -RedirectStandardOutput $StandardOutputPath -RedirectStandardError $StandardErrorPath -PassThru
  $startedAt = Get-Date
  $observedChildIds = @()
  $deadline = $startedAt.AddSeconds($TimeoutSec)
  while ((Get-Date) -lt $deadline) {
    $process.Refresh()
    if ($process.HasExited) { break }
    Start-Sleep -Milliseconds 200
  }
  $process.Refresh()
  $timedOut = -not $process.HasExited
  $processInventory = @(Get-CypressProcessInventory)
  $observedChildIds = @(Get-CypressOwnedProcessIds -RootProcessId $process.Id -ProcessInventory $processInventory | Select-Object -Unique)
  $processTree = @(Get-CypressProcessTreeSnapshot -RootProcessId $process.Id -ProcessInventory $processInventory)
  $cleanupResult = "not_required"
  if ($timedOut) {
    $cleanupResult = Stop-CypressOwnedProcessTree -RootProcessId $process.Id -ObservedChildProcessIds $observedChildIds -ProcessInventory $processInventory
  } else { $process.WaitForExit() }
  $rootProcessId = $process.Id
  $exitCode = if ($timedOut) { 124 } else { [int]$process.ExitCode }
  $process.Dispose()
  $finishedAt = Get-Date
  $runnerPhase = if ($timedOut) { "execution_timeout" } elseif ($exitCode -eq 0) { "completed" } else { "cypress_failed" }
  $lastOutput = @()
  if ($timedOut) {
    $lastOutput = @(Get-CypressOutputTail $StandardOutputPath $StandardErrorPath $OutputTailLineCount)
  } else {
    $outputDeadline = (Get-Date).AddSeconds(2)
    do {
      $lastOutput = @(Get-CypressOutputTail $StandardOutputPath $StandardErrorPath $OutputTailLineCount)
      if ($lastOutput.Count -gt 0) { break }
      Start-Sleep -Milliseconds 100
    } while ((Get-Date) -lt $outputDeadline)
  }
  return [pscustomobject][ordered]@{
    timed_out = $timedOut; exit_code = $exitCode; runner_phase = $runnerPhase
    root_pid = $rootProcessId; child_pids = @($observedChildIds)
    process_tree = @($processTree)
    elapsed_ms = [int64][Math]::Max(0, [Math]::Round(($finishedAt - $startedAt).TotalMilliseconds))
    last_output = @($lastOutput); cleanup_result = $cleanupResult
    stdout_path = $StandardOutputPath; stderr_path = $StandardErrorPath
  }
}
