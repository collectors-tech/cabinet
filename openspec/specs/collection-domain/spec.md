## Purpose
Define canonical item and instance data model behavior for collection management.

## Requirements
### Requirement: Canonical item records SHALL include required collector metadata
Cabinet SHALL store canonical item fields including brand, category, part number, title, make/model/year/scale/series, description, tags, barcodes, and timestamps.

#### Scenario: Create canonical item
- **WHEN** user submits valid canonical item payload
- **THEN** Cabinet SHALL persist the record with required metadata

### Requirement: One canonical item SHALL support many instances
Cabinet SHALL allow multiple instance records per canonical item.

#### Scenario: Add instance to item
- **WHEN** user adds instance to existing item
- **THEN** Cabinet SHALL store instance under item linkage

### Requirement: Auto-merge SHALL require explicit user confirmation
Cabinet SHALL not merge item records implicitly.

#### Scenario: Potential duplicate insertion
- **WHEN** incoming record matches existing identity keys
- **THEN** Cabinet SHALL require explicit user action before merge

### Requirement: Status and grading vocab SHALL be configurable per database
Cabinet SHALL support configurable status and grading enums with admin-only management.

#### Scenario: Enum configuration update
- **WHEN** admin updates status or grading lists
- **THEN** new values SHALL become available in item workflows
