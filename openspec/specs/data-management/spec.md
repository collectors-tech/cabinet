## Purpose
Define safe import/export flows and maintenance operations for collection data stores.

## Requirements
### Requirement: Data import SHALL support safe dry-run and explicit conflict choices
Cabinet SHALL support JSON/CSV dry-run preview and merge/create/skip conflict resolution on apply.

#### Scenario: Import dry-run
- **GIVEN** import payload is submitted
- **WHEN** user runs dry-run import
- **THEN** Cabinet SHALL report conflicts without mutating persisted records

### Requirement: Maintenance operations SHALL include reindex and repair
Cabinet SHALL support search reindex and database repair endpoints.

#### Scenario: Reindex operation
- **GIVEN** maintenance operation is requested
- **WHEN** user triggers reindex
- **THEN** Cabinet SHALL execute and report maintenance outcome
