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

#### Scenario: Load workspace state
- **GIVEN** user opens `/market-watch`
- **WHEN** query-set, failure, and provider-health data is loading
- **THEN** UI MUST show a deterministic loading state

#### Scenario: Empty saved-query setup
- **GIVEN** query-set loading succeeds with no saved queries
- **WHEN** screen renders the workspace
- **THEN** UI MUST show create-first guidance for the first provider-scoped query set

#### Scenario: Provider needs attention
- **GIVEN** provider health reports a non-`ok` status
- **WHEN** Market Watch renders saved-query controls
- **THEN** UI MUST show provider/auth attention guidance with the reported health status and recovery direction

#### Scenario: Load failure retry
- **GIVEN** query-set, failure, or provider-health loading fails
- **WHEN** the error state renders
- **THEN** UI MUST show a retry control that reloads Market Watch workspace data

#### Scenario: Output detail with no result rows
- **GIVEN** a saved query has latest run metadata but no visible output rows
- **WHEN** user opens output details from the query table
- **THEN** UI MUST show an explicit no-output state with direction to run the query or adjust provider scope

### Requirement UI-SCREEN-MARKET-WATCH-005: Market Watch SHALL provide query-set table view with run output visibility
Market Watch SHALL provide a table view for query sets so users can find saved queries quickly, inspect latest run outputs, and create new watched queries from the page toolbar.

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
- **THEN** table MUST list query name, query terms, provider scope, schedule/manual state, latest status/error, last run time, result count, and row actions
- **AND** last run status/time/output summary MUST hydrate from durable query-set metadata after reload, not only transient in-memory run state
- **AND** scheduled refresh completion MUST reload the durable query-set snapshots and update the latest-run history summary without requiring a manual browser reload

#### Scenario: Create query from table toolbar
- **GIVEN** user is on `/market-watch`
- **WHEN** user enters new query details and activates the toolbar `+` create action
- **THEN** the query set MUST be created with the selected provider scope
- **AND** the saved query MUST appear in the Market Watch query list

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

### Requirement UI-SCREEN-MARKET-WATCH-007: Market Watch SHALL filter query table rows by operational state
Market Watch SHALL provide query table filters for provider, latest run status, schedule mode, attention state, and result presence so users can narrow large saved-query workspaces without mutating saved query state.

#### Scenario: Filter query table rows
- **GIVEN** saved Market Watch queries exist with different providers, schedule modes, latest statuses, and result counts
- **WHEN** user applies provider, status, schedule, attention, or result filters
- **THEN** Market Watch MUST update the visible query rows and latest-run history summary to only include matching query sets
- **AND** saved query definitions MUST remain unchanged

#### Scenario: Filter no-match recovery
- **GIVEN** saved Market Watch queries exist
- **WHEN** current filters match no query rows
- **THEN** Market Watch MUST show a no-matches state with a reset action
- **AND WHEN** user resets filters
- **THEN** all saved query rows MUST be visible again

### Requirement UI-SCREEN-MARKET-WATCH-008: Market Watch SHALL show output result provenance and handoff state
Market Watch output details SHALL render latest result items in a structured table or panel that exposes provenance and handoff state for each result.

#### Scenario: Inspect output result provenance
- **GIVEN** a Market Watch query has latest output items
- **WHEN** user opens output details from the query table
- **THEN** each result row MUST show provider/source, listing or item title, price/currency when available, source URL or listing identifier when available, stock/status when available, and handoff state
- **AND** existing output-detail handoff actions MUST remain available

#### Scenario: Inspect output with no result items
- **GIVEN** a Market Watch query has no latest output items
- **WHEN** user opens output details from the query table
- **THEN** Market Watch MUST show an explicit no-output state instead of an empty table

### Requirement UI-SCREEN-MARKET-WATCH-009: Market Watch SHALL persist output-detail Wishlist handoff provenance
Market Watch output-detail Wishlist handoff SHALL persist the selected result to Wishlist with Market Watch, provider, query, and scope provenance visible after downstream route reload.

