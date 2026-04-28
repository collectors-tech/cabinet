param(
  [switch]$Rebuild,
  [switch]$FreshData,
  [switch]$Background,
  [switch]$Restart,
  [switch]$AllowParallel
)

$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
. (Join-Path $repoRoot 'scripts\lib\cabinet-console.ps1')

$exePath = Join-Path $repoRoot 'bin\cabinet.exe'
$dataDir = Join-Path $repoRoot 'tmp\demo2\data'
$profile = 'demo2-helper'
$instanceName = 'demo2-helper'
$port = 17882
$url = "http://127.0.0.1:$port"

Write-CabinetBanner -Command "start-demo2" -Summary "Build and start the demo2 review runtime."

if ($Rebuild) {
  Write-CabinetSection "Build"
  Write-CabinetStatus -State "run" -Message "Rebuilding Cabinet before launch."
  & (Join-Path $repoRoot 'scripts\build-cabinet.ps1')
}

if (-not (Test-Path $exePath)) {
  throw "Cabinet executable not found: $exePath`nRun .\scripts\build-cabinet.ps1 first or pass -Rebuild."
}

if ($FreshData -and (Test-Path $dataDir)) {
  Write-CabinetSection "Data"
  Write-CabinetStatus -State "warn" -Message "Removing existing demo2 data directory."
  Write-CabinetKeyValue -Key "Data dir" -Value $dataDir
  Remove-Item -Recurse -Force $dataDir
}

if ($Restart -and $AllowParallel) {
  throw 'Restart cannot be combined with AllowParallel.'
}

$mode = 'reuse-or-attach'
$parallelNote = 'singleton endpoint guard enabled'
if ($Restart) {
  $mode = 'restart'
}
if ($AllowParallel) {
  $mode = 'parallel'
  $parallelNote = 'explicit parallel mode enabled'
}

New-Item -ItemType Directory -Force -Path $dataDir | Out-Null

$args = @(
  '--no-open-browser',
  '--data-dir', $dataDir,
  '--profile', $profile,
  '--instance-name', $instanceName,
  '--seed-sample-data',
  '--port', $port
)

if ($Restart) {
  $args += '--restart'
}
if ($AllowParallel) {
  $args += '--allow-parallel'
}

Write-CabinetSection "Runtime"
Write-CabinetKeyValue -Key "Executable" -Value $exePath
Write-CabinetKeyValue -Key "Data dir" -Value $dataDir
Write-CabinetKeyValue -Key "Port" -Value ([string]$port)
Write-CabinetKeyValue -Key "URL" -Value $url
Write-CabinetKeyValue -Key "Mode" -Value $mode
Write-CabinetKeyValue -Key "Guard" -Value $parallelNote

if ($Background) {
  Write-CabinetStatus -State "run" -Message "Starting Cabinet in the background."
  $proc = Start-Process -FilePath $exePath -ArgumentList $args -WorkingDirectory $repoRoot -PassThru
  Start-Sleep -Seconds 2
  Write-CabinetStatus -State "ok" -Message "Started in background."
  Write-CabinetKeyValue -Key "PID" -Value ([string]$proc.Id)
  Write-CabinetHint -Message "Verify with: Invoke-WebRequest -UseBasicParsing $url/"
  exit 0
}

Write-CabinetStatus -State "run" -Message "Starting Cabinet in the foreground."
& $exePath @args
