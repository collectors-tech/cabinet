param(
  [string]$OutputDir = "bin",
  [string]$BinaryName = "cabinet.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$targetDir = Join-Path $repoRoot $OutputDir
$targetPath = Join-Path $targetDir $BinaryName

if (-not (Test-Path $targetDir)) {
  New-Item -ItemType Directory -Path $targetDir | Out-Null
}

Write-Host "[build-cabinet] Building to $targetPath"
go build -o $targetPath ./cmd/cabinet
if ($LASTEXITCODE -ne 0) {
  throw "go build failed with exit code $LASTEXITCODE"
}

Write-Host "[build-cabinet] Done"
