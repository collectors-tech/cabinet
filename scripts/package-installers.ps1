param(
  [string]$Version = "",
  [switch]$InstallUIDeps
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
. (Join-Path $root "scripts\lib\cabinet-console.ps1")

$versionFile = Join-Path $root "release\cabinet-beta-version.json"
if ([string]::IsNullOrWhiteSpace($Version)) {
  if (-not (Test-Path $versionFile)) {
    throw "Missing canonical beta version file: $versionFile"
  }
  $versionPayload = Get-Content -LiteralPath $versionFile -Raw | ConvertFrom-Json
  $Version = [string]$versionPayload.version
}
$resolvedVersion = $Version.Trim()
if ($resolvedVersion -notmatch '^\d+\.\d+\.\d+-beta\.\d+$') {
  throw "Version must be a semantic private-beta version such as 0.1.0-beta.1; got '$Version'"
}

$out = Join-Path $root "dist"
$stage = Join-Path $out "stage-windows-amd64"
$packageName = "cabinet-$resolvedVersion-windows-amd64-portable.zip"
$packagePath = Join-Path $out $packageName
$checksumPath = "$packagePath.sha256"
$notesPath = Join-Path $out "cabinet-$resolvedVersion-release-notes.md"

Write-CabinetBanner -Command "package-installers" -Summary "Build the Windows portable beta package."
Write-CabinetKeyValue -Key "Version" -Value $resolvedVersion
Write-CabinetKeyValue -Key "Package" -Value $packageName
Write-CabinetHint "This script creates a truthful Windows portable package, not an installer."

New-Item -ItemType Directory -Force -Path $out | Out-Null
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $stage
New-Item -ItemType Directory -Force -Path $stage | Out-Null

Write-CabinetSection "Static UI"
Write-CabinetStatus -State "run" -Message "Building ui.web static bundle first."
& (Join-Path $root "scripts\build-ui-static.ps1") -InstallDeps:$InstallUIDeps
if ($LASTEXITCODE -ne 0) {
  throw "ui.web build failed with exit code $LASTEXITCODE"
}

$buildRevision = (& git -C $root rev-parse HEAD 2>$null).Trim()
$buildDate = (& git -C $root show -s --format=%cI HEAD 2>$null).Trim()
$ldflags = @(
  "-X", "github.com/collectors-tech/cabinet/internal/app.buildVersion=$resolvedVersion",
  "-X", "github.com/collectors-tech/cabinet/internal/app.buildRevision=$buildRevision",
  "-X", "github.com/collectors-tech/cabinet/internal/app.buildDate=$buildDate"
)

Write-CabinetSection "Runtime"
Write-CabinetKeyValue -Key "Revision" -Value $buildRevision
Write-CabinetKeyValue -Key "Build date" -Value $buildDate
$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
try {
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  go build -ldflags ($ldflags -join " ") -o (Join-Path $stage "cabinet.exe") ./cmd/cabinet
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed for windows/amd64"
  }
}
finally {
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
}

Copy-Item -Path (Join-Path $root "README.md") -Destination (Join-Path $stage "README.md") -Force
Copy-Item -Path (Join-Path $root "openspec\migration\windows-portable-beta.md") -Destination (Join-Path $stage "WINDOWS-PORTABLE-BETA.md") -Force

@"
# Cabinet $resolvedVersion private beta

Package: ``$packageName``
Commit: ``$buildRevision``
Build date: ``$buildDate``
Channel: private beta

This artefact is a Windows portable package. Code signing and installer claims are intentionally out of scope until signed installer evidence exists.

Release remains gated on #1864 approval and must not be promoted to ``main`` without explicit approval.
"@ | Set-Content -LiteralPath $notesPath -Encoding utf8

Remove-Item -Force -ErrorAction SilentlyContinue $packagePath, $checksumPath
Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $packagePath -Force
$hash = Get-FileHash -LiteralPath $packagePath -Algorithm SHA256
"$($hash.Hash.ToLowerInvariant())  $packageName" | Set-Content -LiteralPath $checksumPath -Encoding ascii

Write-CabinetStatus -State "ok" -Message "Windows portable package created."
Write-CabinetKeyValue -Key "Package path" -Value $packagePath
Write-CabinetKeyValue -Key "SHA256 path" -Value $checksumPath
Write-CabinetKeyValue -Key "Release notes" -Value $notesPath
