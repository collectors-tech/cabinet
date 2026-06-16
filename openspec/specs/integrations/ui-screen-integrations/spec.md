## Purpose
Define Integrations screen behavior for provider cards, filters, and credential edit dialogs.

## Requirements
### Requirement UI-SCREEN-INTEGRATIONS-001: Integrations screen SHALL support search/filter/sort over provider cards
Integrations screen SHALL support text filter, connection-type filter, and sort controls.

### Requirement UI-SCREEN-INTEGRATIONS-012: Integrations screen SHALL support integration-type selector and rows/cards view toggles
Integrations screen SHALL expose integration-type selector (default `All Integrations`) and explicit `Rows`/`Cards` view toggles.

#### Scenario: Select integration type filter
- **GIVEN** integrations route is loaded
- **WHEN** user opens integration type selector and chooses `All Integrations` or another type
- **THEN** provider list MUST refresh to selected integration type context

#### Scenario: Toggle rows/cards view
- **GIVEN** integrations route is loaded
- **WHEN** user toggles `Rows` and `Cards`
- **THEN** provider presentation MUST switch deterministically and preserve active filter context

### Requirement UI-SCREEN-INTEGRATIONS-013: Integrations screen SHALL hydrate URL-backed filter state on direct route entry
Integrations screen SHALL apply supported query parameters on first render so shared or reloaded route URLs are deterministic.

#### Scenario: Direct route query state hydration
- **GIVEN** an authenticated user opens `/integrations/` with `filter`, `type`, `sort`, and `view` query parameters
- **WHEN** profile, provider registry, and profile settings bootstrap resolves
- **THEN** the screen MUST apply the query-backed filter text, integration type, sort order, and rows/cards view before rendering provider results
- **AND** visible provider results MUST match the query-backed connected/not-connected filter and text filter without requiring a second user action

#### Scenario: Direct route query zero-result state
- **GIVEN** an authenticated user opens `/integrations/` with query parameters that match no providers
- **WHEN** profile, provider registry, and profile settings bootstrap resolves
- **THEN** the screen MUST render a deterministic no-match state in the requested view
- **AND** pagination and route query state MUST remain stable without opening provider dialogs

#### Scenario: Direct route query zero-result cards view
- **GIVEN** an authenticated user opens `/integrations/` with `view=cards` and query parameters that match no providers
- **WHEN** profile, provider registry, and profile settings bootstrap resolves
- **THEN** the cards surface MUST render explicit no-match feedback instead of a blank list
- **AND** route query state MUST remain stable without opening provider dialogs

### Requirement UI-SCREEN-INTEGRATIONS-014: Integrations table rows SHALL expose deterministic row interaction surfaces
Integrations table rows SHALL provide distinct single-click, double-click, and row-action behaviors without ambiguous dialog overlap.

#### Scenario: Table row single-click and double-click surfaces
- **GIVEN** integrations route is loaded in rows view with provider registry data
- **WHEN** user single-clicks a non-interactive row area
- **THEN** a read-only row details dialog MUST open for that provider
- **AND** the current URL MUST preserve the selected provider context
- **WHEN** user double-clicks a non-interactive row area for another provider
- **THEN** an edit row dialog MUST open for that provider without leaving the details dialog open
- **AND** the current URL MUST preserve the newly selected provider context

#### Scenario: Nested row actions do not trigger row dialogs
- **GIVEN** integrations route is loaded in rows view with provider registry data
- **WHEN** user clicks the row action button for a provider
- **THEN** the provider configuration dialog MUST open
- **AND** row details/edit dialogs MUST NOT open from the nested action click

### Requirement UI-SCREEN-INTEGRATIONS-011: Integrations screen SHALL use a paginated full-page table as the primary provider list
Integrations screen SHALL render the primary provider list as a scan-friendly full-page table with pagination and stable operational columns.

#### Scenario: Default integrations table renders stable scan columns
- **GIVEN** integrations route is loaded with provider registry data
- **WHEN** the primary provider list renders
- **THEN** the default presentation MUST be a table, not cards
- **AND** the table MUST include stable columns for provider/name, category/type, connection/config status, action availability, health/last-run state, and row actions
- **AND** row actions MUST keep the provider configuration/details flow reachable

