param(
  [string]$Owner = "collectors-tech",
  [string]$Repo = "cabinet"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not $env:GITHUB_TOKEN) {
  throw "GITHUB_TOKEN is required. Set it in the environment and rerun."
}

$headers = @{
  Authorization = "Bearer $($env:GITHUB_TOKEN)"
  Accept        = "application/vnd.github+json"
  "X-GitHub-Api-Version" = "2022-11-28"
  "User-Agent"  = "cabinet-backlog-bootstrap"
}

function Get-Milestones {
  param([string]$State = "all")
  Invoke-RestMethod -Method Get -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/milestones?state=$State&per_page=100" -ErrorAction Stop
}

function Ensure-Milestone {
  param(
    [string]$Title,
    [string]$Description
  )
  $existing = Get-Milestones | Where-Object { $_.PSObject.Properties.Name -contains "title" -and $_.title -eq $Title } | Select-Object -First 1
  if ($existing) {
    Write-Host "Milestone exists: $Title (#$($existing.number))"
    return $existing.number
  }
  $body = @{ title = $Title; description = $Description } | ConvertTo-Json
  try {
    $created = Invoke-RestMethod -Method Post -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/milestones" -Body $body -ErrorAction Stop
    Write-Host "Created milestone: $Title (#$($created.number))"
    return $created.number
  } catch {
    $msg = $_.Exception.Message
    if ($msg -match "already_exists") {
      $existingRetry = Get-Milestones | Where-Object { $_.PSObject.Properties.Name -contains "title" -and $_.title -eq $Title } | Select-Object -First 1
      if ($existingRetry) {
        Write-Host "Milestone exists: $Title (#$($existingRetry.number))"
        return $existingRetry.number
      }
    }
    throw
  }
}

function Get-OpenIssues {
  Invoke-RestMethod -Method Get -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/issues?state=open&per_page=100" -ErrorAction Stop
}

function Get-Labels {
  Invoke-RestMethod -Method Get -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/labels?per_page=100" -ErrorAction Stop
}

function Ensure-Label {
  param(
    [string]$Name,
    [string]$Color = "1d76db",
    [string]$Description = ""
  )
  $existing = Get-Labels | Where-Object { $_.PSObject.Properties.Name -contains "name" -and $_.name -eq $Name } | Select-Object -First 1
  if ($existing) {
    return
  }
  $payload = @{
    name = $Name
    color = $Color
    description = $Description
  } | ConvertTo-Json
  try {
    Invoke-RestMethod -Method Post -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/labels" -Body $payload -ErrorAction Stop | Out-Null
    Write-Host "Created label: $Name"
  } catch {
    $msg = $_.Exception.Message
    if ($msg -match "already_exists") {
      Write-Host "Label exists: $Name"
      return
    }
    throw
  }
}

function Ensure-Issue {
  param(
    [string]$Title,
    [string]$Body,
    [int]$MilestoneNumber,
    [string[]]$Labels
  )
  $existing = Get-OpenIssues | Where-Object { $_.PSObject.Properties.Name -contains "title" -and -not $_.pull_request -and $_.title -eq $Title } | Select-Object -First 1
  if ($existing) {
    Write-Host "Issue exists: $Title (#$($existing.number))"
    return $existing.number
  }
  $payload = @{
    title     = $Title
    body      = $Body
    milestone = $MilestoneNumber
    labels    = $Labels
  } | ConvertTo-Json -Depth 5
  $created = Invoke-RestMethod -Method Post -Headers $headers -Uri "https://api.github.com/repos/$Owner/$Repo/issues" -Body $payload -ErrorAction Stop
  Write-Host "Created issue: $Title (#$($created.number))"
  return $created.number
}

$milestones = @(
  @{
    key = "M1"
    title = "M1 - Foundation and Core Experience"
    description = "Install, profiles, WebAuthn, collection core, photos, search, import/export, backups."
  },
  @{
    key = "M2"
    title = "M2 - Discovery and Matching"
    description = "Scanner engine, provider integration, matching engine, not-in-collection workflows."
  },
  @{
    key = "M3"
    title = "M3 - Intelligence and Commercial"
    description = "Wishlist, price tracking, dashboard, AI assist, licensing, settings."
  },
  @{
    key = "M4"
    title = "M4 - Hardening and Beta Launch"
    description = "Error handling, diagnostics, updater hardening, NFR/security validation, beta readiness."
  }
)

