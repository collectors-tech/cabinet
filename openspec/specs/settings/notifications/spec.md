## Purpose
Define Notifications settings screen behavior for notification channel toggles, guarded controls, persistence feedback, and missing active-profile recovery.

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

### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-003: Notifications screen SHALL expose Retry and Update notifications actions

#### Scenario: Retry notifications load failure
- **GIVEN** notifications section is in error state due to fetch/bootstrap failure
- **WHEN** user clicks `Retry`
- **THEN** notifications section MUST re-attempt load and render deterministic ready/empty/error state

#### Scenario: Update notifications action
- **GIVEN** notifications form is loaded with editable controls
- **WHEN** user clicks `Update notifications`
- **THEN** runtime MUST persist notification preferences and show deterministic success feedback
- **AND** the success feedback MUST be preserved in Notification Inbox history with source, level, category, title, and detail metadata.

### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-004: Notifications screen SHALL preserve editable controls on save failure

#### Scenario: Notifications save failure keeps retry context
- **GIVEN** notifications form is loaded with editable controls
- **WHEN** user changes notification scope/channel preferences and `Update notifications` fails
- **THEN** UI MUST show deterministic error feedback
- **AND** edited notification scope/channel preferences MUST remain selected for retry
- **AND** UI MUST NOT show success feedback for the failed save

### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-005: Notifications screen SHALL block editing when active profile context is missing

#### Scenario: Missing active profile blocks notification edits
- **GIVEN** `/settings/notifications` cannot resolve an active profile
- **WHEN** the screen renders the profile-context blocker
- **THEN** editable notification controls and the `Update notifications` action MUST remain hidden
- **AND** the screen MUST expose `Retry` and `Create or Select Profile` recovery actions without leaving the route

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-NOTIF-01 | Retry notifications load failure | `Retry` re-attempts notifications fetch deterministically | `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` (`UI-SCREEN-SETTINGS-NOTIFICATIONS-003 retries notifications settings load failure without route reload`) |
| UC-SET-NOTIF-02 | Update notifications action | `Update notifications` persists notification settings and records the success event in Notification Inbox history | `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` (`UI-SCREEN-SETTINGS-NOTIFICATIONS-003 updates notifications with deterministic success feedback`) |
| UC-SET-NOTIF-03 | Notifications save failure | Failed save shows error feedback and preserves edited notification choices | `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` (`UI-SCREEN-SETTINGS-NOTIFICATIONS-004 preserves edited notification choices when save fails`) |
| UC-SET-NOTIF-04 | Missing active profile blocker | Profile-context blocker hides editable notification controls and exposes recovery actions | `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` (`UI-SCREEN-SETTINGS-NOTIFICATIONS-005 blocks notification edits when active profile is missing`) |
