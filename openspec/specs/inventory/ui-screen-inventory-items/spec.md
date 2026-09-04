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
- **AND** the Retry action SHALL remain mounted while a retry is in flight and expose its busy/disabled state
- **AND** pointer click, focused Enter, or focused Space SHALL each dispatch exactly one retry request
- **AND** duplicate rapid activation SHALL NOT dispatch a second concurrent retry request

### Requirement UI-SCREEN-INVENTORY-ITEMS-003: Inventory Items SHALL support sample and bulk data usage
The screen SHALL remain usable with both starter and stress datasets.

#### Scenario: Bulk dataset interaction
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user filters/sorts large item lists
- **THEN** interactions SHALL remain responsive and deterministic

### Requirement UI-SCREEN-INVENTORY-ITEMS-004: Inventory Collection layout SHALL keep controls compact and non-duplicated
Inventory Collection layout SHALL avoid duplicate control strips and keep summary context inside the Collection Browser header area.

#### Scenario: Compact summary in browser header and no duplicate command/summary blocks
- **GIVEN** an authenticated user opens `/inventory` with a resolved collection dataset
- **WHEN** the Inventory workspace renders the Collection Browser region
- **THEN** the standalone `Command Row` section MUST NOT render
- **AND** the standalone `Summary Strip` card MUST NOT render
- **AND** the Collection Browser header MUST render a one-line summary with `Folders`, `Items`, `Active Brand`, and `Active Category` directly above the filter bar

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

#### Scenario: Quick-create collection while assigning inventory item
- **GIVEN** user edits inventory item and opens collection picker
- **WHEN** user creates a new collection from picker
- **THEN** collection MUST be created and selected without leaving inventory edit flow

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

### Requirement UI-SCREEN-INVENTORY-ITEMS-013: Inventory item editor sidepanel SHALL keep footer actions visible
The responsive inventory item editor sidepanel SHALL keep its header and footer actions fixed within the panel while the form body scrolls independently.

#### Scenario: Constrained-height editor keeps actions reachable
- **GIVEN** user opens the inventory item editor sidepanel at a constrained viewport height
- **WHEN** the item form body contains more fields, media, notes, pricing, and evidence content than can fit vertically
- **THEN** the sidepanel body SHALL scroll independently
- **AND** footer actions including Previous, Next, Cancel, and Save Changes SHALL remain visible and actionable before and after body scrolling

### Requirement UI-SCREEN-INVENTORY-BONZA-001: Inventory create paste flow SHALL process supported provider URLs
Inventory create paste flow SHALL call provider URL ingestion for supported product URLs and prefill the create-item modal from the normalized item draft.

#### Scenario: Pasted Bonza URL prefills create item modal
- **GIVEN** user opens Inventory and activates the paste create action
- **WHEN** user submits `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **THEN** the create item modal MUST remain in confirm-before-create mode
- **AND** modal fields MUST be prefilled from Bonza normalized data
- **AND** title MUST be populated with `BONZA MUG WHITE`
- **AND** source URL MUST include the pasted Bonza product URL
- **AND** price, category, item metadata, stock, description, and image URL evidence MUST be available for review before create

#### Scenario: Unsupported pasted URL gives actionable feedback
- **GIVEN** user opens the Inventory create paste flow
- **WHEN** user submits a URL that does not match a supported provider product URL
- **THEN** UI MUST show an actionable unsupported-provider or unsupported-page message
- **AND** UI MUST keep the pasted value available for manual item creation
- **AND** UI MUST NOT discard user input

#### Scenario: Duplicate Bonza URL blocks silent create
- **GIVEN** provider ingestion reports an existing item for the pasted Bonza product URL
- **WHEN** the create modal renders the result
- **THEN** UI MUST show duplicate information
- **AND** UI MUST provide an action to open the existing item
- **AND** UI MUST require explicit confirmation before creating another item from the same source

### Requirement UI-SCREEN-INVENTORY-BONZA-002: Inventory create flow SHALL preserve provider provenance for pasted URLs
Inventory item creation from provider-ingested pasted URLs SHALL save source provenance and user-visible evidence with the created item.

#### Scenario: Created item stores Bonza source evidence
- **GIVEN** user confirms creation from a Bonza-ingested product draft
- **WHEN** Cabinet creates the inventory item
- **THEN** the item MUST retain original pasted URL, normalized source URL, provider id, provider family, provider product id, observed timestamp, and extraction method
- **AND** the item detail/editor evidence area MUST be able to display the source link and provider extraction summary

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
| UC-INV-03 | API failure on load | Stable error-state Retry supports pointer and keyboard recovery with one in-flight request and no 500 route | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` (`UI-SCREEN-INVENTORY-ITEMS-002 keeps pointer Retry stable and single-dispatch`; `UI-SCREEN-INVENTORY-ITEMS-002 dispatches one Retry from focused Enter`; `UI-SCREEN-INVENTORY-ITEMS-002 dispatches one Retry from focused Space`; `UI-SCREEN-INVENTORY-ITEMS-002 ignores duplicate rapid Retry activation`) |
| UC-INV-04 | Row click open details | Details drawer opens selected item | planned: `cypress/e2e/ui/inventory.cy.ts` `inventory-row-opens-details` |
| UC-INV-05 | Checkbox bulk select | Selection mode appears, no row-open side effect | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection` |
| UC-INV-06 | Click Add Item | Create-item workflow opens from toolbar | planned: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `inventory-add-item-opens-create-flow` |
| UC-INV-07 | Click Add Folder | Folder creation workflow opens from toolbar | planned: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `inventory-add-folder-opens-create-flow` |
| UC-INV-08 | Toggle Rows/Cards view | View mode changes deterministically | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `renders inventory workspace, supports view toggle and filtering, and avoids 500`; existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection` |
| UC-INV-09 | Open Status/Priority/View controls | Browser controls render and open without errors | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-001/008 covers search, filters, sort, reset, and bulk selection`; existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-011 scopes condition choices by item type and restores compact filters` |
| UC-INV-10 | Review dense inventory rows at desktop width | Right-side columns remain readable and actions are reachable via table scroll | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-012 keeps dense row columns readable` |
| UC-INV-11 | Open Delete from a horizontally scrolled inventory row action menu | Delete confirmation opens for the same row; cancel preserves selected item context | existing: `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-009 persists create-edit save flow and keeps media attach usable` |
| UC-INV-12 | Open responsive item editor sidepanel at constrained height | Body scrolls independently while footer actions stay visible/actionable | existing: `ui.web/cypress/e2e/inventory/inventory-editor-scroll-footer/spec.cy.ts` `UI-SCREEN-INVENTORY-ITEMS-013 keeps editor panel footer visible while the body scrolls` |