$milestoneMap = @{}
foreach ($m in $milestones) {
  $milestoneMap[$m.key] = Ensure-Milestone -Title $m.title -Description $m.description
}

$labels = @(
  @{ name = "feature"; color = "0e8a16"; description = "Feature work" },
  @{ name = "enhancement"; color = "a2eeef"; description = "Enhancement work" },
  @{ name = "core"; color = "0052cc"; description = "Core application" },
  @{ name = "security"; color = "d93f0b"; description = "Security/authentication/secrets" },
  @{ name = "media"; color = "5319e7"; description = "Photo and media pipeline" },
  @{ name = "data"; color = "fbca04"; description = "Import/export/backup/data flows" },
  @{ name = "scanner"; color = "1d76db"; description = "Discovery scanner and matching" },
  @{ name = "provider"; color = "bfdadc"; description = "External provider integration" },
  @{ name = "ux"; color = "c2e0c6"; description = "User experience and dashboard" },
  @{ name = "wishlist"; color = "fef2c0"; description = "Wishlist and target-price features" },
  @{ name = "pricing"; color = "ff9f1c"; description = "Price tracking and trends" },
  @{ name = "ai"; color = "6f42c1"; description = "AI assist features" },
  @{ name = "licensing"; color = "d4c5f9"; description = "License and entitlement features" },
  @{ name = "quality"; color = "ededed"; description = "Quality, reliability, and diagnostics" },
  @{ name = "release"; color = "0b3b8c"; description = "Release and beta readiness" },
  @{ name = "high-priority"; color = "b60205"; description = "High-priority backlog item" }
)
foreach ($l in $labels) {
  Ensure-Label -Name $l.name -Color $l.color -Description $l.description
}

