## Purpose
Define integrations workspace provider listing and wiring behavior.

## Requirements
### Requirement INTEGRATION-013: Integrations workspace MUST present API-backed provider records
Cabinet SHALL display integration providers and details from runtime sources, not static template placeholders.

#### Scenario: Open integrations route
- **GIVEN** authenticated user opens Integrations route
- **WHEN** UI requests `GET /api/providers/registry`
- **THEN** route MUST render provider rows from response payload and handle:
  - `200` ready list
  - `4xx` permissions/config error
  - `5xx` provider-service failure state

### Requirement INTEGRATION-014: Integrations workspace MUST consume canonical provider registry
Cabinet SHALL use `provider-registry` capability as source-of-truth for provider list and capability badges.

#### Scenario: Render provider list from registry
- **GIVEN** provider registry response includes capability metadata
- **WHEN** integrations workspace renders provider rows
- **THEN** UI MUST map badges/toggles directly from registry fields:
  - `capabilities.search`
  - `capabilities.stock_observation`
  - `capabilities.pricing`
  - `state`
