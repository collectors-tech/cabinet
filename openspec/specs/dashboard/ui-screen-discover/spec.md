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

### Requirement UI-SCREEN-DISCOVER-003: Discover SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states.

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
