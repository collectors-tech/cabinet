param(
  [string]$Workflow = "cabinet-validator"
)

$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$artifactDir = Join-Path $repoRoot ".antfarm\artifacts"
$stateDir = Join-Path $repoRoot ".antfarm\state"
$lastFile = Join-Path $stateDir "last_validated_artifact.txt"

New-Item -ItemType Directory -Path $stateDir -Force | Out-Null
if (!(Test-Path $artifactDir)) { Write-Output "No artifacts dir yet."; exit 0 }

$manifests = Get-ChildItem -Path $artifactDir -Filter "release-*.json" | Sort-Object LastWriteTime -Descending
if (-not $manifests -or $manifests.Count -eq 0) { Write-Output "No artifact manifests yet."; exit 0 }

$latest = $null
foreach ($m in $manifests) {
  try {
    $obj = Get-Content $m.FullName -Raw | ConvertFrom-Json
    if ($obj.artifactId -and $obj.gitSha -and $obj.createdAt) { $latest = $m; break }
  } catch {
    continue
  }
}
if (-not $latest) { Write-Output "No valid artifact manifests (missing artifactId/gitSha/createdAt)."; exit 0 }

$latestId = [IO.Path]::GetFileNameWithoutExtension($latest.Name)
$last = if (Test-Path $lastFile) { (Get-Content $lastFile -Raw).Trim() } else { "" }
if ($latestId -eq $last) { Write-Output "Latest artifact already validated: $latestId"; exit 0 }

$runs = node C:\projects\antfarm\dist\cli\cli.js workflow runs | Out-String
$validatorActive = (($runs -split "`r?`n") | Where-Object { $_ -match "\[running" -and $_ -match "cabinet-validator" }).Count -gt 0
if ($validatorActive) { Write-Output "Validator run already active."; exit 0 }

$task = "Validate latest artifact for Cabinet. Artifact manifest: $($latest.FullName)"
node C:\projects\antfarm\dist\cli\cli.js workflow run $Workflow "$task"
