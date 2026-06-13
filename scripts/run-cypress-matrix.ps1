param(
  [string]$SpecGlob = "ui.web/cypress/e2e/**/*.cy.ts",
  [string]$Browser = "electron",
  [int]$BasePort = 17880,
  [int]$LaneCount = 2,
  [int]$MaxWorkers = 2,
  [switch]$RequireE2EHooks,
  [switch]$SkipDependencyPrep,
  [switch]$SkipRuntimeBuild,
  [switch]$UseContainerImage,
  [string]$ContainerImage = "cabinet:e2e",
  [int]$ContainerStartupTimeoutSec = 60,
  [switch]$KeepContainers,
  [switch]$PlanOnly,
  [string]$RunId = "",
  [string]$LogRoot = ".work-agent\logs\cypress-matrix"
)

$ErrorActionPreference = "Stop"

function ConvertTo-SafeSegment([string]$value) {
  if ([string]::IsNullOrWhiteSpace($value)) {
    return "run"
  }
  $safe = $value -replace '[^A-Za-z0-9._-]+', '-'
  $safe = $safe.Trim('-')
  if ([string]::IsNullOrWhiteSpace($safe)) {
    return "run"
  }
  if ($safe.Length -gt 80) {
    return $safe.Substring(0, 80)
  }
  return $safe
}