#### Scenario: Integrations table paginates larger provider lists
- **GIVEN** integrations route has more providers than the table page size
- **WHEN** user uses pagination controls
- **THEN** the visible rows MUST advance between pages
- **AND** pagination status MUST report the visible range and current page
- **AND** filter/type/sort changes MUST reset pagination to the first matching page

#### Scenario: Filter and sort integrations
- **GIVEN** integrations route is loaded with provider cards
- **WHEN** user applies text filter, type filter, and sort order
- **THEN** rendered provider cards MUST reflect active filter/sort state

### Requirement UI-SCREEN-INTEGRATIONS-002: Integrations screen SHALL support connect/edit provider modal workflow
Integrations screen SHALL open provider modal from card action and allow editing connection values.

#### Scenario: Open and edit integration
- **GIVEN** integrations route is loaded and provider card is visible
- **WHEN** user clicks `Connect` or `Edit`
- **THEN** modal MUST open prefilled with provider settings where available

### Requirement UI-SCREEN-INTEGRATIONS-003: Integrations screen SHALL persist provider settings via profile settings API
Integrations screen SHALL persist provider settings to active profile settings endpoint.

#### Scenario: Save integration settings
- **GIVEN** active profile is available and integration edit modal is open
- **WHEN** user saves base URL, token, and marketplace fields
- **THEN** UI MUST call profile settings API and reflect updated connected state on success

### Requirement UI-SCREEN-INTEGRATIONS-004: Integrations screen SHALL enforce write-only credential UX
Integrations screen SHALL never rehydrate clear credential values into UI and SHALL use replace-token workflow.

#### Scenario: Edit provider without exposing existing token
- **GIVEN** provider credentials already exist for active profile
- **WHEN** user opens integration edit panel
- **THEN** token field MUST render masked/empty placeholder state and save flow MUST support explicit token replacement without showing existing token

### Requirement UI-SCREEN-INTEGRATIONS-005: Integrations screen SHALL provide deterministic bootstrap/load/error states
Integrations screen SHALL show explicit loading and actionable error states for profile/registry/settings bootstrap.

#### Scenario: Integrations heading copy resolves for users
- **GIVEN** integrations route loads successfully
- **WHEN** the page header renders
- **THEN** the heading and description MUST render resolved user-facing copy
- **AND** raw translation keys such as `integrations.title` and `integrations.description` MUST NOT be visible

#### Scenario: Registry bootstrap failure
- **GIVEN** integrations route loads and registry request fails
- **WHEN** `GET /api/providers/registry` returns non-`200` or network failure
- **THEN** screen MUST show user-visible error state with retry control and MUST avoid silent failure

#### Scenario: Active-profile bootstrap recovery in route
- **GIVEN** integrations route loads without an active profile context but profile listing is still available
- **WHEN** active profile bootstrap fails with `active_profile_*`
- **THEN** screen MUST expose an in-route recovery path to either select an existing profile or create a new profile and continue bootstrap in place

#### Scenario: Create active profile inline during recovery
- **GIVEN** integrations route loads without an active profile context and no selectable profiles are returned
- **WHEN** the user enters a new profile name and submits the recovery create action
- **THEN** the UI MUST create the profile through the profiles API
- **AND** the UI MUST activate the created profile before reloading provider registry and profile settings
- **AND** the route MUST stay on `/integrations/`, clear the bootstrap error, and render the provider list for the created profile context

### Requirement UI-SCREEN-INTEGRATIONS-006: Integrations screen SHALL use provider-registry endpoint as list source-of-truth
Integrations screen SHALL derive provider cards exclusively from `GET /api/providers/registry` response.

#### Scenario: Screen bootstrap uses registry source
- **GIVEN** integrations route is opened for active profile
- **WHEN** screen initializes provider list data
- **THEN** provider cards MUST be sourced from `/api/providers/registry` and MUST NOT rely on static connector seed arrays

### Requirement UI-SCREEN-INTEGRATIONS-007: Provider detail panel SHALL expose health and action controls
Integrations screen SHALL expose provider health and actionable controls from detail panel.

