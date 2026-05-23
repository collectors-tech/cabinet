## Purpose
Define profile isolation, switching, and profile-scoped storage behavior.

## Requirements
### Requirement PROFILES-001: Cabinet SHALL support profile-isolated storage and secrets
Each profile SHALL have isolated database, settings, API keys, and license state.

#### Scenario: Profile isolation
- **GIVEN** profile A and profile B both exist
- **WHEN** user switches from profile A to profile B
- **THEN** profile A records SHALL not appear in profile B views

### Requirement PROFILES-002: Cabinet SHALL persist active profile context across authenticated shell sections
The active profile SHALL remain the app-wide context for shell workspace state across reloads and authenticated section changes.

#### Scenario: Profile-scoped shell workspace persistence
- **GIVEN** an authenticated user has selected a shell workspace while profile A is active
- **WHEN** the user reloads and navigates from Inventory to Profile Settings
- **THEN** the selected shell workspace SHALL be restored from profile A's persisted context
- **AND** the authenticated section change SHALL NOT reset the active profile workspace context

### Requirement PROFILES-003: Cabinet SHALL surface active profile load failures with retry recovery
The authenticated shell SHALL show actionable guidance when active profile/database context cannot be loaded and SHALL let the user retry without a dead-end screen.

#### Scenario: Active profile unavailable recovery
- **GIVEN** the authenticated shell cannot load the active profile/database context
- **WHEN** the user opens the database switcher
- **THEN** the shell SHALL show profile-unavailable guidance with a retry action
- **AND** retrying after the active profile endpoint recovers SHALL restore the active database label
