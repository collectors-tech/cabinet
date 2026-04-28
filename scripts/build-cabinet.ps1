param(
  [string]$OutputDir = "bin",
  [string]$BinaryName = "cabinet.exe",
  [switch]$InstallUIDeps
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $repoRoot "scripts\lib\cabinet-console.ps1")

$targetDir = Join-Path $repoRoot $OutputDir
$targetPath = Join-Path $targetDir $BinaryName

Write-CabinetBanner -Command "build-cabinet" -Summary "Build static UI assets and the Cabinet runtime binary."

if (-not (Test-Path $targetDir)) {
  New-Item -ItemType Directory -Path $targetDir | Out-Null
}

Write-CabinetSection "Static UI"
Write-CabinetStatus -State "run" -Message "Building ui.web static bundle first."
& (Join-Path $repoRoot "scripts\build-ui-static.ps1") -InstallDeps:$InstallUIDeps
if ($LASTEXITCODE -ne 0) {
  throw "ui.web build failed with exit code $LASTEXITCODE"
}

Write-CabinetSection "Runtime"
Write-CabinetKeyValue -Key "Output" -Value $targetPath
Write-CabinetStatus -State "run" -Message "Building Cabinet executable."
go build -o $targetPath ./cmd/cabinet
if ($LASTEXITCODE -ne 0) {
  throw "go build failed with exit code $LASTEXITCODE"
}

Write-CabinetStatus -State "ok" -Message "Cabinet build complete."
