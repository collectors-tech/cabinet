# CABINET - USE CASES AND SCENARIOS

## Purpose
This document defines end-user behavior for Cabinet v1 and near-term planned flows.
It complements:
- `docs/SPEC.md` (product intent)
- `docs/FULL_FEATURE_LIST.md` (feature inventory)

## Actors
- Collector User: primary user managing collection and market discovery
- Secondary Profile User: another local user on same machine
- System Scheduler: local job runner for scans, pricing, backups
- External Provider: eBay listing source
- AI Provider: OpenAI API for metadata assistance
- License Authority: external issuer of signed license file

## Scope and Assumptions
- Desktop-first app (Windows and macOS)
- Local-first architecture with per-profile storage
- WebAuthn required in v1
- Cloud account not required
- Scanner provider in v1 is eBay

## End-to-End User Journeys
### Journey A: First-Time Setup to First Discovery
1. Install app and launch.
2. Create local profile.
3. Register first WebAuthn credential.
4. Import collection or create first item.
5. Configure scanner query set.
6. Run scan.
7. Review "Not in my collection" suggestions.

Success criteria:
- User reaches first meaningful discovery within one session.

### Journey B: Ongoing Weekly Collector Workflow
1. Unlock app.
2. Review dashboard changes (discoveries, wishlist hits, price drops).
3. Run scheduled or manual scan.
4. Classify scanner candidates.
5. Update collection instances and photos.
6. Review tracked prices, stock counts, and wishlist targets.

Success criteria:
- User can complete weekly review in under 20 minutes.

### Journey C: Recovery and Continuity
1. User loses primary authenticator.
2. Uses recovery flow to restore access.
3. Verifies profile integrity.
4. Restores backup if needed.

Success criteria:
- No irreversible profile lockout with valid recovery setup.

## Use Cases
## UC-01: Install and Launch Application
Actor: Collector User  
Preconditions:
- Installer is downloaded.

Main flow:
1. User runs installer.
2. App installs and registers local storage paths.
3. User launches app.

Alternate flows:
- A1: Existing version detected; app performs signed update path.
- A2: Signature validation fails; installation halts with error.

Postconditions:
- App starts to profile/login screen.

## UC-02: Create Profile and First Credential
Actor: Collector User  
Preconditions:
- No profile exists or user selects "Create profile".

Main flow:
1. User enters profile name.
2. App creates profile-specific database and settings store.
3. App starts WebAuthn registration.
4. User registers credential (Touch ID / Windows Hello / security key).

Alternate flows:
- A1: User cancels credential registration; profile creation is not finalized.
- A2: Authenticator unavailable; user is prompted to connect/enable one.

Postconditions:
- Profile exists with at least one WebAuthn credential.

## UC-03: Unlock App and Auto-Lock
Actor: Collector User  
Preconditions:
- Profile exists with credential.

Main flow:
1. User selects profile.
2. App prompts WebAuthn assertion.
3. User unlocks and accesses dashboard.
4. App auto-locks after inactivity timeout.

Alternate flows:
- A1: Assertion fails; attempt is logged and retry is offered.
- A2: Too many failed attempts; temporary lockout is applied.

Postconditions:
- Session starts securely and lock behavior is enforced.

## UC-04: Recover Access After Credential Loss
Actor: Collector User  
Preconditions:
- Recovery passphrase was configured.
- No usable authenticator is available.

Main flow:
1. User selects recovery path.
2. User enters recovery passphrase.
3. App verifies recovery secret.
4. User registers a new WebAuthn credential.

Alternate flows:
- A1: Invalid recovery attempts exceed threshold; timed cooldown.
- A2: Recovery was not configured; access cannot be restored locally.

Postconditions:
- Access restored with new credential.

## UC-05: Switch Between Local Profiles
Actor: Collector User / Secondary Profile User  
Preconditions:
- Two or more profiles exist.

Main flow:
1. User logs out or locks app.
2. Login screen shows profile list.
3. User selects different profile and authenticates.

Postconditions:
- Active data context changes to selected profile only.

## UC-06: Create Canonical Item and Instances
Actor: Collector User  
Preconditions:
- User is authenticated.

Main flow:
1. User creates canonical item with part number and metadata.
2. User adds one or more instances with condition/status and quantity.
3. App validates required fields.
4. Item appears in search and collection views.

Alternate flows:
- A1: Duplicate part number detected; user chooses create anyway or review existing.
- A2: Missing required fields; inline validation blocks save.

