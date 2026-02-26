## Purpose
Define Inventory Barcodes screen behavior and barcode-resolution use cases.

## Requirements
### Requirement: Inventory Barcodes screen SHALL support manual add and lookup workflows
The screen SHALL support barcode entry, item-linked barcode list, and lookup actions.

#### Scenario: Use case - add and verify barcode
- **WHEN** user adds barcode to item and runs lookup
- **THEN** barcode list and lookup result SHALL reflect updated state

### Requirement: Inventory Barcodes screen SHALL support external search fallback
When local matches are absent, screen SHALL provide external search flow.

#### Scenario: Use case - no local match fallback
- **WHEN** lookup returns no local match
- **THEN** screen SHALL provide external search option and result URL/state
