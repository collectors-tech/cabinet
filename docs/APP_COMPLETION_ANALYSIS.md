# Cabinet App Completion Analysis (Execution Blueprint)

Last updated: 2026-02-26  
Owner: Product + Engineering  
Backlog source of truth: GitHub issues in `collectors-tech/cabinet`

## 1. Purpose
This document defines the full execution model to complete Cabinet to release-ready quality, not just feature breadth.  
It is intentionally detailed so each item can be implemented issue-by-issue without ambiguity.

## 2. Source Inputs
- Product scope: `docs/FULL_FEATURE_LIST.md`
- Product intent: `docs/SPEC.md`
- Use cases: `docs/USE_CASES_AND_SCENARIOS.md`
- UI model: `docs/UI_INTUITIVE_PLANNING.md`
- API/screen parity: `docs/UI_ENDPOINT_PARITY.md`
- Architecture baseline: `docs/ARCHITECTURE.md`

## 3. Release Definition of Done (Global)
Cabinet is done only when all criteria below are true.

1. Every feature section 1-23 in `docs/FULL_FEATURE_LIST.md` is either:
   - implemented and validated, or
   - explicitly deferred with issue, rationale, and signed decision.
2. Every screen has:
   - valid data source,
   - loading/empty/error states,
   - successful mutation path,
   - automated regression.
3. Critical end-to-end journeys pass on Windows and macOS.
4. NFR thresholds and security checks are measured with evidence.
5. No P0/P1 issues remain open.
6. Installer, updates, rollback, and diagnostics support workflow are production-ready.

## 4. Current State (Reality Check)

### 4.1 What is strong
- Broad API surface exists (profiles, auth, items, scanner, pricing, dashboard, licensing, chat, data mgmt, backup).
- Go API tests already cover major modules in `internal/app/*_api_test.go`.
- OpenAPI and runtime docs endpoints exist (`/apidocs`, `/api/openapi.yaml`).
- Modern UI codebase exists in `ui.web` with route scaffolding and reusable components.

### 4.2 What is not yet release-ready
- Several user-visible screens still have partial or sample/template behavior.
- Inventory flow still has known reliability regressions (500 conditions observed by user).
- Cypress baseline stability on Windows is not fully hardened.
- E2E coverage does not yet fully protect critical user workflows.
- NFR and security proof points are not yet complete as hard evidence.

### 4.3 Active known issues (at time of update)
- `#151` Cypress startup failure on Windows
- `#149`, `#147` Inventory non-500 regressions
- `#154` Inventory layout and density cleanup
- `#152` Remove placeholder/template copy
- `#145` Chat API wiring audit
- `#144` Settings API wiring audit
- `#143` Users API wiring audit

## 5. Feature Completion Matrix (1-23)

Legend:
- `Done`: complete and validated
- `Partial`: implementation exists but not done
- `Missing`: implementation not yet present to required scope

| Feature | Status | Primary Gaps | Exit Condition |
| --- | --- | --- | --- |
| 1. Application Core | Partial | Packaging/update hardening | Signed install/update validated on Win+macOS |
| 2. Login/Auth | Partial | Full WebAuthn-first UX completeness | Onboarding+lock+recovery E2E green |
| 3. Profiles | Partial | Full profile switch + isolation UX | Profile data and secrets isolated in E2E |
| 4. Collection | Partial | Field completeness + polished CRUD | Full item/instance lifecycle with regressions covered |
| 5. Photos | Partial | Lightbox/ordering/full upload paths | Upload->view->reorder->delete flows green |
| 6. Barcodes | Partial | Duplicate/variant UX behavior | Local/external lookup fully actionable |
| 7. AI Assist | Partial | Consistent guarded apply UX | Preview+confirm enforced on all mutating actions |
| 8. Scanner | Partial | Scheduler/retry/health UX | Query sets to candidates end-to-end stable |
| 9. Matching | Partial | Explainability and manual overrides | Confidence and state transitions verified |
| 10. Not In Collection | Partial | Triage/action speed and reliability | Ignore/wishlist/track/create all covered |
| 11. Wishlist | Partial | Full table/card and mutation flows | Wishlist full lifecycle E2E green |
| 12. Price Tracking | Partial | Snapshot/trend UX + export reliability | Daily snapshots and trend export validated |
| 13. Dashboard | Partial | Actionable attention center quality | Useful, data-driven, no placeholder widgets |
| 14. Search/Filter | Partial | Saved views and cross-screen consistency | Search/filter state works reliably across views |
| 15. Data Management | Partial | Import conflict UX and repair operations | Dry-run/apply/reindex/repair fully validated |
| 16. Licensing | Partial | Entitlement clarity and failure handling | Free/pro gating deterministic and test-covered |
| 17. Settings | Partial | Route-by-route API parity | All setting pages persist via API |
| 18. Error Handling | Partial | Runtime recovery and user-facing clarity | Known failures handled without dead ends |
| 19. Logging/Diagnostics | Partial | Dedicated diagnostics UX quality | Logs/debug/export/recovery all accessible |
| 20. Future Hooks | Partial | Explicitly disabled but visible contracts | Hooks present and gated without side-effects |
| 21. NFRs | Missing | Measured performance evidence | NFR benchmark suite complete with reports |
| 22. Security/Privacy | Partial | Secret, redaction, lease/offline verification | Security checklist complete and audited |
| 23. Chat Copilot | Partial | Cross-screen utility + attachments/actions | Local chat + tool actions stable and safe |

