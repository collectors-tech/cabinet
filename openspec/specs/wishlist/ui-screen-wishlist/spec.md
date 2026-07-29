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

### Requirement UI-SCREEN-WISHLIST-006: Wishlist table SHALL support title sorting

Wishlist table rows view SHALL expose sortable `Title` column with deterministic ordering behavior.

#### Scenario: Sort wishlist by title

- **GIVEN** wishlist rows view is visible
- **WHEN** user clicks `Title` column sort control
- **THEN** wishlist entries MUST reorder by title with deterministic ascending/descending toggle behavior

### Requirement UI-SCREEN-WISHLIST-004: Wishlist screen SHALL expose compact icon header actions

Wishlist SHALL provide compact icon-only header actions for creating a wishlist entry, creating a collection, and importing wishlist entries.

#### Scenario: Wishlist compact header actions

- **GIVEN** user is on `/wishlist`
- **WHEN** user clicks the new-entry icon action
- **THEN** primary create-wishlist-entry flow MUST open
- **AND** header actions MUST expose accessible labels without visible text labels
- **AND** Wishlist MUST NOT render an adjacent `Create` menu

### Requirement UI-SCREEN-WISHLIST-021: Wishlist screen SHALL show a visible page header title

Wishlist SHALL render a visible page header title that follows the shared Cabinet
page-header pattern without replacing compact header actions.

#### Scenario: Wishlist page header title

- **GIVEN** user is on `/wishlist`
- **WHEN** the page header renders at desktop or mobile width
- **THEN** the header MUST show a visible `Wishlist` title
- **AND** the title MUST include the wishlist page icon
- **AND** the compact wishlist header actions MUST remain available

### Requirement UI-SCREEN-WISHLIST-005: Wishlist collection creation SHALL use the header modal

Wishlist SHALL support collection creation from the header collection icon and MUST NOT render inline collection creation controls inside the table toolbar.

#### Scenario: Create collection from Wishlist header

- **GIVEN** user is on `/wishlist`
- **WHEN** user clicks the create-collection icon action
- **THEN** a collection creation modal MUST open
- **AND** saving a valid collection MUST persist it for shared collection filters

#### Scenario: Blank collection modal submission shows validation

- **GIVEN** Wishlist collection creation modal is open
- **WHEN** user clicks `Save` with an empty collection name
- **THEN** the modal MUST remain open
- **AND** the screen MUST show visible required-field guidance
- **AND** only an explicit cancel/dismiss action MAY close the modal without a create result

### Requirement UI-SCREEN-WISHLIST-007: Wishlist rows SHALL use collection semantics and MUST NOT leak task seed labels

Wishlist rows/cards MUST be sourced from canonical `/api/items?status=wishlist` records with `/api/wishlist` metadata overlays and MUST NOT render generic task IDs or task taxonomy labels.

#### Scenario: Wishlist semantics in rows view

- **GIVEN** `/api/items?status=wishlist` returns canonical wishlist item records and `/api/wishlist` returns wishlist metadata overlays
- **WHEN** user opens wishlist rows view
- **THEN** rendered IDs MUST align to wishlist `item_id` values
- **AND** row titles MUST align to canonical item title/part number
- **AND** rows header semantics MUST render `Item ID`, `Title`, `Watch Status`, and `Target Priority`
- **AND** UI MUST NOT render generic task-template headers such as `Task` or task workflow labels such as `Backlog`

### Requirement UI-SCREEN-WISHLIST-008: Wishlist screen SHALL avoid stale planning summary controls

Wishlist screen SHALL remain table-first and MUST NOT render the old planning summary cards or persisted planning-focus controls.

#### Scenario: Stale planning controls are absent

- **GIVEN** wishlist route is loaded with wishlist metadata overlays
- **WHEN** the screen renders
- **THEN** the old `All planned`, `High priority`, `Below target`, and `Steady watch` summary cards MUST NOT appear
- **AND** `cabinet.wishlistPlanningFocus` MUST NOT be retained in local storage

### Requirement UI-SCREEN-WISHLIST-009: Wishlist rows SHALL capture purchase details without row-menu ownership hacks

Wishlist rows SHALL expose Purchased state and purchase details through dedicated row fields and purchase dialog controls, not through a `Mark owned` row action.

#### Scenario: Add purchase details

- **GIVEN** wishlist route shows a wanted item with wishlist entry metadata
- **WHEN** user clicks the row purchase action
- **THEN** the purchase dialog MUST open with sensible defaults
- **AND** saving MUST persist Purchased state, price paid, quantity, condition, purchase date, and URL
- **AND** Wishlist row actions MUST NOT include `Mark owned`

