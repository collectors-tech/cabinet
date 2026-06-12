## Purpose
Define Appearance settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-APPEARANCE-001: Appearance screen SHALL manage theme and font preferences

#### Scenario: Update appearance preferences
- **GIVEN** user opens `/settings/appearance`
- **WHEN** user changes theme/font and saves
- **THEN** UI MUST apply preferences immediately and persist for next session

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-002: Appearance screen SHALL support language selection including Chinese and Japanese
Language preferences MUST include at minimum English, Chinese, and Japanese options.

#### Scenario: Change language to Chinese or Japanese
- **GIVEN** user opens `/settings/appearance`
- **WHEN** user selects `Chinese` or `Japanese` and saves
- **THEN** UI language MUST switch without requiring account re-login
- **AND** selected language MUST persist for subsequent sessions

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-003: Language switch SHALL provide deterministic fallback behavior

#### Scenario: Missing translation key fallback
- **GIVEN** selected language lacks specific translation key
- **WHEN** UI renders missing key
- **THEN** runtime MUST fallback to default language text deterministically without breaking layout

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-004: First-run theme SHALL default to dark without startup flash

#### Scenario: First-run with no saved preference
- **GIVEN** user opens the app with no `vite-ui-theme` cookie present
- **AND** system color preference resolves to light mode
- **WHEN** initial HTML and app shell render
- **THEN** `<html>` MUST apply `dark` class on first paint
- **AND** UI MUST NOT flash a light theme before hydration
- **AND** no theme cookie MUST be created until user explicitly selects a preference

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-005: Appearance screen SHALL expose explicit Update preferences and Retry actions

#### Scenario: Update preferences action
- **GIVEN** appearance controls are loaded with valid selectable values
- **WHEN** user clicks `Update preferences`
- **THEN** runtime MUST persist appearance settings and UI MUST show deterministic success feedback

#### Scenario: Retry appearance load failure
- **GIVEN** appearance section fails to load due to API/bootstrap error
- **WHEN** user clicks `Retry`
- **THEN** appearance section MUST re-attempt load and render ready/empty/error state deterministically

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-006: Appearance screen SHALL preserve edits without applying unpersisted preferences when save fails

#### Scenario: Appearance save failure
- **GIVEN** appearance controls are loaded with valid selectable values
- **WHEN** user edits theme, font, and language and the settings save request fails
- **THEN** UI MUST show deterministic error feedback and MUST NOT show success feedback
- **AND** the edited control values MUST remain available for retry
- **AND** runtime MUST NOT apply theme, font, or language side effects that were not persisted

### Requirement UI-SCREEN-SETTINGS-APPEARANCE-007: Appearance screen SHALL block editing when active profile context is missing

#### Scenario: Missing active profile context
- **GIVEN** `/settings/appearance` cannot resolve an active profile
- **WHEN** Appearance settings renders the blocker state
- **THEN** editable appearance controls and `Update preferences` MUST be hidden
- **AND** the screen MUST expose deterministic `Retry` and `Create or Select Profile` recovery actions without leaving `/settings/appearance`

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-APP-01 | Update preferences action | `Update preferences` persists theme/language/display selections | implemented: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` (`UI-SCREEN-SETTINGS-APPEARANCE-005 updates preferences with deterministic success feedback`) |
| UC-SET-APP-02 | Retry appearance load failure | `Retry` re-attempts appearance fetch deterministically | implemented: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` (`UI-SCREEN-SETTINGS-APPEARANCE-005 retries appearance settings load failure without route reload`) |
| UC-SET-APP-03 | Appearance save failure | Failed `Update preferences` shows error feedback, keeps edited controls ready for retry, and does not apply unpersisted theme/font/language side effects | implemented: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` (`UI-SCREEN-SETTINGS-APPEARANCE-006 preserves edited appearance controls without applying unpersisted preferences when save fails`) |
| UC-SET-APP-04 | Missing active profile blocker | Profile-context blocker hides appearance controls and exposes retry/profile-selection recovery actions | implemented: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` (`UI-SCREEN-SETTINGS-APPEARANCE-007 blocks appearance edits when active profile is missing`) |
