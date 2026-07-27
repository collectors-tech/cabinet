## 1. Contract

- [x] 1.1 Define the canonical authenticated action-placement contract for
  #1941.
- [x] 1.2 Publish the initial route/action-region matrix before page migration.

## 2. Shared Placement Model

- [x] 2.1 Add typed action-region helpers or fixtures for page header,
  table/list toolbar, row menu, bulk actions, and dialog footer expectations.
- [x] 2.2 Add responsive overflow rules for page-header actions without hiding
  the primary action from keyboard or screen-reader users.
- [x] 2.3 Document or implement the boundary between shell utilities and page
  actions.

## 3. Page Migration

- [x] 3.1 Move Reports Refresh/Export into the canonical page-header action
  region and remove duplicated page title/description content.
- [x] 3.2 Move Market Watch Create/Run actions into the canonical action
  regions while preserving form and toolbar behavior.
- [x] 3.3 Audit and adjust remaining authenticated pages so whole-page, toolbar,
  row, bulk, and dialog actions match the matrix.

## 4. Evidence

- [x] 4.1 Add focused Cypress coverage for representative action regions,
  action order, icons, accessible labels, disabled/loading states, and no
  duplicate controls.
- [x] 4.2 Add desktop and narrow-window coverage for header overflow and
  non-overlap.
- [x] 4.3 Add or update packaged shell contract checks for expected action
  regions by route.
- [ ] 4.4 Run focused action-placement tests, UI build, strict OpenSpec
  validation, and `git diff --check`.
