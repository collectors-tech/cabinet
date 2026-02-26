## Purpose
Define Reports screen behavior for analytics and export workflows.

## Requirements
### Requirement: Reports SHALL render wishlist and pricing summaries
Reports SHALL provide summary metrics for wishlist hits, trends, stats, and sources.

#### Scenario: Reports ready state
- **WHEN** reports data loads
- **THEN** summary panels SHALL render expected analytics output

### Requirement: Reports SHALL support export operations
Reports SHALL allow export of report/pricing history outputs.

#### Scenario: Export report output
- **WHEN** user triggers export
- **THEN** export payload SHALL be generated for selected scope

### Requirement: Reports SHALL support deterministic state handling
Reports SHALL support loading, empty, error, and ready states.

#### Scenario: Reports empty state
- **WHEN** no historical data exists
- **THEN** reports SHALL show empty guidance and next action

## Acceptance Criteria
- UC IDs cover analytics load, export, and state transitions.
- E2E mapping includes export behavior.

## Success Criteria
- Reports provide actionable summary and export without runtime errors.
- Users can understand absence of data from clear empty states.

## Data Profiles
- Sample: 12 months pricing history for 50 tracked items
- Bulk: 24 months pricing history for 2,000 tracked items

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-REP-01 | Load reports | Summary panels render | planned: `cypress/e2e/ui/reports.cy.ts` `reports-ready` |
| UC-REP-02 | Export data | Export output generated | planned: `cypress/e2e/ui/reports.cy.ts` `reports-export` |
| UC-REP-03 | No report data | Empty state guidance appears | planned: `cypress/e2e/ui/reports.cy.ts` `reports-empty-state` |
| UC-REP-04 | Reports API failure | Error + retry appears | planned: `cypress/e2e/ui/reports.cy.ts` `reports-error-state` |