Postconditions:
- Canonical item and linked instances are persisted.

## UC-07: Import Existing Collection
Actor: Collector User  
Preconditions:
- User has JSON or CSV source data.

Main flow:
1. User selects import file.
2. App runs schema/version validation.
3. App presents dry-run preview with conflict summary.
4. User applies chosen conflict actions (merge/create/skip).
5. Import executes and activity is logged.

Alternate flows:
- A1: Schema mismatch; import blocked with migration guidance.
- A2: User cancels at preview stage; no data is changed.

Postconditions:
- Collection is updated according to explicit user decisions.

## UC-08: Manage Photos for Items
Actor: Collector User  
Preconditions:
- Canonical item exists.

Main flow:
1. User uploads image from desktop or mobile browser.
2. App stores original in media folder.
3. App generates thumbnail and preview.
4. User selects primary photo and views full screen.

Alternate flows:
- A1: Corrupt or unsupported image format; upload rejected with reason.
- A2: Thumbnail generation fails; original is kept and rebuild is available.

Postconditions:
- Item has usable media assets with stable references.

## UC-09: Add and Resolve Barcodes
Actor: Collector User  
Preconditions:
- Item exists or candidate listing is open.

Main flow:
1. User scans or manually enters barcode.
2. App checks local barcode index.
3. If no local match, user can query external provider.
4. User links barcode to canonical item.

Alternate flows:
- A1: Duplicate barcode across variants; app asks user to choose mapping intent.
- A2: External lookup unavailable; user continues with manual link.

Postconditions:
- Barcode mapping is stored and searchable.

## UC-10: Configure Scanner Query Sets
Actor: Collector User  
Preconditions:
- User authenticated with scanner-enabled license tier.

Main flow:
1. User creates query set with keywords and exclusions.
2. User sets max price, region, and condition filters.
3. User chooses schedule and saves.

Alternate flows:
- A1: Invalid query syntax; save is blocked with inline validation.
- A2: Provider lacks condition filter support; app downgrades gracefully.

Postconditions:
- Query set is stored and eligible for run.

## UC-11: Execute Scanner and Persist Candidates
Actor: Collector User / System Scheduler  
Preconditions:
- At least one query set exists.

Main flow:
1. User clicks Run Now or scheduled execution starts.
2. App requests provider data under rate limits.
3. Candidates are normalized and deduplicated.
4. Candidate records are persisted with first seen/last seen/status.
5. If available, stock status/count are captured from provider data/page.

Alternate flows:
- A1: Provider throttles requests; app applies retry/backoff.
- A2: Provider failure; run ends with partial results and clear health status.

Postconditions:
- New candidate set available for classification.
- Candidate availability state is available for buy-timing decisions.

## UC-12: Match Candidates to Collection
Actor: System  
Preconditions:
- Candidate records exist.

Main flow:
1. App extracts part number from candidate metadata.
2. App evaluates match confidence against canonical records.
3. App classifies each result as matched, suggested, or not in collection.

Alternate flows:
- A1: Low-confidence candidate remains suggested and requires user review.
- A2: Missing part number uses fallback matching signals (title/tags/brand).

Postconditions:
- Candidate classification is ready for user action.

## UC-13: Act from Not-In-Collection Panel
Actor: Collector User  
Preconditions:
- Candidate classification exists.

Main flow:
1. User filters by query, price, and date.
2. User chooses action: ignore, add to wishlist, track price, or create item.
3. App persists action and updates dashboard signals.

Alternate flows:
- A1: User undoes ignore via reset ignore rules.
- A2: Create item flow requires explicit confirmation before save.

Postconditions:
- Candidate state transitions are recorded and auditable.

## UC-14: Manage Wishlist and Target Pricing
Actor: Collector User  
Preconditions:
- Canonical item exists or candidate converted.

Main flow:
1. User adds item to wishlist.
2. User sets priority and target price.
3. Scanner hits linked to wishlist item are highlighted.
4. Below-target state appears when conditions are met.

Postconditions:
- Wishlist is actionable and connected to scanner output.

## UC-15: Track Prices and Review Trends
Actor: Collector User / System Scheduler  
Preconditions:
- Item marked as tracked.

Main flow:
1. Daily snapshot job records min/median/latest and stock status/count by source.
2. User opens price history graph.
3. User exports price + stock history when needed.

Alternate flows:
- A1: Snapshot missed due to provider outage; app records skipped interval event.
- A2: Sparse data is shown with explicit low-confidence visualization.

