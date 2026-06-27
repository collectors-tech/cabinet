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
$buildRevision = ""
$buildDate = ""
try {
  $buildRevision = (& git -C $repoRoot rev-parse HEAD 2>$null).Trim()
  $buildDate = (& git -C $repoRoot show -s --format=%cI HEAD 2>$null).Trim()
}
catch {
  $buildRevision = ""
  $buildDate = ""
}

$ldflags = @()
if (-not [string]::IsNullOrWhiteSpace($buildRevision)) {
  $ldflags += "-X"
  $ldflags += "github.com/collectors-tech/cabinet/internal/app.buildRevision=$buildRevision"
}
if (-not [string]::IsNullOrWhiteSpace($buildDate)) {
  $ldflags += "-X"
  $ldflags += "github.com/collectors-tech/cabinet/internal/app.buildDate=$buildDate"
}

if ($ldflags.Count -gt 0) {
  Write-CabinetKeyValue -Key "Revision" -Value $buildRevision
  go build -ldflags ($ldflags -join " ") -o $targetPath ./cmd/cabinet
} else {
  go build -o $targetPath ./cmd/cabinet
}
if ($LASTEXITCODE -ne 0) {
  throw "go build failed with exit code $LASTEXITCODE"
}

Write-CabinetStatus -State "ok" -Message "Cabinet build complete."
