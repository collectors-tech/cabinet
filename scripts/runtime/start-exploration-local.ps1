param(
  [switch]$Rebuild,
  [switch]$FreshData,
  [switch]$Background,
  [switch]$Restart,
  [switch]$AllowParallel,
  [int]$Port = 17880
)

$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$exePath = Join-Path $repoRoot 'bin\cabinet.exe'
$dataDir = Join-Path $repoRoot 'tmp\exploration-local\data'
$profile = 'exploration-local'
$instanceName = 'exploration-local'
$url = "http://127.0.0.1:$Port"

if ($Rebuild) {
  Write-Host '[start-exploration-local] Rebuilding Cabinet first...'
  & (Join-Path $repoRoot 'scripts\build-cabinet.ps1')
}

if (-not (Test-Path $exePath)) {
  throw "Cabinet executable not found: $exePath`nRun .\scripts\build-cabinet.ps1 first or pass -Rebuild."
}

if ($FreshData -and (Test-Path $dataDir)) {
  Write-Host "[start-exploration-local] Removing existing exploration-local data dir: $dataDir"
  Remove-Item -Recurse -Force $dataDir
}

if ($Restart -and $AllowParallel) {
  throw 'Restart cannot be combined with AllowParallel.'
}

New-Item -ItemType Directory -Force -Path $dataDir | Out-Null

$args = @(
  '--no-open-browser',
  '--data-dir', $dataDir,
  '--profile', $profile,
  '--instance-name', $instanceName,
  '--port', $Port
)

$mode = 'reuse-or-attach'
$parallelNote = 'singleton endpoint guard enabled'
if ($Restart) {
  $mode = 'restart'
}
if ($AllowParallel) {
  $mode = 'parallel'
  $parallelNote = 'explicit parallel mode enabled'
}
if ($Restart) {
  $args += '--restart'
}
if ($AllowParallel) {
  $args += '--allow-parallel'
}

Write-Host "[start-exploration-local] Executable: $exePath"
Write-Host "[start-exploration-local] Data dir:   $dataDir"
Write-Host "[start-exploration-local] Port:       $Port"
Write-Host "[start-exploration-local] URL:        $url"
Write-Host "[start-exploration-local] Mode:       $mode"
Write-Host "[start-exploration-local] Guard:      $parallelNote"
Write-Host '[start-exploration-local] Expected first-run path:'
Write-Host '  1. Complete setup wizard'
Write-Host '  2. Choose Auth Mode = local'
Write-Host '  3. Sign in with any valid email + password length >= 7'
Write-Host '  4. Use starter data or Showcase DB for authenticated exploration'

if ($Background) {
  $proc = Start-Process -FilePath $exePath -ArgumentList $args -WorkingDirectory $repoRoot -PassThru
  Start-Sleep -Seconds 2
  Write-Host "[start-exploration-local] Started in background. PID: $($proc.Id)"
  Write-Host "[start-exploration-local] Verify with: Invoke-WebRequest -UseBasicParsing $url/"
  exit 0
}

& $exePath @args
