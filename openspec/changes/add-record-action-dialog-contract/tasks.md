## 1. Contract

- [x] 1.1 Define the shared record action menu and CRUD/destructive dialog
  contract for #1938.
- [x] 1.2 Publish a surface/action matrix for current Cabinet tables before
  page migration begins.

## 2. Shared components

- [ ] 2.1 Add a shared record action menu component with capability-driven
  action definitions, stable ordering, accessible icon trigger, tooltips, and
  row-click isolation.
- [ ] 2.2 Add shared create/edit dialog primitives for title/description/icon,
  validation/server errors, dirty-state cancel protection, loading state,
  double-submit prevention, focus trap, initial focus, and focus return.
- [ ] 2.3 Add a shared destructive confirmation primitive that names the record,
  describes the consequence, and distinguishes archive/soft delete from
  permanent deletion.

## 3. Evidence

- [ ] 3.1 Add component tests for menu permissions/capabilities, omitted and
  disabled operations, keyboard/screen-reader affordances, loading, server
  errors, destructive confirmation, and focus return.
- [ ] 3.2 Add story/demo fixtures documenting the standard before dependent page
  migrations start.
- [ ] 3.3 Run strict OpenSpec validation, focused component tests, UI build, and
  `git diff --check`.
