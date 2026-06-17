# Tasks / Operational Surfaces Exploration - 2026-06-18

Issue: #495
Run window: 2026-06-18 02:00 Australia/Sydney / 2026-06-17 16:00 UTC

## Scope

This report audits the current Cabinet tasks and operational work surfaces from live repo/runtime state. The original section scope named task list, task detail/actions, queue/status behavior, and batch/execution surfaces if present.

## Preconditions And Evidence

- Branch at start: `develop`, clean and current with `origin/develop`.
- Work branch: `issue-495-tasks-operational-exploration`.
- Demo2 runtime was healthy at `http://127.0.0.1:17882`.
- `/api/runtime` reported `app_version=rev-d0fc6edc3a0e`, `runtime_port=17882`, and data dir `C:\projects\collectors-tech\cabinet\tmp\demo2\data`.
- Required browser profile `project-cabinet` was running and CDP doctor passed, but `open`, `navigate`, `snapshot`, and `screenshot` calls returned `browser navigation blocked by policy`. This pass therefore does not claim a live browser element-by-element validation.

Evidence logs for this run are under `.work-agent/logs/issue-495-tasks-operational-exploration/20260618-0200/`.

## Route And Surface Inventory

### Standalone Task Route

Status: not present in the current route tree.

Evidence:
- `ui.web/src/routeTree.gen.ts` contains authenticated routes for Dashboard, Wishlist, Users, Scanner, Reports, Purchases, Media, Inventory, Integrations, Inbox, Help Center, Discoveries, Collections, Chats, and Settings.
- `ui.web/src/routes/_authenticated/tasks/index.tsx` is absent.
- `ui.web/src/features/tasks` remains a shared table/drawer implementation used by Inventory and Wishlist semantics, not an independently routed operational task screen.
- `internal/app/ui_template_contract_test.go` explicitly forbids legacy `/_authenticated/tasks/` bindings in migrated Inventory and shared task-table code.

Classification: complete with limitation. There is no current standalone task-list/detail route to audit. Follow-up is not a product bug unless product direction reintroduces a first-class task workspace.

### Settings Operations Route

Status: implemented and traceability-backed.

Route: `/settings/operations`

Observed implementation surfaces from `ui.web/src/features/settings/operations/index.tsx`:
- Runtime card loads `/api/runtime` and `/api/runtime/recovery`, with retry/error state.
- Runtime setup import posts `/api/runtime/setup-import`.
- Auth recovery card supports passphrase setup and reset-session start.
- Diagnostics logs card exports redacted runtime logs.
- JSON import/export card supports export, dry-run, conflict default selection, apply, and summary/error states.
- CSV import/export card supports export, dry-run, custom column mapping, conflict default selection, apply, and summary/error states.
- Queue controls pause and resume worker scheduling through profile settings keys `scanner_schedule` and `operations_queue_resume_schedule`.

Existing coverage and traceability:
- `openspec/specs/settings/ui-screen-settings/spec.md` requires `/settings/operations` deep-link routing and active sidebar state under `UI-SCREEN-SETTINGS-006`.
- `openspec/traceability.md` maps data-management operations to `ui.web/cypress/e2e/settings/operations/spec.cy.ts`.
- `internal/app/ui_template_contract_test.go` includes `TestSettingsOperationsQueueControlsContract` and `TestSettingsOperationsRecoveryWorkflowContract`.
- `ui.web/cypress/e2e/settings/operations/spec.cy.ts` covers runtime/recovery display, retry behavior, queue pause/resume, recovery workflow, JSON import/export dry-run/apply/failure, CSV import/export dry-run/apply/failure/mapping, log export, and setup import states.
- `ui.web/cypress/e2e/settings/ui-screen-settings/spec.cy.ts` covers `/settings/operations` deep-link and keyboard/sidebar navigation.

Classification: covered by existing spec and tests; live browser validation blocked by tool policy in this cron context.

### Other Operational Queues

Operational queue concepts exist in feature-specific specs and routes rather than a single tasks route:
- Notification queue: `openspec/specs/chats/notification-inbox/spec.md`.
- Assistant workflow queued/running/completed/failed states: `openspec/specs/chats/assistant-execution-surfaces/spec.md` and `openspec/specs/chats/assistant-inbox-handoff/spec.md`.
- Media and scanner upload/review queues: `openspec/specs/media/ui-screen-media/spec.md` and `openspec/specs/scanner/ui-screen-card-scanner/spec.md`.
- eBay seller operations and listing lifecycle safety-gated workflows: `INTEGRATION-027` and `INTEGRATION-028`.

Classification: these are covered by their owning section issues, not by #495's standalone task-route scope.

## Scenario Matrix

| Scenario class | Status | Evidence |
| --- | --- | --- |
| Happy path task list | Blocked/not applicable | No standalone `/tasks` route exists in current route tree. |
| Task detail/actions | Blocked/not applicable | No standalone task detail/action screen exists; shared task table is owned by Inventory/Wishlist. |
| Empty task state | Blocked/not applicable | No standalone task screen exists. |
| Queue/status happy path | Covered by tests | Settings Operations queue pause/resume has UI contract and Cypress coverage. |
| Queue/status error path | Covered by tests | Settings Operations runtime/error/retry and import failure states are covered. |
| Batch/execution surfaces | Covered by owning surfaces | Data import apply, CSV apply, logs export, setup import, assistant workflows, scanner/media queues have separate specs/tests. |
| Keyboard path | Covered by tests | Settings operations route keyboard/sidebar navigation is covered by settings Cypress. |
| Refresh/back-forward persistence | Covered by tests where applicable | Cypress asserts route remains `/settings/operations` after key operations. |
| Live browser element pass | Blocked | `project-cabinet` browser doctor passed, but browser actions returned `browser navigation blocked by policy`. |

## Findings And Follow-Up

No new Cabinet product defect was created from this pass.

The only uncovered item is a tooling/precondition limitation: live browser element-by-element validation for #495 could not be completed through the required `project-cabinet` browser profile in this isolated cron context. Since static route/spec/test evidence shows no standalone task route and Settings Operations is already covered, this is recorded as a confidence limit rather than a Cabinet product issue.

If product direction requires a first-class task workspace, open a new implementation issue with a concrete product contract for:
- `/tasks` route ownership and navigation placement
- task list source of truth
- task detail/action model
- queue/status behavior
- Cypress coverage separate from Inventory/Wishlist shared table reuse

## Completeness

Completeness label: Complete with blockers.

Reachable repo/runtime evidence was audited, Settings Operations coverage was reconciled, and the absent standalone task route was explicitly classified. A live browser element-level pass remains blocked by browser tool policy and should be rerun only when `project-cabinet` browser actions can snapshot/navigate normally.

## Next Recommendation

After #495 is reviewed/merged, resume route-ordered exploration at #496 Integrations unless higher-priority Cabinet feedback is added.