#### Scenario: Persist Wishlist handoff from output details
- **GIVEN** a Market Watch output detail has at least one result row
- **WHEN** user activates `Add First Result to Wishlist`
- **THEN** UI MUST post the selected candidate through the durable discovery action with Market Watch query provenance
- **AND** UI MUST show a testable handoff success state for the selected candidate
- **AND WHEN** user opens or reloads `/wishlist`
- **THEN** the Wishlist route MUST render the handed-off result and its Market Watch/provider/query provenance

### Requirement UI-SCREEN-MARKET-WATCH-010: Market Watch SHALL persist output-detail Inventory handoff provenance
Market Watch output-detail Inventory handoff SHALL persist the selected result to Inventory with Market Watch, provider, query, source URL, and scope provenance visible after downstream route reload.

#### Scenario: Persist Inventory handoff from output details
- **GIVEN** a Market Watch output detail has at least one result row
- **WHEN** user activates `Add First Result to Inventory`
- **THEN** UI MUST post the selected candidate through the durable discovery action with Market Watch query provenance
- **AND** UI MUST show a testable handoff success state for the selected candidate
- **AND WHEN** user opens or reloads `/inventory`
- **THEN** the Inventory route MUST render the handed-off result and its Market Watch/provider/query/source provenance

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-MW-01 | Create query set with provider scope | Query set persists with provider metadata from either the form submit action or toolbar `+` create action | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 creates provider-scoped query sets from selector controls` |
| UC-MW-02 | Run scoped query | Runtime payload and results are provider-scoped | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-002 sends provider scope in run payload and shows provider-attributed results`; `UI-SCREEN-MARKET-WATCH-002 runs eBay-only saved searches through the provider route` |
| UC-MW-03 | Handle run failure | Human-readable error + retry shown, then durable failure/query-set state reloads after recovery | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-003 surfaces run failure guidance and retry action` |
| UC-MW-04 | Deterministic workspace states | Loading, empty, provider/auth attention, API load failure retry, and no-output detail states are explicit and testable | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-004 shows deterministic workspace states`; `UI-SCREEN-MARKET-WATCH-004 shows load failure with retry recovery`; `UI-SCREEN-MARKET-WATCH-004 keeps no-output detail state explicit` |
| UC-MW-05 | Query-set table review | Table shows saved queries with terms, provider scope, schedule/manual state, durable status/error, time, result count, and action columns across reloads and scheduled-refresh snapshot reloads | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 renders query table view with saved-query columns for rapid inspection`; `UI-SCREEN-MARKET-WATCH-005 refreshes table run history after scheduled refresh`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-005 runs saved searches now and through scheduled refresh`; `TestScannerRunItemsPerPageSummaryAppliesSafeCap`; `TestDefaultSiteSearchScheduledRefreshPersistsRunSnapshot` |
| UC-MW-06 | Inspect run outputs from table | Row action opens latest run output detail for verification and handoff persistence | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 opens deterministic output details from query-table row action`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-006 hands off saved-search output to discoveries and persisted wishlist flows` |
| UC-MW-07 | Create Bonza watched query AFX | Query persists with provider scope=Bonza and watched metadata | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-create-bonza-afx-query` |
| UC-MW-08 | Run Bonza watched query AFX | Output summary shows page-scan count + aggregated candidate count | planned: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `market-watch-run-bonza-afx-summary` |
| UC-MW-09 | Edit and delete provider-scoped query set | Edited name/keywords/schedule persist while provider scope remains intact, then delete removes the saved query | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 manages saved-query create edit and delete lifecycle` |
| UC-MW-10 | Filter query table and history | Provider/status/schedule/attention/result filters narrow rows and history, with no-match reset recovery | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-007 filters query table rows by provider status schedule attention and result state` |
| UC-MW-11 | Inspect output result provenance | Output detail table shows provider, title, price/currency, source identifier, stock/status, and handoff state while preserving handoff actions | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-008 shows output result provenance and handoff state` |
| UC-MW-12 | Wishlist handoff from output detail | Output detail Wishlist handoff posts selected candidate, reports success, and persists Market Watch provenance to the reloaded Wishlist route | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-009 persists output-detail Wishlist handoff provenance` |
| UC-MW-13 | Inventory handoff from output detail | Output detail Inventory handoff posts selected candidate, reports success, and persists Market Watch provenance to the reloaded Inventory route | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-010 persists output-detail Inventory handoff provenance` |
