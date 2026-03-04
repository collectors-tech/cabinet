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

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-ACC-01 | Update account action | `Update account` persists account values | planned: `ui.web/cypress/e2e/settings/account/spec.cy.ts` `settings-account-update-account` |
| UC-SET-ACC-02 | Retry account load failure | `Retry` re-attempts account fetch deterministically | planned: `ui.web/cypress/e2e/settings/account/spec.cy.ts` `settings-account-retry` |
