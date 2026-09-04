# Collections Exploration Audit - 2026-06-18

Issue: #493 `[Exploration] Collections audit and traceability pass`

Mode: exploratory documentation/audit only. No product code, runtime code, fixture, or test code changed in this slice.

Runtime and profile evidence:
- Browser/profile: OpenClaw `project-cabinet` profile, attached to Brave through CDP port `18801`.
- Route: `http://127.0.0.1:17882/collections/`.
- Demo lane: demo2-helper on port `17882`.
- Runtime commit: `develop` app_version `rev-f5200e08e89e`.
- Health evidence: `.work-agent/logs/issue-493-collections-exploration/20260618-0130/demo2-healthz.txt`.
- Runtime evidence: `.work-agent/logs/issue-493-collections-exploration/20260618-0130/demo2-runtime.json`.
- Initial DOM inventory: `.work-agent/logs/issue-493-collections-exploration/20260618-0130/collections-initial-cdp.json`.
- Interaction evidence: `.work-agent/logs/issue-493-collections-exploration/20260618-0130/collections-interaction-cdp-2.json`.

## Scope Reviewed

Section:
- Collections route and table-driven collection management surface.

Screens/routes:
- `/collections/` with active sample/showcase profile state.
- Row-level `View` navigation was attempted in the live route; it navigated to `/inventory/` as expected, but the first combined CDP run was interrupted by target navigation before returning its full result. The navigation result is recorded by the browser tab state and the interrupted log at `.work-agent/logs/issue-493-collections-exploration/20260618-0130/collections-interaction-cdp.json`.

Component/feature surfaces observed:
- Global Collections header and primary `New collection` action.
- Collections search input and management summary.
- Collections shared table.
- Row selection for `All Items`, `Store 1`, `Store 2`, `Store 3`, and `Store 4`.
- Row `View`, `Edit`, and `Delete` controls for protected and custom rows.
- Collection members panel.
- Members table search input, row list, and pagination.
- Create collection dialog.
- Edit collection side panel.
- Delete collection confirmation dialog.
- Protected `All Items` edit/delete attempts.

## Live Outcome

The reachable Collections route rendered the expected shared table management surface and lower members table. The active persisted context opened on `Store 1`, and selecting `All Items` updated the members table to the live inventory total of 102 records. Selecting `Store 1` showed its single assigned member. Members pagination moved from page 1 to page 2 and back while preserving the `All Items` active context.

Create, edit, and delete transient surfaces opened from their row/header controls and cancelled without visible table mutation. Protected `All Items` edit and delete attempts opened the same edit/delete surfaces and were cancelled; destructive completion was not executed against shared demo data in this pass. Existing focused Cypress coverage under #1078 remains the mutation-proof authority for protected-row submit behavior.

No new product defect was confirmed in this exploratory slice. No new GitHub issue was opened.

## Scenario Matrix

| Scenario class | Result | Evidence and traceability |
| --- | --- | --- |
| Happy path route render | Validated | `/collections/` loaded with table surface, 5 collections, active context, tag page identity, and members panel. Covered by `UI-SCREEN-COLLECTIONS-001`, `002`, `009`, `016`, `018`. |
| Row selection | Validated | `Store 1` active context showed one assigned member; `All Items` restored 102 live members. Covered by `UI-SCREEN-COLLECTIONS-002`, `016`, `027`. |
| Create dialog open/cancel | Validated for cancellation | `New collection` opened `collections-create-dialog`; Cancel closed it without adding `Cron Cancel Shelf`. Covered by `UI-SCREEN-COLLECTIONS-003`, `022`, `023`, `025`. |
| Edit side panel open/cancel | Validated for cancellation | `Store 1` edit opened `collections-edit-panel`; Cancel closed it and row remained unchanged. Covered by `UI-SCREEN-COLLECTIONS-004`, `025`, `026`. |
| Delete dialog open/cancel | Validated for cancellation | `Store 1` delete opened `collections-delete-dialog`; Cancel closed it and row remained. Covered by `UI-SCREEN-COLLECTIONS-005`, `025`. |
| Protected default row attempts | Partially validated live, fully covered by Cypress | `All Items` edit/delete surfaces opened and were cancelled. Destructive submit was not executed against shared demo data. Protected no-mutation submit behavior is covered by `UI-SCREEN-COLLECTIONS-030` and `ui.web/cypress/e2e/collections/collections-protected-all-items/spec.cy.ts`. |
| Members pagination | Validated | `All Items` members page 2 displayed later inventory records, then page 1 returned, while active context stayed `All Items`. Covered by `UI-SCREEN-COLLECTIONS-029`. |
| Filtering and sorting | Not manually re-proven in shared demo | CDP value injection changed input values but did not reliably trigger React filtering events in this automation pass, so no product conclusion was drawn from that attempt. Existing focused Cypress coverage remains authoritative for `UI-SCREEN-COLLECTIONS-006`, `017`, `024`, `031`. |
| View row navigation | Partially validated | `Store 1` row View navigated to `/inventory/`; the combined CDP evaluation was interrupted by the target navigation before returning structured post-click data. Existing Cypress coverage remains authoritative for `UI-SCREEN-COLLECTIONS-010`. |
| Empty/loading/error states | Traceability covered, not re-executed live | The shared demo profile had populated data and healthy runtime. Empty/filter states, cancellation no-mutation, profile isolation, and error-path coverage are represented in existing Collections specs/tests. |
| Persistence/data outcome | Covered by tests; mutation-heavy live execution intentionally limited | This audit did not create, rename, delete, or move persisted shared demo data. Persistence claims remain bound to Cypress/API evidence in `openspec/traceability.md`. |

