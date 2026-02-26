## ADDED Requirements

### Requirement: Users screen SHALL be API-backed for list and mutations
The system SHALL load user records from API and SHALL execute add/update role/invite actions through API-backed workflows.

#### Scenario: Users list load
- **WHEN** the user opens the Users screen
- **THEN** the screen SHALL render data returned by Cabinet user APIs

#### Scenario: Users mutation
- **WHEN** the user performs add or role update action
- **THEN** the action SHALL persist through API and reflect in refreshed state

### Requirement: Settings screens SHALL persist profile configuration via API
The system SHALL load and save all settings sections through profile-scoped API endpoints.

#### Scenario: Settings persistence
- **WHEN** a user updates a setting and reloads the screen
- **THEN** the updated value SHALL remain persisted from API state

### Requirement: Chat screen SHALL use local thread/message/action APIs
The system SHALL use Cabinet chat APIs for thread listing, message history, attachments, and action preview/apply flows.

#### Scenario: Chat thread lifecycle
- **WHEN** a user creates and reopens a thread
- **THEN** thread and message data SHALL be retrieved from API storage
