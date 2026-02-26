## Purpose
Define canonical barcode capture, validation, and attachment behavior for item records.

## Requirements
### Requirement: Cabinet SHALL support manual and image-assisted barcode workflows
Cabinet SHALL support manual barcode entry and barcode detection from uploaded media.

#### Scenario: Manual barcode add
- **GIVEN** an item is selected for barcode operations
- **WHEN** user submits a barcode
- **THEN** Cabinet SHALL attach barcode to the item's barcode set

### Requirement: Duplicate barcode handling SHALL support variant-aware resolution
Cabinet SHALL support duplicate barcode resolution across variants without destructive overwrite.

#### Scenario: Duplicate barcode attach attempt
- **GIVEN** barcode already exists on variant-linked records
- **WHEN** user attempts to attach the same barcode
- **THEN** Cabinet SHALL preserve existing links and require explicit resolution action
