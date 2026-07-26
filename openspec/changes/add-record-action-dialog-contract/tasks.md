## 1. Contract

- [x] 1.1 Define the shared record action menu and CRUD/destructive dialog
  contract for #1938.
- [x] 1.2 Publish a surface/action matrix for current Cabinet tables before
  page migration begins.

## 2. Shared components

- [x] 2.1 Add a shared record action menu component with capability-driven
  action definitions, stable ordering, accessible icon trigger, tooltips, and
  row-click isolation.
- [x] 2.2 Add shared create/edit dialog primitives for title/description/icon,
  validation/server errors, dirty-state cancel protection, loading state,
  double-submit prevention, focus trap, initial focus, and focus return.
- [x] 2.3 Add a shared destructive confirmation primitive that names the record,
  describes the consequence, and distinguishes archive/soft delete from
  permanent deletion.

## 3. Evidence

- [x] 3.1 Add component tests for menu permissions/capabilities, omitted and
  disabled operations, keyboard/screen-reader affordances, loading, server
  errors, destructive confirmation, and focus return.
- [x] 3.2 Add story/demo fixtures documenting the standard before dependent page
  migrations start.
- [x] 3.3 Run strict OpenSpec validation, focused component tests, UI build, and
  `git diff --check`.

Evidence note: `ui.web/cypress/component/data-table/record-edit-dialog.cy.tsx`
now covers the shared create/edit dialog title/description/icon contract,
validation and server errors, dirty cancel protection, loading/double-submit
prevention, initial focus, and focus return. Task 3.1 remains open until the
destructive confirmation primitive and keyboard/screen-reader matrix coverage
are also in place.

Evidence note: `ui.web/cypress/component/data-table/record-destructive-confirm-dialog.cy.tsx`
now covers the shared destructive confirmation primitive naming the record,
distinguishing archive/soft-delete/permanent-delete consequences, preventing
duplicate destructive submissions, and returning focus after close.

Evidence note: `ui.web/src/components/data-table/record-action-contract-demo.tsx`
documents the standard record action menu, create/edit dialog, and destructive
confirmation fixtures for dependent page migrations. Its Cypress component
coverage lives in `ui.web/cypress/component/data-table/record-action-contract-demo.cy.tsx`.

Evidence note: the full focused component matrix passed with 6/6 tests across
`record-action-menu.cy.tsx`, `record-edit-dialog.cy.tsx`,
`record-destructive-confirm-dialog.cy.tsx`, and
`record-action-contract-demo.cy.tsx`. `npm run build`, strict OpenSpec
validation, and `git diff --check` also passed before PR handoff.
