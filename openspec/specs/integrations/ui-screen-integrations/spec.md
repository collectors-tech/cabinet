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
| UC-INT-UI-16 | Recover missing active profile inline | Active-profile bootstrap error exposes selectable profile recovery, activates chosen profile, reloads registry/settings, and clears the route error | `cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` `UI-SCREEN-INTEGRATIONS-005: recovers active-profile bootstrap inline by selecting or creating profile context` |
