## Purpose
Define collector grading, packaging grades, and classification metadata for inventory items.

## Requirements
### Requirement INVENTORY-GRADING-001: Inventory SHALL support configurable condition and packaging grade enums
Cabinet SHALL support admin-managed enum sets for car condition grade and packaging grade.

#### Scenario: Persist grading enums
- **GIVEN** admin user opens grading settings with enum management permission
- **WHEN** admin saves grade lists for car and packaging
- **THEN** updated enum values MUST be available in inventory create/edit forms

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
