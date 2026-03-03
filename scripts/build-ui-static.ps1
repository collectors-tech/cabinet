param(
  [switch]$InstallDeps
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$uiRoot = Join-Path $repoRoot "ui.web"

if (-not (Test-Path $uiRoot)) {
  throw "ui.web directory not found at $uiRoot"
}

$nodeModules = Join-Path $uiRoot "node_modules"
$installRequired = $InstallDeps -or -not (Test-Path $nodeModules)

Push-Location $uiRoot
try {
  if ($installRequired) {
    Write-Host "[build-ui-static] Installing ui.web dependencies (npm ci)"
    npm ci
    if ($LASTEXITCODE -ne 0) {
      throw "npm ci failed with exit code $LASTEXITCODE"
    }
  }

  Write-Host "[build-ui-static] Building ui.web into internal/ui/static"
  npm run build
  if ($LASTEXITCODE -ne 0) {
    throw "npm run build failed with exit code $LASTEXITCODE"
  }
}
finally {
  Pop-Location
}

Write-Host "[build-ui-static] Done"
