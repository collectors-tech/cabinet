# App Shell / Navigation Exploration Audit - 2026-06-17

Issue: #489
Run window: 2026-06-17 23:30 Australia/Sydney / 2026-06-17 13:30 UTC
Branch: `issue-489-app-shell-navigation`

## Scope

This exploration pass reviewed the authenticated Cabinet app shell and navigation section in live-runtime mode:

- sidebar route navigation and route title transitions
- shell workspace switcher for Navigation, Assistant, and Inbox
- header controls and assistant workspace activation
- navigation edit-mode inventory of reorder, visibility, and drag handles
- database/profile switcher presence and active demo profile labelling
- shell runtime metadata and current demo runtime preconditions

Evidence logs for this run are under `.work-agent/logs/issue-489-app-shell-navigation/20260617-2330/`.

## Runtime Preconditions

- Cabinet demo2 runtime checked at `http://127.0.0.1:17882`.
- `/healthz` returned HTTP 200 `ok`.
- `/api/runtime` returned `app_version` `rev-71043eccd440`, `runtime_port` `17882`, and pid `65700`.
- Local repo branch for the report was `issue-489-app-shell-navigation`, created from clean `develop`.
- OpenClaw browser profile `project-cabinet` was started successfully through CDP on port `18801`.
- The higher-level browser navigation API was policy-blocked for direct navigation, so evidence was collected through raw CDP against the same `project-cabinet` profile. This was treated as tooling friction, not a product defect.
- Local password sign-in succeeded for an exploration-only account and reached `/dashboard` with title `Cabinet - Home`.
- Active profile after sign-in was `Showcase DB` with visible `Showcase sample data` labelling.

## Section Outcome

Completeness label: Complete with documented limitations.

No new product defects or spec gaps were found in the reachable app shell/navigation surface. The live surface matched the existing OpenSpec and Cypress traceability for the sampled contracts. Profile creation/switching mutation paths and drag/drop reorder mutation were inventoried but not executed in this exploration pass because they are already covered by focused Cypress suites and this pass avoided changing the shared demo profile state.

## Screen and Component Evidence

### Authenticated shell baseline

Requirements:
- `UI-FOUNDATION-SHELL-NAVIGATION-001`
- `UI-FOUNDATION-SHELL-NAVIGATION-004`
- `UI-FOUNDATION-SHELL-NAVIGATION-005`
- `UI-FOUNDATION-SHELL-NAVIGATION-008`
- `UI-FOUNDATION-SHELL-NAVIGATION-010`
- `UI-FOUNDATION-SHELL-NAVIGATION-011`
- `UI-FOUNDATION-SHELL-NAVIGATION-012`

Screens/routes:
- `/dashboard`

Elements found:
- skip-to-main link
- active database/profile switcher showing `Showcase DB`
- workspace switcher controls: Nav, Assistant, Inbox
- primary navigation links: Dashboard, Inventory, Media, Collections, Wishlist, Discoveries, Market Watch, Integrations, Purchases, Chats, Users, Reports
- secondary navigation links: Settings, Storage, Help Center
- signed-in account/menu area
- runtime metadata: version `rev-71043eccd440`, build date `2026-06-17T13:02:39Z`
- header controls: Toggle Sidebar, Search/Ctrl+K, Home, Language, Toggle theme, account/avatar, Signal hub

Scenarios:
- Validated: signed-in shell rendered sidebar, header, active demo database context, and runtime metadata.
- Validated: primary and secondary nav groups were visible and scan-readable.
- Validated: Chats nav badge rendered as a trailing count affordance.
- Validated: shell header remained product/action focused and did not expose inline collection/profile context copy.
- Uncovered in this live pass: viewport-width and scroll ownership were not remeasured manually; existing Cypress traceability covers the fixed-nav and full-width layout contracts.

Issues created or linked:
- No new issue. Existing requirement/test traceability covers the observed behavior.

### Route transition navigation

Requirements:
- `UI-FOUNDATION-SHELL-NAVIGATION-013`
- `UI-ROUTE-COVERAGE-006`

Screens/routes:
- `/dashboard`
- `/inventory`
- `/collections`
- `/wishlist`
- `/integrations`
- `/chats`

Elements found:
- sidebar links for each route above
- centered route headers/page titles where the destination page supplies them
- active route document titles

Scenarios:
- Validated: Dashboard link landed on `/dashboard` with title `Cabinet - Home`.
- Validated: Inventory link landed on `/inventory` with title `Cabinet - Inventory`.
- Validated: Collections link landed on `/collections` with title `Cabinet - Collections`.
- Validated: Wishlist link landed on `/wishlist` with title `Cabinet - Wishlist`.
- Validated: Integrations link landed on `/integrations` with title `Cabinet - Integrations`.
- Validated: Chats link landed on `/chats` with title `Cabinet - Chats`.
- Validated: route transitions occurred through the shell without losing the active signed-in session.
- Not applicable: invalid-route recovery was covered by #488 public-entry exploration and existing `UI-ROUTE-COVERAGE-005` tests.

Issues created or linked:
- No new issue.

### Shell workspace switcher

Requirements:
- `UI-SHELL-WORKSPACES-001`
- `UI-SHELL-WORKSPACES-002`
- `UI-SHELL-WORKSPACES-003`
- `UI-SHELL-WORKSPACES-004`
- `UI-SHELL-WORKSPACES-005`
- `UI-SHELL-WORKSPACES-006`

Screens/routes:
- `/dashboard`
- `/chats`

Elements found:
- `shell-workspace-navigation`
- `shell-workspace-assistant`
- `shell-workspace-inbox`
- Assistant workspace panel and route-aware assistant content
- Inbox workspace panel with catch-up copy, Refresh, Open Chats, and Open Assistant Workspace controls
- header assistant/chat toggle

