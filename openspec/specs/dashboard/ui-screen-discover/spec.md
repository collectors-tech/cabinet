## Purpose
Define Discover screen behavior for not-in-collection triage actions.

## Requirements
### Requirement UI-SCREEN-DISCOVER-001: Discover SHALL support filterable candidate triage
Discover SHALL support query/price/date filtering and list rendering.

#### Scenario: Filtered triage list
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user applies filters
- **THEN** discover list SHALL render filtered candidates

#### Scenario: Apply Filters control triggers deterministic query update
- **GIVEN** discover filter inputs are populated
- **WHEN** user clicks `Apply Filters`
- **THEN** screen MUST execute filtered query and refresh candidate list without route transition

### Requirement UI-SCREEN-DISCOVER-002: Discover SHALL support all primary candidate actions
Discover SHALL support ignore, wishlist, track, and create-item actions.

#### Scenario: Candidate action apply
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user chooses action on candidate
- **THEN** candidate state and downstream linkage SHALL update

### Requirement UI-SCREEN-DISCOVER-003: Discover SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states.

#### Scenario: Discover error state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** discover API request fails
- **THEN** screen SHALL present actionable retry state

## Acceptance Criteria
- UC IDs cover filtering and each primary action class.
- E2E mappings include action outcomes.

## Success Criteria
- Discover triage can be completed without route/context loss.
- High-volume candidate lists remain manageable.

## Data Profiles
- Sample: 50 candidates
- Bulk: 10,000 candidates

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-DIS-01 | Filter discover list | Filtered candidates displayed | planned: `cypress/e2e/ui/discover.cy.ts` `discover-filtering` |
| UC-DIS-02 | Ignore action | Candidate state updates to ignored | planned: `cypress/e2e/ui/discover.cy.ts` `discover-ignore` |
| UC-DIS-03 | Add to wishlist action | Wishlist linkage created | planned: `cypress/e2e/ui/discover.cy.ts` `discover-wishlist` |
| UC-DIS-04 | Track/create action | Pricing track or item create executes | planned: `cypress/e2e/ui/discover.cy.ts` `discover-track-create` |
| UC-DIS-05 | Discover API failure | Error + retry appears | planned: `cypress/e2e/ui/discover.cy.ts` `discover-error-state` |
| UC-DIS-06 | Apply Filters action | `Apply Filters` triggers deterministic filtered-query refresh | planned: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `discover-apply-filters` |
