param(
  [string]$Version = "",
  [string]$ExpectedCommit = "",
  [string]$OutputDirectory = "dist",
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

$buildRevision = (& git -C $root rev-parse HEAD 2>$null).Trim().ToLowerInvariant()
if ($LASTEXITCODE -ne 0 -or $buildRevision -notmatch '^[0-9a-f]{40}$') {
  throw "Unable to resolve a full source commit for Cabinet packaging."
}
if (-not [string]::IsNullOrWhiteSpace($ExpectedCommit)) {
  $resolvedExpectedCommit = $ExpectedCommit.Trim().ToLowerInvariant()
  if ($resolvedExpectedCommit -notmatch '^[0-9a-f]{40}$') {
    throw "ExpectedCommit must be a full 40-character commit SHA."
  }
  if ($buildRevision -ne $resolvedExpectedCommit) {
    throw "Expected commit $resolvedExpectedCommit but checked out $buildRevision."
  }
}
$worktreeState = @(& git -C $root status --porcelain --untracked-files=all 2>$null)
if ($LASTEXITCODE -ne 0) {
  throw "Unable to verify the Cabinet packaging worktree."
}
if ($worktreeState.Count -gt 0) {
  throw "Cabinet packaging requires a clean source worktree."
}
$buildDate = (& git -C $root show -s --format=%cI HEAD 2>$null).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($buildDate)) {
  throw "Unable to resolve the source commit build date."
}

$out = if ([System.IO.Path]::IsPathRooted($OutputDirectory)) {
  [System.IO.Path]::GetFullPath($OutputDirectory)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $root $OutputDirectory))
}
$stage = Join-Path $out "stage-windows-amd64"
$packageName = "cabinet-$resolvedVersion-windows-amd64-portable.zip"
$packagePath = Join-Path $out $packageName
$checksumPath = "$packagePath.sha256"
$notesPath = Join-Path $out "cabinet-$resolvedVersion-release-notes.md"
$manifestPath = Join-Path $out "cabinet-release-manifest.json"
$disclosurePath = Join-Path $root "release\cabinet-beta-disclosure.json"

Write-CabinetBanner -Command "package-installers" -Summary "Build the Windows portable beta package."
Write-CabinetKeyValue -Key "Version" -Value $resolvedVersion
Write-CabinetKeyValue -Key "Package" -Value $packageName
Write-CabinetHint "This script creates a truthful Windows portable package, not an installer."

New-Item -ItemType Directory -Force -Path $out | Out-Null
Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $stage
New-Item -ItemType Directory -Force -Path $stage | Out-Null
foreach ($artifactPath in @($packagePath, $checksumPath, $notesPath, $manifestPath)) {
  if (Test-Path -LiteralPath $artifactPath) {
    throw "Refusing to overwrite existing release output: $artifactPath"
  }
}

Write-CabinetSection "Static UI"
Write-CabinetStatus -State "run" -Message "Building ui.web static bundle first."
& (Join-Path $root "scripts\build-ui-static.ps1") -InstallDeps:$InstallUIDeps
if ($LASTEXITCODE -ne 0) {
  throw "ui.web build failed with exit code $LASTEXITCODE"
}

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
  go build -ldflags ($ldflags -join " ") -o (Join-Path $stage "cabinet-mcp.exe") ./cmd/cabinet-mcp
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed for windows/amd64 MCP launcher"
  }
}
finally {
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
}

Copy-Item -Path (Join-Path $root "README.md") -Destination (Join-Path $stage "README.md") -Force
Copy-Item -Path (Join-Path $root "openspec\migration\windows-portable-beta.md") -Destination (Join-Path $stage "WINDOWS-PORTABLE-BETA.md") -Force

$disclosureNotes = & node (Join-Path $root "scripts\render-beta-disclosure.mjs") --format release-notes --source $disclosurePath
if ($LASTEXITCODE -ne 0) {
  throw "Failed to render governed Cabinet beta disclosure."
}

@"
# Cabinet $resolvedVersion private beta

Package: ``$packageName``
Commit: ``$buildRevision``
Build date: ``$buildDate``
Channel: private beta

This artefact is a Windows portable package. It is not an installer. Code signing and installer claims are intentionally out of scope until signed installer evidence exists.

Release remains gated on #1864 approval and must not be promoted to ``main`` without explicit approval.

$disclosureNotes
"@ | Set-Content -LiteralPath $notesPath -Encoding utf8

Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $packagePath
$hash = Get-FileHash -LiteralPath $packagePath -Algorithm SHA256
$checksumLine = "$($hash.Hash.ToLowerInvariant())  $packageName`n"
[System.IO.File]::WriteAllText($checksumPath, $checksumLine, [System.Text.UTF8Encoding]::new($false))

$stagePrefixLength = $stage.TrimEnd('\', '/').Length + 1
$packageFiles = @(
  Get-ChildItem -LiteralPath $stage -Recurse -File |
    Sort-Object FullName |
    ForEach-Object {
      $fileHash = Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256
      [ordered]@{
        path = $_.FullName.Substring($stagePrefixLength).Replace('\', '/')
        size_bytes = [int64]$_.Length
        sha256 = $fileHash.Hash.ToLowerInvariant()
      }
    }
)
$releaseManifest = [ordered]@{
  schema_version = 1
  product = "Cabinet"
  channel = "private-beta"
  version = $resolvedVersion
  source_commit = $buildRevision
  build_date = $buildDate
  publication_state = "private_candidate_not_published"
  artifact = [ordered]@{
    target = "windows-amd64"
    kind = "portable_zip"
    filename = $packageName
    sha256_filename = "$packageName.sha256"
    sha256 = $hash.Hash.ToLowerInvariant()
    size_bytes = [int64](Get-Item -LiteralPath $packagePath).Length
  }
  release_notes_filename = [System.IO.Path]::GetFileName($notesPath)
  package_files = $packageFiles
}
$releaseManifest | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $manifestPath -Encoding utf8

Write-CabinetStatus -State "ok" -Message "Windows portable package created."
Write-CabinetKeyValue -Key "Package path" -Value $packagePath
Write-CabinetKeyValue -Key "SHA256 path" -Value $checksumPath
Write-CabinetKeyValue -Key "Release notes" -Value $notesPath
Write-CabinetKeyValue -Key "Release manifest" -Value $manifestPath
