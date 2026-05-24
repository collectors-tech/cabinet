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
- **AND** retrying after the active profile endpoint recovers SHALL restore the active profile name and database/sample classification label

### Requirement PROFILES-004: Cabinet SHALL create and activate database profiles from the shell switcher
The authenticated database/profile switcher SHALL let users create a new database profile and SHALL make that profile the active app-wide data context immediately after creation.

#### Scenario: Create active database profile from switcher
- **GIVEN** an authenticated user has an active database profile
- **WHEN** the user creates a new database profile from the shell switcher
- **THEN** Cabinet SHALL create the profile through the profile API
- **AND** Cabinet SHALL set the created profile as the active profile before reloading shell data context
- **AND** the active database label SHALL show the created profile name after reload

### Requirement PROFILES-005: Cabinet SHALL keep existing profile selection app-wide across core sections
Selecting an existing database profile from the authenticated shell switcher SHALL activate that profile as the app-wide data context for core authenticated sections.

#### Scenario: Select existing database profile across app sections
- **GIVEN** an authenticated user has at least two database profiles
- **WHEN** the user selects an existing non-active database profile from the shell switcher
- **THEN** Cabinet SHALL persist that profile through the active profile API
- **AND** Inventory, Wishlist, Collections, Settings, Chats, and Integrations SHALL show the selected database profile as the active shell context

### Requirement PROFILES-006: Cabinet SHALL distinguish showcase sample profiles from working profiles
Showcase, demo, or sample-data profiles SHALL be visibly labelled as sample context anywhere the authenticated shell presents them beside normal working database profiles.

#### Scenario: Showcase profile context is explicit in switcher
- **GIVEN** an authenticated user can choose between `Primary DB` and `Showcase DB`
- **WHEN** the user opens the database/profile switcher
- **THEN** `Primary DB` SHALL be labelled as a normal database
- **AND** `Showcase DB` SHALL be labelled as showcase sample data before and after selection