## 6. Detailed Completion Requirements by Domain

## 6.1 Platform and Runtime
Required capabilities:
- deterministic boot
- robust embedded static UI serving
- `.env`-driven address/port
- local LAN mode policy
- signed update verification

Acceptance criteria:
- App starts with configured address in `.env` and no bootstrap errors.
- API docs route resolves and OpenAPI spec is served.
- Installer and update signatures verify in release flow.

Tests:
- API smoke (`/healthz`, `/api/runtime`, `/api/runtime/recovery`, `/apidocs`).
- Startup integration test with env overrides.
- Packaging smoke tests for Win/macOS artifacts.

## 6.2 Identity, Profiles, and Access
Required capabilities:
- WebAuthn-first local auth
- optional cloud ownership bootstrap where enabled
- auto-lock and unlock
- recovery passphrase fallback
- profile-scoped databases and secrets

Acceptance criteria:
- First run cannot proceed to workspace without credential registration.
- Locked session blocks protected APIs.
- Profile A data never appears in Profile B routes.

Tests:
- API auth lifecycle tests (register/login/lock/validate/recovery).
- E2E onboarding and lock/unlock flows.
- Profile isolation regression across items/wishlist/scanner/pricing/chat.

## 6.3 Inventory and Wishlist Core
Required capabilities:
- row and card view parity
- details drawer on row click
- thumbnail lightbox with carousel across filtered/sorted data source
- bulk mode via checkbox only
- status lifecycle with soft-delete and recycle

Acceptance criteria:
- No inventory/wishlist route 500 for empty, seeded, or large datasets.
- Row click opens details except when interaction target is control element.
- Deleted records hidden from default lists but recoverable.

Tests:
- E2E for create/edit/status/delete/restore on inventory and wishlist.
- E2E for details drawer and lightbox interactions.
- API tests for status transitions and linked-record delete protection.

## 6.4 Discovery, Matching, and Pricing
Required capabilities:
- query-set CRUD, scheduled scans, retry failed scans
- provider health and structured failure logs
- candidate classification matched/suggested/not-owned
- stock count extraction support and confidence notes
- daily price snapshots and trends

Acceptance criteria:
- User can run scan, inspect candidates, act, and see dashboard effects.
- Price history exports and by-source trends remain stable for large sets.
- Failures are visible and recoverable.

Tests:
- API tests for scanner run/schedule/failure/retry/provider health.
- E2E for scan -> candidate -> action pipeline.
- Integration test for pricing snapshot and trend endpoints.

## 6.5 Dashboard and Reports
Required capabilities:
- "What needs attention now" panel
- watchlist hits, price moves, discovery and recovery signals
- compact information density with clear action buttons

Acceptance criteria:
- Dashboard shows live data, not sample placeholders.
- Actions from dashboard deep-link correctly to destination context.
- Refresh and quick actions are deterministic and responsive.

Tests:
- E2E for dashboard cards and action buttons.
- API contract checks for `/api/dashboard` payload shape.
- Visual regression snapshot for primary dashboard layout states.

## 6.6 Settings, Users, Integrations, Chat
Required capabilities:
- all settings pages API-backed
- user list and add/invite endpoints wired to data layer
- integration details/edit actions reliable
- chat panel open/close and context-aware actions

Acceptance criteria:
- Settings changes persist after reload.
- Users and integrations pages never fall back to static sample content.
- Chat action preview/apply requires confirmation and logs results.

Tests:
- E2E per settings section (account/appearance/display/notifications).
- E2E for users add/edit role flow.
- API and E2E chat message/thread/attachment/action flows.