$issues = @(
  @{
    title = "[Backlog] Application Core and Runtime"
    milestone = "M1"
    labels = @("feature", "core", "high-priority")
    body = @"
## Scope
Build application core for desktop (Windows + macOS), embedded web UI, local SQLite, media folder, signed updates, and stable/beta channels.

## Acceptance Criteria
- [ ] Single installer path works on Windows and macOS.
- [ ] Embedded web UI loads from app shell.
- [ ] Local SQLite initializes with migrations.
- [ ] Signed update verification exists.
- [ ] Stable/beta channel selection is supported.
"@
  },
  @{
    title = "[Backlog] WebAuthn Login, Locking, and Recovery"
    milestone = "M1"
    labels = @("feature", "security", "high-priority")
    body = @"
## Scope
Implement mandatory WebAuthn login in v1, app lock behavior, inactivity auto-lock, and credential recovery/reset flows.

## Acceptance Criteria
- [ ] First launch requires creating WebAuthn credential.
- [ ] Unlock uses WebAuthn assertion.
- [ ] Inactivity auto-lock is enforced.
- [ ] Recovery passphrase fallback path is available when configured.
- [ ] Credential reset flow exists for no-usable-authenticator scenario.
"@
  },
  @{
    title = "[Backlog] Multi-Profile Local User Support"
    milestone = "M1"
    labels = @("feature", "core")
    body = @"
## Scope
Add multi-profile support with isolated DB, settings, API keys, and license per profile.

## Acceptance Criteria
- [ ] Profile creation and switching from login screen.
- [ ] Per-profile storage isolation.
- [ ] Per-profile key/license separation.
"@
  },
  @{
    title = "[Backlog] Collection Core (Canonical Items + Instances)"
    milestone = "M1"
    labels = @("feature", "core", "high-priority")
    body = @"
## Scope
Implement canonical item and instance data model and CRUD flows.

## Acceptance Criteria
- [ ] Canonical fields from spec are implemented.
- [ ] Instance fields and status values are enforced.
- [ ] One canonical item can have many instances.
- [ ] No auto-merge without explicit confirmation.
"@
  },
  @{
    title = "[Backlog] Photo System (Upload, Derivatives, Primary, Fullscreen)"
    milestone = "M1"
    labels = @("feature", "media")
    body = @"
## Scope
Support desktop/mobile upload, local originals, thumbnail/preview generation, primary image management, and rebuild flow.

## Acceptance Criteria
- [ ] Upload from desktop and mobile browser.
- [ ] Original stored locally with thumbnail+preview derivatives.
- [ ] Primary image can be set and changed.
- [ ] Thumbnail rebuild job available.
"@
  },
  @{
    title = "[Backlog] Barcode System and Variant Handling"
    milestone = "M1"
    labels = @("feature", "core")
    body = @"
## Scope
Implement barcode detection/entry, multiple barcodes per item, local lookup, external lookup, and duplicate variant handling.

## Acceptance Criteria
- [ ] Manual barcode entry and storage.
- [ ] Multiple barcodes per canonical item.
- [ ] Local barcode match lookup.
- [ ] External provider lookup path.
- [ ] Duplicate barcode handling prompts explicit mapping choice.
"@
  },
  @{
    title = "[Backlog] Search, Filters, and Saved Views"
    milestone = "M1"
    labels = @("feature", "core")
    body = @"
## Scope
Build FTS search with filters, saved filters, and sort options.

## Acceptance Criteria
- [ ] Full-text search across item/instance data.
- [ ] Filters for brand/condition/status/tags/scale.
- [ ] Saved filter CRUD.
- [ ] Sort by date added/price/part number.
"@
  },
  @{
    title = "[Backlog] Data Management (Import/Export, Reindex, Repair)"
    milestone = "M1"
    labels = @("feature", "data", "high-priority")
    body = @"
## Scope
Implement JSON/CSV import-export with dry-run and conflict resolution, plus reindex/repair tools.

## Acceptance Criteria
- [ ] JSON full backup export + CSV item export.
- [ ] JSON snapshot import + CSV mapping import.
- [ ] Dry-run preview before apply.
- [ ] Conflict resolution actions (merge/create/skip).
- [ ] Schema version checks.
- [ ] Reindex and repair operations available.
"@
  },
  @{
    title = "[Backlog] Local Backup and Restore Reliability"
    milestone = "M1"
    labels = @("feature", "data", "security")
    body = @"
## Scope
Implement automatic backup frequency controls and restore workflow with validation.

## Acceptance Criteria
- [ ] Automatic local backups by schedule.
- [ ] Restore flow validates snapshot integrity.
- [ ] Restore tested on Windows and macOS.
"@
  },
  @{
    title = "[Backlog] Scanner Engine (Query Sets, Scheduling, Rate Limits)"
    milestone = "M2"
    labels = @("feature", "scanner", "high-priority")
    body = @"
## Scope
Build scanner query model, manual/scheduled execution, rate limiting, retry/backoff, and candidate persistence.

## Acceptance Criteria
- [ ] Query sets support keywords/exclusions/max-price/region/condition.
- [ ] Manual Run Now and scheduled runs.
- [ ] Rate limiting and retry/backoff.
- [ ] Candidate lifecycle stores first seen/last seen/status.
"@
  },
  @{
    title = "[Backlog] eBay Provider v1 Integration"
    milestone = "M2"
    labels = @("feature", "scanner", "provider")
    body = @"
## Scope
Integrate eBay search provider for scanner and external barcode lookup.

## Acceptance Criteria
- [ ] Provider credentials/config in settings.
- [ ] Query execution returns normalized candidates.
- [ ] Provider health indicator and failure logging.
"@
  },
  @{
    title = "[Backlog] Matching Engine and Confidence Rules"
    milestone = "M2"
    labels = @("feature", "scanner", "high-priority")
    body = @"
## Scope
Match candidates to canonical items using part number extraction and confidence scoring.

## Acceptance Criteria
- [ ] Candidate classification states: matched/suggested/not-in-collection.
- [ ] Confidence computation is persisted and visible.
- [ ] Low-confidence candidates require user review.
"@
  },
  @{
    title = "[Backlog] Not In My Collection Panel and Actions"
    milestone = "M2"
    labels = @("feature", "scanner", "ux")
    body = @"
## Scope
Deliver panel for candidate triage with filters and actions.

## Acceptance Criteria
- [ ] Filters by price/query/date.
- [ ] Actions: ignore, add to wishlist, track price, create item.
- [ ] Ignore rules can be reset from settings.
"@
  },
  @{
    title = "[Backlog] Wishlist and Target Price Signals"
    milestone = "M3"
    labels = @("feature", "wishlist")
    body = @"
## Scope
Implement wishlist linked to canonical items with target price and hit highlighting.

## Acceptance Criteria
- [ ] Wishlist CRUD with priority and notes.
- [ ] Scanner hits linked and highlighted.
- [ ] Below-target indicator shown when applicable.
"@
  },
  @{
    title = "[Backlog] Price Tracking and History Export"
    milestone = "M3"
    labels = @("feature", "pricing", "high-priority")
    body = @"
## Scope
Implement tracked-item daily snapshots and trend views.

## Acceptance Criteria
- [ ] Daily snapshots capture min/median/latest.
- [ ] Graph view for item history.
- [ ] Per-source breakdown supported.
- [ ] Price history export available.
"@
  },
  @{
    title = "[Backlog] Dashboard (Discoveries, Hits, Drops, Stats)"
    milestone = "M3"
    labels = @("feature", "ux")
    body = @"
## Scope
Build dashboard summarizing weekly collector workflow signals.

## Acceptance Criteria
- [ ] Shows new discoveries, wishlist hits, price drops, recently added items.
- [ ] Shows total items/instances (+ optional estimated value).
- [ ] Cards deep-link to relevant detailed views.
"@
  },
  @{
    title = "[Backlog] AI Assist (OpenAI) with Confirmation Guardrails"
    milestone = "M3"
    labels = @("feature", "ai", "security")
    body = @"
## Scope
Add OpenAI-based assist for photo/title metadata suggestions with confidence and explicit confirmation.

## Acceptance Criteria
- [ ] API key configuration and connectivity test.
- [ ] Photo identify, part number extraction, normalized field suggestions.
- [ ] Confidence score shown for every suggestion.
- [ ] No auto-create/update without user confirmation.
- [ ] AI can be disabled per profile.
"@
  },
  @{
    title = "[Backlog] Licensing System and Feature Gating"
    milestone = "M3"
    labels = @("feature", "licensing", "high-priority")
    body = @"
## Scope
Implement free-tier limits, Pro unlock via signed license file, and offline validation.

## Acceptance Criteria
- [ ] License import and signature verification.
- [ ] Offline entitlement validation.
- [ ] Free-tier item cap enforced.
- [ ] Pro gates for scanner automation, price tracking, AI assist.
- [ ] Clear license status UI for valid/invalid/expired states.
"@
  },
  @{
    title = "[Backlog] Settings Surface and Secret Storage"
    milestone = "M3"
    labels = @("feature", "security", "core")
    body = @"
## Scope
Complete settings for scanner schedule, update channel, credentials, backups, DB location, and maintenance actions.

## Acceptance Criteria
- [ ] Per-profile settings persist.
- [ ] API keys use OS-backed secure storage.
- [ ] Rebuild thumbnails and reset ignore rules actions exist.
"@
  },
  @{
    title = "[Backlog] Error Handling and Crash Recovery UX"
    milestone = "M4"
    labels = @("feature", "quality")
    body = @"
## Scope
Standardize failure handling for scanner, AI, and provider failures, including crash recovery prompt.

## Acceptance Criteria
- [ ] Scanner and AI failures logged with actionable messages.
- [ ] Provider health indicator surfaces degraded state.
- [ ] Retry flow for failed scans.
- [ ] Crash recovery prompt appears after abnormal shutdown.
"@
  },
  @{
    title = "[Backlog] Logging, Diagnostics, and Redaction"
    milestone = "M4"
    labels = @("feature", "quality", "security")
    body = @"
## Scope
Implement activity logging, diagnostic export, debug mode, and sensitive data redaction.

## Acceptance Criteria
- [ ] Activity log view available.
- [ ] Export logs bundle for support.
- [ ] Debug mode toggle available.
- [ ] Secrets and credentials are redacted from logs/exports.
"@
  },
  @{
    title = "[Backlog] Beta Hardening and NFR Gate Validation"
    milestone = "M4"
    labels = @("feature", "release", "high-priority")
    body = @"
## Scope
Finalize performance/security/reliability gates for beta readiness.

## Acceptance Criteria
- [ ] Startup < 2.5s on baseline hardware with 5k instances.
- [ ] Search < 200ms for 5k instances.
- [ ] Scanner 10 query sets < 8 minutes.
- [ ] Backup restore validated on Windows and macOS.
- [ ] Crash-free beta session rate target > 99%.
"@
  },
  @{
    title = "[Backlog] Future Hooks Scaffold (Disabled by Default)"
    milestone = "M4"
    labels = @("enhancement", "architecture")
    body = @"
## Scope
Introduce extension points for future providers and commerce hooks without enabling user-facing behavior.

## Acceptance Criteria
- [ ] AI provider abstraction includes placeholder for additional providers.
- [ ] Scanner provider abstraction supports additional sources.
- [ ] For-sale flag and structured offers remain disabled in v1.
"@
  }
)

foreach ($i in $issues) {
  Ensure-Issue -Title $i.title -Body $i.body -MilestoneNumber $milestoneMap[$i.milestone] -Labels $i.labels | Out-Null
}

Write-Host "Backlog provisioning complete."
