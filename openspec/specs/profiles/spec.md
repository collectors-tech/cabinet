## Purpose
Define profile isolation, switching, and profile-scoped storage behavior.

## Requirements
### Requirement PROFILES-001: Cabinet SHALL support profile-isolated storage and secrets
Each profile SHALL have isolated database, settings, API keys, and license state.

#### Scenario: Profile isolation
- **GIVEN** profile A and profile B both exist
- **WHEN** user switches from profile A to profile B
- **THEN** profile A records SHALL not appear in profile B views