#### Scenario: Provider dialog sync action is not inert
- **GIVEN** a provider detail panel is open and provider discovery runs are started from Market Watch query sets
- **WHEN** the detail panel renders the Sync affordance
- **THEN** Sync MUST NOT be an enabled inert action
- **AND** the panel MUST explain that sync runs from Market Watch query sets until an in-dialog sync flow exposes progress and completion state

### Requirement UI-SCREEN-INTEGRATIONS-008: Provider validation SHALL reconcile visible health state
Integrations screen SHALL keep validation completion feedback and visible provider health metadata consistent.

#### Scenario: Successful validation updates visible provider health metadata
- **GIVEN** a provider detail panel is open and currently shows stale or unknown health metadata
- **WHEN** the user triggers `Validate` and the health endpoint returns successfully
- **THEN** the UI MUST expose an in-progress validating state while the request is active
- **AND** the completion message MUST include the resulting health status
- **AND** the provider detail panel MUST update health, last-run, and last-checked metadata from the validation result
- **AND** the provider card MUST use the same reconciled health state after the result is applied

#### Scenario: eBay validation displays readiness aliases and recovery guidance
- **GIVEN** the eBay provider detail panel is open
- **WHEN** `/api/provider/health?provider=ebay` returns ready, missing-credential, invalid-credential, or backoff readiness payloads
- **THEN** the panel MUST display the returned readiness `state`, message, `last_error` when present, `retry_after_seconds` when present, and `next_action` when present
- **AND** missing or invalid credential states MUST direct the operator to credential and health recovery instead of implying a Market Watch run succeeded

### Requirement UI-SCREEN-INTEGRATIONS-009: Integrations UI SHALL display provider API family support mapping
Integrations screen SHALL show API support mapping per provider (Woo/Boost/Algolia/custom) in cards and detail panel.

#### Scenario: API family badges in provider cards
- **GIVEN** provider registry payload includes `api_family` mapping
- **WHEN** integrations cards render
- **THEN** each card MUST show API family badge/label derived from registry field

#### Scenario: API support details in provider panel
- **GIVEN** provider detail panel is opened
- **WHEN** panel renders support metadata
- **THEN** panel MUST show `api_family` and `api_support_profile` values with deterministic formatting

#### Scenario: Open provider detail panel from card
- **GIVEN** provider card is visible on integrations route
- **WHEN** user clicks `Connect` or `Edit`
- **THEN** detail panel MUST render:
  - setup instructions
  - health status and last-run status
  - `Validate` action
  - disabled/explained `Sync` affordance when sync is unavailable in the dialog
  - `Save Integration` action

### Requirement UI-SCREEN-INTEGRATIONS-010: Provider edit dialog inputs SHALL have visible and programmatic labels
Integrations provider configuration dialogs MUST render stable visible labels associated with each editable field instead of relying on placeholder text alone.

#### Scenario: Config fields remain labeled after values are populated
- **GIVEN** provider detail/edit dialog opens for a configurable provider
- **WHEN** base URL, marketplace/region, items-per-page, or token fields render
- **THEN** each field MUST have a visible label
- **AND** each label MUST be programmatically associated with its input by matching `htmlFor` and input `id`
- **AND** placeholder text MUST NOT be the only source of field meaning

