param(
  [string]$ClerkPublishableKey,
  [switch]$Rebuild,
  [switch]$FreshData,
  [switch]$Background,
  [int]$Port = 17883
)

$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$exePath = Join-Path $repoRoot 'bin\cabinet.exe'
$dataDir = Join-Path $repoRoot 'tmp\exploration-clerk\data'
$profile = 'exploration-clerk'
$instanceName = 'exploration-clerk'
$url = "http://127.0.0.1:$Port"

if ([string]::IsNullOrWhiteSpace($ClerkPublishableKey)) {
  $ClerkPublishableKey = $env:VITE_CLERK_PUBLISHABLE_KEY
}

if ([string]::IsNullOrWhiteSpace($ClerkPublishableKey)) {
  throw 'Clerk publishable key is required. Pass -ClerkPublishableKey or set VITE_CLERK_PUBLISHABLE_KEY before launching Clerk exploration.'
}

$previousClerkKey = $env:VITE_CLERK_PUBLISHABLE_KEY
$previousIdentityMode = $env:CABINET_AUTH_IDENTITY_MODE

$env:VITE_CLERK_PUBLISHABLE_KEY = $ClerkPublishableKey
$env:CABINET_AUTH_IDENTITY_MODE = 'clerk'

try {
  if ($Rebuild) {
    Write-Host '[start-exploration-clerk] Rebuilding Cabinet first...'
    & (Join-Path $repoRoot 'scripts\build-cabinet.ps1')
  }

  if (-not (Test-Path $exePath)) {
    throw "Cabinet executable not found: $exePath`nRun .\scripts\build-cabinet.ps1 first or pass -Rebuild."
  }

  if ($FreshData -and (Test-Path $dataDir)) {
    Write-Host "[start-exploration-clerk] Removing existing exploration-clerk data dir: $dataDir"
    Remove-Item -Recurse -Force $dataDir
  }

  New-Item -ItemType Directory -Force -Path $dataDir | Out-Null

  $existing = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
  if ($existing) {
    $pids = @($existing | Select-Object -ExpandProperty OwningProcess -Unique)
    throw "Port $Port is already in use by PID(s): $($pids -join ', '). Stop the existing listener first."
  }

  $args = @(
    '--allow-parallel',
    '--no-open-browser',
    '--data-dir', $dataDir,
    '--profile', $profile,
    '--instance-name', $instanceName,
    '--port', $Port
  )

  Write-Host "[start-exploration-clerk] Executable: $exePath"
  Write-Host "[start-exploration-clerk] Data dir:   $dataDir"
  Write-Host "[start-exploration-clerk] Port:       $Port"
  Write-Host "[start-exploration-clerk] URL:        $url"
  Write-Host '[start-exploration-clerk] Effective auth env:'
  Write-Host '  CABINET_AUTH_IDENTITY_MODE=clerk'
  Write-Host ('  VITE_CLERK_PUBLISHABLE_KEY=' + $ClerkPublishableKey)
  Write-Host '[start-exploration-clerk] Expected Clerk exploration path:'
  Write-Host '  1. Use a Clerk-allowed origin for the runtime URL'
  Write-Host '  2. Complete setup wizard with Auth Mode = clerk when first-run config is required'
  Write-Host '  3. Use /clerk/sign-in for Clerk route validation'
  Write-Host '  4. Treat invalid-domain/passkey failures as auth-environment setup issues first'

  if ($Background) {
    $proc = Start-Process -FilePath $exePath -ArgumentList $args -WorkingDirectory $repoRoot -PassThru
    Start-Sleep -Seconds 2
    Write-Host "[start-exploration-clerk] Started in background. PID: $($proc.Id)"
    Write-Host "[start-exploration-clerk] Verify with: Invoke-WebRequest -UseBasicParsing $url/api/auth/provider-options"
    exit 0
  }

  & $exePath @args
}
finally {
  $env:VITE_CLERK_PUBLISHABLE_KEY = $previousClerkKey
  $env:CABINET_AUTH_IDENTITY_MODE = $previousIdentityMode
}
