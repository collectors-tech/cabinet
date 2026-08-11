# Row Action and Dialog Migration Matrix

Issue: #1939
Baseline contract: #1938, `openspec/changes/add-record-action-dialog-contract/`
Run timestamp: 2026-08-12 00:14 Australia/Sydney / 2026-08-11 14:14 UTC

## Purpose

This matrix converts the #1938 shared row-action and dialog contract into the
#1939 page-migration checklist. It records the current page surfaces, the
truthful row operations each surface can expose, the dialog/API touchpoints that
must be preserved, and the focused verification target for each migration.

This is an implementation planning artifact for dependent page slices. It does
not claim that every page already uses the shared component.

## Migration Rules

- Put exactly one shared record action trigger in the final actionable row
  column for each table-like record surface.
- Use capability-driven actions; omit unsupported operations unless a disabled
  item has a truthful short reason.
- Keep page actions in headers, table/view/filter controls in toolbars, and
  bulk operations outside row menus.
- Preserve row single-click, double-click, checkbox, card, and detail-panel
  semantics when adding the shared trigger.
- Route create/edit and destructive confirmation flows through the shared
  dialog contracts when the page opens a row-scoped mutation.
- Verify persistence by API-visible state, refreshed table state, or downstream
  surface state after every save, delete, archive, restore, handoff, or queue
  action.

## Surface Matrix

