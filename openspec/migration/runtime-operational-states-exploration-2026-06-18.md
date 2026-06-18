# Runtime / Operational States Exploration Audit - 2026-06-18

Issue: #500 `[Exploration] Runtime / operational states audit and traceability pass`

Mode: exploratory documentation/traceability slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-500-runtime-operational-states-audit`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-d24ccca5cdd2`, `runtime_port=17882`, pid `70808`, `bind_mode=lan`.
- Local repo: clean `develop` at `d24ccca5cdd2850f1d370f684ace8a197862e209` before branch creation.
- Browser/profile: OpenClaw `project-cabinet` profile doctor passed through CDP port `18801`.
- Browser limitation: OpenClaw browser `open` for `http://127.0.0.1:17882/settings/operations` returned `browser navigation blocked by policy`. This pass therefore uses runtime health plus repo/source/spec/test evidence and does not claim fresh live element-by-element browser interaction.

Evidence logs for this run are under `.work-agent/logs/issue-500-runtime-operational-states-audit/20260618-1530/`.

## Scope Reviewed

Section:
- Runtime / operational states.

Screens/routes:
- Runtime endpoints: `/healthz`, `/api/runtime`, `/api/runtime/recovery`, `/api/runtime/setup-status`, `/api/runtime/setup-complete`, `/api/runtime/setup-import`, `/api/runtime/setup-storage-validate`, `/api/runtime/update/install`, and `/api/runtime/shutdown`.
- Setup and recovery-adjacent UI: first-run setup, route guard/error pages, active-profile recovery, profile switcher, and shell runtime metadata.
- Operational settings routes: `/settings/storage` and `/settings/operations`.
- Representative app surfaces with explicit loading, empty, error, retry, and unavailable states: Dashboard, Inventory, Media, Notification Inbox, Integrations, Market Watch, Scanner, Users, Discoveries, Reports, Chat/Assistant, and Purchases package inbox.

Component/feature surfaces inventoried from source/spec/test evidence:
- Runtime health metadata, build version, port/bind mode, data directory, update channel, recovery flag, setup config import/export/validation, startup/shutdown lifecycle metadata, structured runtime/access/error logs, PID attach/stale recovery, fallback port negotiation, and Cypress stale-runtime guard.
- Route-level 401/403/404/500/503 error components and authenticated route error handling.
- Loading spinners/text, empty states, retry buttons, unavailable cards, provider-health panels, scanner failure retry, profile-load retry, storage degradation, backup/restore/integrity failure feedback, notification retry, and assistant policy/setup error feedback.

## Findings And Changes

No new product defect or uncovered runtime-operational contract was found in the reviewed no-browser-interaction evidence set.

The durable change for #500 is this audit artifact. It records that current runtime/operational state surfaces are already represented by focused OpenSpec requirements and executable Go/Cypress coverage for the non-live-provider scope. The report also records the isolated-browser limitation so #500 is not treated as fresh live UI proof.

## Screen And Component Coverage

### Runtime Health And Diagnostics

Surfaces: `/healthz`, `/api/runtime`, startup output, runtime logs, setup metadata, PID attach/recovery, and Cypress runner runtime preflight.

Classification: covered by existing spec/tests and fresh HTTP proof for demo2.

Requirement anchors:
- `RUNTIME-CORE-001` through `RUNTIME-CORE-018`.
- `RUNTIME-CONFIG-ENV-001` and `RUNTIME-CONFIG-ENV-002`.
- `RUNTIME-NETWORK-LAN-001` and `RUNTIME-NETWORK-LAN-002`.
- `RUNTIME-MULTI-001` through `RUNTIME-MULTI-008`.

Evidence:
- `internal/app/app.go` implements `/healthz`, `/api/runtime`, `/api/runtime/recovery`, setup, update, shutdown, and runtime logging handlers.
- `internal/app/logging_recovery_api_test.go` covers `/api/runtime`, `/api/runtime/recovery`, and recovery-required state.
- `tests/runtime_script_policy_test.go` and `tests/api_contract_smoke_test.go` cover Cypress stale-runtime app-version rejection and API smoke runtime metadata.
- `openspec/specs/general/runtime-core/spec.md`, `runtime-config-env/spec.md`, `runtime-network-lan/spec.md`, and `runtime-multi-instance/spec.md` define startup, health, diagnostics, config, LAN, parallel startup, and orchestration requirements.
- Fresh demo2 proof from this run: `/healthz` HTTP 200 `ok`; `/api/runtime` reports `rev-d24ccca5cdd2` on port `17882`.

