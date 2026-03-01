## Purpose
Define Help Center screen baseline behavior for in-app documentation entry.

## Requirements
### Requirement UI-SCREEN-HELP-CENTER-001: Help Center route SHALL provide deterministic placeholder state
Help Center route SHALL render a stable placeholder/coming-soon experience until full help content is implemented.

#### Scenario: Open help center route
- **GIVEN** authenticated user navigates to `/help-center`
- **WHEN** route loads
- **THEN** UI MUST render a non-error placeholder state with route-level shell controls still functional

### Requirement UI-SCREEN-HELP-CENTER-002: Help Center route SHALL preserve shell-level controls
Help Center route SHALL preserve global shell controls and layout contracts.

#### Scenario: Use shell controls on help center
- **GIVEN** help center route is active
- **WHEN** user uses global search, theme, language, or profile controls
- **THEN** controls MUST behave the same as other authenticated routes

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-HLP-01 | Open help center | Coming-soon placeholder loads without 500/404 | planned: `cypress/e2e/ui/help-center.cy.ts` `help-center-placeholder` |
| UC-HLP-02 | Use header controls from help center | Search/theme/profile controls remain usable | planned: `cypress/e2e/ui/help-center.cy.ts` `help-center-shell-controls` |
