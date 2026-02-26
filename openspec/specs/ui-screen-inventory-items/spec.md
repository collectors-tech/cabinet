## Purpose
Define Inventory Items screen behavior for high-volume item browsing, editing, and bulk operations.

## Requirements
### Requirement: Inventory Items SHALL support list/card browsing with consistent interactions
Inventory Items SHALL support row-details behavior, selection mode, and filter/sort workflows.

#### Scenario: Row details open
- **WHEN** user clicks non-interactive row area
- **THEN** item details drawer SHALL open for selected row

### Requirement: Inventory Items SHALL support deterministic state handling
Inventory Items SHALL support loading, empty, error, and ready states.

#### Scenario: Inventory empty state
- **WHEN** no items match active filters
- **THEN** screen SHALL render empty state with add/search guidance

#### Scenario: Inventory error state
- **WHEN** items API fails
- **THEN** screen SHALL render inline retry/error without fatal route failure

### Requirement: Inventory Items SHALL support sample and bulk data usage
The screen SHALL remain usable with both starter and stress datasets.

#### Scenario: Bulk dataset interaction
- **WHEN** user filters/sorts large item lists
- **THEN** interactions SHALL remain responsive and deterministic

## Acceptance Criteria
- UC IDs cover browse, edit, selection, and state transitions.
- E2E mapping includes non-500 regression behavior.

## Success Criteria
- Inventory remains stable for empty/seeded/bulk datasets.
- Row, checkbox, and control-click behaviors are predictable.

## Data Profiles
- Sample: 100 items, 200 instances, 300 photos
- Bulk: 10,000 items, 50,000 instances

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-INV-01 | Open inventory ready state | List/table data renders | planned: `cypress/e2e/ui/inventory.cy.ts` `inventory-ready` |
| UC-INV-02 | Empty filtered results | Empty state appears | existing/planned: `cypress/e2e/regression/inventory-empty-non500.cy.ts` `inventory-empty-non500` |
| UC-INV-03 | API failure on load | Error state + retry, no 500 route | existing/planned: `cypress/e2e/regression/inventory-non500.cy.ts` `inventory-load-error-non500` |
| UC-INV-04 | Row click open details | Details drawer opens selected item | planned: `cypress/e2e/ui/inventory.cy.ts` `inventory-row-opens-details` |
| UC-INV-05 | Checkbox bulk select | Selection mode appears, no row-open side effect | planned: `cypress/e2e/ui/inventory.cy.ts` `inventory-bulk-checkbox-mode` |
