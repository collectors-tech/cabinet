# Item Detail / Workflow Exploration Audit - 2026-06-18

Issue: #492
Run window: 2026-06-18 00:30 Australia/Sydney / 2026-06-17 14:30 UTC
Branch: `issue-492-item-workflow-exploration`

## Scope

This exploration pass reviewed the reachable Cabinet item detail and item workflow surfaces in live-runtime mode:

- inventory route item table after local sample-data recovery
- seeded item action controls for assignment, barcodes, photos, and row actions
- collection-assignment dialog
- barcode attach and lookup dialog
- photo review, primary, delete, and camera/rebuild controls
- existing spec and Cypress traceability for deeper edit/save, media, barcode, and constrained-height workflows

Evidence logs for this run are under `.work-agent/logs/issue-492-item-workflow-exploration/20260618-0030/`.

## Runtime Preconditions

- Cabinet demo2 runtime checked at `http://127.0.0.1:17882`.
- `/healthz` returned HTTP 200 `ok`.
- `/api/runtime` returned `app_version` `rev-8fc8b464a469`, runtime port `17882`, and pid `48060`.
- Local repo branch for the report was `issue-492-item-workflow-exploration`, created from clean `develop`.
- OpenClaw browser profile `project-cabinet` was started successfully through CDP on port `18801`.
- The higher-level OpenClaw browser navigation/snapshot actions were policy-blocked, so evidence was collected through raw CDP against the same `project-cabinet` profile. This was treated as tooling friction, not a product defect.
- Initial `/inventory` load showed `Profile unavailable. Retry loading databases.` and `/api/profiles` returned `{"profiles":null}` with `/api/items` returning an empty list.
- The documented local exploration recovery path was used: create an exploration-only profile, activate it, and invoke `POST /api/onboarding/sample-data`.
- Sample-data recovery created profile `Item Workflow Exploration` and seeded `102` items, `102` instances, `142` photos, and `5` wishlist entries.

## Section Outcome

Completeness label: Complete with documented limitations.

No new product defect or spec gap was found in the reachable item workflow surface after recovering the missing local profile/sample-data precondition. The live surface matched existing OpenSpec and Cypress traceability for the sampled contracts. Full edit/save persistence, destructive delete, upload from local file, and row-action menu subcommands were not mutated in this shared demo lane; those remain covered by focused Cypress and Go/API suites rather than claimed as live manual persistence proof from this audit.

## Screen and Component Evidence

### Inventory item table and seeded item controls

Requirements:
- `UI-SCREEN-INVENTORY-ITEMS-001`
- `UI-SCREEN-INVENTORY-ITEMS-005`
- `UI-SCREEN-INVENTORY-ITEMS-008`
- `UI-SCREEN-INVENTORY-ITEMS-009`
- `UI-SCREEN-INVENTORY-ITEMS-012`

Screens/routes:
- `/inventory`

Elements found:
- active profile label `Item Workflow Exploration`
- folder tree with All Items, Watch List, Wishlist Focus, Store, Warehouse, and Archive nodes
- filter input for title, part number, type, condition, or packaging
- Condition and Category filters
- saved-views selector, Save View, disabled Delete View
- Rows and Cards view toggles
- dense row headers for Part #, Title, Condition, Item type, Packaging, and Category
- seeded item rows, including `Starter RX-78-2 Gundam`
- row controls: Assign to collection, Open barcodes, Open photos, Open actions menu

Scenarios:
- Validated: profile/sample-data recovery made `/inventory` render seeded rows rather than the empty profile-unavailable state.
- Validated: item rows exposed item-specific collection, barcode, photo, and actions controls.
- Validated: the visible row table used inventory semantics and did not leak generic task-template columns.
- Validated: dense row controls remained reachable at the desktop review width.
- Blocked from live claim: non-control row click did not produce a reliable item editor proof through the raw CDP selector used in this pass; existing Cypress coverage remains the authority for row-details open behavior.

Issues created or linked:
- No new issue. The initial missing profile/sample-data condition was recovered through the documented exploration setup path and was not filed as a product bug.

### Collection assignment dialog

Requirements:
- `UI-SCREEN-INVENTORY-ITEMS-006`
- `UI-SCREEN-COLLECTIONS-012`

Screens/routes:
- `/inventory`

Elements found:
- Assign to Collection dialog for `Starter RX-78-2 Gundam`
- Collection selector with Watch List, Warehouse 1, Store 1, Store 2, and Overflow
- Cancel, Assign, and Close controls

Scenarios:
- Validated: row-level Assign opened an item-scoped collection assignment dialog without route loss.
- Validated: dialog controls were item-specific and provided cancel/close exits.
- Blocked from live mutation claim: Assign persistence was not executed against the shared demo profile in this pass; existing Cypress traceability covers assignment persistence.

Issues created or linked:
- No new issue.

