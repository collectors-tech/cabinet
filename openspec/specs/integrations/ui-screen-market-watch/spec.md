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

#### Scenario: Friendly saved-watch cadence controls
- **GIVEN** user creates or edits a saved Market Watch query
- **WHEN** user chooses a friendly cadence such as manual, hourly, every 6 hours, daily, or weekly
- **THEN** Market Watch MUST persist the corresponding schedule metadata without requiring raw cron syntax in the default flow
- **AND WHEN** user pauses the watch while editing
- **THEN** the saved watch MUST persist as disabled and show the paused cadence state in the visible list
- **AND** custom cron entry MAY remain available as an advanced compatibility path

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

#### Scenario: Activate saved-watch rows and table actions
- **GIVEN** saved Market Watch query rows are visible in table view
- **WHEN** user double-clicks a row or focuses the row and presses Enter
- **THEN** Market Watch MUST open the saved watch output/detail panel for that row
- **AND WHEN** user activates table row actions for run now, pause/resume, edit, inspect output, or delete
- **THEN** each action MUST use the same saved-query contracts as the card view without requiring raw field editing

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

### Requirement UI-SCREEN-MARKET-WATCH-015: Market Watch SHALL persist output-detail Purchase handoff provenance
Market Watch output-detail purchase handoff SHALL persist the selected result as a purchase lifecycle entry with Market Watch, provider, query, source URL, price, currency, and scope provenance.

#### Scenario: Persist Purchase handoff from output details
- **GIVEN** a Market Watch output detail has at least one result row
- **WHEN** user activates `Mark First Result Purchased`
- **THEN** UI MUST post the selected candidate through the durable discovery action with Market Watch query provenance
- **AND** backend MUST create or update a profile-scoped purchase lifecycle entry and expected-arrival record for the matched or newly created item
- **AND** the selected result status MUST persist as `purchase_candidate` so Market Watch and Discoveries stop treating it as an unresolved new result unless filtered
- **AND WHEN** user opens or reloads `/purchases`
- **THEN** the Purchases route MUST render the handed-off purchase with provider/listing/query provenance available to the row or detail view

### Requirement UI-SCREEN-MARKET-WATCH-016: Market Watch SHALL expose a filterable result inbox lifecycle surface
Market Watch output details SHALL render the latest provider results as a result inbox with visible match context, lifecycle status, and filters that keep downstream decisions discoverable instead of hiding old or acted-on results.

#### Scenario: Filter result inbox by lifecycle and match provenance
- **GIVEN** a Market Watch query has latest output results with provider, matched watch, match target, first/last seen timestamps, lifecycle status, and wishlist-match metadata
- **WHEN** user opens output details from the query table
- **THEN** Market Watch MUST show a result inbox summary and table columns for provider, result, matched watch, match target/reason, price, shipping, total price, source, first seen, last seen, stock, status, and handoff state
- **AND WHEN** user filters by result status, provider, match target, or wishlist matches
- **THEN** the visible inbox rows MUST narrow without mutating result decision state
- **AND** dismissed, duplicate, expired, purchased, wishlist, and inventory candidate states MUST remain reachable through filters

#### Scenario: Persist result inbox lifecycle decisions through API reload
- **GIVEN** a Market Watch query has persisted provider results for the active profile
- **WHEN** the result inbox API lists candidates with status/provider filters and pagination
- **THEN** the API MUST return the matching result rows plus total, page, and page-size metadata without dropping source attribution
- **AND WHEN** a user decision updates a result lifecycle status such as dismissed, purchased, added to wishlist, added to inventory, watching, duplicate, expired, or failed to refresh
- **THEN** the status update MUST be profile-scoped, persisted on the candidate record, and visible through a subsequent filtered API reload

### Requirement UI-SCREEN-MARKET-WATCH-017: Market Watch SHALL expose provider health and persisted run history with recovery guidance
Market Watch SHALL show provider health with an explanation and next action, and SHALL list persisted provider run records so users can inspect run outcomes after reload.

#### Scenario: Unknown provider health remains actionable
- **GIVEN** provider health has not been checked for a Market Watch provider
- **WHEN** Market Watch renders the provider health strip
- **THEN** the screen MUST show that the provider has not been checked yet
- **AND** it MUST show a next step for collecting health evidence instead of displaying bare `unknown`

