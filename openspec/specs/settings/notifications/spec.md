## Purpose
Define Notifications settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-001: Notifications screen SHALL support notification scope and channel toggles

#### Scenario: Save notification preferences
- **GIVEN** user opens `/settings/notifications`
- **WHEN** user updates notification scope/toggles and saves
- **THEN** runtime MUST persist editable controls and show deterministic success state

### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-002: Notifications screen SHALL enforce guarded controls

#### Scenario: Guarded toggle remains immutable
- **GIVEN** a control is marked immutable (for example security notification)
- **WHEN** user attempts to modify it
- **THEN** UI MUST block change and preserve required value
