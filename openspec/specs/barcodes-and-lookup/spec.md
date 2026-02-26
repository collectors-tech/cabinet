## Purpose
Define barcode capture, lookup, and variant-aware duplicate handling behavior.

## Requirements
### Requirement: Cabinet SHALL support manual and image-assisted barcode workflows
Cabinet SHALL support manual barcode entry and barcode detection from uploaded media.

#### Scenario: Manual barcode add
- **WHEN** user submits barcode for an item
- **THEN** Cabinet SHALL attach barcode to item barcode set

### Requirement: Barcode lookup SHALL support local and external resolution
Cabinet SHALL support local match lookup and external search integrations.

#### Scenario: Local barcode lookup
- **WHEN** user requests barcode lookup
- **THEN** Cabinet SHALL return local matches when present

#### Scenario: External barcode search
- **WHEN** local match is absent and external search is invoked
- **THEN** Cabinet SHALL return provider search results or failure state

### Requirement: Duplicate barcode handling SHALL support variant-aware resolution
Cabinet SHALL support duplicate barcode resolution across variants without destructive overwrite.

#### Scenario: Duplicate barcode attach attempt
- **WHEN** barcode already exists on variant-linked records
- **THEN** Cabinet SHALL preserve existing links and require explicit resolution action
