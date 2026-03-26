## Purpose
Define Help Center screen behavior for in-app documentation discovery and reading.

## Requirements
### Requirement UI-SCREEN-HELP-CENTER-001: Help Center route SHALL surface an in-app article library
Help Center route SHALL render the available Help Center article set as a browsable in-app library rather than only a placeholder state.

#### Scenario: Open help center route
- **GIVEN** authenticated user navigates to `/help-center`
- **WHEN** route loads
- **THEN** UI MUST render a visible Help Center article library with selectable article entries and MUST avoid route-level error fallback state

### Requirement UI-SCREEN-HELP-CENTER-002: Help Center route SHALL render selected article content in-app
Help Center route SHALL let the user open readable article content inside the Help Center route without leaving the shell.

#### Scenario: Open article from help center library
- **GIVEN** help center route is active with article entries visible
- **WHEN** user selects a Help Center article
- **THEN** the route MUST show that article title and readable guide content in the in-app article viewer

### Requirement UI-SCREEN-HELP-CENTER-003: Help Center route SHALL preserve shell-level controls
Help Center route SHALL preserve global shell controls and layout contracts while the article library and reader are active.

#### Scenario: Use shell controls on help center
- **GIVEN** help center route is active
- **WHEN** user uses global search, theme, language, sidebar, or profile controls
- **THEN** controls MUST behave the same as other authenticated routes

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-HLP-01 | Open help center | Article library loads without 500/404 fallback | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-001 renders article library on help-center route` |
| UC-HLP-02 | Open article from help center | Selected article content renders in the in-app reader | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-002 renders selected article content in-app` |
| UC-HLP-03 | Use shell controls from help center | Search/theme/profile/sidebar controls remain usable | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-003 preserves shell controls on help-center route` |
