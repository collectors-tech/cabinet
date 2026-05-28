## Purpose
Define collector grading, packaging grades, and classification metadata for inventory items.

## Requirements
### Requirement INVENTORY-GRADING-001: Inventory SHALL support configurable condition and packaging grade enums
Cabinet SHALL support admin-managed enum sets for packaging grades and admin-managed condition scales scoped by single-select item type.

#### Scenario: Persist grading enums
- **GIVEN** admin user opens grading settings with enum management permission
- **WHEN** admin saves grade lists for car and packaging
- **THEN** updated enum values MUST be available in inventory create/edit forms

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
