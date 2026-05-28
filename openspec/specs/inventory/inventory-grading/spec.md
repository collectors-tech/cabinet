## Purpose
Define collector grading, packaging grades, and classification metadata for inventory items.

## Requirements
### Requirement INVENTORY-GRADING-001: Inventory SHALL support configurable condition and packaging grade enums
Cabinet SHALL support admin-managed enum sets for packaging grades and admin-managed condition scales scoped by single-select item type.

#### Scenario: Persist grading enums
- **GIVEN** admin user opens grading settings with enum management permission
- **WHEN** admin saves grade lists for car and packaging
- **THEN** updated enum values MUST be available in inventory create/edit forms

#### Scenario: Manage packaging grades from taxonomy settings
- **GIVEN** admin user opens Settings > Categories for the active profile
- **WHEN** admin adds or removes packaging grade values and saves taxonomy settings
- **THEN** Cabinet MUST persist those values in the profile-scoped packaging grade enum setting
- **AND** Inventory item grading controls MUST use the saved packaging grade values without losing existing item records

#### Scenario: Persist item type condition scale
- **GIVEN** admin user opens inventory taxonomy settings
- **WHEN** admin saves an item type with ordered condition values
- **THEN** updated item type and condition values MUST be available in inventory create/edit forms

#### Scenario: Seed collector condition scales
- **WHEN** a profile has no custom item type condition scale settings
- **THEN** Cabinet MUST provide Slot Cars and Trading Cards item type defaults
- **AND** Slot Cars MUST include numeric condition values from `10+` through `1`
- **AND** Trading Cards MUST include `Mint (M)`, `Near Mint (NM)`, `Excellent (EX)`, `Good (GD)`, `Light Played (LP)`, `Played (PL)`, and `Poor (PO)`

#### Scenario: Scope condition choices by item type
- **GIVEN** an inventory item is assigned an item type
- **WHEN** user edits condition for that item
- **THEN** condition choices MUST come from the selected item type condition scale
- **AND** changing item type MUST update the available condition choices

#### Scenario: Use configured taxonomy in inventory editor
- **GIVEN** profile-scoped item type condition scales and packaging grade enums exist
- **WHEN** user creates or edits an inventory item
- **THEN** the editor MUST expose item type, condition, and packaging grade controls from the configured taxonomy
- **AND** saving the item MUST persist the selected item type and packaging grade values on the inventory record

#### Scenario: Preserve taxonomy from wishlist planning
- **GIVEN** profile-scoped item type condition scales and packaging grade enums exist
- **WHEN** user creates or edits a wishlist entry before owning the item
- **THEN** the wishlist editor MUST expose item type, condition, and packaging grade controls from the configured taxonomy
- **AND** saving the wishlist entry MUST persist the selected taxonomy values on the linked inventory record so later conversion keeps the same classification

#### Scenario: Reject invalid taxonomy values
- **GIVEN** profile-scoped item type condition scales and packaging grade enums exist
- **WHEN** API clients create, update, or bulk edit inventory items with item type, condition, or packaging grade values outside the configured taxonomy
- **THEN** Cabinet MUST reject the request with an actionable `invalid_taxonomy_value` error that names the invalid field
- **AND** valid taxonomy values MUST continue to save without breaking existing records that omit optional taxonomy fields

### Requirement INVENTORY-GRADING-002: Item records SHALL support grading and collector classification fields
Cabinet SHALL persist grading fields per item/instance including grading status and collector classification.

#### Scenario: Save graded item metadata
- **GIVEN** inventory item edit form is open
- **WHEN** user submits grading fields (`grading_status`, `grader`, `grade_numeric`, `slabbed`, `collector_classification`)
- **THEN** API MUST return `200` and persisted item MUST include submitted grading values

### Requirement INVENTORY-GRADING-003: Default values SHALL align with collector workflow
Cabinet SHALL apply default values for newly created inventory records where configured.

#### Scenario: Create new item with defaults
- **GIVEN** profile default grading/priority configuration exists
- **WHEN** user creates a new inventory item without explicit grading fields
- **THEN** runtime MUST apply configured defaults (for example `priority=medium`, `grading_status=ungraded`)
