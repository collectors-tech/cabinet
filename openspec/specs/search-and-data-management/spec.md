## Purpose
Define search, filter, saved view, import/export, and maintenance operations.

## Requirements
### Requirement: Search SHALL support full-text query and structured filtering
Cabinet SHALL support full-text search with brand, condition, status, tags, and scale filters.

#### Scenario: Filtered search
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user applies query and filter set
- **THEN** Cabinet SHALL return filtered/sorted search results

### Requirement: Saved filters SHALL be profile-scoped and reusable
Cabinet SHALL support create, update, delete, and reuse for saved search filters.

#### Scenario: Save and reload filter
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user saves filter and reloads route
- **THEN** saved filter SHALL be available for reuse

### Requirement: Data import SHALL support safe dry-run and explicit conflict choices
Cabinet SHALL support JSON/CSV dry-run preview and merge/create/skip resolution on apply.

#### Scenario: Import dry-run
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user executes dry-run import
- **THEN** Cabinet SHALL report conflicts without mutating persisted records

### Requirement: Maintenance operations SHALL include reindex and repair
Cabinet SHALL support search reindex and database repair endpoints.

#### Scenario: Reindex operation
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user triggers reindex
- **THEN** Cabinet SHALL execute and report maintenance outcome