Postconditions:
- Historical pricing + stock data supports collector decision-making.

## UC-16: Search, Filter, and Sort Collection
Actor: Collector User  
Preconditions:
- Collection contains data.

Main flow:
1. User enters full-text query.
2. User applies filters (brand, condition, status, tags, scale).
3. User saves filter for later reuse.
4. User sorts by date, price, or part number.

Postconditions:
- User can quickly locate and segment collection records.

## UC-17: Dashboard Weekly Review
Actor: Collector User  
Preconditions:
- Scanner and pricing data exist.

Main flow:
1. User opens dashboard.
2. User reviews discoveries, wishlist hits, price drops, low-stock alerts, and restock alerts.
3. User checks recently added items and collection totals.
4. User navigates into detailed views from dashboard cards.

Postconditions:
- Dashboard functions as control center for weekly workflow.

## UC-18: Configure AI Assist and Apply Suggestions
Actor: Collector User  
Preconditions:
- Valid OpenAI API key provided.

Main flow:
1. User enables AI assist.
2. User requests metadata extraction from photo or title.
3. App returns suggested fields with confidence.
4. User accepts, edits, or rejects suggestions.

Alternate flows:
- A1: API key invalid; connection test fails with remediation guidance.
- A2: AI service unavailable; user can continue manual workflow.

Postconditions:
- AI only assists; user retains final control over persisted changes.

## UC-19: Manage License State and Feature Gating
Actor: Collector User  
Preconditions:
- App is running with free or pro entitlement.

Main flow:
1. User views current license status.
2. User imports signed license file.
3. App verifies signature and updates entitlement.
4. Pro features unlock immediately after valid activation.

Alternate flows:
- A1: Invalid signature; license is rejected with reason.
- A2: Expired/unsupported license format; status remains unchanged.

Postconditions:
- Feature gates reflect valid local entitlement state.

## UC-20: Export Logs and Diagnostics for Support
Actor: Collector User  
Preconditions:
- App has activity/error logs.

Main flow:
1. User opens diagnostics.
2. User exports logs bundle.
3. App redacts sensitive values before export.

Alternate flows:
- A1: Debug mode is enabled to increase detail before export.
- A2: Export path unavailable; user chooses alternate location.

Postconditions:
- Support-ready diagnostic package is produced without secret leakage.

## UC-21: Restore Backup After Failure
Actor: Collector User  
Preconditions:
- Backup files exist.

Main flow:
1. App detects abnormal shutdown or user initiates restore.
2. User selects backup snapshot.
3. App validates snapshot integrity and version compatibility.
4. User confirms restore.
5. App restores profile data and restarts.

Alternate flows:
- A1: Corrupt snapshot; restore blocked with details.
- A2: Version incompatibility; migration path is offered.

Postconditions:
- User can return to a known good local state.

## UC-22: Upgrade Application Safely
Actor: Collector User  
Preconditions:
- New signed version available.

Main flow:
1. App checks selected update channel.
2. App downloads update metadata and verifies signature.
3. User confirms install (or auto-install policy applies).
4. App updates binary and migrates schema if needed.
5. App reopens profile and validates health.

Alternate flows:
- A1: Signature mismatch; update is blocked.
- A2: Migration failure; app rolls back and prompts diagnostics export.

Postconditions:
- Upgrade completes without data loss.

## Scenario Matrix (Critical Edge Scenarios)
1. WebAuthn unavailable on fresh install.
Expected outcome:
- User cannot bypass credential requirement; actionable setup guidance shown.

2. Primary credential lost and recovery not configured.
Expected outcome:
- Access denied; profile remains encrypted/locked; prevention guidance documented.

3. Import file contains 10k rows with mixed duplicates.
Expected outcome:
- Dry-run conflict summary appears; no destructive automatic merge.

4. eBay API throttles mid-run.
Expected outcome:
- Backoff + retry; partial results preserved; health indicator warns user.

5. AI returns low-confidence part number.
Expected outcome:
- Suggestion shown but never auto-applied.

6. License expires while using Pro-only features.
Expected outcome:
- Existing data preserved; gated actions blocked with clear upgrade/relicense path.

7. Crash occurs during active scan + import operations.
Expected outcome:
- Recovery prompt appears; user can restore from latest valid backup.

## v1 Acceptance Summary
For v1 completion, all UC-01 through UC-22 must satisfy:
- Main flow executable end to end.
- At least one alternate/failure path handled gracefully.
- User-visible errors are actionable.
- Behavior is covered by test plan and release checklist.
