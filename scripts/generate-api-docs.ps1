$ErrorActionPreference = "Stop"

npx --yes @redocly/cli@latest build-docs docs/api/openapi.yaml -o docs/api/index.html
Write-Host "Generated docs/api/index.html"
