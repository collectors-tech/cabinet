## Purpose
Define Market Watch screen behavior for provider-scoped query set creation, execution, and result triage.

## Requirements
### Requirement UI-SCREEN-MARKET-WATCH-001: Market Watch SHALL expose provider selector and saved-query lifecycle controls
Market Watch SHALL require selecting at least one provider before creating a query set, and SHALL let users edit and delete saved query sets without losing provider scope metadata.

#### Scenario: Create provider-scoped query set
- **GIVEN** user is on `/market-watch`
- **WHEN** user enters query name/keywords and selects provider scope
- **THEN** query set MUST persist with provider scope metadata

#### Scenario: Edit and delete provider-scoped query set
- **GIVEN** a provider-scoped query set exists
- **WHEN** user edits name, keywords, or schedule and saves
- **THEN** query set MUST persist the edited fields while preserving provider scope metadata
- **AND WHEN** user deletes the saved query set
- **THEN** the query set MUST be removed from the visible Market Watch list

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
- **AND WHEN** retry is requested for the failed query set
- **THEN** UI MUST show retry-request feedback, reload durable query-set/failure state, clear stale failure entries when recovered, and show the recovered latest-run snapshot

### Requirement UI-SCREEN-MARKET-WATCH-004: Market Watch SHALL support deterministic states
Screen SHALL support loading, empty, ready, and error states with retry controls.

### Requirement UI-SCREEN-MARKET-WATCH-005: Market Watch SHALL provide query-set table view with run output visibility
Market Watch SHALL provide a table view for query sets so users can find saved queries quickly and inspect latest run outputs.

### Requirement UI-SCREEN-MARKET-WATCH-006: Market Watch SHALL support saved watched-query execution for Bonza `AFX`
Market Watch SHALL support creating and running provider-scoped watched query `AFX` for Bonza and surface aggregated cross-page candidate outputs.

#### Scenario: Create watched AFX query for Bonza
- **GIVEN** user is on `/market-watch`
- **WHEN** user creates query name `AFX` with provider scope `Bonza`
- **THEN** query set MUST persist with watched flag and provider scope metadata

#### Scenario: Run watched AFX query
- **GIVEN** watched Bonza query `AFX` exists
- **WHEN** user clicks `Run Now`
- **THEN** Market Watch MUST display run output summary including total pages scanned and candidate count

#### Scenario: Query table inspection
- **GIVEN** query sets exist
- **WHEN** user switches to `Table` view
- **THEN** table MUST list query name, provider scope, last run status, last run time, and latest output summary
- **AND** last run status/time/output summary MUST hydrate from durable query-set metadata after reload, not only transient in-memory run state
- **AND** scheduled refresh completion MUST reload the durable query-set snapshots and update the latest-run history summary without requiring a manual browser reload

#### Scenario: Open run output details from table
- **GIVEN** query row has prior run output
- **WHEN** user opens output detail action from table row
- **THEN** UI MUST show deterministic run output details for testing/verification
- **AND** output detail view MUST include provider attribution and run timestamp
- **AND** Wishlist handoff from the output detail MUST persist enough state for the Wishlist route to render the handed-off result after reload

#### Scenario: No query sets yet
- **GIVEN** no query sets exist
- **WHEN** screen loads
- **THEN** empty state SHALL guide user to create first provider-scoped query set

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-MW-01 | Create query set with provider scope | Query set persists with provider metadata | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 creates provider-scoped query sets from selector controls` |
| UC-MW-02 | Run scoped query | Runtime payload and results are provider-scoped | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-002 sends provider scope in run payload and shows provider-attributed results`; `UI-SCREEN-MARKET-WATCH-002 runs eBay-only saved searches through the provider route` |
| UC-MW-03 | Handle run failure | Human-readable error + retry shown, then durable failure/query-set state reloads after recovery | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-003 surfaces run failure guidance and retry action` |
| UC-MW-04 | Empty state | Create-first guidance shown | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-empty-state` |
| UC-MW-05 | Query-set table review | Table shows saved queries with durable status/time/output summary columns across reloads and scheduled-refresh snapshot reloads | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 renders query table view with saved-query columns for rapid inspection`; `UI-SCREEN-MARKET-WATCH-005 refreshes table run history after scheduled refresh`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-005 runs saved searches now and through scheduled refresh`; `TestScannerRunItemsPerPageSummaryAppliesSafeCap`; `TestDefaultSiteSearchScheduledRefreshPersistsRunSnapshot` |
| UC-MW-06 | Inspect run outputs from table | Row action opens latest run output detail for verification and handoff persistence | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 opens deterministic output details from query-table row action`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-006 hands off saved-search output to discoveries and persisted wishlist flows` |
| UC-MW-07 | Create Bonza watched query AFX | Query persists with provider scope=Bonza and watched metadata | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-create-bonza-afx-query` |
| UC-MW-08 | Run Bonza watched query AFX | Output summary shows page-scan count + aggregated candidate count | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-run-bonza-afx-summary` |
| UC-MW-09 | Edit and delete provider-scoped query set | Edited name/keywords/schedule persist while provider scope remains intact, then delete removes the saved query | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 manages saved-query create edit and delete lifecycle` |
