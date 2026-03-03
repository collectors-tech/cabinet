## Purpose
Define Home screen behavior as the primary command center with testable, action-first use-cases.

## Requirements
### Requirement UI-SCREEN-HOME-001: Home SHALL render a command-center summary with actionable priorities
Home SHALL present actionable discovery, pricing, and health signals above the fold.

#### Scenario: Home priority view ready state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home screen data loads successfully
- **THEN** users SHALL see actionable priority signals with direct actions

### Requirement UI-SCREEN-HOME-002: Home SHALL support deterministic state handling
Home SHALL support loading, empty, error, and ready states for each major panel.

#### Scenario: Home loading state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home data is still loading
- **THEN** Home SHALL render loading placeholders without layout shift

#### Scenario: Home empty state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** no actionable items exist
- **THEN** Home SHALL render calm-state guidance with at least one next action

#### Scenario: Home error state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home API fetch fails
- **THEN** Home SHALL render actionable retry/error messaging without route crash

### Requirement UI-SCREEN-HOME-003: Home quick actions SHALL route to correct workflows
Each quick action SHALL navigate to the correct destination with preserved context.

#### Scenario: Quick action navigation
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user clicks a Home quick action
- **THEN** app SHALL open the expected screen/workflow context

### Requirement UI-SCREEN-HOME-004: Home onboarding rail SHALL expose deterministic starter actions
Home SHALL provide explicit starter actions for first-run workflows from the onboarding panel.

#### Scenario: Starter onboarding actions available
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home renders the starter onboarding panel
- **THEN** users SHALL see `Start Setup`, `Import Existing Collection`, and `Use Sample Data` actions

#### Scenario: Starter onboarding step navigation controls
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user interacts with onboarding step controls
- **THEN** `Back Step` and `Next Step` controls SHALL move onboarding state deterministically

### Requirement UI-SCREEN-HOME-005: Home toolbar SHALL support explicit refresh action
Home SHALL expose a `Refresh Dashboard` action that re-fetches Home data and preserves route context.

#### Scenario: Refresh dashboard action
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user clicks `Refresh Dashboard`
- **THEN** Home data SHALL refresh and render updated panel state without route transition

## Acceptance Criteria
- Every Home critical flow has UC ID and deterministic expected outcome.
- E2E mapping exists for command-center render and quick actions.
- Starter onboarding actions and refresh behavior are explicitly covered by requirements.

## Success Criteria
- Users can identify and start next action from Home in one interaction.
- No placeholder-only content remains in Home ready state.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-HOME-01 | Open Home with active data | Actionable priority cards render | planned: `cypress/e2e/ui/home.cy.ts` `home-renders-priority-cards` |
| UC-HOME-02 | Open Home with no pending actions | Calm empty state with CTA renders | planned: `cypress/e2e/ui/home.cy.ts` `home-empty-state` |
| UC-HOME-03 | Home data fetch failure | Inline error + retry appears | planned: `cypress/e2e/ui/home.cy.ts` `home-error-retry` |
| UC-HOME-04 | Trigger quick action | Correct navigation/context opens | planned: `cypress/e2e/ui/home.cy.ts` `home-quick-action-routing` |
| UC-HOME-05 | Use starter onboarding actions | Setup/import/sample actions are visible and actionable | planned: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `home-starter-onboarding-actions` |
| UC-HOME-06 | Click Refresh Dashboard | Home data re-fetches without route transition | planned: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `home-refresh-dashboard` |
| UC-HOME-07 | Navigate onboarding steps | Back/Next step controls update onboarding state deterministically | planned: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `home-onboarding-step-nav` |
