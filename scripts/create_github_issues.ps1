param(
  [string]$Owner = "collectors-tech",
  [string]$Repo = "cabinet"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not $env:GITHUB_TOKEN) {
  throw "GITHUB_TOKEN is required."
}

$headers = @{
  Authorization = "Bearer $($env:GITHUB_TOKEN)"
  Accept = "application/vnd.github+json"
  "X-GitHub-Api-Version" = "2022-11-28"
  "User-Agent" = "cabinet-issues-bootstrap"
}

function ApiGet([string]$Url) {
  $resp = Invoke-WebRequest -Method Get -Headers $headers -Uri $Url -ErrorAction Stop
  $data = $resp.Content | ConvertFrom-Json
  return @($data)
}

function ApiPost([string]$Url, [object]$BodyObj) {
  $body = $BodyObj | ConvertTo-Json -Depth 10
  Invoke-RestMethod -Method Post -Headers $headers -Uri $Url -Body $body -ErrorAction Stop
}

function Ensure-Label([string]$Name, [string]$Color, [string]$Description) {
  $labels = @(ApiGet "https://api.github.com/repos/$Owner/$Repo/labels?per_page=100")
  $exists = @($labels | Where-Object { $_.name -eq $Name }).Count -gt 0
  if ($exists) { return }
  try {
    ApiPost "https://api.github.com/repos/$Owner/$Repo/labels" @{
      name = $Name
      color = $Color
      description = $Description
    } | Out-Null
  } catch {
    if ($_.Exception.Message -notmatch "already_exists") { throw }
  }
}

function Ensure-Issue([string]$Title, [string]$Body, [int]$Milestone, [string[]]$Labels) {
  $open = @(ApiGet "https://api.github.com/repos/$Owner/$Repo/issues?state=open&per_page=100")
  $existing = $open | Where-Object {
    $_.PSObject.Properties.Name -contains "title" -and
    $_.title -eq $Title -and
    -not ($_.PSObject.Properties.Name -contains "pull_request")
  } | Select-Object -First 1
  if ($existing) {
    Write-Host "Issue exists: #$($existing.number) $Title"
    return
  }
  $created = ApiPost "https://api.github.com/repos/$Owner/$Repo/issues" @{
    title = $Title
    body = $Body
    milestone = $Milestone
    labels = $Labels
  }
  Write-Host "Created issue: #$($created.number) $Title"
}

$milestones = @(ApiGet "https://api.github.com/repos/$Owner/$Repo/milestones?state=all&per_page=100")
function Resolve-MilestoneNumber([object[]]$Milestones, [string]$Title) {
  $matches = @($Milestones | Where-Object { $_.title -eq $Title })
  if ($matches.Count -eq 0) { return $null }
  return [int]$matches[0].number
}

$m1 = Resolve-MilestoneNumber -Milestones $milestones -Title "M1 - Foundation and Core Experience"
$m2 = Resolve-MilestoneNumber -Milestones $milestones -Title "M2 - Discovery and Matching"
$m3 = Resolve-MilestoneNumber -Milestones $milestones -Title "M3 - Intelligence and Commercial"
$m4 = Resolve-MilestoneNumber -Milestones $milestones -Title "M4 - Hardening and Beta Launch"

if (-not $m1 -or -not $m2 -or -not $m3 -or -not $m4) {
  throw "Missing one or more milestones. Create milestones first."
}

