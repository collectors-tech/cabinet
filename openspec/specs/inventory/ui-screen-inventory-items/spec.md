## Purpose
Define Inventory Items screen behavior for high-volume item browsing, editing, and bulk operations.

## Requirements
### Requirement UI-SCREEN-INVENTORY-ITEMS-001: Inventory Items SHALL support list/card browsing with consistent interactions
Inventory Items SHALL support row-details behavior, selection mode, and filter/sort workflows.

#### Scenario: Row details open
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user clicks non-interactive row area
- **THEN** item details drawer SHALL open for selected row

#### Scenario: Checkbox bulk selection mode
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user selects one or more inventory rows using row checkboxes or the select-all checkbox
- **THEN** the bulk actions toolbar SHALL appear without opening the item editor
- **AND** clearing selection by keyboard or clear action SHALL remove the toolbar and selected row state

### Requirement UI-SCREEN-INVENTORY-ITEMS-002: Inventory Items SHALL support deterministic state handling
Inventory Items SHALL support loading, empty, error, and ready states.

#### Scenario: Inventory empty state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** no items match active filters
- **THEN** screen SHALL render empty state with add/search guidance

#### Scenario: Inventory error state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** items API fails
- **THEN** screen SHALL render inline retry/error without fatal route failure

### Requirement UI-SCREEN-INVENTORY-ITEMS-003: Inventory Items SHALL support sample and bulk data usage
The screen SHALL remain usable with both starter and stress datasets.

#### Scenario: Bulk dataset interaction
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user filters/sorts large item lists
- **THEN** interactions SHALL remain responsive and deterministic

### Requirement UI-SCREEN-INVENTORY-ITEMS-004: Inventory Collection layout SHALL keep controls compact and non-duplicated
Inventory Collection layout SHALL avoid duplicate control strips and keep summary context inside the Collection Browser header area.

### Requirement UI-SCREEN-INVENTORY-ITEMS-005: Inventory screen SHALL expose dedicated New action and adjacent Create menu
Inventory SHALL provide a dedicated `New` button for primary inventory entry creation and an adjacent `Create` menu for quick create actions.

#### Scenario: Inventory New + Create menu
- **GIVEN** user is on `/inventory`
- **WHEN** user clicks `New`
- **THEN** primary create-item flow MUST open
- **WHEN** user clicks adjacent `Create` menu
- **THEN** menu MUST show quick-create actions relevant to inventory context

### Requirement UI-SCREEN-INVENTORY-ITEMS-006: Inventory detail collection picker SHALL support inline quick-create
Inventory item details collection picker MUST support `+ New Collection` inline create.

### Requirement UI-SCREEN-INVENTORY-ITEMS-007: Inventory toolbar SHALL expose explicit create and folder actions
Inventory toolbar SHALL expose primary create-item and folder-creation actions (`Add Item`, `Add Folder`) with deterministic behavior.

#### Scenario: Add Item action
- **GIVEN** user is on `/inventory`
- **WHEN** user clicks `Add Item`
- **THEN** create-item workflow MUST open without route crash

#### Scenario: Add Folder action
- **GIVEN** user is on `/inventory`
- **WHEN** user clicks `Add Folder`
- **THEN** folder creation workflow MUST open and add folder entry to collection browser on success

### Requirement UI-SCREEN-INVENTORY-ITEMS-008: Inventory browser controls SHALL support filter/sort/view switching
Inventory browser controls SHALL expose `Status`, `Priority`, `View`, and list/card toggles with deterministic state changes.

#### Scenario: Filter and view controls available
- **GIVEN** user is on `/inventory`
- **WHEN** collection browser renders
- **THEN** `Status`, `Priority`, and `View` controls MUST be available and operable
- **AND** `Rows`/`Cards` mode toggles MUST switch presentation consistently

### Requirement UI-SCREEN-INVENTORY-ITEMS-009: Inventory rows SHALL use inventory item semantics and MUST NOT leak task-template headers
Inventory rows view MUST render Cabinet inventory item semantics instead of generic task-template columns.