function Resolve-Specs([string]$repoRoot, [string]$glob) {
  $normalized = $glob.Replace("/", "\")
  $wildcardRoot = $normalized
  $wildcardIndex = $normalized.IndexOf("*")
  if ($wildcardIndex -ge 0) {
    $wildcardRoot = $normalized.Substring(0, $wildcardIndex)
  }
  $searchRoot = Split-Path -Parent (Join-Path $repoRoot $wildcardRoot)
  while (-not [string]::IsNullOrWhiteSpace($searchRoot) -and -not (Test-Path $searchRoot)) {
    $parent = Split-Path -Parent $searchRoot
    if ($parent -eq $searchRoot) {
      break
    }
    $searchRoot = $parent
  }
  if ([string]::IsNullOrWhiteSpace($searchRoot) -or -not (Test-Path $searchRoot)) {
    throw "Spec search root does not exist for glob: $glob"
  }

  $regex = "^" + [regex]::Escape($normalized).Replace("\*\*", ".*").Replace("\*", "[^\\]*") + "$"
  Get-ChildItem -Path $searchRoot -Recurse -File -Filter "*.cy.ts" |
    ForEach-Object {
      Resolve-Path -Relative $_.FullName
    } |
    ForEach-Object {
      $_.TrimStart(".\").Replace("/", "\")
    } |
    Where-Object {
      $_ -match $regex
    } |
    Sort-Object
}

function Wait-MatrixJobSlot([System.Collections.ArrayList]$jobs, [int]$maxWorkers) {
  while ($jobs.Count -ge $maxWorkers) {
    $finished = Wait-Job -Job $jobs -Any
    [void]$jobs.Remove($finished)
    Receive-Job -Job $finished
    Remove-Job -Job $finished
  }
}

function Get-SourceCommit([string]$repoRoot) {
  Push-Location $repoRoot
  try {
    $commit = (& git rev-parse HEAD 2>$null)
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($commit)) {
      return "unknown"
    }
    return $commit.Trim()
  }
  finally {
    Pop-Location
  }
}

function ConvertTo-ContainerSegment([string]$value) {
  $safe = ConvertTo-SafeSegment $value
  $safe = $safe.ToLowerInvariant() -replace '[^a-z0-9_.-]+', '-'
  $safe = $safe.Trim('-')
  if ([string]::IsNullOrWhiteSpace($safe)) {
    return "lane"
  }
  return $safe
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$sourceCommit = Get-SourceCommit $repoRoot
$runStamp = if ([string]::IsNullOrWhiteSpace($RunId)) { Get-Date -Format "yyyyMMdd-HHmmss" } else { ConvertTo-SafeSegment $RunId }
$containerRunSegment = ConvertTo-ContainerSegment $runStamp
$resolvedLogRoot = if ([System.IO.Path]::IsPathRooted($LogRoot)) { $LogRoot } else { Join-Path $repoRoot $LogRoot }
$runLogDir = Join-Path $resolvedLogRoot $runStamp
$summaryPath = Join-Path $runLogDir "matrix.summary.json"

if ($LaneCount -lt 1) {
  throw "LaneCount must be at least 1."
}
if ($MaxWorkers -lt 1) {
  throw "MaxWorkers must be at least 1."
}
$workerLimit = [Math]::Min($LaneCount, $MaxWorkers)

New-Item -ItemType Directory -Force -Path $runLogDir | Out-Null

$specs = @(Resolve-Specs $repoRoot $SpecGlob)
if ($specs.Count -eq 0) {
  throw "No Cypress specs matched $SpecGlob"
}

$laneSpecs = @()
for ($laneIndex = 0; $laneIndex -lt $LaneCount; $laneIndex++) {
  $laneSpecs += ,@()
}
for ($specIndex = 0; $specIndex -lt $specs.Count; $specIndex++) {
  $laneIndex = $specIndex % $LaneCount
  $laneSpecs[$laneIndex] += $specs[$specIndex]
}

Write-Host "[cypress-matrix] Run log dir: $runLogDir"
Write-Host "[cypress-matrix] Run summary: $summaryPath"
Write-Host "[cypress-matrix] Specs: $($specs.Count); lanes: $LaneCount; workers: $workerLimit; base port: $BasePort; commit: $sourceCommit"
if ($UseContainerImage) {
  Write-Host "[cypress-matrix] Container image lanes enabled: image=$ContainerImage startup_timeout_sec=$ContainerStartupTimeoutSec keep_containers=$($KeepContainers.IsPresent)"
}

$lanePlans = @()
for ($laneIndex = 0; $laneIndex -lt $LaneCount; $laneIndex++) {
  $laneNumber = $laneIndex + 1
  $lanePort = $BasePort + $laneIndex
  $laneLogDir = Join-Path $runLogDir "lane-$laneNumber"
  $containerName = "cabinet-cypress-$containerRunSegment-lane-$laneNumber"
  $containerVolume = "$containerName-data"
  $lanePlans += [pscustomobject]@{
    lane = $laneNumber
    port = $lanePort
    base_url = "http://127.0.0.1:$lanePort"
    data_dir = Join-Path $repoRoot ".tmp\cypress-runtime-$lanePort"
    profile = "e2e-cypress-$lanePort"
    instance_name = "cypress-$lanePort"
    use_container_image = $UseContainerImage.IsPresent
    container_image = if ($UseContainerImage) { $ContainerImage } else { $null }
    container_name = if ($UseContainerImage) { $containerName } else { $null }
    container_volume = if ($UseContainerImage) { $containerVolume } else { $null }
    source_commit = $sourceCommit
    log_dir = $laneLogDir
    specs = @($laneSpecs[$laneIndex])
  }
}
$specCountsByLane = @($lanePlans | ForEach-Object { $_.specs.Count })
$activeLaneCount = @($specCountsByLane | Where-Object { $_ -gt 0 }).Count
$emptyLaneCount = $LaneCount - $activeLaneCount

if ($PlanOnly) {
  $summary = [ordered]@{
    timestamp = (Get-Date).ToString("o")
    exit_code = 0
    plan_only = $true
    source_commit = $sourceCommit
    spec_glob = $SpecGlob
    spec_count = $specs.Count
    base_port = $BasePort
    lane_count = $LaneCount
    max_workers = $MaxWorkers
    worker_limit = $workerLimit
    use_container_image = $UseContainerImage.IsPresent
    container_image = if ($UseContainerImage) { $ContainerImage } else { $null }
    container_startup_timeout_sec = if ($UseContainerImage) { $ContainerStartupTimeoutSec } else { $null }
    keep_containers = $KeepContainers.IsPresent
    active_lane_count = $activeLaneCount
    empty_lane_count = $emptyLaneCount
    spec_counts_by_lane = $specCountsByLane
    log_dir = $runLogDir
    lanes = $lanePlans
  }
  $summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $summaryPath -Encoding UTF8
  Write-Host "[cypress-matrix] Plan-only summary written: $summaryPath"
  exit 0
}

$jobs = [System.Collections.ArrayList]::new()
$laneResults = @()
for ($laneIndex = 0; $laneIndex -lt $LaneCount; $laneIndex++) {
  $laneNumber = $laneIndex + 1
  $assignedSpecs = @($laneSpecs[$laneIndex])
  if ($assignedSpecs.Count -eq 0) {
    continue
  }
  $laneResults += Wait-MatrixJobSlot $jobs $workerLimit
  $lanePlan = $lanePlans[$laneIndex]
  $lanePort = $lanePlan.port
  $laneLogDir = $lanePlan.log_dir
  New-Item -ItemType Directory -Force -Path $laneLogDir | Out-Null
  $job = Start-Job -Name "cypress-lane-$laneNumber" -ArgumentList @(
    $repoRoot,
    $assignedSpecs,
    $Browser,
    $lanePort,
    $laneLogDir,
    $RequireE2EHooks.IsPresent,
    $SkipDependencyPrep.IsPresent,
    $SkipRuntimeBuild.IsPresent,
    $UseContainerImage.IsPresent,
    $ContainerImage,
    $ContainerStartupTimeoutSec,
    $KeepContainers.IsPresent,
    $laneNumber,
    $lanePlan.data_dir,
    $lanePlan.profile,
    $lanePlan.instance_name,
    $lanePlan.container_name,
    $lanePlan.container_volume,
    $sourceCommit
  ) -ScriptBlock {
    param(
      [string]$repoRoot,
      [string[]]$assignedSpecs,
      [string]$browser,
      [int]$lanePort,
      [string]$laneLogDir,
      [bool]$requireE2EHooks,
      [bool]$skipDependencyPrep,
      [bool]$skipRuntimeBuild,
      [bool]$useContainerImage,
      [string]$containerImage,
      [int]$containerStartupTimeoutSec,
      [bool]$keepContainers,
      [int]$laneNumber,
      [string]$laneDataDir,
      [string]$laneProfile,
      [string]$laneInstanceName,
      [string]$containerName,
      [string]$containerVolume,
      [string]$sourceCommit
    )

    function Test-LaneHealth([string]$baseUrl) {
      try {
        $response = Invoke-WebRequest -Uri "$baseUrl/healthz" -UseBasicParsing -TimeoutSec 2
        return $response.StatusCode -eq 200
      } catch {
        return $false
      }
    }

    $containerStarted = $false
    if ($useContainerImage) {
      docker rm -f $containerName *> $null
      docker volume rm $containerVolume *> $null
      $dockerArgs = @(
        "run", "-d",
        "--name", $containerName,
        "-p", "$($lanePort):17880",
        "-v", "$($containerVolume):/data",
        $containerImage,
        "--no-open-browser",
        "--listen", "0.0.0.0:17880",
        "--data-dir", "/data",
        "--profile", $laneProfile,
        "--instance-name", $laneInstanceName,
        "--allow-parallel"
      )
      & docker @dockerArgs | Out-Host
      if ($LASTEXITCODE -ne 0) {
        throw "Failed to start container lane $laneNumber from image $containerImage"
      }
      $containerStarted = $true
      $deadline = (Get-Date).AddSeconds($containerStartupTimeoutSec)
      while ((Get-Date) -lt $deadline) {
        if (Test-LaneHealth "http://127.0.0.1:$lanePort") {
          break
        }
        Start-Sleep -Seconds 1
      }
      if (-not (Test-LaneHealth "http://127.0.0.1:$lanePort")) {
        throw "Container lane $laneNumber did not become healthy at http://127.0.0.1:$lanePort within $containerStartupTimeoutSec seconds."
      }
    }

    $laneResults = @()
    try {
      foreach ($spec in $assignedSpecs) {
      $specRelativeToUi = $spec
      if ($specRelativeToUi.StartsWith("ui.web\")) {
        $specRelativeToUi = $specRelativeToUi.Substring("ui.web\".Length)
      }
      $logName = "lane-$laneNumber-$([System.IO.Path]::GetFileNameWithoutExtension($specRelativeToUi))"
      $args = @(
        "-NoLogo",
        "-NoProfile",
        "-File", (Join-Path $repoRoot "cypress.ps1"),
        "-Spec", $specRelativeToUi,
        "-Browser", $browser,
        "-BaseUrl", "http://127.0.0.1:$lanePort",
        "-LogDir", $laneLogDir,
        "-LogName", $logName
      )
      if ($requireE2EHooks) {
        $args += "-RequireE2EHooks"
      }
      if ($skipDependencyPrep) {
        $args += "-SkipDependencyPrep"
      }
      if ($skipRuntimeBuild -or $useContainerImage) {
        $args += "-SkipRuntimeBuild"
      }
      if ($useContainerImage) {
        $args += "-ReuseServer"
      }

      & pwsh @args 2>&1 | ForEach-Object {
        Write-Host $_
      }
      $exitCode = $LASTEXITCODE
      $laneResults += [pscustomobject]@{
        spec = $spec
        base_url = "http://127.0.0.1:$lanePort"
        exit_code = $exitCode
      }
      if ($exitCode -ne 0) {
        break
      }
    }
    }
    finally {
      if ($containerStarted -and -not $keepContainers) {
        docker rm -f $containerName *> $null
        docker volume rm $containerVolume *> $null
      }
    }

    [pscustomobject]@{
      lane = $laneNumber
      port = $lanePort
      base_url = "http://127.0.0.1:$lanePort"
      data_dir = $laneDataDir
      profile = $laneProfile
      instance_name = $laneInstanceName
      use_container_image = $useContainerImage
      container_image = if ($useContainerImage) { $containerImage } else { $null }
      container_name = if ($useContainerImage) { $containerName } else { $null }
      container_volume = if ($useContainerImage) { $containerVolume } else { $null }
      container_started = $containerStarted
      container_kept = $keepContainers
      source_commit = $sourceCommit
      log_dir = $laneLogDir
      results = $laneResults
      exit_code = if (($laneResults | Where-Object { $_.exit_code -ne 0 }).Count -gt 0) { 1 } else { 0 }
    }
  }
  [void]$jobs.Add($job)
}

foreach ($job in @($jobs)) {
  Wait-Job -Job $job | Out-Null
  $laneResults += Receive-Job -Job $job
  Remove-Job -Job $job
}

$cleanLaneResults = @(
  $laneResults | ForEach-Object {
    [pscustomobject]@{
      lane = $_.lane
      port = $_.port
      base_url = $_.base_url
      data_dir = $_.data_dir
      profile = $_.profile
      instance_name = $_.instance_name
      use_container_image = $_.use_container_image
      container_image = $_.container_image
      container_name = $_.container_name
      container_volume = $_.container_volume
      container_started = $_.container_started
      container_kept = $_.container_kept
      source_commit = $_.source_commit
      log_dir = $_.log_dir
      results = $_.results
      exit_code = $_.exit_code
    }
  }
)
$exitCode = if (($cleanLaneResults | Where-Object { $_.exit_code -ne 0 }).Count -gt 0) { 1 } else { 0 }
$completedActiveLaneCount = @($cleanLaneResults | Where-Object { $_.results.Count -gt 0 }).Count
$completedEmptyLaneCount = $LaneCount - $completedActiveLaneCount
$summary = [ordered]@{
  timestamp = (Get-Date).ToString("o")
  exit_code = $exitCode
  plan_only = $false
  source_commit = $sourceCommit
  spec_glob = $SpecGlob
  spec_count = $specs.Count
  base_port = $BasePort
  lane_count = $LaneCount
  max_workers = $MaxWorkers
  worker_limit = $workerLimit
  use_container_image = $UseContainerImage.IsPresent
  container_image = if ($UseContainerImage) { $ContainerImage } else { $null }
  container_startup_timeout_sec = if ($UseContainerImage) { $ContainerStartupTimeoutSec } else { $null }
  keep_containers = $KeepContainers.IsPresent
  active_lane_count = $completedActiveLaneCount
  empty_lane_count = $completedEmptyLaneCount
  spec_counts_by_lane = $specCountsByLane
  log_dir = $runLogDir
  lanes = $cleanLaneResults
}
$summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $summaryPath -Encoding UTF8
Write-Host "[cypress-matrix] Summary written: $summaryPath"
exit $exitCode