Element inventory and classification:
- Health endpoint: covered and freshly checked.
- Runtime metadata endpoint: covered and freshly checked.
- Recovery endpoint and clean-shutdown flag: covered by Go tests.
- Setup status/import/complete/storage-validate endpoints: covered by setup and storage-focused tests.
- Update install errors and channel mismatch responses: covered by runtime/update API code and OpenSpec runtime requirements.
- Shutdown endpoint: source verified; live shutdown was intentionally not executed against demo2.
- Runtime logs/access logs/error logs: covered by `RUNTIME-CORE-014` and runtime logging implementation.
- Cypress stale-runtime preflight: covered by runner policy tests and traceability row `RUNTIME-CORE-018`.

Scenario status:
- Happy path: Validated by fresh `/healthz` and `/api/runtime` proof plus Go tests.
- Cancellation/abort path: Covered for pre-canceled startup context by `RUNTIME-CORE-016`; runtime shutdown was not live-executed in this exploration pass.
- Invalid input path: Covered by runtime config/env and setup validation requirements/tests.
- Empty state: Not applicable to raw health endpoints.
- Loading state: Not applicable to raw health endpoints; UI route loading states are covered below.
- Error/failure state: Covered by recovery, config validation, stale runtime, update-install, and startup failure requirements/tests.
- Keyboard-only path: Not applicable to raw endpoints.
- Refresh/back-forward persistence path: Not applicable to raw endpoints.
- Post-action verification path: Covered by runtime metadata and health response assertions.
- Permission/role variance: LAN auth protections are specified in `RUNTIME-NETWORK-LAN-002`; protected endpoint behavior remains covered by auth/API tests.

Residual limit:
- This pass did not intentionally trigger shutdown, stale PID, update install, or unsafe recovery mutations against demo2.

### Setup, Recovery, And Route Error States

Surfaces: first-run setup wizard, authenticated route guards, Clerk bootstrap fallback, route error pages, active profile recovery, and shell profile switcher.

Classification: covered by existing spec/tests for deterministic setup/error/retry contracts.

Requirement anchors:
- `UI-SCREEN-ONBOARDING-AUTH-*`.
- `UI-LOGIN-SESSION-*`.
- `UI-FOUNDATION-SHELL-NAVIGATION-*`.
- `PROFILES-003` / active-profile recovery rows in traceability.
- Runtime setup requirements in `RUNTIME-CORE-*`.

Evidence:
- `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` covers setup wizard first-run behavior.
- `ui.web/cypress/e2e/general/route-guards-errors/spec.cy.ts` covers route guard and error surfaces.
- `ui.web/cypress/e2e/general/profile-context-recovery/spec.cy.ts` covers active-profile load failure and retry recovery.
- `ui.web/src/routes/(errors)/401.tsx`, `403.tsx`, `404.tsx`, `500.tsx`, and `503.tsx` render route-level operational error pages.
- `ui.web/src/routes/clerk/_authenticated/route.tsx` renders identity loading and bootstrap failure feedback.
- `ui.web/src/components/layout/team-switcher.tsx` exposes profile loading, unavailable, and retry states.

Element inventory and classification:
- First-run setup step state, import, storage validation, and setup completion: covered.
- Route 401/403/404/500/503 screens: covered by route error specs/source.
- Clerk identity loading/bootstrap error: source verified and covered by auth/session specs.
- Profile unavailable/retry control in team switcher: covered by profile-context recovery Cypress.
- Shell runtime metadata block: covered by shell navigation/runtime metadata traceability.

