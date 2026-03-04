param(
  [string]$Workflow = "cabinet",
  [string]$RepoSlug = "collectors-tech/cabinet",
  [int]$LookbackMinutes = 20
)

$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$stateDir = Join-Path $repoRoot ".antfarm\state"
$lockPath = Join-Path $stateDir "dispatcher.lock"

New-Item -ItemType Directory -Path $stateDir -Force | Out-Null

function Release-Lock {
  if (Test-Path $lockPath) { Remove-Item $lockPath -Force -ErrorAction SilentlyContinue }
}

if (Test-Path $lockPath) {
  $ageMin = ((Get-Date) - (Get-Item $lockPath).LastWriteTime).TotalMinutes
  if ($ageMin -lt $LookbackMinutes) {
    Write-Output "Dispatcher lock exists ($([math]::Round($ageMin,1))m). Exiting."
    exit 0
  }
  Remove-Item $lockPath -Force -ErrorAction SilentlyContinue
}

New-Item -ItemType File -Path $lockPath -Force | Out-Null
try {
  Push-Location $repoRoot

  gh auth status | Out-Null

  $runsText = node C:\projects\antfarm\dist\cli\cli.js workflow runs 2>$null | Out-String
  $activeRun = (($runsText -split "`r?`n") | Where-Object { $_ -match "\[running" -and $_ -match [regex]::Escape($Workflow) }).Count -gt 0
  if ($activeRun) {
    Write-Output "Active $Workflow run detected. Skipping dispatch."
    exit 0
  }

  $issuesJson = gh issue list --repo $RepoSlug --state open --limit 100 --json number,title,labels,updatedAt,url
  $issues = $issuesJson | ConvertFrom-Json

  $ready = $issues | Where-Object {
    $labelNames = @($_.labels | ForEach-Object { $_.name.ToLowerInvariant() })
    $hasReady = ($labelNames -contains "ready")
    $hasPriority = ($labelNames | Where-Object { $_ -match '^priority:p[0-9]+$' } | Measure-Object).Count -gt 0
    $hasHighPriority = ($labelNames -contains "high-priority")
    ($hasReady -or $hasPriority -or $hasHighPriority) -and -not ($labelNames -contains "blocked")
  }

  if (-not $ready -or $ready.Count -eq 0) {
    Write-Output "No ready issues."
    exit 0
  }

  function PriorityScore($issue) {
    $names = @($issue.labels | ForEach-Object { $_.name.ToLowerInvariant() })
    if ($names -contains "p0") { return 0 }
    if ($names -contains "p1") { return 1 }
    if ($names -contains "p2") { return 2 }
    if ($names -contains "p3") { return 3 }
    return 9
  }

  $selected = $ready |
    Sort-Object @{Expression={ PriorityScore $_ }}, @{Expression={ [datetime]$_.updatedAt }} |
    Select-Object -First 1

  $task = @"
Issue-driven execution for $RepoSlug.
Target issue: #$($selected.number) - $($selected.title)
Issue URL: $($selected.url)

Mandatory:
- Issue -> Spec -> Validate -> Commit
- UI intent/form/layering validation where applicable
- openspec validate --all before completion
- Post evidence back to the issue
"@

  $runOut = node C:\projects\antfarm\dist\cli\cli.js workflow run $Workflow "$task" | Out-String
  Write-Output $runOut

  try {
    gh issue comment $selected.number --repo $RepoSlug --body "🤖 Antfarm dispatched workflow '$Workflow' for this issue.\n\nDispatcher host: $env:COMPUTERNAME"
  } catch {
    Write-Warning "Could not comment on issue: $($_.Exception.Message)"
  }
}
finally {
  Pop-Location
  Release-Lock
}
