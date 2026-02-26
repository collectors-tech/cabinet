## Purpose
Define full-text and structured filtering behavior for collection query workflows.

## Requirements
### Requirement: Search SHALL support full-text query and structured filtering
Cabinet SHALL support full-text search with brand, condition, status, tags, and scale filters.

#### Scenario: Filtered search
- **GIVEN** indexed collection data exists
- **WHEN** user applies query and filter set
- **THEN** Cabinet SHALL return filtered and sorted results

### Requirement: Saved filters SHALL be profile-scoped and reusable
Cabinet SHALL support create, update, delete, and reuse for saved filters.

#### Scenario: Save and reload filter
- **GIVEN** a filter definition is saved for active profile
- **WHEN** route reloads
- **THEN** saved filter SHALL remain available for reuse
