# Cabinet - 90 Day Build Roadmap (13 Weeks)

## Scope Baseline
Feature scope source of truth:
- `docs/FULL_FEATURE_LIST.md`

Product intent source of truth:
- `docs/SPEC.md`

## Delivery Principles
- Ship vertical slices weekly, not isolated subsystems
- Keep local-first and offline behavior as default
- WebAuthn is required in v1 login flows
- No destructive auto-merge of collection records
- Every week ends with a demo build + regression checklist

## Definition of Done (Global)
- Feature implemented behind stable UI flow
- Unit tests added for business logic changes
- Integration test added for DB + API path
- Errors are user-visible and logged
- Feature is documented in `docs/FULL_FEATURE_LIST.md` if scope changed

## Quality Gates Before Beta
- Startup time under 2.5s on baseline hardware with 5k instances
- Scanner run of 10 query sets completes under 8 minutes without errors
- No data loss during crash/restart in active import/scan flow
- Backup restore tested on Windows and macOS
- Installer signature and update signature verified
- Security review completed for WebAuthn, API key storage, license validation

## Workstreams
1. Platform and Runtime: shell, installer, updater, profile bootstrapping
2. Identity and Security: WebAuthn, local credential recovery, secret storage
3. Collection Core: canonical items, instances, photos, barcodes, search
4. Discovery Intelligence: scanner, matching, not-in-collection, wishlist
5. Pricing and Insights: snapshots, trending, dashboard signals
6. Commercial Readiness: licensing, diagnostics, beta operations

## Week-by-Week Plan
### Week 1 - Foundation and Project Skeleton
Objectives:
- Finalize UI framework decision (React or Svelte) and lock toolchain
- Establish Go app shell, embedded web UI, SQLite bootstrap, migrations
- Set CI for lint, test, build on Windows and macOS

Exit criteria:
- App boots with profile selector placeholder
- Migration system can create fresh DB and upgrade schema
- CI produces signed nightly artifacts

### Week 2 - Profile Model and WebAuthn Base
Objectives:
- Implement multi-profile layout (separate DB, settings, keys, license)
- Implement first-run local user creation with required WebAuthn credential
- Implement lock screen and inactivity auto-lock

Exit criteria:
- New profile can be created, locked, unlocked with WebAuthn
- At least one secondary credential can be added
- Failed auth events are logged with non-sensitive details

### Week 3 - Recovery, Secret Storage, and Settings Spine
Objectives:
- Implement recovery passphrase fallback flow (explicit opt-in)
- Add secure storage for API keys and sensitive profile material
- Build settings scaffolding (theme, DB location, backup frequency)

Exit criteria:
- Credential loss can be recovered through documented fallback
- Secrets are never stored in plaintext in SQLite
- Settings persist per profile and survive restart

### Week 4 - Collection Core v1
Objectives:
- Implement canonical item CRUD with required fields
- Implement instances CRUD with status/condition model
- Implement tags and basic validation rules

Exit criteria:
- One item to many instances behavior verified
- No auto-merge occurs without explicit confirmation
- Core collection screens usable end to end

### Week 5 - Photos and Media Pipeline
Objectives:
- Desktop and mobile-browser upload support
- Store originals locally and generate thumbnail/preview derivatives
- Primary photo selection, delete, full-screen preview

Exit criteria:
- Photo ingestion succeeds for typical formats and large files
- Thumbnail rebuild command restores missing derivatives
- Media references remain valid after app restart

### Week 6 - Search, Filters, and Barcode Foundation
Objectives:
- Implement SQLite FTS-based search
- Add filter/sort and saved filter support
- Add manual barcode entry and local barcode matching

Exit criteria:
- Search returns relevant item + instance results in under 200ms for 5k instances
- Saved filters can be created, edited, and removed
- Multiple barcodes per canonical item are supported

### Week 7 - Data Management and Backups
Objectives:
- Implement JSON full export/import and CSV export/import mappings
- Add import conflict preview and conflict resolution actions
- Implement automated local backups and restore flow

Exit criteria:
- Import can run dry-run with conflict summary
- Backup restore reproduces a known dataset without data loss
- Reindex and repair actions complete with diagnostics output

### Week 8 - Scanner Framework and Provider Integration
Objectives:
- Implement query-set model, manual run, and scheduled execution
- Add rate limiting, retry policy, provider health indicator
- Integrate eBay provider v1 and persist candidate records

Exit criteria:
- Query sets run manually and on schedule
- Candidate record lifecycle tracks first seen/last seen/status
- Provider errors are retriable and clearly surfaced

### Week 9 - Matching Engine and Not-In-Collection
Objectives:
- Implement part number extraction + normalized candidate model
- Implement matching confidence rules and output states
- Build not-in-collection panel with ignore/wishlist/track/create actions

Exit criteria:
- Candidate can be classified as matched/suggested/not-in-collection
- Ignore rules persist and can be reset
- False positive review workflow exists before item creation

### Week 10 - Wishlist, Price Tracking, and Dashboard
Objectives:
- Implement wishlist model with priority and target price
- Implement tracked item daily snapshot job (min/median/latest)
- Build dashboard signals (discoveries, wishlist hits, price drops, stats)

Exit criteria:
- Below-target signals are generated accurately
- Price history graph and export work for tracked items
- Dashboard updates after scanner and pricing jobs

### Week 11 - AI Assist and Safety Controls
Objectives:
- Add OpenAI key configuration and connection checks
- Implement photo identification + title normalization + metadata suggestions
- Enforce confirmation before any create/update action from AI output

Exit criteria:
- AI responses include confidence and are reviewable before apply
- AI failures are logged and shown with actionable messages
- AI can be fully disabled per profile

### Week 12 - Licensing, Packaging, and Update Readiness
Objectives:
- Implement license import, signature verification, offline validation
- Implement free-tier enforcement and Pro feature gating
- Finalize installer behavior and update strategy for beta channel

Exit criteria:
- Expired/invalid/missing license states are handled cleanly
- Feature gates enforce limits without data corruption
- Install and upgrade tested on clean Windows and macOS machines

### Week 13 - Hardening and Beta Launch
Objectives:
- Full regression pass, performance tuning, bug burn-down
- Finalize diagnostics export, activity logs, and support playbook
- Launch beta to 20-50 collectors with feedback loop

Exit criteria:
- Quality gates met
- Known issues documented with mitigations
- Beta telemetry process and triage cadence active

## Dependency Map
- WebAuthn and profile model must land before broader settings and key management
- Collection schema stability is required before scanner matching confidence work
- Query sets and candidate persistence must exist before not-in-collection panel
- Price tracking depends on stable candidate ingestion and canonical matching
- Licensing gates should land after core workflows to avoid blocking development velocity

## Risk Register and Mitigations
1. WebAuthn cross-platform edge cases
- Mitigation: run platform matrix tests weekly and keep recovery passphrase path
2. Provider instability or throttling (eBay)
- Mitigation: robust backoff, caching, health indicator, manual rerun controls
3. Import data corruption risk
- Mitigation: dry-run conflict preview, immutable backups before apply
4. Matching false positives
- Mitigation: confidence thresholds + explicit user confirmation path
5. Beta support burden
- Mitigation: diagnostics bundle export, issue templates, weekly bug triage ritual

## Beta Success Criteria
- 20-50 active beta users with weekly usage
- At least 70 percent of users complete first scan within 48 hours
- At least 50 percent of users import or create 100+ instances
- Crash-free session rate above 99 percent during beta window
