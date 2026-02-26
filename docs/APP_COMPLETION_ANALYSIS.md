# Cabinet App Completion Analysis (Execution Blueprint)

Last updated: 2026-02-26  
Owner: Product + Engineering (single backlog source: GitHub issues)

## 1. Objective
Define the exact path to take Cabinet from current implementation to release-ready across the full scope in `docs/FULL_FEATURE_LIST.md`.

This document is the execution blueprint for:
- feature completion
- UI/API parity
- test completeness
- release readiness

## 2. What "Done" Means
Cabinet is only "done" when all of the following are true:
1. Every feature section (1-23) in `docs/FULL_FEATURE_LIST.md` is implemented and user-accessible.
2. Every user-facing workflow has matching UI + API + persistence + diagnostics.
3. Every feature has automated regression coverage (API + E2E where user-visible).
4. No P0/P1 open defects remain.
5. Release gates (alpha, beta, GA) pass with evidence.

## 3. Current-State Snapshot (Evidence-Based)

### 3.1 Strengths Already in Place
- Broad API surface implemented in `internal/app/app.go`.
- Significant API test coverage exists in `internal/app/*_api_test.go`.
- OpenAPI route + apidocs endpoint exist.
- Template migration to `ui.web` established with modern route structure and shadcn-style primitives.
- Core nav structure exists for major app areas (dashboard, inventory, wishlist, integrations, settings, users, chat).

### 3.2 Active Gaps and Risks (Open Issues)
- `#154` Inventory layout cleanup.
- `#152` Remove placeholder/template sample copy.
- `#151` Cypress startup failure on Windows.
- `#149`, `#147` Inventory 500-regression coverage gaps.
- `#145` Chat API wiring audit.
- `#144` Settings API wiring audit.
- `#143` Users API wiring audit.

### 3.3 Delivery Risk Summary
- Highest risk is not backend API breadth; it is frontend workflow completion and E2E reliability.
- UI foundation is present, but many screens still need production-level wiring/UX hardening.
- Test-first discipline is not consistently reflected by current UX defects reaching runtime.

## 4. Full Feature Completion Matrix (Sections 1-23)

Legend:
- `Done`: implemented, validated, and tested end-to-end
- `Partial`: implementation exists but missing UX/test/reliability/completion elements
- `Missing`: not implemented to required scope

| Feature Section | Status | Backend | UI | Tests | Key Gaps to Close |
| --- | --- | --- | --- | --- | --- |
| 1. Application Core | Partial | Mostly present | Partial | Partial | Installer/update hardening, release verification on both OSes |
| 2. Login/Auth (WebAuthn-first) | Partial | Present (WebAuthn + lock/recovery + cloud hooks) | Partial | Partial | Complete first-run/lock UX, edge-case recovery UX, full E2E |
| 3. Multi-profile support | Partial | Present | Partial | Partial | Profile-switch UX and permissions clarity |
| 4. Collection management | Partial | Present | Partial | Partial | Full field coverage, polished CRUD UX, no-regression E2E |
| 5. Photo system | Partial | Present | Partial | Partial | Full lightbox workflow, upload edge handling, UX polish |
| 6. Barcode system | Partial | Present | Partial | Partial | Confidence/duplicate variant UX, E2E paths |
| 7. AI Assist | Partial | Present | Partial | Partial | Confirm-before-apply UX consistency + diagnostics parity |
| 8. Scanner engine | Partial | Present | Partial | Partial | Scheduler/health/failure UX + reliability tests |
| 9. Matching engine | Partial | Present | Partial | Partial | Confidence explainability and action flow UX |
| 10. Not-in-collection | Partial | Present | Partial | Partial | Production-grade triage/action UX + E2E |
| 11. Wishlist | Partial | Present | Partial | Partial | Complete table/card workflows + mutation regression tests |
| 12. Price tracking | Partial | Present | Partial | Partial | Snapshot cadence UX, trends fidelity, export UX/tests |
| 13. Dashboard | Partial | Present | Partial | Partial | Actionable attention model, compact layout, meaningful data cards |
| 14. Search/filtering | Partial | Present | Partial | Partial | Saved views UX, cross-screen filter consistency |
| 15. Data management | Partial | Present | Partial | Partial | Full import conflict UX, repair/reindex operator flow |
| 16. Licensing | Partial | Present | Partial | Partial | Entitlement UX clarity, billing/cloud ownership paths |
| 17. Settings | Partial | Present | Partial | Partial | Full screen wiring audit + persistence verification |
| 18. Error handling | Partial | Present | Partial | Partial | Runtime-safe recovery UX + non-500 guarantees |
| 19. Logging/diagnostics | Partial | Present | Partial | Partial | Dedicated diagnostics screen quality + export UX |
| 20. Future hooks | Partial | Scaffolded | Minimal | Minimal | Explicitly gated UX + extension contracts |
| 21. NFRs | Missing | Partial | N/A | Missing | Measured startup/search/scan performance gates |
| 22. Security/privacy | Partial | Partial | N/A | Partial | Secret handling audits + log redaction verification |
| 23. In-app chat copilot | Partial | Present | Partial | Partial | Cross-screen chat workflow, action preview/apply confidence UX |

## 5. Gap Analysis by Workstream

### 5.1 Platform and Runtime
Must finish:
- deterministic startup/runtime health behavior
- binary packaging + signed update validation flow
- local-LAN access behavior and security profile definition

### 5.2 Identity and Access
Must finish:
- fully guided WebAuthn-first onboarding
- lock/unlock/recovery edge cases
- cloud ownership and local profile lease behavior clarity (where enabled)