## 6.7 Diagnostics, Backup, Data Management
Required capabilities:
- dedicated diagnostics screen
- activity logs and export
- backup run/list/restore
- import dry-run with conflict summary and apply path
- reindex and repair operations with clear outcomes

Acceptance criteria:
- Diagnostics flows are accessible outside dashboard.
- Backup restore reproduces known dataset.
- Data import never mutates records during dry-run.

Tests:
- API tests for backup/data operations.
- E2E for diagnostics and backup/restore flows.
- CSV/JSON import scenario tests with conflict permutations.

## 7. Screen-by-Screen Semantic Structure Target

All major authenticated screens should follow this structure:

1. `AppShell`
2. `Sidebar` (collapsible, reorderable nav where required)
3. `TopHeader` (sticky in content column)
4. `PageHeader` (title, subtitle, primary actions)
5. `ContextStrip` (active DB/profile/context)
6. `FilterBar` (only where list/grid pages require it)
7. `ContentBody` (table/grid/cards/drawers/modals)
8. `DiagnosticsState` (loading/empty/error)
9. `PersistentFootnotes` (version/build/runtime state in nav footer)

Per-page required content:

- Dashboard:
  - page title + quick actions
  - attention cards
  - summary metrics
  - recent activity panel
- Inventory:
  - collection browser
  - compact summary line
  - shared filter bar
  - table or cards toggle
  - details drawer and image lightbox
- Wishlist:
  - same interaction model as inventory
  - pricing and schedule signals
- Scanner/Discoveries:
  - query set controls
  - candidate list with action rail
  - failure/health summary
- Pricing:
  - tracked items, trend chart, source breakdown, export
- Settings:
  - tabbed sections with persistent save feedback
- Diagnostics:
  - health cards, logs, export, recovery actions

## 8. API-to-UI Parity Completion Rules
For each user-facing endpoint:
1. At least one screen reads/writes it.
2. Empty/loading/error states exist and are tested.
3. Mutations provide success and failure feedback.
4. Endpoint appears in parity matrix and regression coverage list.

Parity audits currently required:
- Users screen parity audit (`#143`)
- Settings screen parity audit (`#144`)
- Chat screen parity audit (`#145`)

## 9. Testing Blueprint (Detailed)

## 9.1 Test layers
- Unit tests:
  - validation, mappers, reducers, utility functions
- API integration tests (Go):
  - handler + service + persistence behavior
- E2E tests (Cypress):
  - user-visible workflows
- Contract tests:
  - OpenAPI parity and route binding
- Performance and soak:
  - startup, search, scan runtime, large dataset interactions

## 9.2 Mandatory E2E suites to reach release
1. Auth and onboarding
2. Inventory lifecycle
3. Wishlist lifecycle
4. Scanner and discoveries
5. Pricing tracking and exports
6. Dashboard actions
7. Settings persistence
8. Users and integrations operations
9. Chat workflows with guarded actions
10. Diagnostics and backup/recovery

## 9.3 Minimum anti-regression policy
- Any user-reported bug gets:
  - failing E2E or API test first
  - fix
  - passing regression test in CI
- No issue closure without test evidence command output.

## 10. Data Scale and Performance Test Plan

Baseline data packages:
- Small:
  - 500 items, 2k instances
- Medium:
  - 5k items, 20k instances
- Large:
  - 10k items, 50k instances
- Extreme discovery:
  - 200k candidates, 2k tracked items with 365 snapshots each

Per-screen scale checks:
- Dashboard:
  - render and interaction responsiveness under large aggregates
- Inventory/Wishlist:
  - filter/sort/paginate within target latency
  - detail drawer and lightbox maintain smooth interactions
- Scanner:
  - candidate filtering and action throughput
- Pricing:
  - chart and table rendering without lockups

NFR target checks:
- Startup under 2.5s baseline hardware
- Search responses under 200ms for target dataset
- 10-query-set scanner run under 8 minutes

## 11. Security and Privacy Readiness Checklist
- Secrets in secure storage, not plaintext DB.
- Token/key redaction in logs and diagnostics export.
- Auth events logged without sensitive data.
- License verification works offline.
- LAN access mode documented with trust boundaries.
- Recovery flow abuse cases tested.

Required evidence:
- security checklist markdown signed by engineer
- redaction test logs
- secret storage verification tests

## 12. Release Gates and Evidence Pack

## 12.1 Alpha Gate
- Core workflows are functional.
- No blocking UX dead ends.
- Critical E2E suites pass.

Evidence:
- test command logs
- issue closure list for blocker set
- manual smoke checklist

## 12.2 Beta Gate
- NFR measurements completed.
- install/update flow verified on Win+macOS.
- support diagnostics flow ready.

