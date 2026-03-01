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

## Acceptance Criteria
- Every Home critical flow has UC ID and deterministic expected outcome.
- E2E mapping exists for command-center render and quick actions.

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
