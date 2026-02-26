## Purpose
Define per-profile settings persistence and secure credential storage behavior.

## Requirements
### Requirement: Settings SHALL persist per-profile configuration
Cabinet SHALL persist profile settings including theme, scanner schedule, update channel, backup frequency, and database location.

#### Scenario: Save settings
- **GIVEN** profile settings are modified
- **WHEN** user saves settings
- **THEN** values SHALL persist across restart for active profile

### Requirement: Settings SHALL manage provider and AI credentials
Cabinet SHALL support profile-scoped provider and AI credential configuration.

#### Scenario: Update provider credential
- **GIVEN** profile settings context is active
- **WHEN** user saves provider credential
- **THEN** credential SHALL be retrievable via secure profile secret path

