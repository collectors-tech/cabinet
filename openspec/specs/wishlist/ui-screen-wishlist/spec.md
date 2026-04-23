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

### Requirement UI-SCREEN-WISHLIST-006: Wishlist table SHALL support title sorting
Wishlist table rows view SHALL expose sortable `Title` column with deterministic ordering behavior.

#### Scenario: Sort wishlist by title
- **GIVEN** wishlist rows view is visible
- **WHEN** user clicks `Title` column sort control
- **THEN** wishlist entries MUST reorder by title with deterministic ascending/descending toggle behavior

#### Scenario: Select multiple wishlist entries
- **GIVEN** wishlist rows/cards are visible
- **WHEN** user selects multiple entries
- **THEN** bulk action controls MUST appear and selected state MUST remain consistent through pagination changes

### Requirement UI-SCREEN-WISHLIST-004: Wishlist screen SHALL expose dedicated New action and adjacent Create menu
Wishlist SHALL provide a dedicated `New` button for primary wishlist entry creation and an adjacent `Create` menu for quick create actions.

#### Scenario: Wishlist New + Create menu
- **GIVEN** user is on `/wishlist`
- **WHEN** user clicks `New`
- **THEN** primary create-wishlist-entry flow MUST open
- **WHEN** user clicks adjacent `Create` menu
- **THEN** menu MUST show quick-create actions relevant to wishlist context

### Requirement UI-SCREEN-WISHLIST-005: Wishlist detail collection picker SHALL support inline quick-create
Wishlist detail collection picker MUST support `+ New Collection` inline create.

#### Scenario: Quick-create collection while assigning wishlist entry
- **GIVEN** user edits wishlist entry and opens collection picker
- **WHEN** user creates a new collection from picker
- **THEN** collection MUST be created and selected without leaving wishlist edit flow

#### Scenario: Blank inline collection submission shows validation
- **GIVEN** wishlist inline collection create state is open
- **WHEN** user clicks `Save` with an empty `Collection name`
- **THEN** the inline create state MUST remain open
- **AND** the screen MUST show visible required-field guidance
- **AND** only an explicit cancel/dismiss action MAY close the inline create state without a create result

### Requirement UI-SCREEN-WISHLIST-007: Wishlist rows SHALL use collection semantics and MUST NOT leak task seed labels
Wishlist rows/cards MUST be sourced from canonical `/api/items?status=wishlist` records with `/api/wishlist` metadata overlays and MUST NOT render generic task IDs or task taxonomy labels.

#### Scenario: Wishlist semantics in rows view
- **GIVEN** `/api/items?status=wishlist` returns canonical wishlist item records and `/api/wishlist` returns wishlist metadata overlays
- **WHEN** user opens wishlist rows view
- **THEN** rendered IDs MUST align to wishlist `item_id` values
- **AND** row titles MUST align to canonical item title/part number
- **AND** rows header semantics MUST render `Item ID`, `Title`, `Watch Status`, and `Target Priority`
- **AND** UI MUST NOT render generic task-template headers such as `Task` or task workflow labels such as `Backlog`

### Requirement UI-SCREEN-WISHLIST-008: Wishlist screen SHALL support planning focus controls
Wishlist screen SHALL show planning metadata and persist the selected planning focus for the wishlist route.

#### Scenario: Persist wishlist planning focus
- **GIVEN** wishlist route is loaded with wishlist metadata overlays
- **WHEN** user selects a planning focus
- **THEN** the selected focus MUST be visually active
- **AND** the selected focus MUST persist across refresh and route return

### Requirement UI-SCREEN-WISHLIST-009: Wishlist screen SHALL explicitly move acquired items to Inventory
Wishlist row actions SHALL expose a deliberate `Mark owned` action that calls the explicit wishlist conversion API.

#### Scenario: Mark wishlist row owned
- **GIVEN** wishlist route shows a wanted item with wishlist entry metadata
- **WHEN** user selects `Mark owned` from the row action menu
- **THEN** UI MUST submit `POST /api/wishlist/convert-owned` with the wishlist entry `id`
- **AND** the item MUST disappear from Wishlist after refresh
- **AND** the item MUST be visible in Inventory

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-WSH-01 | Filter wishlist and switch views | List updates and view mode persists | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-filter-view-toggle` |
| UC-WSH-02 | Open create/import | Correct dialog/drawer opens | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-create-import` |
| UC-WSH-03 | Bulk select wishlist entries | Bulk controls appear with stable selection state | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-bulk-actions` |
| UC-WSH-04 | Sort wishlist by title | Title sort control reorders rows deterministically | planned: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `wishlist-title-sort` |
