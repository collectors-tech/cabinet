param(
  [string]$OutputDir = "bin",
  [string]$BinaryName = "cabinet.exe",
  [switch]$SkipUIBuild,
  [switch]$InstallUIDeps
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$targetDir = Join-Path $repoRoot $OutputDir
$targetPath = Join-Path $targetDir $BinaryName

if (-not (Test-Path $targetDir)) {
  New-Item -ItemType Directory -Path $targetDir | Out-Null
}

if (-not $SkipUIBuild) {
  Write-Host "[build-cabinet] Building ui.web static bundle first"
  & (Join-Path $repoRoot "scripts\build-ui-static.ps1") -InstallDeps:$InstallUIDeps
  if ($LASTEXITCODE -ne 0) {
    throw "ui.web build failed with exit code $LASTEXITCODE"
  }
}

Write-Host "[build-cabinet] Building to $targetPath"
go build -o $targetPath ./cmd/cabinet
if ($LASTEXITCODE -ne 0) {
  throw "go build failed with exit code $LASTEXITCODE"
}

Write-Host "[build-cabinet] Done"