| Surface | Current contract state | Supported row operations | Dialog/API touchpoints | #1939 migration gap | Verification target |
| --- | --- | --- | --- | --- | --- |
| Inventory items | Existing task/inventory rows use a kebab menu for edit/delete, with restore separated for deleted records. | Open/detail, Edit, Delete, Restore, Permanent delete where record state permits; duplicate/favorite remain unavailable unless implemented truthfully. | Task/inventory mutation APIs, item editor modal, delete confirmation, media attach dialogs. | Replace page-specific action definitions with shared `RecordActionMenu`; move restore/permanent-delete capabilities into the same final-column trigger while preserving checkbox and double-click guards. | `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts`; `ui.web/cypress/e2e/inventory/inventory-context-menu-actions/spec.cy.ts`; scoped Go item/task API persistence tests. |
| Wishlist | Wishlist rows use the task row menu for edit/delete or permanent delete, with restore separated for deleted entries. | Open/detail, Edit, Delete, Restore, Permanent delete; purchase state remains a dedicated field/dialog, not a row action. | Wishlist task APIs, wishlist edit drawer, delete confirmation, purchase detail dialog. | Adopt the shared action definitions and confirm restore/permanent-delete are capability-driven without reintroducing `Mark owned` as a row menu item. | `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts`; `ui.web/cypress/e2e/wishlist/wishlist-row-side-panel/spec.cy.ts`; wishlist API persistence tests. |
| Users | Users table uses a kebab menu; protected owner rows omit unsafe delete actions. | View/details, Edit, Delete where role/capability allows; resend invite remains outside the generic row menu unless exposed as a user capability. | Users API, Add/Edit User dialog, delete confirmation, invite/resend flows. | Replace feature-local row action copy with shared menu definitions and explicit capability omission/disabled reason for protected users. | `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts`; `internal/app/users_api_test.go`. |
| Collections | Collections row cells still expose separate icon buttons for View, Edit, and Delete. | View/Open, Edit, Delete with destination handling for protected/all-items constraints. | Collections API, create/edit collection dialog, delete collection dialog with destination selector, side panel route state. | Consolidate the icon cluster into one shared final-column trigger and route delete copy/focus through shared destructive confirmation while keeping destination selection. | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts`; `ui.web/cypress/e2e/collections/collections-row-side-panel/spec.cy.ts`; collection API tests. |
| Media | Media rows/cards expose Open, Analyze, Assign, Archive, and metadata edit controls. | Open/detail, Edit metadata, Analyze, Assign, Archive/Restore where asset state supports it. | Media API, add/edit metadata dialog, analyze dialog, assign dialog, archive state mutation. | Model analyze/assign/archive availability as row capabilities, add shared destructive/state-change copy for archive, and ensure card actions do not drift from row actions. | `ui.web/cypress/e2e/media/ui-screen-media/spec.cy.ts`; media API/storage tests. |
| Purchases | Purchase workflows rely on detail-pane and inbox action buttons rather than one compact row menu. | Open/detail, Review, Reconcile, Link/Unlink package, queue follow-up, and line-item actions where the target record supports them. | Purchases, forwarding package, package-link, landed-cost, and queue-action APIs; draft/review/reconcile dialogs. | Separate order-level and line-item action capabilities, name the target type in dialogs, and avoid moving bulk or planner actions into row menus. | `ui.web/cypress/e2e/purchases/purchase-inbox/spec.cy.ts`; `internal/app/forwarder_package_api_test.go`; purchase API tests. |
| Market Watch | Query rows expose multiple inline buttons for run, pause/resume, edit, inspect output, and delete. | Inspect/Open output, Run now, Pause/Resume, Edit, Delete. | Market Watch query/run APIs, query create/edit dialog, output detail panel, delete/state mutation APIs. | Convert the inline button cluster into shared menu items ordered by view/edit/state/delete semantics; add confirmation or truthful copy for delete and state changes. | `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts`; scanner/Market Watch API tests. |
| Discoveries | Discovery rows expose several icon actions for review, restore, ignore/archive, promote, purchase follow-up, and inventory handoff. | Review source, Restore for review, Ignore/Archive, Promote to Wishlist, Purchase follow-up, Inventory handoff when evidence supports it. | Discoveries/scanner APIs, wishlist promotion, purchase and inventory handoff persistence. | Map each handoff/state transition to capabilities with omitted unsupported paths and confirm archive/restore labels remain domain-specific. | `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts`; scanner API tests. |
| Integrations | Configured provider rows expose Edit or Connect while row click opens details and double-click opens edit. | View/Open details, Edit/Connect, Test, Repair, Disable/Disconnect only where registry capabilities and provider state allow. | Provider registry/config APIs, provider detail modal, schema-driven setup dialog, provider health/test endpoints. | Use shared trigger without breaking row click/double-click separation; only promote manifest actions into row menu after capability mapping is explicit. | `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts`; `ui.web/cypress/e2e/integrations/ui-screen-integrations-schema-form/spec.cy.ts`; provider registry/API tests. |
| Settings skills | Skills rows expose enable/disable and details controls. | Details, Enable, Disable, Import/Remove only where local skill state supports the operation. | Agent skills registry/import/apply APIs, details sheet, import panel. | Treat skills as record-like rows only inside the skills list; keep global settings forms outside row-menu migration. | `ui.web/cypress/e2e/settings/agent-skills/spec.cy.ts`; `internal/app/agent_skills_api_test.go`. |
| Settings storage | Backup rows expose download/restore style controls. | Download, Restore, Delete backup only if supported by the storage service. | Storage backup/export/restore APIs and restore confirmation dialog. | Keep export/create backup in page/header controls; migrate repeated backup-row actions to shared menu and preserve restore confirmation copy. | `ui.web/cypress/e2e/settings/storage/spec.cy.ts`; `ui.web/cypress/e2e/settings/storage-export/spec.cy.ts`; storage API tests. |
| Settings categories | Category rows/lists expose inline remove controls before save. | Remove pending category changes; restore/cancel via form state rather than row API. | Settings/category form state and save API. | Do not force a table menu unless categories are rendered as repeated actionable records; keep inline removal if it remains form-local before a save. | `ui.web/cypress/e2e/settings/categories/spec.cy.ts`; settings API tests. |

## Slice Order

1. Migrate Collections first because it has the clearest duplicate inline row
   action cluster and existing row-side-panel coverage.
2. Migrate Media next so archive/removal semantics and card/table parity are
   made explicit before broader destructive flows.
3. Migrate Market Watch and Discoveries after capability mapping because they
   have the densest handoff/state-change action sets.
4. Reconcile Inventory, Wishlist, and Users after the shared menu adoption path
   is proven so existing compliant behavior remains intact.
5. Treat Purchases and Settings as per-surface follow-ups because their row-like
   actions overlap with workflow/detail/form contexts.

## Acceptance Mapping

- #1939 matrix acceptance is satisfied by this file plus
  `openspec/traceability.md` linking the issue to the shared contract and
  planned verification targets.
- Per-page migration acceptance remains open until each surface replaces
  duplicate action implementations, passes the named validation target, and
  posts persistence evidence.
- The #1938 shared component contract remains the implementation source of truth
  for menu ordering, focus return, row-click isolation, dialog validation, busy
  state, server errors, and destructive confirmation copy.