#### Scenario: Provider health taxonomy is consistent and actionable
- **GIVEN** provider health reports healthy, not checked, setup-required, reauthentication, rate-limited, provider-unavailable, failed, or partial-failure outcomes
- **WHEN** Market Watch renders provider health and provider attention states
- **THEN** the API response MUST include a stable category, user-facing label, and recovery guidance for the outcome
- **AND** the UI MUST show that label and guidance rather than relying on raw provider status strings alone
- **AND** rate-limited states MUST preserve retry timing when supplied
- **AND** partial-failure states MUST explain that successful provider results remain available while failed provider details need review

#### Scenario: Persisted provider run history is visible
- **GIVEN** one or more Market Watch runs have been persisted for the active profile
- **WHEN** Market Watch loads run history
- **THEN** the screen MUST show provider, trigger type, status, finished time, total result count, new result count, and failure or retry guidance for each recent run
- **AND** failed runs MUST keep provider-specific recovery guidance visible without marking unrelated provider runs as failed

#### Scenario: Provider health transitions persist after run outcomes
- **GIVEN** an eBay Market Watch run fails due to missing setup, invalid credentials, rate limiting, or provider outage
- **WHEN** the scanner records provider health for the run
- **THEN** the persisted provider health status MUST distinguish setup-required, reauthentication, rate-limited, and provider-unavailable outcomes
- **AND** retry timing MUST remain available for rate-limited responses
- **AND WHEN** a later run succeeds
- **THEN** provider health MUST return to healthy state and clear stale failure guidance

### Requirement UI-SCREEN-MARKET-WATCH-011: Market Watch SHALL bootstrap saved-query creation from route handoff state
Market Watch SHALL translate route handoff context into editable saved-query fields before persistence so handoffs from barcode/search surfaces remain deterministic.

#### Scenario: Route barcode handoff creates saved query
- **GIVEN** user opens `/scanner/?barcode=<barcode>`
- **WHEN** Market Watch loads
- **THEN** query name and keyword fields MUST be prefilled from the barcode
- **AND** route handoff guidance MUST explain that provider scope should be reviewed before creation
- **AND WHEN** user creates the query
- **THEN** the saved query MUST persist the prefilled barcode keyword with the selected provider scope

### Requirement UI-SCREEN-MARKET-WATCH-012: Market Watch SHALL preserve output-detail Discoveries handoff context
Market Watch output-detail Discoveries handoff SHALL use the saved watch keyword context when requesting Discoveries candidates and SHALL expose a testable handoff result state.

#### Scenario: Handoff output detail context to Discoveries
- **GIVEN** a Market Watch output detail is open for a saved watch with keywords
- **WHEN** user activates `Open Discoveries Handoff`
- **THEN** UI MUST request Discoveries candidates using the saved watch keyword as the query
- **AND** UI MUST show a testable Discoveries handoff status with the returned item count

### Requirement UI-SCREEN-MARKET-WATCH-013: Market Watch SHALL persist run and result records within profile and watch scope
Market Watch SHALL store saved-watch executions as durable run records and SHALL dedupe discovered results within the current profile and saved watch so private watch state, result decisions, and rerun metadata survive reloads without leaking across profiles.

#### Scenario: Durable run record for saved watch execution
- **GIVEN** a saved Market Watch query exists for the active profile
- **WHEN** the query is run manually or by schedule
- **THEN** the backend MUST create a durable run record with run ID, profile ID, query-set ID, provider, trigger type, started/finished timestamps, status, result count, new-result count, and actionable error/retry metadata when the run fails
- **AND** saved-query reloads MUST derive latest run status, latest run timestamp, latest run message, candidate count, and computed next scheduled run timestamp from durable run/watch state rather than transient in-memory execution state

#### Scenario: Scheduled partial failure preserves unrelated watch state
- **GIVEN** multiple enabled scheduled Market Watch queries exist for the active profile
- **WHEN** one provider-backed scheduled query fails and another scheduled query succeeds in the same scheduled refresh
- **THEN** the failed query MUST store a durable failed scheduled run with actionable retry guidance
- **AND** the successful query MUST still run, store durable result records, and reload as succeeded without being corrupted by the failed query

