param(
  [string]$Spec = "cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts",
  [string]$Browser = "electron",
  [switch]$Headed,
  [switch]$NoServer,
  [switch]$RequireE2EHooks,
  [string]$BaseUrl = "http://127.0.0.1:17880",
  [int]$StartupTimeoutSec = 45
)

$ErrorActionPreference = "Stop"

function Write-Step([string]$msg) {
  Write-Host "[cypress.ps1] $msg"
}

function Test-Health([string]$url) {
  try {
    $res = Invoke-WebRequest -Uri "$url/healthz" -UseBasicParsing -TimeoutSec 2
    return $res.StatusCode -eq 200
  }
  catch {
    return $false
  }
}

function Test-E2EHooks([string]$url) {
  try {
    $res = Invoke-WebRequest -Uri "$url/api/test/reset" -Method Post -Body "{}" -ContentType "application/json" -UseBasicParsing -TimeoutSec 2
    return $res.StatusCode -ge 200 -and $res.StatusCode -lt 300
  }
  catch {
    return $false
  }
}

function Stop-PortListener([string]$url) {
  $uri = [Uri]$url
  $port = $uri.Port
  try {
    $listeners = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction Stop
    foreach ($listener in $listeners) {
      if ($listener.OwningProcess -gt 0) {
        Write-Step "Stopping process $($listener.OwningProcess) on port $port"
        Stop-Process -Id $listener.OwningProcess -Force -ErrorAction SilentlyContinue
      }
    }
  }
  catch {
    Write-Step "Unable to inspect port listeners for $port; continuing."
  }
}

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$uiRoot = Join-Path $repoRoot "ui.web"
$configPath = Join-Path $uiRoot "cypress.config.ts"
$specPath = if ([System.IO.Path]::IsPathRooted($Spec)) { $Spec } else { Join-Path $uiRoot $Spec }

if (-not (Test-Path $configPath)) {
  throw "Missing Cypress config: $configPath"
}
if (-not (Test-Path $specPath)) {
  throw "Missing Cypress spec: $specPath"
}

$serverProc = $null
$startedServer = $false
$exitCode = 1

try {
  if (-not $NoServer) {
    $alreadyRunning = Test-Health $BaseUrl
    $canReuse = $alreadyRunning
    if ($alreadyRunning -and $RequireE2EHooks) {
      $canReuse = Test-E2EHooks $BaseUrl
      if (-not $canReuse) {
        Write-Step "Existing server lacks E2E hooks; recycling runtime with CABINET_E2E_MODE=1."
        Stop-PortListener $BaseUrl
      }
    }
    if ($canReuse) {
      Write-Step "Reusing existing server at $BaseUrl"
    }
    else {
      Write-Step "Starting Cabinet server..."
      $serverProc = Start-Process -FilePath "go" -ArgumentList @(
        "run",
        "./cmd/cabinet"
      ) -WorkingDirectory $repoRoot -Environment @{ CABINET_E2E_MODE = "1" } -PassThru
      $startedServer = $true

      $deadline = (Get-Date).AddSeconds($StartupTimeoutSec)
      while ((Get-Date) -lt $deadline) {
        if (Test-Health $BaseUrl) {
          break
        }
        Start-Sleep -Seconds 1
      }
      if (-not (Test-Health $BaseUrl)) {
        throw "Server did not become healthy at $BaseUrl within $StartupTimeoutSec seconds."
      }
      if ($RequireE2EHooks -and -not (Test-E2EHooks $BaseUrl)) {
        throw "Server is healthy but E2E hooks are unavailable at $BaseUrl/api/test/reset."
      }
      Write-Step "Server is healthy."
    }
  }

  Write-Step "Running Cypress spec: $specPath"
  $args = @(
    "cypress", "run",
    "--browser", $Browser,
    "--config-file", $configPath,
    "--spec", $specPath
  )
  if ($Headed) {
    $args += "--headed"
  }

  Push-Location $uiRoot
  try {
    $hadElectronRunAsNode = Test-Path Env:ELECTRON_RUN_AS_NODE
    $originalElectronRunAsNode = $env:ELECTRON_RUN_AS_NODE
    if ($hadElectronRunAsNode) {
      Write-Step "Temporarily clearing ELECTRON_RUN_AS_NODE for Cypress runtime compatibility."
    }
    Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue
    & npx @args
    $exitCode = $LASTEXITCODE
  }
  finally {
    if ($hadElectronRunAsNode) {
      $env:ELECTRON_RUN_AS_NODE = $originalElectronRunAsNode
    } else {
      Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue
    }
    Pop-Location
  }
}
finally {
  if ($startedServer -and $serverProc) {
    Write-Step "Stopping server process $($serverProc.Id)"
    Stop-Process -Id $serverProc.Id -Force -ErrorAction SilentlyContinue
  }
}

exit $exitCode