$labelDefs = @(
  @{ n="feature"; c="0e8a16"; d="Feature work" },
  @{ n="enhancement"; c="a2eeef"; d="Enhancement work" },
  @{ n="core"; c="0052cc"; d="Core application" },
  @{ n="security"; c="d93f0b"; d="Security/authentication/secrets" },
  @{ n="media"; c="5319e7"; d="Photo and media pipeline" },
  @{ n="data"; c="fbca04"; d="Import/export/backup/data flows" },
  @{ n="scanner"; c="1d76db"; d="Discovery scanner and matching" },
  @{ n="provider"; c="bfdadc"; d="External provider integration" },
  @{ n="ux"; c="c2e0c6"; d="User experience and dashboard" },
  @{ n="wishlist"; c="fef2c0"; d="Wishlist and target-price features" },
  @{ n="pricing"; c="ff9f1c"; d="Price tracking and trends" },
  @{ n="ai"; c="6f42c1"; d="AI assist features" },
  @{ n="licensing"; c="d4c5f9"; d="License and entitlement features" },
  @{ n="quality"; c="ededed"; d="Quality, reliability, and diagnostics" },
  @{ n="release"; c="0b3b8c"; d="Release and beta readiness" },
  @{ n="high-priority"; c="b60205"; d="High-priority backlog item" },
  @{ n="architecture"; c="0366d6"; d="Architecture work" }
)
foreach ($l in $labelDefs) { Ensure-Label $l.n $l.c $l.d }

