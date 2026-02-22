param(
  [Parameter(Mandatory = $true)][string]$Version
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$out = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $out | Out-Null

$targets = @(
  @{ GOOS = "windows"; GOARCH = "amd64"; EXE = "cabinet.exe"; PACKAGE = "cabinet-$Version-windows-amd64.zip" },
  @{ GOOS = "darwin"; GOARCH = "amd64"; EXE = "cabinet"; PACKAGE = "cabinet-$Version-macos-amd64.zip" },
  @{ GOOS = "darwin"; GOARCH = "arm64"; EXE = "cabinet"; PACKAGE = "cabinet-$Version-macos-arm64.zip" }
)

foreach ($t in $targets) {
  $stage = Join-Path $out ("stage-" + $t.GOOS + "-" + $t.GOARCH)
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $stage
  New-Item -ItemType Directory -Force -Path $stage | Out-Null

  $env:GOOS = $t.GOOS
  $env:GOARCH = $t.GOARCH
  & "C:\Program Files\Go\bin\go.exe" build -o (Join-Path $stage $t.EXE) ./cmd/cabinet
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed for $($t.GOOS)/$($t.GOARCH)"
  }

  Copy-Item -Path (Join-Path $root "README.md") -Destination (Join-Path $stage "README.md") -Force
  Compress-Archive -Path (Join-Path $stage "*") -DestinationPath (Join-Path $out $t.PACKAGE) -Force
}

Write-Host "Installers packaged under $out"
