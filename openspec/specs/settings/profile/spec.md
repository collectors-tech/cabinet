## Purpose
Define Profile settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-PROFILE-001: Profile screen SHALL support validated profile editing
Profile screen SHALL allow editing username, display email, bio, and URL list with inline validation.

#### Scenario: Save profile details
- **GIVEN** user opens `/settings/profile`
- **WHEN** user submits valid profile values
- **THEN** runtime MUST persist values and UI MUST show deterministic success state
- **AND** the screen title/description MUST resolve to user-facing copy instead of raw translation keys

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
| UC-SET-PROF-01 | Retry profile load failure | `Retry` re-attempts profile fetch deterministically | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-003 retries profile settings load failure without route reload` |
| UC-SET-PROF-02 | Add URL action | `Add URL` appends editable URL row | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001 persists profile values through Cabinet settings API` |
| UC-SET-PROF-03 | Update profile action | `Update profile` persists profile values | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001 persists profile values through Cabinet settings API` |