Scenarios:
- Validated: Assistant workspace activated from the left workspace switcher without changing the current route.
- Validated: Inbox workspace activated from the left workspace switcher and showed actionable empty-state controls.
- Validated: returning to Navigation restored the sidebar navigation region.
- Validated: `/chats` remained a distinct route with its own page title and did not replace the left-rail Assistant/Inbox semantics.
- Uncovered in this live pass: reload persistence was not claimed from live CDP evidence because the secondary persistence script returned no usable payload; existing Cypress traceability covers profile-scoped workspace persistence across reload and route changes.

Issues created or linked:
- No new issue. Persistence remains covered by `ui.web/cypress/e2e/general/ui-shell-workspaces/spec.cy.ts`.

### Navigation edit mode

Requirements:
- `UI-FOUNDATION-SHELL-NAVIGATION-002`
- `UI-FOUNDATION-SHELL-NAVIGATION-007`

Screens/routes:
- `/dashboard`

Elements found:
- nav edit toggle
- edit rows for Dashboard, Inventory, Media, Collections, Wishlist, Discoveries, Market Watch, Integrations, Purchases, Chats, Users, and Reports
- drag handles for every editable primary nav row
- move-up and move-down buttons for every editable primary nav row
- visibility toggles for every editable primary nav row

Scenarios:
- Validated: edit mode opened and exposed row-level reorder, drag, and visibility controls for the full primary navigation list.
- Validated: closing edit mode returned to the normal shell without route loss.
- Uncovered in this live pass: reorder/visibility persistence was intentionally not executed against the shared demo profile; existing Cypress traceability covers live reorder order, drag insertion feedback, visibility persistence, and reload behavior.

Issues created or linked:
- No new issue.

### Database/profile switcher

Requirements:
- `UI-FOUNDATION-SHELL-NAVIGATION-005`
- `UI-FOUNDATION-SHELL-NAVIGATION-008`
- `UI-FOUNDATION-SHELL-NAVIGATION-009`
- `UI-FOUNDATION-SHELL-NAVIGATION-010`
- `UI-FOUNDATION-SHELL-NAVIGATION-014`
- `UI-FOUNDATION-SHELL-NAVIGATION-015`
- `UI-FOUNDATION-SHELL-NAVIGATION-016`
- `PROFILES-004`
- `PROFILES-005`
- `PROFILES-006`

Screens/routes:
- `/dashboard`

Elements found:
- top database/profile switcher trigger
- active profile name `Showcase DB`
- visible plan/status text `Showcase sample data`

Scenarios:
- Validated: the shell top area showed database/profile context only, not collection-management controls.
- Validated: the active demo profile was visibly distinguished as sample data.
- Uncovered in this live pass: profile creation, profile switching, and load-failure recovery were not executed to avoid mutating the shared demo profile state; existing Cypress suites cover those contracts.

Issues created or linked:
- No new issue.

## Scenario Matrix

- Happy path: Validated route navigation from Dashboard to Inventory, Collections, Wishlist, Integrations, Chats, and back to Dashboard.
- Cancellation/abort path: Validated nav edit mode can open and close without a route change or visible shell loss.
- Invalid input path: Not applicable for passive shell navigation; profile-switcher create/error input is covered by existing Cypress.
- Empty state: Validated Inbox workspace empty state shows Refresh, Open Chats, and Open Assistant Workspace actions.
- Loading state: Not directly observed in this fast live pass; covered by route and profile Cypress contracts where applicable.
- Error/failure state: Not directly observed in this pass; active profile load failure and route errors are covered by existing traceability.
- Keyboard/accessibility path: Partially covered by visible header search shortcut (`Ctrl+K`), skip-to-main, labeled workspace controls, and existing Cypress accessibility/keyboard suites. A full keyboard-only traversal was not claimed from this pass.
- Refresh/back-forward persistence path: Not live-claimed because the persistence CDP helper returned no usable payload; existing Cypress covers workspace and nav preference persistence.
- Post-action verification path: Validated route URL/title state after each sidebar navigation action; validated edit-mode close restored normal shell.
- Permission/role variance: Not applicable to the current local exploration account; no alternate role was available in this pass.

## Traceability Summary

Current requirement/test mappings were checked through `openspec/traceability.md`, OpenSpec files, and Cypress specs:

- `UI-FOUNDATION-SHELL-NAVIGATION-001` through `UI-FOUNDATION-SHELL-NAVIGATION-016` -> `openspec/specs/general/ui-foundation-shell-navigation/spec.md` and focused Cypress under `ui.web/cypress/e2e/general/`.
- `UI-SHELL-WORKSPACES-001` through `UI-SHELL-WORKSPACES-006` -> `openspec/specs/general/ui-shell-workspaces/spec.md` and `ui.web/cypress/e2e/general/ui-shell-workspaces/spec.cy.ts`.
- `UI-ROUTE-COVERAGE-005` / `UI-ROUTE-COVERAGE-006` -> `openspec/specs/general/ui-route-coverage/spec.md` and `ui.web/cypress/e2e/general/route-guards-errors/spec.cy.ts`.
- `PROFILES-004` through `PROFILES-006` -> `openspec/specs/users/profiles/spec.md` and profile-switcher Cypress suites.

No new spec gap was identified for the reachable app shell/navigation scope. Persistence and mutation-heavy profile/nav-edit scenarios are covered by existing Cypress and were not repeated manually against the shared demo profile during this exploration pass.

## Next Recommendation

#489 can close after this report is merged if no reviewer requires a separate mutation-heavy manual pass. The next route-ordered exploration issue is #490 Dashboard / home, unless newer higher-priority product feedback is added.
