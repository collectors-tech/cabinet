## Purpose
Define settings persistence and integrations workspace behavior.

## Requirements
### Requirement: Settings SHALL persist per-profile configuration
Cabinet SHALL persist profile settings including theme, scanner schedule, update channel, backup frequency, and database location.

#### Scenario: Save settings
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user updates settings and saves
- **THEN** values SHALL persist across restart for active profile

### Requirement: Settings SHALL manage provider and AI credentials
Cabinet SHALL support profile-scoped eBay and OpenAI credential configuration.

#### Scenario: Update provider credential
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user saves provider credential in settings
- **THEN** credential SHALL be retrievable via secure profile secret path

### Requirement: Integrations workspace SHALL present API-backed provider records
Cabinet SHALL display integration providers and details from runtime sources, not static template placeholders.

#### Scenario: Open integrations route
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user navigates to integrations workspace
- **THEN** route SHALL load provider states from API-backed data path

### Requirement: Integrations workspace SHALL consume canonical provider registry
Cabinet SHALL use `provider-registry` capability as source-of-truth for provider listing, capability badges, and provider-specific configuration routes.

#### Scenario: Render provider list from registry
- **GIVEN** provider registry definitions are available
- **WHEN** integrations workspace renders provider list
- **THEN** provider rows SHALL reflect registry-defined provider IDs and capability metadata