## Element Inventory and Classification

| Element/surface | Classification | Notes |
| --- | --- | --- |
| Collections route header, tag icon, shell title | Covered and observed working | `UI-SCREEN-COLLECTIONS-009`; header showed Collections identity and tag icon. |
| Primary `New collection` action | Covered and observed working for dialog open/cancel | `UI-SCREEN-COLLECTIONS-003`, `022`, `023`, `025`. |
| Collections search input | Covered by spec/tests; live automation inconclusive | `UI-SCREEN-COLLECTIONS-006`, `024`; no product issue opened because synthetic CDP input did not provide reliable manual evidence. |
| Collections shared table | Covered and observed working | `UI-SCREEN-COLLECTIONS-001`; rows and management columns rendered. |
| Collection row selection | Covered and observed working | `UI-SCREEN-COLLECTIONS-002`, `016`, `027`; active context and members panel updated. |
| Row View actions | Covered; partially observed | Existing `UI-SCREEN-COLLECTIONS-010` Cypress coverage remains the full proof. |
| Row Edit actions and edit panel | Covered and observed for open/cancel | `UI-SCREEN-COLLECTIONS-004`, `025`, `026`, `030`. |
| Row Delete actions and confirmation dialog | Covered and observed for open/cancel | `UI-SCREEN-COLLECTIONS-005`, `025`, `030`. |
| `All Items` protected row | Covered; submit path not executed live | `UI-SCREEN-COLLECTIONS-030` guards no-mutation submit behavior. |
| Members panel and table | Covered and observed working | `UI-SCREEN-COLLECTIONS-016`, `017`, `018`, `027`, `029`. |
| Members search input | Covered by spec/tests; live automation inconclusive | `UI-SCREEN-COLLECTIONS-017`, `024`, `029`. |
| Members pagination | Covered and observed working | `UI-SCREEN-COLLECTIONS-029`; page 2 records rendered and page 1 returned. |
| Cross-route Inventory/Wishlist sync | Covered by specs/tests; not re-executed in this route-only audit | `UI-SCREEN-COLLECTIONS-012`, `013`, `014`, `015`. |
| Legacy in-route assignment controls | Retired by design | `UI-SCREEN-COLLECTIONS-019` is documented as retired/replaced. |
| Layout/overflow/long cells | Covered by specs/tests; not visually stress-tested beyond normal demo data | `UI-SCREEN-COLLECTIONS-020`, `021`. |

## Traceability Check

Existing OpenSpec and Cypress traceability already cover the audited surface:
- `openspec/specs/collections/ui-screen-collections/spec.md` defines `UI-SCREEN-COLLECTIONS-001` through `031`, with `019` documented as retired/replaced.
- `openspec/traceability.md` maps the active Collections requirements to the focused Cypress specs under `ui.web/cypress/e2e/collections/`.
- #1078 was validated done after restoring and validating the missing `011` through `021` coverage; this audit did not find a new untracked requirement gap.

## Follow-up Disposition

New issues opened: none.

Existing relevant artifacts:
- #1078 closed as validated done for focused Collections QA coverage.
- #537 and #534 remain historical implementation follow-ups referenced by #493; the current table/members surface now has OpenSpec and Cypress coverage.

Recommended next action:
- Merge this #493 exploration audit if accepted.
- Close #493 after merge/deploy evidence because the Collections exploration umbrella has now been reconciled with current OpenSpec/Cypress traceability.
- Continue route-ordered exploration with #495 Tasks / operational surfaces unless a higher-priority Cabinet issue is added.

Completeness label: Complete with explicit live-mutation limits. All reachable non-destructive scenarios were accounted for, mutation-heavy save/delete/move/submit proof remains delegated to existing focused Cypress tests to avoid modifying shared demo data during the exploratory audit.
