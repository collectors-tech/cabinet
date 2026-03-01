## Purpose
Define Wishlist screen behavior built on shared task/collection table workflows.

## Requirements
### Requirement UI-SCREEN-WISHLIST-001: Wishlist screen SHALL support table filters and row/card view toggle
Wishlist screen SHALL support status/priority filtering and rows/cards view mode toggle.

#### Scenario: Toggle wishlist view mode
- **GIVEN** wishlist route is loaded
- **WHEN** user switches between `Rows` and `Cards`
- **THEN** selected view mode MUST be applied and persisted for wishlist route context

### Requirement UI-SCREEN-WISHLIST-002: Wishlist screen SHALL support create and import entry workflows
Wishlist screen SHALL expose primary actions for create and import dialogs.

#### Scenario: Open wishlist create/import actions
- **GIVEN** wishlist route is loaded
- **WHEN** user clicks `Create` or `Import`
- **THEN** corresponding drawer/dialog MUST open and allow user action completion or cancel

### Requirement UI-SCREEN-WISHLIST-003: Wishlist screen SHALL support selection and bulk action affordances
Wishlist screen SHALL support row/card selection and bulk action controls.

#### Scenario: Select multiple wishlist entries
- **GIVEN** wishlist rows/cards are visible
- **WHEN** user selects multiple entries
- **THEN** bulk action controls MUST appear and selected state MUST remain consistent through pagination changes

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-WSH-01 | Filter wishlist and switch views | List updates and view mode persists | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-filter-view-toggle` |
| UC-WSH-02 | Open create/import | Correct dialog/drawer opens | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-create-import` |
| UC-WSH-03 | Bulk select wishlist entries | Bulk controls appear with stable selection state | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-bulk-actions` |
