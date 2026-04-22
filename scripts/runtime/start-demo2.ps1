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
$exePath = Join-Path $repoRoot 'bin\cabinet.exe'
$dataDir = Join-Path $repoRoot 'tmp\demo2\data'
$profile = 'demo2-helper'
$instanceName = 'demo2-helper'
$port = 17882
$url = "http://127.0.0.1:$port"

if ($Rebuild) {
  Write-Host '[start-demo2] Rebuilding Cabinet first...'
  & (Join-Path $repoRoot 'scripts\build-cabinet.ps1')
}

if (-not (Test-Path $exePath)) {
  throw "Cabinet executable not found: $exePath`nRun .\scripts\build-cabinet.ps1 first or pass -Rebuild."
}

if ($FreshData -and (Test-Path $dataDir)) {
  Write-Host "[start-demo2] Removing existing demo2 data dir: $dataDir"
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
  '--port', $port
)

if ($Restart) {
  $args += '--restart'
}
if ($AllowParallel) {
  $args += '--allow-parallel'
}

Write-Host "[start-demo2] Executable: $exePath"
Write-Host "[start-demo2] Data dir:   $dataDir"
Write-Host "[start-demo2] Port:       $port"
Write-Host "[start-demo2] URL:        $url"
Write-Host "[start-demo2] Mode:       $mode"
Write-Host "[start-demo2] Guard:      $parallelNote"

if ($Background) {
  $proc = Start-Process -FilePath $exePath -ArgumentList $args -WorkingDirectory $repoRoot -PassThru
  Start-Sleep -Seconds 2
  Write-Host "[start-demo2] Started in background. PID: $($proc.Id)"
  Write-Host "[start-demo2] Verify with: Invoke-WebRequest -UseBasicParsing $url/"
  exit 0
}

& $exePath @args