Evidence:
- benchmark report
- installer test report
- backup/restore validation report

## 12.3 GA Gate
- No open P0/P1 defects.
- Feature matrix status moved to Done or explicit defer.
- Production runbook complete.

Evidence:
- final release checklist
- closed issue report
- signed artifacts and rollback plan

## 13. Critical Path Milestones (Execution Order)

## M0: Stability Recovery
Close:
- `#151`, `#149`, `#147`, `#154`, `#152`

## M1: Screen/API Parity
Close:
- `#143`, `#144`, `#145`
- diagnostics page completion issue (create if missing)

## M2: Core UX Completion
Close:
- onboarding wizard completion issues
- inventory/wishlist interaction model issues
- table/card parity issues

## M3: Intelligence Completion
Close:
- scanner/matching/not-in-collection/pricing dashboard integration issues

## M4: Hardening and Release
Close:
- NFR benchmark issues
- security verification issues
- packaging/release readiness issues

## 14. Risk Register (Detailed)
| Risk | Impact | Likelihood | Mitigation | Owner |
| --- | --- | --- | --- | --- |
| UI regressions due to rapid template migration | High | High | Mandatory E2E per affected flow | UI lead |
| Scanner provider instability | High | Medium | Retry/backoff + health indicators + provider mocks | Backend lead |
| Data corruption on import/repair | High | Medium | Dry-run only preview + pre-apply backup | Data owner |
| Auth edge-case lockouts | High | Medium | Recovery fallback + auth lifecycle tests | Auth owner |
| Performance drops at scale | High | Medium | Dataset benchmarks in CI gates | Platform owner |

## 15. Issue Template for Remaining Work
Each implementation issue must contain:
- scope statement
- acceptance criteria
- subtasks checklist
- explicit tests to add/update
- rollout/rollback notes where relevant

Commit format:
- `#<issue> <type>(<scope>): <description>`

Issue closure requirements:
- all subtasks checked
- tests run and passing
- commit pushed
- evidence comment posted

## 16. Immediate Action List (Practical Next Steps)
1. Stabilize Cypress startup and inventory non-500 regressions.
2. Remove all placeholder/template copy from live routes.
3. Complete users/settings/chat API parity audits.
4. Expand E2E to cover onboarding, inventory, wishlist, integrations, and diagnostics.
5. Implement NFR benchmark automation and capture baseline reports.

## 17. Summary
Cabinet is feature-rich at API level but not yet release-ready at UX reliability level.  
The fastest route to done is strict issue-driven execution with test-first validation and hard release evidence.

## 18. Feature-by-Feature Acceptance and Test Map
| Section | Required User Outcome | Minimum Acceptance Criteria | Mandatory Automated Tests |
| --- | --- | --- | --- |
| 1 Core | App installs and boots reliably | Signed install/update works; runtime healthy | API smoke + startup integration + packaging smoke |
| 2 Auth | User registers and unlocks with WebAuthn | First-run gate enforced; lock/unlock works | Auth API suite + onboarding/lock E2E |
| 3 Profiles | User can isolate data by profile | No cross-profile leakage in items/settings/keys | Profile isolation API + E2E switch tests |
| 4 Collection | User can manage canonical items + instances | Full CRUD with validations and status transitions | Items API + inventory lifecycle E2E |
| 5 Photos | User can upload/manage photos safely | Upload, derivatives, primary, delete, fullscreen | Photos API + lightbox/upload E2E |
| 6 Barcodes | User resolves barcodes accurately | Manual + lookup + duplicate handling supported | Barcode API + barcode flow E2E |
| 7 AI Assist | User gets suggestions with guardrails | Confidence + confirmation required for apply | AI API + guarded apply E2E |
| 8 Scanner | User runs/schedules scans and sees candidates | Query sets and failures/retry operate reliably | Scanner API + scanner pipeline E2E |
| 9 Matching | User can trust match confidence outputs | Candidate states and confidence are explainable | Matching API + match review E2E |
| 10 Discoveries | User can triage new non-owned items | Ignore/wishlist/track/create fully functional | Discoveries API + triage E2E |
| 11 Wishlist | User tracks wanted items with targets | Target, status, notes, hits and deletes work | Wishlist API + wishlist lifecycle E2E |
| 12 Pricing | User tracks trends and exports history | Snapshot/trend/by-source/export operate correctly | Pricing API + pricing E2E |
| 13 Dashboard | User sees what needs action now | Attention cards are actionable and data-backed | Dashboard API + dashboard action E2E |
| 14 Search | User can quickly find/filter/sort | Filters/saved views persist and restore | Search API + filter persistence E2E |
| 15 Data Mgmt | User can import/export safely | Dry-run + apply + conflict handling + repair | Data mgmt API + import/export E2E |
| 16 Licensing | User sees entitlement state clearly | Free/pro gates enforce and display correctly | Licensing API + entitlement UI E2E |
| 17 Settings | User can configure app reliably | Every settings screen persists via API | Settings API parity + settings E2E |
| 18 Errors | User recovers from failures cleanly | No dead-end errors for known scenarios | Error regression E2E + recovery API |
| 19 Diagnostics | User/support can export diagnostics | Logs/debug/export/recovery all accessible | Logging API + diagnostics E2E |
| 20 Future Hooks | Disabled hooks do not break app | Hooks visible as disabled/extensible contracts | Contract tests for disabled behavior |
| 21 NFRs | App is fast and stable at scale | Measured against startup/search/scan targets | Benchmark scripts + threshold assertions |
| 22 Security | Sensitive data remains protected | No plaintext secrets, logs redacted | Security checks + redaction tests |
| 23 Chat | User can chat and apply safe actions | Thread, files, preview/apply and logs function | Chat API + chat action E2E |

