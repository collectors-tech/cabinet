## Purpose
Define integrations workspace provider-card listing and runtime wiring behavior.

## Requirements
### Requirement INTEGRATION-013: Integrations workspace MUST present API-backed provider records
Cabinet SHALL display integration providers and details from runtime sources, not static template placeholders.

#### Scenario: Open integrations route
- **GIVEN** authenticated user with an active profile opens Integrations route in provider-focused workspace mode
- **WHEN** UI requests `GET /api/providers/registry`
- **THEN** route MUST render provider cards from response payload and handle:
  - `200` ready list
  - `4xx` permissions/config error
  - `5xx` provider-service failure state

### Requirement INTEGRATION-014: Integrations workspace MUST consume canonical provider registry
Cabinet SHALL use `provider-registry` capability as source-of-truth for provider cards and capability badges.

#### Scenario: Render provider cards from registry
- **GIVEN** provider registry response includes capability metadata and provider operational snapshots
- **WHEN** integrations workspace renders provider cards
- **THEN** UI MUST map badges/toggles directly from registry fields:
  - `provider_id`
  - `display_name`
  - `capabilities.search`
  - `capabilities.stock_observation`
  - `capabilities.pricing`
  - `has_token`
  - `health.status`
  - `last_run.status`
  - `state`

### Requirement INTEGRATION-020: Integrations cards MUST open actionable provider detail panels
Cabinet SHALL open provider detail modal/drawer from card action to support instructions, credential fields, and provider actions.

#### Scenario: Open provider details and edit keys
- **GIVEN** integrations workspace shows provider cards with stable `provider_id`
- **WHEN** user clicks card `Connect` or `Edit`
- **THEN** details panel MUST open with:
  - provider setup instructions
  - credential form fields (write-only token input)
  - health and last-run status
  - provider actions (`Validate`, `Sync`, `Save`)
  - save action that persists provider settings

### Requirement INTEGRATION-022: Integrations screen MUST default to card interactions and keep URL-backed view state
Cabinet SHALL default Integrations to cards and SHALL encode non-default view mode in route search state.

#### Scenario: Open integrations without explicit view query
- **GIVEN** authenticated user opens `/integrations` without `view` search parameter
- **WHEN** integrations screen initializes
- **THEN** cards view MUST render as default and rows view MUST only activate when `view=rows` is present
