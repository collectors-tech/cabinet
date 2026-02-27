## Purpose
Define Amazon provider contract for discovery and pricing ingestion.

## Requirements
### Requirement INTEGRATION-008: Amazon provider MUST declare integration mode
Cabinet SHALL classify Amazon integration mode (official API, affiliate feed, or web ingestion) with explicit availability status and constraints.

#### Scenario: Resolve Amazon integration mode
- **GIVEN** integrations UI requests provider registry for active profile and Amazon provider is configured in registry metadata
- **WHEN** `GET /api/providers/registry` resolves Amazon entry
- **THEN** payload MUST include:
  - status `200` on successful lookup
  - `provider_id: "amazon"`
  - `integration_mode` (`program_api|disabled`)
  - `eligibility_required` (boolean)
  - `policy_scope_note` (string)
  - error outcomes: `4xx` for validation/auth conflicts and `5xx` for unexpected runtime failures

### Requirement INTEGRATION-009: Amazon provider MUST normalize listing candidates when enabled
Cabinet SHALL normalize Amazon listing payloads into candidate schema when provider mode is enabled for scanning.

#### Scenario: Ingest Amazon candidates
- **GIVEN** Amazon provider mode is `program_api`, credentials are valid, and query set `q1` has keywords and region
- **WHEN** scanner executes Amazon run for `q1`
- **THEN** provider response MUST return `200` and MUST normalize to shared candidate fields (`listing_id`, `title`, `price.amount`, `price.currency`, `url`, `seller`, `source.provider_id`)

### Requirement INTEGRATION-010: Amazon provider MUST expose unsupported-state diagnostics when disabled
Cabinet SHALL return explicit unsupported/disabled diagnostics when Amazon integration mode is unavailable.

#### Scenario: Amazon mode unavailable
- **GIVEN** Amazon integration mode is disabled for active profile
- **WHEN** user attempts Amazon scan execution through provider route/adapter
- **THEN** runtime MUST return `409` with payload:
  - `error_code: "PROVIDER_DISABLED"`
  - `provider: "amazon"`
  - `message`
  - `next_action`
