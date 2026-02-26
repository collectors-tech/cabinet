## Purpose
Define settings persistence and integrations workspace behavior.

## Requirements
### Requirement: Settings SHALL persist per-profile configuration
Cabinet SHALL persist profile settings including theme, scanner schedule, update channel, backup frequency, and database location.

#### Scenario: Save settings
- **WHEN** user updates settings and saves
- **THEN** values SHALL persist across restart for active profile

### Requirement: Settings SHALL manage provider and AI credentials
Cabinet SHALL support profile-scoped eBay and OpenAI credential configuration.

#### Scenario: Update provider credential
- **WHEN** user saves provider credential in settings
- **THEN** credential SHALL be retrievable via secure profile secret path

### Requirement: Integrations workspace SHALL present API-backed provider records
Cabinet SHALL display integration providers and details from runtime sources, not static template placeholders.

#### Scenario: Open integrations route
- **WHEN** user navigates to integrations workspace
- **THEN** route SHALL load provider states from API-backed data path