#### Scenario: Scoped result dedupe preserves user decisions
- **GIVEN** a provider returns the same listing ID or source URL across repeated runs
- **WHEN** Market Watch persists the run results
- **THEN** results MUST dedupe within the same profile, saved watch, provider, and listing/source URL key
- **AND** existing user decision status such as ignored, archived, wishlisted, purchase candidate, or inventory candidate MUST be preserved while refreshed result metadata is updated
- **AND** preserved decisions MUST remain available to downstream Discoveries, Wishlist, Purchases, and Inventory views through stable candidate status and provenance fields after reload or rerun
- **AND** the same provider listing may be tracked independently by another profile or another saved watch without corrupting the original watch state

#### Scenario: Result lifecycle details expose decision history
- **GIVEN** a Market Watch result has a lifecycle transition such as new to dismissed or new to purchased
- **WHEN** the result is updated through the candidate lifecycle API or reloaded in the result inbox
- **THEN** the transition MUST be recorded as durable profile-scoped decision history with previous status, next status, reason, and timestamp
- **AND** the Market Watch output details MUST expose the latest decision history so dismissed, purchased, wishlist, and inventory handoff decisions remain auditable after reload

### Requirement UI-SCREEN-MARKET-WATCH-014: Market Watch SHALL keep scanner capture secondary to saved searches
Market Watch SHALL focus its primary dashboard on saved provider searches while keeping scanner/manual listing capture available behind a secondary action.

#### Scenario: Reveal manual listing capture as a secondary action
- **GIVEN** user is on `/market-watch`
- **WHEN** Market Watch first renders
- **THEN** Quick Scan, manual card entry, and Recent Unlinked Scans controls MUST NOT occupy the main dashboard body
- **AND** the page MUST expose a secondary `Add listing manually` action
- **AND WHEN** user activates the secondary action
- **THEN** Market Watch MUST reveal Quick Scan, manual entry queueing, and Recent Unlinked Scans review controls without disrupting saved-query creation

### Requirement UI-SCREEN-MARKET-WATCH-018: Market Watch SHALL present a saved integration search dashboard
Market Watch SHALL frame the primary route as a collector-facing saved integration search dashboard instead of a raw query-set control panel.

#### Scenario: Saved integration search dashboard shell
- **GIVEN** user opens `/market-watch`
- **WHEN** Market Watch workspace data loads
- **THEN** the page header MUST explain saved searches across integrations, discoveries, and provider recovery
- **AND** the top summary MUST show active watches, new discoveries, wishlist matches, provider issues, last run, and next run
- **AND** the primary create controls MUST use saved-watch/search terminology rather than raw query-set terminology
- **AND** provider health, saved-watch table controls, result inbox controls, and secondary manual listing capture MUST be visible or reachable from the same dashboard surface

### Requirement UI-SCREEN-MARKET-WATCH-019: Beta Market Watch providers SHALL fail closed until live proof exists
Market Watch provider registry projection SHALL distinguish live-validated, setup-required, beta-limited, manual-capture-only, and disabled providers so beta users are not shown unsupported or unproven provider paths as connected production-ready integrations.

#### Scenario: Beta provider registry status fails closed
- **GIVEN** Cabinet is preparing the 0.1 beta Market Watch provider proof path
- **WHEN** `/api/providers/registry` projects Market Watch-capable providers
- **THEN** eBay MUST remain `setup_required` with unavailable Market Watch actions until credentials and live proof are present
- **AND** disabled or unsupported providers MUST expose disabled actions and a next safe provider-selection action
- **AND** public storefront providers without attached live proof MUST declare `beta_limited` or `manual_url_capture_only` status rather than `available_live_validated`
- **AND** live evidence state MUST be visible in the registry response for release evidence review

#### Scenario: Bonza live proof upgrades beta provider status
- **GIVEN** a Bonza-scoped saved Market Watch query runs successfully against the public Store API
- **WHEN** Cabinet persists normalized candidates and records provider health for the Bonza registry provider
- **THEN** `/api/providers/registry` MUST project Bonza as `available_live_validated`
- **AND** Bonza MUST expose `live_evidence_state=validated` and an available `market_watch.run` action
- **AND** the release evidence MUST remain non-secret and identify the access method, rate-limit/cache boundary, observed result provenance, and log path

