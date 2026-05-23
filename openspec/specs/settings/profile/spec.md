## Purpose
Define Profile settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-PROFILE-001: Profile screen SHALL support validated profile editing
Profile screen SHALL allow editing username, display email, bio, URL list, and Telegram catalog capture authorization values with inline validation.

#### Scenario: Save profile details
- **GIVEN** user opens `/settings/profile`
- **WHEN** user submits valid profile values
- **THEN** runtime MUST persist values and UI MUST show deterministic success state
- **AND** the screen title/description MUST resolve to user-facing copy instead of raw translation keys
- **AND** Telegram catalog capture sender/chat authorization values MUST persist through the same profile settings API

#### Scenario: Save Telegram catalog capture authorization
- **GIVEN** user opens `/settings/profile`
- **WHEN** user enters a Telegram sender ID and chat ID and submits the profile form
- **THEN** runtime MUST persist `telegram.catalog_capture.sender_id` and `telegram.catalog_capture.chat_id`
- **AND** reloading the Profile settings screen MUST show the saved Telegram authorization values

### Requirement UI-SCREEN-SETTINGS-PROFILE-002: Profile screen SHALL handle deterministic error states

#### Scenario: Profile save failure
- **GIVEN** profile API update fails
- **WHEN** user submits profile form
- **THEN** UI MUST show actionable error state and preserve entered values for retry

### Requirement UI-SCREEN-SETTINGS-PROFILE-003: Profile screen SHALL expose retry action for failed profile bootstrap/load

#### Scenario: Retry profile load
- **GIVEN** profile section is in error state after fetch/bootstrap failure
- **WHEN** user clicks `Retry`
- **THEN** runtime MUST re-attempt profile fetch and render ready/empty/error state deterministically

### Requirement UI-SCREEN-SETTINGS-PROFILE-004: Profile screen SHALL expose explicit Add URL and Update profile actions

#### Scenario: Add URL action
- **GIVEN** user is editing profile links list
- **WHEN** user clicks `Add URL`
- **THEN** UI MUST append a new editable URL row without losing existing form state

#### Scenario: Update profile action
- **GIVEN** profile form contains valid values
- **WHEN** user clicks `Update profile`
- **THEN** profile changes MUST persist and success feedback MUST render deterministically

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-PROF-01 | Retry profile load failure | `Retry` re-attempts profile fetch deterministically | planned: `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `settings-profile-retry` |
| UC-SET-PROF-02 | Add URL action | `Add URL` appends editable URL row | planned: `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `settings-profile-add-url` |
| UC-SET-PROF-03 | Update profile action | `Update profile` persists profile values | planned: `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `settings-profile-update-profile` |
| UC-SET-PROF-04 | Update Telegram catalog capture authorization | `Update profile` persists Telegram sender/chat authorization values | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001` |
