param(
  [switch]$InstallDeps
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $repoRoot "scripts\lib\cabinet-console.ps1")

$uiRoot = Join-Path $repoRoot "ui.web"

Write-CabinetBanner -Command "build-ui-static" -Summary "Compile the web app into internal/ui/static."

if (-not (Test-Path $uiRoot)) {
  throw "ui.web directory not found at $uiRoot"
}

$nodeModules = Join-Path $uiRoot "node_modules"
$installRequired = $InstallDeps -or -not (Test-Path $nodeModules)

Push-Location $uiRoot
try {
  if ($installRequired) {
    Write-CabinetSection "Dependencies"
    Write-CabinetStatus -State "run" -Message "Installing ui.web dependencies with npm ci."
    npm ci
    if ($LASTEXITCODE -ne 0) {
      throw "npm ci failed with exit code $LASTEXITCODE"
    }
  }

  Write-CabinetSection "Build"
  Write-CabinetKeyValue -Key "Source" -Value $uiRoot
  Write-CabinetKeyValue -Key "Output" -Value (Join-Path $repoRoot "internal\ui\static")
  Write-CabinetStatus -State "run" -Message "Running npm build."
  npm run build
  if ($LASTEXITCODE -ne 0) {
    throw "npm run build failed with exit code $LASTEXITCODE"
  }
}
finally {
  Pop-Location
}

Write-CabinetStatus -State "ok" -Message "Static UI build complete."
