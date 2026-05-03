$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $repoRoot "scripts\lib\cabinet-console.ps1")

Write-CabinetBanner -Command "validate-openapi" -Summary "Lint the generated OpenAPI contract."
Write-CabinetSection "OpenAPI"
Write-CabinetKeyValue -Key "Spec" -Value (Join-Path $repoRoot "docs\api\openapi.yaml")
Write-CabinetStatus -State "run" -Message "Running Redocly lint."
npx --yes @redocly/cli@latest lint docs/api/openapi.yaml
Write-CabinetStatus -State "ok" -Message "OpenAPI lint passed."
