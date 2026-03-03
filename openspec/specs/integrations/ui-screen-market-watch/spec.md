## Purpose
Define Market Watch screen behavior for provider-scoped query set creation, execution, and result triage.

## Requirements
### Requirement UI-SCREEN-MARKET-WATCH-001: Market Watch SHALL expose provider selector for query creation
Market Watch SHALL require selecting at least one provider before creating a query set.

#### Scenario: Create provider-scoped query set
- **GIVEN** user is on `/market-watch`
- **WHEN** user enters query name/keywords and selects provider scope
- **THEN** query set MUST persist with provider scope metadata

### Requirement UI-SCREEN-MARKET-WATCH-002: Market Watch SHALL execute runs scoped to selected provider(s)
`Run Now` MUST execute only against query set provider scope.

#### Scenario: Run provider-scoped query
- **GIVEN** query set has provider scope
- **WHEN** user clicks `Run Now`
- **THEN** runtime MUST submit provider-scoped request payload and return provider-attributed results

### Requirement UI-SCREEN-MARKET-WATCH-003: Market Watch SHALL render provider-attributed result/error states
Results and errors SHALL clearly indicate source provider and actionable recovery.

#### Scenario: Provider run failure messaging
- **GIVEN** provider run fails
- **WHEN** failure response is returned
- **THEN** UI MUST show human-readable error guidance (not raw error keys) with retry path

### Requirement UI-SCREEN-MARKET-WATCH-004: Market Watch SHALL support deterministic states
Screen SHALL support loading, empty, ready, and error states with retry controls.

### Requirement UI-SCREEN-MARKET-WATCH-005: Market Watch SHALL provide query-set table view with run output visibility
Market Watch SHALL provide a table view for query sets so users can find saved queries quickly and inspect latest run outputs.

#### Scenario: Query table inspection
- **GIVEN** query sets exist
- **WHEN** user switches to `Table` view
- **THEN** table MUST list query name, provider scope, last run status, last run time, and latest output summary

#### Scenario: Open run output details from table
- **GIVEN** query row has prior run output
- **WHEN** user opens output detail action from table row
- **THEN** UI MUST show deterministic run output details for testing/verification

#### Scenario: No query sets yet
- **GIVEN** no query sets exist
- **WHEN** screen loads
- **THEN** empty state SHALL guide user to create first provider-scoped query set

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-MW-01 | Create query set with provider scope | Query set persists with provider metadata | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-create-provider-scoped-query` |
| UC-MW-02 | Run scoped query | Runtime payload and results are provider-scoped | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-run-provider-scoped` |
| UC-MW-03 | Handle run failure | Human-readable error + retry shown | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-run-failure-guidance` |
| UC-MW-04 | Empty state | Create-first guidance shown | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-empty-state` |
| UC-MW-05 | Query-set table review | Table shows saved queries with status/time/output summary columns | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-query-table` |
| UC-MW-06 | Inspect run outputs from table | Row action opens latest run output detail for verification | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-output-detail` |
