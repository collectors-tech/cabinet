$ErrorActionPreference = "Stop"

npx --yes @redocly/cli@latest lint docs/api/openapi.yaml
Write-Host "OpenAPI lint passed."