$issues = @(
  @{ t="[Backlog] Application Core and Runtime"; m=$m1; l=@("feature","core","high-priority"); b="## Scope`nBuild application core for desktop (Windows + macOS), embedded web UI, local SQLite, media folder, signed updates, and stable/beta channels.`n`n## Acceptance Criteria`n- [ ] Single installer path works on Windows and macOS.`n- [ ] Embedded web UI loads from app shell.`n- [ ] Local SQLite initializes with migrations.`n- [ ] Signed update verification exists.`n- [ ] Stable/beta channel selection is supported." },
  @{ t="[Backlog] WebAuthn Login, Locking, and Recovery"; m=$m1; l=@("feature","security","high-priority"); b="## Scope`nImplement mandatory WebAuthn login in v1, app lock behavior, inactivity auto-lock, and credential recovery/reset flows.`n`n## Acceptance Criteria`n- [ ] First launch requires creating WebAuthn credential.`n- [ ] Unlock uses WebAuthn assertion.`n- [ ] Inactivity auto-lock is enforced.`n- [ ] Recovery passphrase fallback path is available when configured.`n- [ ] Credential reset flow exists for no-usable-authenticator scenario." },
  @{ t="[Backlog] Multi-Profile Local User Support"; m=$m1; l=@("feature","core"); b="## Scope`nAdd multi-profile support with isolated DB, settings, API keys, and license per profile.`n`n## Acceptance Criteria`n- [ ] Profile creation and switching from login screen.`n- [ ] Per-profile storage isolation.`n- [ ] Per-profile key/license separation." },
  @{ t="[Backlog] Collection Core (Canonical Items + Instances)"; m=$m1; l=@("feature","core","high-priority"); b="## Scope`nImplement canonical item and instance data model and CRUD flows.`n`n## Acceptance Criteria`n- [ ] Canonical fields from spec are implemented.`n- [ ] Instance fields and status values are enforced.`n- [ ] One canonical item can have many instances.`n- [ ] No auto-merge without explicit confirmation." },
  @{ t="[Backlog] Photo System (Upload, Derivatives, Primary, Fullscreen)"; m=$m1; l=@("feature","media"); b="## Scope`nSupport desktop/mobile upload, local originals, thumbnail/preview generation, primary image management, and rebuild flow.`n`n## Acceptance Criteria`n- [ ] Upload from desktop and mobile browser.`n- [ ] Original stored locally with thumbnail+preview derivatives.`n- [ ] Primary image can be set and changed.`n- [ ] Thumbnail rebuild job available." },
  @{ t="[Backlog] Barcode System and Variant Handling"; m=$m1; l=@("feature","core"); b="## Scope`nImplement barcode detection/entry, multiple barcodes per item, local lookup, external lookup, and duplicate variant handling.`n`n## Acceptance Criteria`n- [ ] Manual barcode entry and storage.`n- [ ] Multiple barcodes per canonical item.`n- [ ] Local barcode match lookup.`n- [ ] External provider lookup path.`n- [ ] Duplicate barcode handling prompts explicit mapping choice." },
  @{ t="[Backlog] Search, Filters, and Saved Views"; m=$m1; l=@("feature","core"); b="## Scope`nBuild FTS search with filters, saved filters, and sort options.`n`n## Acceptance Criteria`n- [ ] Full-text search across item/instance data.`n- [ ] Filters for brand/condition/status/tags/scale.`n- [ ] Saved filter CRUD.`n- [ ] Sort by date added/price/part number." },
  @{ t="[Backlog] Data Management (Import/Export, Reindex, Repair)"; m=$m1; l=@("feature","data","high-priority"); b="## Scope`nImplement JSON/CSV import-export with dry-run and conflict resolution, plus reindex/repair tools.`n`n## Acceptance Criteria`n- [ ] JSON full backup export + CSV item export.`n- [ ] JSON snapshot import + CSV mapping import.`n- [ ] Dry-run preview before apply.`n- [ ] Conflict resolution actions (merge/create/skip).`n- [ ] Schema version checks.`n- [ ] Reindex and repair operations available." },
  @{ t="[Backlog] Local Backup and Restore Reliability"; m=$m1; l=@("feature","data","security"); b="## Scope`nImplement automatic backup frequency controls and restore workflow with validation.`n`n## Acceptance Criteria`n- [ ] Automatic local backups by schedule.`n- [ ] Restore flow validates snapshot integrity.`n- [ ] Restore tested on Windows and macOS." },
  @{ t="[Backlog] Scanner Engine (Query Sets, Scheduling, Rate Limits)"; m=$m2; l=@("feature","scanner","high-priority"); b="## Scope`nBuild scanner query model, manual/scheduled execution, rate limiting, retry/backoff, and candidate persistence.`n`n## Acceptance Criteria`n- [ ] Query sets support keywords/exclusions/max-price/region/condition.`n- [ ] Manual Run Now and scheduled runs.`n- [ ] Rate limiting and retry/backoff.`n- [ ] Candidate lifecycle stores first seen/last seen/status." },
  @{ t="[Backlog] eBay Provider v1 Integration"; m=$m2; l=@("feature","scanner","provider"); b="## Scope`nIntegrate eBay search provider for scanner and external barcode lookup.`n`n## Acceptance Criteria`n- [ ] Provider credentials/config in settings.`n- [ ] Query execution returns normalized candidates.`n- [ ] Provider health indicator and failure logging." },
  @{ t="[Backlog] Matching Engine and Confidence Rules"; m=$m2; l=@("feature","scanner","high-priority"); b="## Scope`nMatch candidates to canonical items using part number extraction and confidence scoring.`n`n## Acceptance Criteria`n- [ ] Candidate classification states: matched/suggested/not-in-collection.`n- [ ] Confidence computation is persisted and visible.`n- [ ] Low-confidence candidates require user review." },
  @{ t="[Backlog] Not In My Collection Panel and Actions"; m=$m2; l=@("feature","scanner","ux"); b="## Scope`nDeliver panel for candidate triage with filters and actions.`n`n## Acceptance Criteria`n- [ ] Filters by price/query/date.`n- [ ] Actions: ignore, add to wishlist, track price, create item.`n- [ ] Ignore rules can be reset from settings." },
  @{ t="[Backlog] Wishlist and Target Price Signals"; m=$m3; l=@("feature","wishlist"); b="## Scope`nImplement wishlist linked to canonical items with target price and hit highlighting.`n`n## Acceptance Criteria`n- [ ] Wishlist CRUD with priority and notes.`n- [ ] Scanner hits linked and highlighted.`n- [ ] Below-target indicator shown when applicable." },
  @{ t="[Backlog] Price Tracking and History Export"; m=$m3; l=@("feature","pricing","high-priority"); b="## Scope`nImplement tracked-item daily snapshots and trend views.`n`n## Acceptance Criteria`n- [ ] Daily snapshots capture min/median/latest.`n- [ ] Graph view for item history.`n- [ ] Per-source breakdown supported.`n- [ ] Price history export available." },
  @{ t="[Backlog] Dashboard (Discoveries, Hits, Drops, Stats)"; m=$m3; l=@("feature","ux"); b="## Scope`nBuild dashboard summarizing weekly collector workflow signals.`n`n## Acceptance Criteria`n- [ ] Shows new discoveries, wishlist hits, price drops, recently added items.`n- [ ] Shows total items/instances (+ optional estimated value).`n- [ ] Cards deep-link to relevant detailed views." },
  @{ t="[Backlog] AI Assist (OpenAI) with Confirmation Guardrails"; m=$m3; l=@("feature","ai","security"); b="## Scope`nAdd OpenAI-based assist for photo/title metadata suggestions with confidence and explicit confirmation.`n`n## Acceptance Criteria`n- [ ] API key configuration and connectivity test.`n- [ ] Photo identify, part number extraction, normalized field suggestions.`n- [ ] Confidence score shown for every suggestion.`n- [ ] No auto-create/update without user confirmation.`n- [ ] AI can be disabled per profile." },
  @{ t="[Backlog] Licensing System and Feature Gating"; m=$m3; l=@("feature","licensing","high-priority"); b="## Scope`nImplement free-tier limits, Pro unlock via signed license file, and offline validation.`n`n## Acceptance Criteria`n- [ ] License import and signature verification.`n- [ ] Offline entitlement validation.`n- [ ] Free-tier item cap enforced.`n- [ ] Pro gates for scanner automation, price tracking, AI assist.`n- [ ] Clear license status UI for valid/invalid/expired states." },
  @{ t="[Backlog] Settings Surface and Secret Storage"; m=$m3; l=@("feature","security","core"); b="## Scope`nComplete settings for scanner schedule, update channel, credentials, backups, DB location, and maintenance actions.`n`n## Acceptance Criteria`n- [ ] Per-profile settings persist.`n- [ ] API keys use OS-backed secure storage.`n- [ ] Rebuild thumbnails and reset ignore rules actions exist." },
  @{ t="[Backlog] Error Handling and Crash Recovery UX"; m=$m4; l=@("feature","quality"); b="## Scope`nStandardize failure handling for scanner, AI, and provider failures, including crash recovery prompt.`n`n## Acceptance Criteria`n- [ ] Scanner and AI failures logged with actionable messages.`n- [ ] Provider health indicator surfaces degraded state.`n- [ ] Retry flow for failed scans.`n- [ ] Crash recovery prompt appears after abnormal shutdown." },
  @{ t="[Backlog] Logging, Diagnostics, and Redaction"; m=$m4; l=@("feature","quality","security"); b="## Scope`nImplement activity logging, diagnostic export, debug mode, and sensitive data redaction.`n`n## Acceptance Criteria`n- [ ] Activity log view available.`n- [ ] Export logs bundle for support.`n- [ ] Debug mode toggle available.`n- [ ] Secrets and credentials are redacted from logs/exports." },
  @{ t="[Backlog] Beta Hardening and NFR Gate Validation"; m=$m4; l=@("feature","release","high-priority"); b="## Scope`nFinalize performance/security/reliability gates for beta readiness.`n`n## Acceptance Criteria`n- [ ] Startup < 2.5s on baseline hardware with 5k instances.`n- [ ] Search < 200ms for 5k instances.`n- [ ] Scanner 10 query sets < 8 minutes.`n- [ ] Backup restore validated on Windows and macOS.`n- [ ] Crash-free beta session rate target > 99%." },
  @{ t="[Backlog] Future Hooks Scaffold (Disabled by Default)"; m=$m4; l=@("enhancement","architecture"); b="## Scope`nIntroduce extension points for future providers and commerce hooks without enabling user-facing behavior.`n`n## Acceptance Criteria`n- [ ] AI provider abstraction includes placeholder for additional providers.`n- [ ] Scanner provider abstraction supports additional sources.`n- [ ] For-sale flag and structured offers remain disabled in v1." }
)

foreach ($i in $issues) {
  Ensure-Issue -Title $i.t -Body $i.b -Milestone $i.m -Labels $i.l
}

Write-Host "Issue bootstrap complete."