## 19. Screen Data Requirements for Scalability Testing
| Screen | Dataset | User Behaviors to Validate |
| --- | --- | --- |
| Dashboard | 1k alerts + 10k items summaries | refresh, card action latency, navigation deep-links |
| Inventory | 10k items, 50k instances, 30k photos | search, sort, filters, row click drawer, lightbox carousel |
| Wishlist | 5k entries, mixed statuses | list/card toggle, status filtering, quick actions |
| Scanner | 50 query sets, 200k candidates | run now, candidate filters, bulk triage, failure retry |
| Discoveries | 100k unresolved candidates | triage throughput, action feedback reliability |
| Pricing | 2k tracked items, 365 snapshots each | graph loading, source breakdown, export latency |
| Settings | 100+ persisted preferences | edit/save/reload consistency |
| Users | 500 local users (simulated) | role updates, search, add/invite workflows |
| Chat | 1k threads, 20k messages, 2k attachments | thread switching, message load, preview/apply actions |
| Diagnostics | 100k log lines | filtering, export, debug mode toggles |

## 20. API Domain Checklist (Implementation Tracking)
| Domain | Core Routes | Completion Conditions |
| --- | --- | --- |
| Runtime | `/healthz`, `/api/runtime`, `/api/runtime/recovery` | stable health payloads and recovery semantics |
| Profiles | `/api/profiles*` | create/list/activate/settings/secrets/license stable |
| Auth | `/api/auth/*` | register/login/lock/validate/recovery/cloud hooks stable |
| Items | `/api/items*` | CRUD, instances, barcodes, photos, bulk edit stable |
| Search | `/api/search/items` | filter and sort behavior deterministic |
| Scanner | `/api/scanner/*` | query-set CRUD, run/schedule, failures/retry stable |
| Matching | `/api/matching/*` | run/results confidence payloads stable |
| Discoveries | `/api/discovery/*` | list + action workflow stable |
| Wishlist | `/api/wishlist*` | create/list/delete/hits stable |
| Pricing | `/api/pricing/*` | track, snapshot, trend, export stable |
| Dashboard | `/api/dashboard` | cards/summary payload stable |
| Logs | `/api/logs/*` | activity/export/debug stable |
| Licensing | `/api/license/*` | import/status and enforcement stable |
| AI | `/api/ai/*` | toggle/test/suggest endpoints stable |
| Chat | `/api/chat/*` | threads/messages/attachments/action preview+apply stable |
| Data Mgmt | `/api/data/*` | export/import/reindex/repair stable |
| Backup | `/api/backup/*` | run/list/restore stable |

## 21. Documentation and Operational Readiness Deliverables
Before GA, maintain these docs as release artifacts:
- `docs/APP_COMPLETION_ANALYSIS.md`
- `docs/FULL_FEATURE_LIST.md` with final status annotation
- `docs/UI_ENDPOINT_PARITY.md` synchronized to shipped UI routes
- `docs/USE_CASES_AND_SCENARIOS.md` reflecting final UX
- `docs/api/openapi.yaml` matching runtime handlers
- release runbook (install, update, rollback, support)
- support diagnostics playbook

## 22. Execution Control Rule
Any issue that changes user-visible behavior must include:
1. failing test first
2. implementation
3. passing test evidence
4. updated docs/parity entry when applicable
5. closure checklist completion
