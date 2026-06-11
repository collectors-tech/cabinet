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

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-ACC-01 | Update account action | `Update account` persists account values | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-001 persists account fields across reload` |
| UC-SET-ACC-02 | Retry account load failure | `Retry` re-attempts account fetch deterministically | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-004 retries account settings load failure without route reload` |
| UC-SET-ACC-03 | Account save failure | Save failure shows error feedback and preserves edited field values | `ui.web/cypress/e2e/settings/account/spec.cy.ts` `UI-SCREEN-SETTINGS-ACCOUNT-005 preserves edited account fields when save fails` |
