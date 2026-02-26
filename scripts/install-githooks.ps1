$ErrorActionPreference = "Stop"

Write-Host "Configuring repository git hooks path to .githooks ..."
git config core.hooksPath .githooks

Write-Host "Hook path configured. Active hooks:"
Get-ChildItem .githooks | ForEach-Object { Write-Host " - $($_.Name)" }