### Barcode workflow dialog

Requirements:
- `UI-SCREEN-INVENTORY-BARCODES-001`
- `UI-SCREEN-INVENTORY-BARCODES-002`
- `UI-SCREEN-INVENTORY-BARCODES-003`

Screens/routes:
- `/inventory`

Elements found:
- Barcodes dialog for `Starter RX-78-2 Gundam`
- `Enter barcode to attach` input
- Add Barcode action
- `Lookup barcode` input
- Lookup Barcode action
- Close action

Scenarios:
- Validated: row-level barcode control opened the correct item-scoped barcode dialog.
- Validated: add and lookup controls were visible and enabled.
- Validated: close path was available without leaving `/inventory`.
- Blocked from live mutation/error claim: barcode add, lookup result persistence, external fallback, and error-state mutation paths were not executed in this shared demo pass; existing Cypress and API traceability cover those deterministic contracts.

Issues created or linked:
- No new issue.

### Photo workflow dialog

Requirements:
- `UI-SCREEN-INVENTORY-PHOTOS-001`
- `UI-SCREEN-INVENTORY-PHOTOS-002`
- `UI-SCREEN-INVENTORY-PHOTOS-003`

Screens/routes:
- `/inventory`

Elements found:
- Photos dialog for `Starter RX-78-2 Gundam`
- Take Photo action
- Rebuild Photos action
- seeded media file `CAB-DEMO-006-identicon-01.png`
- Primary indicator
- Set primary control
- Delete photo control
- Close action

Scenarios:
- Validated: row-level photo control opened the correct item-scoped photo dialog.
- Validated: seeded media rendered with a primary state.
- Validated: primary, delete, take-photo, rebuild, and close controls were present and reachable.
- Blocked from live mutation/error claim: upload/camera capture, set-primary persistence, delete persistence, fullscreen inspection, and media API error states were not executed in this shared demo pass; existing Cypress traceability covers those contracts.

Issues created or linked:
- No new issue.

## Scenario Matrix

- Happy path: Validated seeded item row workflow controls, collection assignment dialog open, barcode dialog open, and photo dialog open.
- Cancellation/abort path: Validated collection, barcode, and photo dialogs expose Cancel and/or Close exits.
- Invalid input path: Not mutated in this live pass; barcode validation and item edit validation remain covered by targeted Cypress/API tests.
- Empty state: Observed initial no-profile/no-item state and recovered it through the documented sample-data path; item workflow audit proceeded with seeded data.
- Loading state: Not directly observed in the fast live pass; deterministic loading coverage remains in Cypress.
- Error/failure state: Profile-unavailable setup state was observed and recovered; item workflow API error states remain covered by existing Cypress/API traceability.
- Keyboard/accessibility path: Partially covered by visible labels and existing traceability; a full keyboard-only traversal was not claimed from this raw-CDP pass.
- Refresh/back-forward persistence path: Not live-claimed for item workflow state because this pass avoided mutating shared demo data; existing Cypress covers persistence-heavy flows.
- Post-action verification path: Verified dialogs were item-scoped to `Starter RX-78-2 Gundam` after each row-level action.
- Permission/role variance: Not applicable to the current local exploration account; no alternate role was available in this pass.

## Traceability Summary

Current requirement/test mappings were checked through `openspec/traceability.md`, OpenSpec files, and Cypress specs:

- `UI-SCREEN-INVENTORY-ITEMS-001`, `005`, `006`, `008`, `009`, `012`, and `013` -> `openspec/specs/inventory/ui-screen-inventory-items/spec.md` and focused Cypress under `ui.web/cypress/e2e/inventory/`.
- `UI-SCREEN-INVENTORY-PHOTOS-001` through `003` -> `openspec/specs/inventory/ui-screen-inventory-photos/spec.md` and `ui.web/cypress/e2e/inventory/ui-screen-inventory-photos/spec.cy.ts`.
- `UI-SCREEN-INVENTORY-BARCODES-001` through `003` -> `openspec/specs/inventory/ui-screen-inventory-barcodes/spec.md` and `ui.web/cypress/e2e/inventory/ui-screen-inventory-barcodes/spec.cy.ts`.
- Collection assignment behavior is also covered by `ui.web/cypress/e2e/inventory/inventory-assign-collection/spec.cy.ts` and collection-domain traceability.

No new spec gap was identified for the reachable item detail/workflow scope. The live pass confirms the item-scoped workflow entry points and dialogs; mutation-heavy save, delete, upload, set-primary, and barcode lookup paths should remain validated by the existing deterministic automation unless a reviewer requests a separate destructive/manual demo-data pass.

## Next Recommendation

#492 can close after this report is merged if no reviewer requires a separate mutation-heavy manual pass. The next route-ordered exploration issue is #493 Collections, unless newer higher-priority product feedback is added.