### Requirement UI-SCREEN-WISHLIST-020: Wishlist rows SHALL expose Purchased and Category workflow fields without a Delivered table column

Wishlist rows, cards, filters, forms, and detail surfaces SHALL use `Purchased` wording instead of `Owned`, SHALL preserve Category when wishlist items move into downstream purchase and inventory records, and SHALL NOT expose `Delivered` as a first-class Wishlist table column. Delivery/received reconciliation belongs in Purchases/Inventory flows or an explicit detail workflow, not the default Wishlist rows table.

#### Scenario: Edit purchase-to-delivery workflow fields

- **GIVEN** wishlist route shows a wanted item with wishlist entry metadata and canonical item category
- **WHEN** user edits Purchased, Delivered, and Category fields
- **THEN** the UI MUST persist Purchased and Delivered values through `/api/wishlist`
- **AND** the UI MUST persist Category on the canonical item record
- **AND** Delivered MUST either set Purchased or block the save with clear validation guidance
- **AND** downstream Purchases and Inventory views MUST be able to show the resulting records with wishlist provenance

#### Scenario: Hide Delivered from default rows table

- **GIVEN** wishlist route shows rows view
- **WHEN** the default Wishlist table renders
- **THEN** the table headers MUST include `Purchased`
- **AND** the table headers MUST NOT include `Delivered`

### Requirement UI-SCREEN-WISHLIST-016: Wishlist rows and cards SHALL render stable date context
Wishlist rows and cards SHALL render date context without implying that normal edits refreshed price data.

#### Scenario: Wishlist date context
- **GIVEN** Wishlist API metadata includes a wishlist entry creation timestamp
- **AND** pricing history includes a latest snapshot or trend date for the item
- **WHEN** the Wishlist rows or cards render
- **THEN** the UI MUST show `Date added` from the wishlist entry creation timestamp
- **AND** the UI MUST show `Updated` from the latest pricing history or refresh date
- **AND** missing legacy date context MUST render a non-misleading empty value

### Requirement UI-SCREEN-WISHLIST-017: Wishlist cost and quantity fields SHALL expose stable stepper controls

Wishlist rows view SHALL expose accessible decrement/increment controls around inline Cost and Quantity numeric fields while preserving direct keyboard entry.

#### Scenario: Edit cost and quantity with stepper controls

- **GIVEN** wishlist rows view is visible with wishlist metadata overlays
- **WHEN** user edits Cost or Quantity by typing numeric values or using adjacent decrement/increment controls
- **THEN** the row MUST persist the corresponding wishlist metadata update
- **AND** Cost MUST NOT decrement below zero
- **AND** Quantity MUST NOT decrement below its valid minimum
- **AND** native browser number spinner affordances MUST be hidden where supported
- **AND** the controls MUST remain fixed-width in the table layout

### Requirement UI-SCREEN-WISHLIST-018: Wishlist rows SHALL render compact deterministic thumbnails

Wishlist rows view SHALL render a compact thumbnail before each item title. When API thumbnail media is missing, the row MUST render a deterministic generated identicon-style fallback derived from the stable item identifier.

#### Scenario: Wishlist row thumbnails render without duplicating the title

- **GIVEN** wishlist rows view is loaded with canonical wishlist item records
- **WHEN** an item has no thumbnail media
- **THEN** the row MUST show a compact deterministic generated thumbnail before the title
- **AND** fallback thumbnails for distinct sample rows MUST be visually distinct and stable for the same item identifier
- **AND** the thumbnail MUST be decorative for assistive technology and MUST NOT duplicate the item title
- **AND** the dense wishlist table layout MUST keep the title and notes readable on desktop and mobile widths

### Requirement UI-SCREEN-WISHLIST-024: Wishlist cards SHALL render asset thumbnails in a compact responsive grid

Wishlist cards view SHALL render the best available wishlist item thumbnail when API media is available, render an intentional no-asset placeholder when media is missing, and keep desktop cards compact enough for browsing density.

#### Scenario: Wishlist card thumbnails and density

- **GIVEN** wishlist cards view is loaded with both media-backed and media-missing entries
- **WHEN** the card grid renders at desktop width
- **THEN** media-backed cards MUST show the item thumbnail
- **AND** media-missing cards MUST show an intentional no-asset placeholder
- **AND** the grid MUST support four columns where desktop viewport width allows
- **AND** card notes and long titles MUST be clamped so cards do not stretch into large empty panels

### Requirement UI-SCREEN-WISHLIST-019: Wishlist create SHALL persist title-only drafts with defaults

Wishlist create SHALL accept a draft with only `Title` populated, create the backing canonical wishlist item with generated part-number metadata and default planning values, then create the wishlist metadata entry without surfacing generic save failure copy.

#### Scenario: Create title-only wishlist entry