Scenario status:
- Happy path: Validated by setup/auth/profile Cypress coverage.
- Cancellation/abort path: Covered where setup/profile flows expose controlled navigation or retry instead of destructive mutation.
- Invalid input path: Covered by setup storage validation and auth/session validation tests.
- Empty state: Covered for missing/empty profile contexts and signed-out guidance.
- Loading state: Covered for identity/profile loading.
- Error/failure state: Covered for route errors, bootstrap failures, and profile recovery.
- Keyboard-only path: Covered by shell/auth/menu shortcut and navigation Cypress suites where route controls are interactive.
- Refresh/back-forward persistence path: Covered by route guard/navigation tests where applicable; not freshly browser-exercised due policy.
- Post-action verification path: Covered by setup/profile recovery assertions.
- Permission/role variance: Covered by auth lock and route guard tests.

Residual limit:
- Browser-policy blocking prevented a fresh manual pass through the setup and error routes in this run.

### Settings Storage And Operations

Routes: `/settings/storage` and `/settings/operations`.

Classification: covered by existing spec/tests for operational diagnostics, degraded states, recovery, and data-action feedback.

Requirement anchors:
- `UI-SCREEN-SETTINGS-STORAGE-001` through `011`.
- Settings Operations route contracts reconciled under #495.
- `DATA-MANAGEMENT-*` traceability rows for import/export failure recovery.

Evidence:
- `openspec/specs/settings/storage/spec.md` defines storage path, degraded storage info, retry, maintenance, backup table, restore failure, integrity check, and export download contracts.
- `ui.web/cypress/e2e/settings/storage/spec.cy.ts` and `settings/storage-export/spec.cy.ts` cover storage info error/retry, maintenance action completion/failure, backup restore failure, integrity check failure, and disabled export downloads while degraded.
- `ui.web/cypress/e2e/settings/operations/spec.cy.ts` covers operations import/export/logs/queue controls from the #495 reconciliation.
- `internal/app/app.go` implements storage validation, backup, restore, integrity, import/export, and runtime recovery APIs.

Element inventory and classification:
- Storage info ready/degraded panels: covered.
- Retry storage fetch: covered.
- Media path write-check and validation: covered.
- Reindex Search and Rebuild Thumbnails maintenance actions: covered.
- Backup table, sort, download, restore, and restore failure feedback: covered.
- Run Integrity Check healthy/failure feedback: covered.
- JSON/CSV export ready/degraded state: covered.
- Operations import/export/logs/queue controls: covered by #495 audit and Cypress evidence.

Scenario status:
- Happy path: Validated by storage and operations Cypress/Go coverage.
- Cancellation/abort path: Covered for backup restore confirmation behavior.
- Invalid input path: Covered for invalid storage path and import failure feedback.
- Empty state: Covered by degraded/no-backup/no-queue operational states where applicable.
- Loading state: Covered by storage/operations Cypress and source.
- Error/failure state: Validated for storage info failure, maintenance failure, restore failure, integrity failure, and import/export failure.
- Keyboard-only path: Covered indirectly by native button/form controls; no fresh manual keyboard pass claimed.
- Refresh/back-forward persistence path: Route stability is covered for failure states; browser history was not freshly exercised due policy.
- Post-action verification path: Validated by Cypress assertions that maintenance, restore, integrity, and export state changes produce deterministic feedback without false success.
- Permission/role variance: Active-profile context blocking is covered by settings/profile tests.

Residual limit:
- No live destructive restore, shutdown, or data import was performed against demo2 in this pass.

### Representative App Loading, Empty, Error, Retry, And Unavailable States

Surfaces: Dashboard, Inventory, Media, Notification Inbox, Integrations, Market Watch, Scanner, Users, Discoveries, Reports, Chat/Assistant, and Purchases package inbox.

Classification: covered by existing focused specs/tests for representative operational state patterns.

