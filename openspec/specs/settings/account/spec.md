## Purpose
Define Account settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-ACCOUNT-001: Account screen SHALL support name, date-of-birth, and language updates

#### Scenario: Save account fields
- **GIVEN** user opens `/settings/account`
- **WHEN** user submits valid account values
- **THEN** runtime MUST persist values and screen MUST reflect saved values on reload

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-002: Account screen SHALL validate required fields

#### Scenario: Invalid account submission
- **GIVEN** required account field is missing/invalid
- **WHEN** user submits form
- **THEN** UI MUST block save and show field-level validation errors

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-003: Account screen SHALL expose explicit Update account action

#### Scenario: Update account action
- **GIVEN** account form has valid values
- **WHEN** user clicks `Update account`
- **THEN** runtime MUST persist account fields and render deterministic success feedback

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-004: Account screen SHALL expose retry action on account load failure

#### Scenario: Retry account section load
- **GIVEN** account section is in error state due to failed fetch/bootstrap
- **WHEN** user clicks `Retry`
- **THEN** account section MUST re-attempt load and render ready/empty/error state deterministically

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-005: Account screen SHALL preserve editable fields on save failure
When account settings save fails, the screen SHALL show deterministic error feedback, keep the edited name/date/language values available for retry, and avoid displaying success feedback until persistence succeeds.

#### Scenario: Account save failure keeps retry context
- **GIVEN** user changes account settings fields
- **WHEN** runtime rejects the account settings update
- **THEN** UI MUST render deterministic save-error feedback
- **AND** edited account field values MUST remain in the form for retry
- **AND** success feedback MUST NOT render

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-006: Account screen SHALL avoid persistence calls for invalid submissions
When account settings validation fails in the browser, the screen SHALL keep the user on `/settings/account`, show field-level validation feedback, and avoid calling the profile settings update API.

#### Scenario: Invalid account submission does not save
- **GIVEN** user clears a required account field
- **WHEN** user submits the account form
- **THEN** UI MUST show field-level validation feedback
- **AND** runtime MUST NOT receive a profile settings update request
- **AND** route MUST remain `/settings/account`

### Requirement UI-SCREEN-SETTINGS-ACCOUNT-007: Account screen SHALL block editing when active profile context is missing
When the authenticated `/settings/account` route cannot resolve an active profile, the screen SHALL render the shared profile-context blocker, hide editable account controls, expose retry/profile-selection recovery actions, and avoid account settings mutation calls.

#### Scenario: Missing active profile blocks account edits
- **GIVEN** user opens `/settings/account`
- **AND** runtime reports no active profile context
- **WHEN** the account settings section renders
- **THEN** UI MUST show the active-profile-required blocker
- **AND** editable account controls and `Update account` MUST NOT render
- **AND** `Retry` and `Create or Select Profile` recovery actions MUST render
- **AND** clicking `Retry` MUST re-attempt active profile resolution without leaving `/settings/account`

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-ACC-01 | Update account action | `Update account` persists account values | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-001 persists account fields across reload` |
| UC-SET-ACC-02 | Retry account load failure | `Retry` re-attempts account fetch deterministically | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-004 retries account settings load failure without route reload` |
| UC-SET-ACC-03 | Account save failure | Save failure shows error feedback and preserves edited field values | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-005 preserves edited account fields when save fails` |
| UC-SET-ACC-04 | Invalid account submission no-save | Invalid required fields show validation feedback without calling the save API | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-006 blocks invalid account submission without calling save` |
| UC-SET-ACC-05 | Missing active profile blocker | Profile-context blocker hides account controls and exposes retry/profile-selection recovery actions | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-007 blocks account edits when active profile is missing` |