## Acceptance Criteria
- Provider cards are sourced from runtime registry, not static seed list.
- Provider detail panel shows instructions, health, and last-run data.
- Provider validation feedback and visible health metadata reconcile after successful validation.
- Token handling is write-only with replace-token flow.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-INT-UI-01 | Filter/sort integration cards | Card list updates correctly | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-001 + UI-SCREEN-INTEGRATIONS-006 + INTEGRATION-022: defaults to cards and supports filter/sort/view using registry data` |
| UC-INT-UI-02 | Open connect/edit modal | Modal opens with provider context | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-002 + UI-SCREEN-INTEGRATIONS-007 + INTEGRATION-020: opens provider detail panel with actions and status` |
| UC-INT-UI-03 | Save provider values | Settings persist and card status updates | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-003 + UI-SCREEN-INTEGRATIONS-004: persists settings with write-only replace-token flow` |
| UC-INT-UI-04 | Edit existing provider token | Clear token is never shown and replace-token works | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-003 + UI-SCREEN-INTEGRATIONS-004: persists settings with write-only replace-token flow` |
| UC-INT-UI-05 | Registry/bootstrap failure | Error state with retry appears | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-005: renders deterministic bootstrap error with retry control` |
| UC-INT-UI-06 | Registry-backed provider list | Cards derive from runtime registry response | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-001 + UI-SCREEN-INTEGRATIONS-006 + INTEGRATION-022: defaults to cards and supports filter/sort/view using registry data` |
| UC-INT-UI-07 | Provider detail actions visible | Validate/Save controls appear with health/last-run; dialog Sync is disabled with Market Watch guidance when unsupported | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-002 + UI-SCREEN-INTEGRATIONS-007 + INTEGRATION-020: opens provider detail panel with actions and status` |
| UC-INT-UI-08 | Use integration type selector | Provider list updates for selected type (`All Integrations` default supported) | `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-012 + UC-INT-UI-08: filters rows by integration type selector` |
| UC-INT-UI-09 | Toggle rows/cards view | Provider presentation switches deterministically while preserving active filter/type query context | `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-012 + UC-INT-UI-09: toggles rows and cards while preserving active filter context` |
| UC-INT-UI-10 | Provider API family badge display | Cards show API family labels from registry mapping | `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-009 + UC-INT-UI-10: cards show provider API family badges from registry mapping` |
| UC-INT-UI-11 | Provider API support detail display | Detail panel shows API family + support profile metadata | `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-009 + UC-INT-UI-11 + INTEGRATION-024: detail panel shows API family + support profile metadata from registry` |
| UC-INT-UI-12 | Provider edit fields are labeled | Dialog config fields have visible labels associated by `htmlFor`/`id` | `internal/app/ui_template_contract_test.go` `TestIntegrationsProviderConfigInputsHaveLabels` |
| UC-INT-UI-13 | Validate provider health | Validation shows progress, reconciles health/last-run/last-checked, and reports resulting status | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-003 + UI-SCREEN-INTEGRATIONS-004 + UI-SCREEN-INTEGRATIONS-008: persists settings and reconciles validation health state` |
| UC-INT-UI-14 | Primary integrations table | Default list renders full-page table columns with provider identity, type, connection, actions, health/last-run, and row actions | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-001 + UI-SCREEN-INTEGRATIONS-006 + INTEGRATION-022: defaults to table and supports filter/sort/view using registry data` |
| UC-INT-UI-15 | Integrations table pagination | Larger provider lists page through stable table rows and reset to page 1 when filters change | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-011 + #1112: paginates the full-page integrations table` |
| UC-INT-UI-16 | Recover missing active profile inline | Active-profile bootstrap error exposes selectable profile recovery or inline profile creation, activates the resulting profile, reloads registry/settings, stays on `/integrations/`, and clears the route error | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-005: recovers active-profile bootstrap inline by selecting or creating profile context`, `UI-SCREEN-INTEGRATIONS-005 + UC-INT-UI-16: creates a missing active profile inline and reloads integrations` |
| UC-INT-UI-17 | Hydrate direct route query state | Shared `/integrations/` URL applies `filter`, `type`, `sort`, and `view` on first render with matching provider results | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-17: applies direct route query state on first render` |
| UC-INT-UI-18 | Use row interaction surfaces | Table row single-click opens details, double-click opens edit, selected context is URL-backed, and nested row actions do not trigger row dialogs | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-014 + UC-INT-UI-18: separates row details edit and action dialogs` |
| UC-INT-UI-19 | Hydrate direct route empty filter state | Shared `/integrations/` URL with no matching providers shows deterministic no-match table state, stable zero-result pagination, and preserved query context | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-19: shows deterministic empty state for direct route filters` |
| UC-INT-UI-20 | Hydrate direct route empty cards state | Shared `/integrations/` URL with no matching providers in cards view shows explicit no-match feedback and preserved query context | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-20: shows deterministic empty state for direct route filters in cards view` |
