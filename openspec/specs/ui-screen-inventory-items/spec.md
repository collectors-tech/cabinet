## Purpose
Define Inventory Items screen behavior and item-management use cases.

## Requirements
### Requirement: Inventory Items screen SHALL support list/card browsing and CRUD entry points
The screen SHALL provide usable item browse/create/edit flows with filter and sort support.

#### Scenario: Use case - find and open item details
- **WHEN** user filters and selects an inventory row
- **THEN** details drawer SHALL open for selected item

### Requirement: Inventory Items screen SHALL enforce consistent row/selection interaction rules
Row click SHALL open details; checkbox selection SHALL drive bulk mode; interactive controls SHALL not trigger row open.

#### Scenario: Use case - bulk select without opening details
- **WHEN** user selects checkboxes for multiple rows
- **THEN** selection state SHALL update and details SHALL not auto-open
