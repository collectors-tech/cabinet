## Purpose
Define integrations workspace provider listing and wiring behavior.

## Requirements
### Requirement: Integrations workspace SHALL present API-backed provider records
Cabinet SHALL display integration providers and details from runtime sources, not static template placeholders.

#### Scenario: Open integrations route
- **GIVEN** integrations workspace route is reachable
- **WHEN** user navigates to integrations workspace
- **THEN** provider states SHALL load from API-backed data path

### Requirement: Integrations workspace SHALL consume canonical provider registry
Cabinet SHALL use `provider-registry` capability as source-of-truth for provider list and capability badges.

#### Scenario: Render provider list from registry
- **GIVEN** provider registry definitions are available
- **WHEN** integrations workspace renders provider list
- **THEN** provider rows SHALL reflect registry-defined provider IDs and metadata