#### Scenario: Inventory semantics in rows view
- **GIVEN** `/api/items` returns canonical inventory item metadata
- **WHEN** user opens inventory rows view
- **THEN** rows header semantics MUST render `Part #`, `Title`, `Condition`, and `Category`
- **AND** row identity MUST align to inventory `part_number` values when available
- **AND** UI MUST NOT render generic task-template headers such as `Task` or `Priority`

#### Scenario: Quick-create collection while assigning inventory item
- **GIVEN** user edits inventory item and opens collection picker
- **WHEN** user creates a new collection from picker
- **THEN** collection MUST be created and selected without leaving inventory edit flow

#### Scenario: Compact summary in browser header and no duplicate command/summary blocks
- **GIVEN** an authenticated user opens `/inventory` with a resolved collection dataset
- **WHEN** the Inventory workspace renders the Collection Browser region
- **THEN** the standalone `Command Row` section MUST NOT render
- **AND** the standalone `Summary Strip` card MUST NOT render
- **AND** the Collection Browser header MUST render a one-line summary with `Folders`, `Items`, `Active Brand`, and `Active Category` directly above the filter bar

### Requirement UI-SCREEN-INVENTORY-ITEMS-012: Inventory rows SHALL keep dense item columns readable
Inventory rows view SHALL preserve readable Part #, Title, Condition, Item type, Packaging, Category, and action columns at normal desktop review widths by allocating stable column widths and using horizontal scrolling when the full dense table exceeds the available viewport.

#### Scenario: Dense inventory columns remain readable
- **GIVEN** `/api/items` returns representative inventory records with long part numbers, titles, conditions, item types, packaging grades, and categories
- **WHEN** user opens `/inventory` in rows view at a normal desktop review width
- **THEN** right-side inventory headers and row values SHALL NOT overlap
- **AND** action controls SHALL remain reachable without covering data columns
- **AND** the table MAY expose horizontal scrolling instead of compressing columns into unreadable text

#### Scenario: Horizontally scrolled row actions remain deterministic
- **GIVEN** `/api/items` returns representative inventory records and the row table is horizontally scrolled to the right-side action column
- **WHEN** user opens a specific row action menu and chooses `Delete`
- **THEN** the Delete confirmation SHALL open for that same row
- **AND** canceling the confirmation SHALL close the dialog without changing the selected inventory context

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
| UC-INV-02 | Empty filtered results | Empty state appears | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `renders empty inventory state without global 500 fallback` |
| UC-INV-03 | API failure on load | Error state + retry, no 500 route | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-002 shows inline error state and recovers on retry` |
| UC-INV-04 | Row click open details | Details drawer opens selected item | planned: `cypress/e2e/ui/inventory.cy.ts` `inventory-row-opens-details` |
| UC-INV-05 | Checkbox bulk select | Selection mode appears, no row-open side effect | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection` |
| UC-INV-06 | Click Add Item | Create-item workflow opens from toolbar | planned: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `inventory-add-item-opens-create-flow` |
| UC-INV-07 | Click Add Folder | Folder creation workflow opens from toolbar | planned: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `inventory-add-folder-opens-create-flow` |
| UC-INV-08 | Toggle Rows/Cards view | View mode changes deterministically | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `renders inventory workspace, supports view toggle and filtering, and avoids 500`; existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection` |
| UC-INV-09 | Open Status/Priority/View controls | Browser controls render and open without errors | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection`; existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-011 scopes condition choices by item type and restores compact filters` |
| UC-INV-10 | Review dense inventory rows at desktop width | Right-side columns remain readable and actions are reachable via table scroll | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-012 keeps dense row columns readable` |
| UC-INV-11 | Open Delete from a horizontally scrolled inventory row action menu | Delete confirmation opens for the same row; cancel preserves selected item context | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-009 persists create-edit save flow and keeps media attach usable` |