### 5.3 Core Data Model and Workflows
Must finish:
- inventory and wishlist production CRUD UX (tables/cards/details/lightbox)
- status lifecycle (`active -> deleted -> recycle`) with hard-link protection in recycle purge
- enum configurability governance (admin-only)

### 5.4 Discovery/Pricing Intelligence
Must finish:
- scanner reliability + provider health + stock count capture consistency
- confidence-based matching review UX
- dashboard "what needs attention now" as primary operations surface

### 5.5 Settings/Diagnostics/Admin
Must finish:
- complete settings route-to-API parity audit
- users/admin API wiring and role behavior
- first-class diagnostics page replacing homepage debug links

### 5.6 UI Foundation and UX Consistency
Must finish:
- strict adoption of shadcn-admin baseline shell patterns
- sticky top header and fixed/collapsible nav behavior
- remove all template/sample placeholder content
- density, i18n readiness, and keyboard/command consistency

### 5.7 Quality and Test Infrastructure
Must finish:
- Windows-stable Cypress startup (`#151`)
- test-first enforcement for all remediations
- no feature merges without passing E2E for critical path screens

## 6. Milestones to Completion

## M0: Stability Recovery (Immediate)
Scope:
- close `#151`, `#149`, `#147`, `#154`, `#152`
- eliminate known 500 pathways in inventory/wishlist/integrations
- ensure deterministic app boot + smoke E2E

Exit criteria:
- Cypress startup fixed on Windows
- inventory and wishlist navigation never produce 500 in tested flows
- placeholder/template copy removed from user-facing UI

## M1: Foundation Parity
Scope:
- close audits: `#143`, `#144`, `#145`
- ensure all nav screens perform real API reads/writes
- convert diagnostics links into dedicated diagnostics screen

Exit criteria:
- users/settings/chat screens confirmed API-backed
- no fake/sample runtime data in production routes
- E2E route parity suite green

## M2: Core Collector UX Completion
Scope:
- onboarding wizard complete (identity -> starter data -> first item -> preferences)
- inventory + wishlist details interactions standardized
- table/card toggle, row-click drawer, thumbnail lightbox carousel

Exit criteria:
- first-run user reaches usable workspace in <5 steps
- data operations are predictable with keyboard + mouse + mobile behavior
- CRUD flows fully covered by E2E

## M3: Discovery + Pricing Operational Readiness
Scope:
- scanner query sets, candidate review, matching confidence UX
- not-in-collection triage workflows
- pricing snapshots, trend cards, and per-source visibility

Exit criteria:
- user can run scanner -> triage -> wishlist/track/create without dead ends
- pricing and watchlist signals populate dashboard attention center
- regression suite covers end-to-end discovery path

## M4: Commercial + Security + Release
Scope:
- licensing + cloud ownership gates
- security/privacy verification
- NFR benchmark evidence
- release packaging/documentation finalization

Exit criteria:
- alpha -> beta -> GA gates passed with evidence package
- no P0/P1 defects open
- release runbook complete

## 7. Testing Strategy Required to Finish

### 7.1 Mandatory Test Layers
1. API tests (Go): persistence, validation, authz, profile isolation.
2. E2E tests (Cypress): user workflows and route reliability.
3. Contract tests: OpenAPI parity and route contract checks.
4. Release checks: startup/build/package/smoke on Windows + macOS.

### 7.2 Critical E2E Journeys (Must Be Green)
1. First launch -> WebAuthn registration -> unlock -> onboarding completion.
2. Inventory create/edit/delete/recycle/restore.
3. Wishlist + pricing track + dashboard attention signal.
4. Scanner run -> candidate review -> not-in-collection actions.
5. Settings persistence, backups, diagnostics export.
6. Chat open/close/thread/action preview/apply guardrails.

### 7.3 Scalability Test Data Plan (Minimum)
- Inventory:
  - 100 folders
  - 10,000 items
  - 50,000 instances
  - 30,000 photos (thumb+preview)
- Scanner/pricing:
  - 50 query sets
  - 200,000 candidate records
  - 365 daily price snapshots per 2,000 tracked items
- Users/profiles:
  - 20 profiles
  - profile-isolated datasets with no cross-leakage

Expected validation:
- no UI lockups in list virtualized views
- search/filter under target thresholds
- page interactions remain responsive under large datasets

## 8. Release Gates

## Alpha Gate
- Core workflows functional without blocking defects.
- Critical E2E journeys green.
- API docs and diagnostics available.

## Beta Gate
- NFR benchmarks measured and within targets.
- installer/update flow validated on Windows and macOS.
- support workflow ready (logs export, recovery path, known issues).

## GA Gate
- zero open P0/P1 issues
- all feature sections 1-23 at `Done` or explicitly deferred with stakeholder sign-off
- signed release artifacts + rollback plan

## 9. Backlog Alignment Rules
- Keep one source of truth: GitHub issues in `collectors-tech/cabinet`.
- Every feature gap in this doc must map to an issue with:
  - acceptance criteria
  - explicit subtasks
  - test evidence requirements
- Close issue only when:
  - implementation complete
  - tests pass in-session
  - commit/push done
  - checklist fully checked

## 10. Immediate Next Execution Queue (Ordered)
1. Finish reliability blockers: `#151`, `#149`, `#147`.
2. Finish UI cleanup parity: `#154`, `#152`.
3. Complete API wiring audits: `#143`, `#144`, `#145`.
4. Create/close missing issues for feature sections still in `Partial`/`Missing` state (especially NFR/security/release hardening).

## 11. Final Note
Cabinet is close on breadth of capability, but not yet release-complete on UX reliability and proof of quality.  
The fastest path to done is strict issue-driven execution against this matrix with mandatory test evidence at each closure.
