## Purpose
Define Discoveries screen behavior for candidate-item triage actions. `/discoveries`
is the inbox for already found candidate items awaiting user decision; Market Watch and
provider integration screens own query execution and import configuration.

## Requirements
### Requirement UI-SCREEN-DISCOVER-001: Discover SHALL support filterable candidate triage
Discover SHALL support query/price/date filtering and list rendering.

#### Scenario: Filtered triage list
- **GIVEN** an authenticated collector session is active on `/discoveries` with seeded discovery candidates
- **WHEN** user applies filters
- **THEN** discover list SHALL render filtered candidates

#### Scenario: Apply Filters control triggers deterministic query update
- **GIVEN** discover filter inputs are populated
- **WHEN** user clicks `Apply Filters`
- **THEN** screen MUST execute filtered query and refresh candidate list without route transition

### Requirement UI-SCREEN-DISCOVER-002: Discover SHALL support all primary candidate actions
Discover SHALL support ignore, wishlist, track, and create-item actions.

#### Scenario: Candidate action apply
- **GIVEN** authenticated Discoveries list includes one candidate row
- **WHEN** user chooses action on candidate
- **THEN** candidate state and downstream linkage SHALL update

#### Scenario: Candidate action failure
- **GIVEN** authenticated Discoveries list includes one candidate row
- **WHEN** a candidate action request fails
- **THEN** Discoveries SHALL surface deterministic failure feedback
- **AND** the current route and candidate list SHALL remain stable without a success reload

### Requirement UI-SCREEN-DISCOVER-003: Discover SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states.

#### Scenario: Discover loading state
- **GIVEN** an authenticated user opens `/discoveries`
- **WHEN** the candidate list request is still pending
- **THEN** the screen SHALL show an explicit loading state in the triage list
- **AND** the loading state SHALL resolve to the loaded candidate list without route transition

### Requirement UI-SCREEN-DISCOVER-004: Discoveries SHALL remain a triage workspace and MUST NOT expose provider query-run controls
Discoveries MUST focus on triage actions for already discovered candidates and must not duplicate Market Watch query-set creation/run capabilities.

#### Scenario: Discoveries boundary enforcement
- **GIVEN** user is on Discoveries screen
- **WHEN** screen actions render
- **THEN** triage actions (ignore/wishlist/track/create item) MUST be available
- **AND** provider query-set creation/run controls MUST NOT be present

#### Scenario: Discoveries to Market Watch handoff
- **GIVEN** user needs provider query-run workflow from Discoveries context
- **WHEN** user selects handoff action
- **THEN** app MUST route user to Market Watch with preserved context where applicable

#### Scenario: Discover error state
- **GIVEN** authenticated Discoveries route receives non-`200` from `GET /api/discovery/not-in-collection`
- **WHEN** discover API request fails
- **THEN** screen SHALL present actionable retry state

### Requirement UI-SCREEN-DISCOVER-005: Discoveries SHALL explain candidate purpose and provenance in-page
Discoveries SHALL make the page purpose, source provenance, status, and next action clear
without presenting the surface as a generic search or query history page.

#### Scenario: Candidate inbox purpose
- **GIVEN** authenticated user opens `/discoveries`
- **WHEN** the page renders candidate rows or an empty state
- **THEN** the screen MUST describe Discoveries as pending found-item triage
- **AND** it MUST distinguish candidate findings from owned Inventory, wanted Wishlist records, and Market Watch query history

#### Scenario: Empty candidate inbox remains actionable without mutation controls
- **GIVEN** authenticated Discoveries route receives an empty candidate list
- **WHEN** the empty state renders
- **THEN** the screen MUST explain that no pending found-item candidates are available
- **AND** candidate-specific mutation controls MUST NOT render
- **AND** the Market Watch handoff MUST remain reachable without leaving `/discoveries/` until selected

#### Scenario: Candidate row provenance and actions
- **GIVEN** a discovery candidate has source metadata
- **WHEN** the candidate row renders
- **THEN** row content MUST expose source/provider label, source result link when available, candidate title, price/currency when available, first-seen or last-seen recency, triage status, and confidence or review signal
- **AND** row actions MUST provide clear paths for Wishlist promotion, Inventory/Purchase handoff where applicable, ignore/archive, and source-result review

