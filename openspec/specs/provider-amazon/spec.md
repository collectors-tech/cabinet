## Purpose
Define Amazon provider contract for discovery and pricing ingestion.

## Requirements
### Requirement INTEGRATION-008: Amazon provider MUST declare integration mode
Cabinet SHALL classify Amazon integration mode (official API, affiliate feed, or web ingestion) with explicit availability status and constraints.

#### Scenario: Resolve Amazon integration mode
- **GIVEN** provider registry request is made for `amazon`
- **WHEN** `GET /api/providers/registry` resolves Amazon entry
- **THEN** payload MUST include:
- API outcome MUST be explicit: `200` on success, `4xx` for validation/auth conflicts, and `5xx` for unexpected failures
  - `provider_id: "amazon"`
  - `integration_mode` (`program_api|disabled`)
  - `eligibility_required` (boolean)
  - `policy_scope_note` (string)

### Requirement INTEGRATION-009: Amazon provider MUST normalize listing candidates when enabled
Cabinet SHALL normalize Amazon listing payloads into candidate schema when provider mode is enabled for scanning.

#### Scenario: Ingest Amazon candidates
- **GIVEN** Amazon mode is `program_api` and credentials/config are valid
- **WHEN** scanner executes Amazon query set
- **THEN** provider response MUST normalize into candidate fields used by common candidate contract and return `200`

### Requirement INTEGRATION-010: Amazon provider MUST expose unsupported-state diagnostics when disabled
Cabinet SHALL return explicit unsupported/disabled diagnostics when Amazon integration mode is unavailable.

#### Scenario: Amazon mode unavailable
- **GIVEN** Amazon integration mode is disabled for active profile
- **WHEN** user attempts Amazon scan execution
- **THEN** runtime MUST return `409` with payload:
  - `error_code: "PROVIDER_DISABLED"`
  - `provider: "amazon"`
  - `message`
  - `next_action`