Requirement anchors and evidence:
- Dashboard Home: `UI-SCREEN-HOME-003` covers error and retry recovery in `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts`.
- Discoveries: `UI-SCREEN-DISCOVER-003` covers loading and retryable error states.
- Reports: `UI-SCREEN-REPORTS-003` covers deterministic loading/empty/error states.
- Inventory Items: `UI-SCREEN-INVENTORY-ITEMS-002` covers inline API failure and retry.
- Inventory Photos/Barcodes/AI Assist: focused Cypress suites cover loading, empty, error, retry, and recovered-ready states.
- Media: `UI-SCREEN-MEDIA-008` covers API ready/empty/error/retry/unlinked states.
- Notification Inbox: `UI-SCREEN-NOTIFICATION-INBOX-002`, `005`, and `006` cover contextual empty states, loading, retryable API error, and failed update preservation.
- Integrations: `UI-SCREEN-INTEGRATIONS-005`, `008`, `013`, and provider specs cover bootstrap error/retry, provider health reconciliation, direct-route empty states, and capability blockers.
- Market Watch and Scanner: `UI-SCREEN-MARKET-WATCH-003`, `004`, `UI-SCREEN-SCANNER-002`, and `003` cover provider failure guidance, retry, workspace states, empty/no-output states, and scanner failure retry.
- Users: `UI-SCREEN-USERS-006` covers loading and empty users table states.
- Chat/Assistant: chat and assistant workspace specs cover loading, empty, setup-needed, policy blocked, preview/apply failure, and retry guidance.
- Purchases package inbox: `INTEGRATION-029` through `062` cover package import/list/reconciliation, suggestion empty/error states, and retry-safe failures.

Element inventory and classification:
- Page-level loading placeholders: covered across representative Cypress suites.
- Contextual empty-state cards/tables: covered across Dashboard, Reports, Users, Notification Inbox, Media, Market Watch, Scanner, and Purchases.
- Retry buttons/actions: covered for dashboard, profile, storage, notification inbox, integrations, scanner, market watch, and media.
- Provider/auth unavailable panels and capability blockers: covered for Integrations, Market Watch, Scanner, Assistant, and eBay seller workflows.
- Toast/feedback surfaces for mutation failures: covered where mutating flows exist, especially settings, storage, users, assistant, and provider workflows.

Scenario status:
- Happy path: Validated by each focused suite for its ready-state contract.
- Cancellation/abort path: Covered on surfaces with confirmations/dialogs; not applicable to passive loading/empty states.
- Invalid input path: Covered for forms/imports/settings/provider workflows, not applicable to read-only empty states.
- Empty state: Validated by focused Cypress coverage.
- Loading state: Validated by focused Cypress/source evidence.
- Error/failure state: Validated by focused Cypress/Go coverage.
- Keyboard-only path: Covered in broader shell/navigation/accessibility suites, but not freshly browser-exercised in this run.
- Refresh/back-forward persistence path: Covered where route/query state is central, such as Integrations and Settings; not freshly exercised due browser policy.
- Post-action verification path: Covered by mutation/retry assertions in focused suites.
- Permission/role variance: Covered for active-profile/auth lock/provider capability boundaries.

Residual limit:
- This pass samples representative operational state coverage from existing specs/tests and does not claim exhaustive live route interaction.

## Traceability Summary

Existing traceability already maps the reviewed runtime/operational contracts:
- Runtime health/startup/config/LAN/Cypress-runner contracts -> `RUNTIME-CORE-*`, `RUNTIME-CONFIG-ENV-*`, `RUNTIME-NETWORK-LAN-*`, and `RUNTIME-MULTI-*`.
- Settings Storage and Operations recovery flows -> focused storage/operations Cypress and Go/API tests.
- App surface loading/empty/error/retry patterns -> focused Cypress suites under `ui.web/cypress/e2e/` and Go tests listed in `openspec/traceability.md`.
- Provider/assistant unavailable and capability-blocked states -> Integrations, Market Watch, Scanner, Assistant, and eBay traceability rows.

No traceability rows were changed in this run because the current rows already point to focused executable coverage and no new gap was found.

## Follow-Up Recommendation

#500 can close after this report is merged if reviewer accepts repo/source/spec/test evidence plus the recorded browser-policy limitation. The next route-ordered exploration issue is #501 Cross-cutting UX checks, unless newer higher-priority product feedback is added.

Completeness label: Complete with browser-policy limitation. Runtime health was freshly checked, all reviewed operational state surfaces were tied to existing specs/tests, no uncovered product gap was found, and fresh live browser interaction remains blocked by OpenClaw browser policy in this isolated cron context.