### Requirement UI-SCREEN-DISCOVER-006: Wishlist promotion SHALL preserve wanted-state boundaries in the UI
Discoveries SHALL let a user promote a candidate into Wishlist without implying the item
has already been purchased or delivered.

#### Scenario: Promote candidate to Wishlist
- **GIVEN** a discovery candidate is visible with source provenance and price context
- **WHEN** user chooses `Promote to Wishlist`
- **THEN** Discoveries MUST submit `add_to_wishlist` for the candidate
- **AND** the resulting Wishlist row/card MUST show the promoted title, category, notes/provenance context, and target price
- **AND** the resulting Wishlist row/card MUST show `Purchased: No` and `Delivered: No`
- **AND** the row MUST keep purchase action controls available for the later purchase workflow

### Requirement UI-SCREEN-DISCOVER-007: Discoveries SHALL render a dashboard-first found-deals review surface
Discoveries SHALL prioritize collector-useful found deals and source outputs while
preserving the candidate-inbox destination workflow.

#### Scenario: Dashboard summary and source filters
- **GIVEN** authenticated Discoveries route receives candidates with wishlist, deal, source, and triage metadata
- **WHEN** `/discoveries` renders
- **THEN** the screen MUST show summary counts for best deals, wishlist matches, new findings, Market Watch outputs, and provider/store attention
- **AND** the screen MUST expose source filter tabs for all discoveries, wishlist matches, great prices, Market Watch, stores/providers, other public or shared inventories, and ignored or archived candidates

#### Scenario: Ranked deal table
- **GIVEN** discovery candidates include wishlist match, target price, baseline price, source provenance, stock, recency, confidence, and status metadata where available
- **WHEN** the results surface renders
- **THEN** the main content MUST render as a table-style review surface rather than a loose candidate card list
- **AND** wishlist and deal candidates MUST be ranked ahead of lower-signal candidates
- **AND** the table MUST expose deterministic sort controls for deal priority and recency without mutating candidate state
- **AND** rows MUST show match reason, source/provider, price, target price, baseline, savings delta, availability, seller/source, first/last seen, triage status, confidence/review signal, and source-result access where available
- **AND** ignored or archived candidates MUST be hidden from the default view and reachable through an explicit filter
- **AND** source-result, Wishlist, Purchase follow-up, Inventory handoff, and ignore/archive actions MUST remain contextual and accessible without claiming ownership by default

#### Scenario: Contextual action set by candidate state
- **GIVEN** Discoveries renders wishlist-match, new non-wishlist, promoted, and ignored or archived candidates
- **WHEN** row actions are shown
- **THEN** wishlist-match rows MUST expose source review, purchase follow-up, and ignore/archive actions without duplicate Wishlist or Inventory promotion controls
- **AND** new non-wishlist rows MUST expose source review, Wishlist promotion, Purchase follow-up, Inventory handoff, and ignore/archive actions
- **AND** promoted rows MUST expose linked destination access where available and suppress duplicate promotion controls
- **AND** ignored or archived rows MUST be reachable only through the explicit ignored/archived filter and expose restore-for-review only when the data model can return them to the review queue

