## Purpose
Define Profile settings screen behavior for identity details, profile links, Telegram setup handoff, persistence feedback, and missing active-profile recovery.

## Requirements
### Requirement UI-SCREEN-SETTINGS-PROFILE-001: Profile screen SHALL support validated profile editing
Profile screen SHALL allow editing username, display email, bio, and URL list with inline validation, while directing Telegram authorization to its governed Integrations pairing flow.

#### Scenario: Save profile details
- **GIVEN** user opens `/settings/profile`
- **WHEN** user submits valid profile values
- **THEN** runtime MUST persist values and UI MUST show deterministic success state
- **AND** the screen title/description MUST resolve to user-facing copy instead of raw translation keys
- **AND** profile save MUST NOT overwrite connector-owned Telegram sender/chat pairing state

#### Scenario: Open governed Telegram pairing setup
- **GIVEN** user opens `/settings/profile`
- **WHEN** the Telegram catalog capture card renders
- **THEN** it MUST explain that private pairing replaces manual sender/chat entry
- **AND** it MUST link to the Telegram Integrations setup
- **AND** it MUST NOT render editable sender-id or chat-id controls

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

#### Scenario: Invalid profile URL submission
- **GIVEN** user adds a profile URL row with an invalid URL value
- **WHEN** user clicks `Update profile`
- **THEN** UI MUST block the save, show a field-level URL validation error, and avoid sending a profile settings update request

#### Scenario: Update profile action
- **GIVEN** profile form contains valid values
- **WHEN** user clicks `Update profile`
- **THEN** profile changes MUST persist and success feedback MUST render deterministically

### Requirement UI-SCREEN-SETTINGS-PROFILE-005: Profile screen SHALL block editing when active profile context is missing

#### Scenario: Missing active profile blocks profile edits
- **GIVEN** `/settings/profile` cannot resolve an active profile
- **WHEN** the screen renders the profile-context blocker
- **THEN** editable profile controls, Telegram catalog capture controls, URL controls, `Add URL`, and `Update profile` MUST remain hidden
- **AND** the screen MUST expose `Retry` and `Create or Select Profile` recovery actions without leaving the route

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-PROF-01 | Retry profile load failure | `Retry` re-attempts profile fetch deterministically | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-003 retries profile settings load failure without route reload` |
| UC-SET-PROF-02 | Add URL action | `Add URL` appends editable URL row | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001 persists profile values through Cabinet settings API` |
| UC-SET-PROF-03 | Update profile action | `Update profile` persists profile values | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001 persists profile values through Cabinet settings API` |
| UC-SET-PROF-04 | Open Telegram pairing setup | Profile shows no manual sender/chat inputs and links to Telegram Integrations setup | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001` |
| UC-SET-PROF-05 | Invalid profile URL submission | Profile URL validation blocks save and keeps profile settings API untouched | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-001 blocks invalid profile URL submission before save` |
| UC-SET-PROF-06 | Missing active profile blocker | Profile-context blocker hides editable profile controls and exposes recovery actions | `ui.web/cypress/e2e/settings/profile/spec.cy.ts` `UI-SCREEN-SETTINGS-PROFILE-005 blocks profile edits when active profile is missing` |
