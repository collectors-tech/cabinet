param(
  [string]$OutPath = "C:\projects\collectors-tech\cabinet\.antfarm\state\validator-metrics.json"
)

$ErrorActionPreference = 'Stop'
$repo = 'C:\projects\collectors-tech\cabinet'
$migration = Join-Path $repo 'openspec\migration'
New-Item -ItemType Directory -Force -Path (Split-Path $OutPath) | Out-Null

function Get-LatestJson($pattern) {
  $f = Get-ChildItem -Path $migration -Filter $pattern -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending | Select-Object -First 1
  if ($f) { return $f.FullName }
  return $null
}

$controls = @{ discovered = 0; matched = 0; unmatched = 0 }
$fields = @{ discovered = 0; validated = 0; failing = 0 }
$layering = @{ pass = 0; fail = 0 }

# Best-effort parse from known summary markdown/json files
$summaryFiles = Get-ChildItem -Path $migration -File -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending | Select-Object -First 30
foreach ($f in $summaryFiles) {
  $txt = Get-Content $f.FullName -Raw
  if ($txt -match 'Discovered:\s*(\d+)') { $controls.discovered = [int]$Matches[1] }
  if ($txt -match 'Matched(?: to requirement IDs)?:\s*(\d+)') { $controls.matched = [int]$Matches[1] }
  if ($txt -match 'Unmatched:\s*(\d+)') { $controls.unmatched = [int]$Matches[1] }
  if ($txt -match 'Form fields discovered:\s*(\d+)') { $fields.discovered = [int]$Matches[1] }
  if ($txt -match 'Form fields validated:\s*(\d+)') { $fields.validated = [int]$Matches[1] }
  if ($txt -match 'Form fields failing:\s*(\d+)') { $fields.failing = [int]$Matches[1] }
  if ($txt -match 'Layering checks pass/fail counts') {
    if ($txt -match 'Pass:\s*(\d+)') { $layering.pass = [int]$Matches[1] }
    if ($txt -match 'Fail:\s*(\d+)') { $layering.fail = [int]$Matches[1] }
  }
}

$obj = [ordered]@{
  generatedAt = (Get-Date).ToString('o')
  controls = $controls
  fields = $fields
  layering = $layering
}

$obj | ConvertTo-Json -Depth 6 | Set-Content -Path $OutPath -Encoding UTF8
Write-Output $OutPath