#### Scenario: Dashboard API fields and archived opt-in
- **GIVEN** scanner candidates include wishlist linkage, saved-search provenance, price snapshots, provider health, thumbnails, and ignored or archived status
- **WHEN** Discoveries requests `/api/discovery/not-in-collection?include_archived=true`
- **THEN** the API response MUST include dashboard display fields for currency, triage status, seller/source labels, thumbnail URL, match type/reason, wishlist IDs, target price, market baseline, price delta amount and percent, deal score, source trust status, and availability
- **AND** ignored or archived candidates MUST be returned only when the request explicitly opts into archived records
- **AND** the default API request without the archived opt-in MUST continue hiding ignored or archived candidates

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
| UC-DIS-01 | Filter discover list | Filtered candidates displayed | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-001 renders filterable candidate triage list` |
| UC-DIS-02 | Ignore action | Candidate state updates to ignored | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-002 + UC-DIS-02..04 submits ignore wishlist track and create action payloads` |
| UC-DIS-03 | Add to wishlist action | Wishlist linkage created | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-002 + UC-DIS-02..04 submits ignore wishlist track and create action payloads`; `UI-SCREEN-DISCOVER-006 promotes a candidate to Wishlist without purchased state` |
| UC-DIS-04 | Track/create action | Pricing track or item create executes | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-002 + UC-DIS-02..04 submits ignore wishlist track and create action payloads` |
| UC-DIS-05 | Discover API failure | Error + retry appears | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-003 shows retryable error state when discover API fails` |
| UC-DIS-06 | Apply Filters action | `Apply Filters` triggers deterministic filtered-query refresh | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-001 + UC-DIS-06 applies query price and date filters without route transition` |
| UC-DIS-07 | Boundary enforcement | Discoveries shows triage actions only; no provider query/run controls | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-004 keeps Discoveries as triage-only and excludes Market Watch query/run controls` |
| UC-DIS-08 | Market Watch handoff | Discoveries can route to Market Watch handoff without losing context | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-004 provides explicit handoff action to Market Watch with preserved context` |
| UC-DIS-09 | Candidate inbox purpose | Discoveries purpose and empty/list states distinguish found-item triage from Inventory, Wishlist, and Market Watch query history | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-005 explains candidate inbox purpose` |
| UC-DIS-10 | Candidate provenance row | Candidate row exposes source/provider, source-result link, title, price/currency, recency, status, confidence/review signal, and destination actions | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-005 renders candidate provenance and destination actions` |
| UC-DIS-11 | Promote candidate to Wishlist | Wishlist promotion creates wanted-state UI proof without purchased/delivered state | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-006 promotes a candidate to Wishlist without purchased state` |
| UC-DIS-12 | Discover loading state | Pending candidate-list request shows loading feedback and resolves to loaded candidates without route transition | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-003 shows loading state before candidate list resolves` |
| UC-DIS-13 | Candidate action failure | Failed candidate action surfaces deterministic feedback without losing route or reloading the candidate list as success | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-002 + UC-DIS-13 keeps candidate list stable when an action fails` |
| UC-DIS-14 | Empty candidate inbox | Empty Discoveries state explains no pending found-item candidates, suppresses candidate mutation controls, and keeps Market Watch handoff reachable | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-005 + UC-DIS-14 renders empty candidate inbox without mutation controls` |
| UC-DIS-15 | Discoveries deal dashboard | Dashboard summary, source filters, ranked table rows, wishlist/deal priority, and ignored/archive review filter render from candidate metadata | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-007 + #1533 renders dashboard summary source filters and ranked deal table` |
| UC-DIS-16 | Discoveries dashboard API fields | Backend returns deal ranking, wishlist linkage, provider trust, display labels, and archived candidates only through explicit opt-in | implemented: `internal/discovery/service_test.go` `TestListNotInCollectionDashboardFieldsAndArchivedOptIn`; `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-007 + #1533 renders dashboard summary source filters and ranked deal table` |
| UC-DIS-17 | Source filters and ranking stability | Source filters and table sort controls narrow or reorder visible rows without posting discovery actions or mutating candidate state, while wishlist/deal candidates remain ranked ahead of lower-signal rows by default | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-007 keeps ranking and source filters deterministic without mutating candidates`; `UI-SCREEN-DISCOVER-007 sorts the dashboard table by deal and recency without mutating candidates` |
| UC-DIS-18 | Provider attention, no-match, and promoted states | Provider-attention, unmatched/no-match, and already-promoted candidates render distinct review states, and promoted rows expose destination access instead of duplicate promotion actions | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-007 renders provider-attention no-match and promoted destination states` |
| UC-DIS-19 | Contextual discovery actions | Wishlist-match, new non-wishlist, promoted, and ignored/archived candidates expose only state-safe actions, and restore returns archived candidates to review | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` `UI-SCREEN-DISCOVER-007 + #1556 shows contextual actions for wishlist, new, and archived candidates`; `internal/discovery/service_test.go` `TestReviewActionRestoresIgnoredCandidateForDefaultQueue` |
