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

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-APP-01 | Update preferences action | `Update preferences` persists theme/language/display selections | planned: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` `settings-appearance-update-preferences` |
| UC-SET-APP-02 | Retry appearance load failure | `Retry` re-attempts appearance fetch deterministically | planned: `ui.web/cypress/e2e/settings/appearance/spec.cy.ts` `settings-appearance-retry` |