#### Scenario: BigCommerce HTML storefront live proof upgrades beta provider status
- **GIVEN** a Voglers-scoped saved Market Watch query runs successfully against the public BigCommerce storefront search page
- **WHEN** Cabinet parses storefront HTML product cards, persists normalized candidates, and records provider health for the Voglers registry provider
- **THEN** `/api/providers/registry` MUST project Voglers as `available_live_validated`
- **AND** Voglers MUST expose `live_evidence_state=validated` and an available `market_watch.run` action
- **AND** the proof path MUST stay public storefront only without login, cart, checkout, payment, or private/admin API use

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-MW-01 | Create query set with provider scope | Query set persists with provider metadata from either the form submit action or toolbar `+` create action | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 creates provider-scoped query sets from selector controls`; `UI-SCREEN-MARKET-WATCH-001 + #1542 manages saved-watch create edit cadence pause and delete lifecycle` |
| UC-MW-02 | Run scoped query | Runtime payload and results are provider-scoped | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-002 sends provider scope in run payload and shows provider-attributed results`; `UI-SCREEN-MARKET-WATCH-002 runs eBay-only saved searches through the provider route` |
| UC-MW-03 | Handle run failure | Human-readable error + retry shown, then durable failure/query-set state reloads after recovery | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-003 surfaces run failure guidance and retry action` |
| UC-MW-04 | Deterministic workspace states | Loading, empty, provider/auth attention, API load failure retry, and no-output detail states are explicit and testable | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-004 shows deterministic workspace states`; `UI-SCREEN-MARKET-WATCH-004 shows load failure with retry recovery`; `UI-SCREEN-MARKET-WATCH-004 keeps no-output detail state explicit` |
| UC-MW-05 | Query-set table review | Table shows saved queries with terms, provider scope, schedule/manual state, durable status/error, time, result count, and action columns across reloads and scheduled-refresh snapshot reloads | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 renders query table view with saved-query columns for rapid inspection`; `UI-SCREEN-MARKET-WATCH-005 refreshes table run history after scheduled refresh`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-005 runs saved searches now and through scheduled refresh`; `TestScannerRunItemsPerPageSummaryAppliesSafeCap`; `TestDefaultSiteSearchScheduledRefreshPersistsRunSnapshot` |
| UC-MW-06 | Inspect run outputs from table | Row action opens latest run output detail for verification and handoff persistence | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-005 opens deterministic output details from query-table row action`; `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts` `DEFAULT-SITE-SEARCH-006 hands off saved-search output to discoveries and persisted wishlist flows` |
| UC-MW-07 | Create Bonza watched query AFX | Query persists with provider scope=Bonza and watched metadata | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-006 runs Bonza AFX query and surfaces aggregated run summary` |
| UC-MW-08 | Run Bonza watched query AFX | Output summary shows page-scan count + aggregated candidate count | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-006 runs Bonza AFX query and surfaces aggregated run summary` |
| UC-MW-09 | Edit and delete provider-scoped query set | Edited name/keywords/schedule persist while provider scope remains intact, then delete removes the saved query | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-001 manages saved-query create edit and delete lifecycle` |
| UC-MW-10 | Filter query table and history | Provider/status/schedule/attention/result filters narrow rows and history, with no-match reset recovery | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-007 filters query table rows by provider status schedule attention and result state` |
| UC-MW-11 | Inspect output result provenance | Output detail table shows provider, title, price/currency, source identifier, stock/status, and handoff state while preserving handoff actions | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-008 shows output result provenance and handoff state` |
| UC-MW-12 | Wishlist handoff from output detail | Output detail Wishlist handoff posts selected candidate, reports success, and persists Market Watch provenance to the reloaded Wishlist route | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-009 persists output-detail Wishlist handoff provenance` |
| UC-MW-13 | Inventory handoff from output detail | Output detail Inventory handoff posts selected candidate, reports success, and persists Market Watch provenance to the reloaded Inventory route | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-010 persists output-detail Inventory handoff provenance` |
| UC-MW-14 | Route handoff bootstrap | Barcode handoff route pre-fills saved-query fields and persists the selected provider scope when created | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-011 creates saved query from route barcode handoff state` |
| UC-MW-15 | Discoveries handoff from output detail | Output detail Discoveries handoff queries Discoveries with the saved watch keyword and reports returned item count | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-012 hands output-detail context to Discoveries with saved-watch keyword` |
| UC-MW-16 | Persist saved-watch run/result records | Manual and scheduled runs create durable run records; scheduled partial failures preserve unrelated watch state; results dedupe by profile/watch/provider/listing or source URL while preserving decision state; saved-query reloads expose durable latest-run and next scheduled run state | implemented: `internal/scanner/service_test.go` `TestRunNowPersistsDurableRunRecordAndDedupesResults`; `TestRunNowPreservesDownstreamDecisionStatuses`; `TestRunNowDedupeIsScopedByProfileAndWatch`; `TestQuerySetRunSnapshotUsesDurableRunRecordsAndComputesNextRun`; `TestRunScheduledRecordsPartialFailureWithoutBlockingOtherWatches` |
| UC-MW-17 | Reveal manual listing capture | Quick Scan/manual entry/Recent Unlinked Scans stay out of the primary Market Watch dashboard until the secondary manual-listing action is opened | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-014 keeps scanner capture behind a secondary action` |
| UC-MW-18 | Purchase handoff from output detail | Output detail Purchase handoff posts selected candidate, persists purchase lifecycle and expected-arrival records, and keeps Market Watch/Discoveries status synchronized | implemented: `internal/discovery/service_test.go` `TestApplyActionMarkPurchasedCreatesCommerceHandoff`; `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-015 persists output-detail Purchase handoff provenance`; `ui.web/src/features/scanner/index.tsx` `scanner-handoff-purchase-*` |
| UC-MW-19 | Result inbox lifecycle review | Output detail result inbox shows status/provider/match/wishlist filters, match rationale, seen timestamps, total price, lifecycle status, and decision history without losing dismissed or downstream-handoff results; candidate API supports status/provider pagination, profile-scoped lifecycle status persistence, and durable transition history | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-016 + #1548 renders result inbox lifecycle filters and match provenance`; `internal/scanner/service_test.go` `TestCandidateResultInboxFiltersPaginationAndLifecycleUpdate`; `internal/app/scanner_api_test.go` `TestScannerCandidatesResultInboxFiltersPaginationAndLifecycleAPI`; `internal/app/openapi_parity_test.go` `TestOpenAPIDocumentsEbaySavedSearchHandoffContract`; `ui.web/src/features/scanner/index.tsx` `market-watch-results-inbox-*`; `docs/api/openapi.yaml` `CandidateDecisionHistoryRecord` |
| UC-MW-20 | Provider health and run history | Unknown provider health stays actionable, provider health taxonomy includes label/guidance/retry timing, health transitions persist after run outcomes, and persisted run history lists provider outcomes after reload | implemented: `internal/scanner/service_test.go` `TestRunNowPersistsClassifiedProviderHealthTransitions`; `internal/app/openapi_parity_test.go` `TestOpenAPIDocumentsEbayProviderHealthContract`; `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-017 shows actionable provider health and persisted run history`; `UI-SCREEN-MARKET-WATCH-017 shows provider health taxonomy labels guidance and retry timing` |
| UC-MW-21 | Saved integration search dashboard shell | Header, top summary, create controls, provider health, saved-watch table, result inbox, and manual listing entry present Market Watch as a saved integration search dashboard | implemented: `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts` `UI-SCREEN-MARKET-WATCH-018 + #1540 presents saved integration search dashboard shell` |
| UC-MW-22 | Beta provider registry fail-closed status | Registry marks eBay setup-required without credentials, disables unsupported provider actions, and labels unproven storefronts beta-limited/manual-capture-only until live evidence is attached | implemented: `internal/app/integration_migration_regression_test.go` `TestBetaMarketWatchProviderRegistryFailsClosedWithoutLiveProof` |
| UC-MW-23 | Bonza public provider beta proof | Successful Bonza Store API Market Watch run records provider health and upgrades registry status to live validated with non-secret release evidence | implemented: `internal/app/provider_bonza_run_api_test.go` `TestBonzaRunRecordsLiveProviderProofForBetaRegistry`; `openspec/migration/market-watch-beta-provider-proof.md` |
| UC-MW-24 | Voglers BigCommerce public provider beta proof | Successful public BigCommerce storefront HTML Market Watch run records provider health and upgrades registry status to live validated with non-secret release evidence | implemented: `internal/app/provider_bigcommerce_api_test.go` `TestBigCommerceRunStorefrontHTMLModeRecordsLiveProof`; `openspec/migration/market-watch-beta-provider-proof.md` |
