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
