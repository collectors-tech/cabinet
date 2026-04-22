param(
  [string]$RepoRoot = "d:\projects\collectors-tech\cabinet",
  [string]$LanIP = "192.168.1.8",
  [int]$Port = 17880,
  [switch]$Restart,
  [switch]$AllowParallel
)

$ErrorActionPreference = "Stop"

$exePath = Join-Path $RepoRoot "bin\cabinet.exe"
if (-not (Test-Path $exePath)) {
  throw "Cabinet executable not found: $exePath"
}

if ($Restart -and $AllowParallel) {
  throw "Restart cannot be combined with AllowParallel."
}

$env:CABINET_BIND_MODE = "lan"
$env:CABINET_HOST = "0.0.0.0"
$env:CABINET_PORT = "$Port"
$env:CABINET_WEBAUTHN_RP_ID = $LanIP
$env:CABINET_WEBAUTHN_ORIGIN = "http://$LanIP`:$Port"

$args = @()
$mode = "reuse-or-attach"
if ($Restart) {
  $args += "--restart"
  $mode = "restart"
}
if ($AllowParallel) {
  $args += "--allow-parallel"
  $mode = "parallel"
}

Write-Output "Cabinet LAN launcher mode=$mode port=$Port host=0.0.0.0"
Start-Process -FilePath $exePath -ArgumentList $args -WorkingDirectory $RepoRoot | Out-Null
Start-Sleep -Seconds 2

$proc = Get-Process cabinet -ErrorAction SilentlyContinue | Select-Object -First 1
if (-not $proc) {
  throw "Cabinet process failed to start."
}

$listener = Get-NetTCPConnection -OwningProcess $proc.Id -State Listen -ErrorAction SilentlyContinue |
  Where-Object { $_.LocalPort -eq $Port } |
  Select-Object -First 1

if (-not $listener) {
  throw "Cabinet started (pid=$($proc.Id)) but no listener found on port $Port."
}

Write-Output "Cabinet started. PID=$($proc.Id) LISTEN=$($listener.LocalAddress):$($listener.LocalPort)"
Write-Output "Open locally:   http://127.0.0.1:$Port/"
Write-Output "Open from LAN:  http://$LanIP`:$Port/"