- **GIVEN** wishlist route is loaded and the create panel is open
- **WHEN** user enters only a title and clicks `Save changes`
- **THEN** Cabinet MUST create a canonical item with status `wishlist`, the submitted title, generated part number metadata, and default priority
- **AND** Cabinet MUST create the wishlist metadata entry linked to that item
- **AND** the created title MUST appear in the refreshed wishlist rows/cards
- **AND** the create panel MUST close after persistence succeeds
- **AND** the UI MUST NOT show generic `Wishlist save failed` copy for this valid title-only draft

### Requirement UI-SCREEN-WISHLIST-022: Wishlist edit panel navigation SHALL update in place

Wishlist edit side panel Previous/Next navigation SHALL keep the active side
panel open while replacing the form data and highlighted row with the newly
selected wishlist entry.

#### Scenario: Navigate wishlist edit panel entries

- **GIVEN** wishlist rows view is loaded with at least two visible wishlist entries
- **AND** the user opens the `Edit Wishlist Entry` side panel for one row
- **WHEN** the user clicks `Previous` or `Next` in the side panel
- **THEN** the same side panel instance MUST remain open without closing and reopening
- **AND** the form fields MUST update to the newly selected wishlist entry
- **AND** the highlighted wishlist row MUST move to the entry currently shown in the side panel
- **AND** existing close and save behavior MUST remain available

### Requirement UI-SCREEN-WISHLIST-023: Wishlist deletion SHALL soft-delete active rows before permanent removal

Wishlist default deletion SHALL hide active wishlist rows without immediately
hard-deleting the data, expose a Deleted filter/view for review and restore,
and require an explicit permanent-delete confirmation from the Deleted view.

#### Scenario: Soft delete, review, restore, and permanently delete wishlist rows

- **GIVEN** wishlist rows view is visible with an active wishlist entry
- **WHEN** user deletes the active row from the normal Wishlist view
- **THEN** Cabinet MUST soft-delete the wishlist entry and hide it from the default active list
- **AND** the deleted row MUST remain available from a `Deleted` filter or view
- **AND** the Deleted view MUST expose a restore/undo action for the soft-deleted row
- **AND** deleting the same row from the Deleted view MUST open a permanent-delete confirmation modal
- **AND** the modal MUST explain that the wishlist row and linked wishlist metadata/history will be removed
- **AND** the modal MUST explain that inventory records already created from the wishlist item will not be deleted
- **AND** Cabinet MUST NOT permanently delete the row until the user confirms the modal

## Use-Case IDs and E2E Mapping

| UC ID     | Flow                             | Expected Result                                                                                                        | E2E Mapping                                                                                                                                            |
| --------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| UC-WSH-01 | Filter wishlist and switch views | List updates and view mode persists                                                                                    | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-filter-view-toggle`                                                                                 |
| UC-WSH-02 | Open create/import               | Correct dialog/drawer opens                                                                                            | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-create-import`                                                                                      |
| UC-WSH-03 | Bulk select wishlist entries     | Bulk controls appear with stable selection state                                                                       | planned: `cypress/e2e/ui/wishlist.cy.ts` `wishlist-bulk-actions`                                                                                       |
| UC-WSH-04 | Sort wishlist by title           | Title sort control reorders rows deterministically                                                                     | planned: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `wishlist-title-sort`                                                             |
| UC-WSH-17 | Edit wishlist Cost and Quantity  | Inline numeric fields support keyboard entry plus accessible fixed-width stepper controls with lower-bound constraints | implemented: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `UI-SCREEN-WISHLIST-017 edits cost and quantity with stable stepper controls` |
| UC-WSH-18 | Show row thumbnails              | Rows render stable decorative thumbnails with deterministic fallback styling                                            | implemented: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `UI-SCREEN-WISHLIST-018 renders compact deterministic row thumbnails`        |
| UC-WSH-24 | Show compact card thumbnails     | Cards render media thumbnails or a no-asset placeholder in a compact responsive grid                                    | implemented: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `UI-SCREEN-WISHLIST-024 renders compact card thumbnails and placeholders`    |
| UC-WSH-19 | Create title-only wishlist entry | Title-only drafts persist with generated metadata and no generic failure copy                                           | implemented: `ui.web/cypress/e2e/wishlist/ui-screen-wishlist/spec.cy.ts` `UI-SCREEN-WISHLIST-019 creates a title-only wishlist entry`                  |
| UC-WSH-22 | Navigate edit panel entries      | Previous/Next updates the side-panel form in place and moves the active row highlight                                  | implemented: `ui.web/cypress/e2e/wishlist/wishlist-row-side-panel/spec.cy.ts` `opens a right-side edit panel on double click and navigates visible records` |
