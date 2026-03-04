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

### Requirement UI-SCREEN-SETTINGS-NOTIFICATIONS-003: Notifications screen SHALL expose Retry and Update notifications actions

#### Scenario: Retry notifications load failure
- **GIVEN** notifications section is in error state due to fetch/bootstrap failure
- **WHEN** user clicks `Retry`
- **THEN** notifications section MUST re-attempt load and render deterministic ready/empty/error state

#### Scenario: Update notifications action
- **GIVEN** notifications form is loaded with editable controls
- **WHEN** user clicks `Update notifications`
- **THEN** runtime MUST persist notification preferences and show deterministic success feedback

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-NOTIF-01 | Retry notifications load failure | `Retry` re-attempts notifications fetch deterministically | planned: `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` `settings-notifications-retry` |
| UC-SET-NOTIF-02 | Update notifications action | `Update notifications` persists notification settings | planned: `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` `settings-notifications-update` |
