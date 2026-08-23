## Purpose
Define measurable UI performance thresholds and resilience-state contracts.

## Requirements
### Requirement UI-PERFORMANCE-001: Screen-level interaction targets SHALL be measurable and enforceable
Key UI interactions SHALL have defined threshold targets for S2 and non-crash constraints for S3.

#### Scenario: Validate S2 interaction thresholds
- **GIVEN** `POST /api/test/scale/bootstrap` is executed with `profile="S2"` and deterministic `seed`
- **WHEN** UI performance probe executes repeated calls to:
  - `GET /api/items?status=active`
  - `GET /api/search/items?q=scale&limit=20`
  - `GET /api/scanner/candidates?query_set_id=<id>`
- **THEN** each call SHALL return `200`
- **AND** median response duration for each workflow SHALL remain below 1500 ms
- **AND** the workflow SHALL complete without route crash or global `500` fallback UI

### Requirement UI-PERFORMANCE-002: Scale validation SHALL enforce UI resilience states
Large-data operations SHALL expose deterministic loading, empty, error, and ready states.

#### Scenario: S3 delayed response handling
- **GIVEN** an E2E local workspace profile is bootstrapped
- **AND** the inventory list request `GET /api/items*` is delayed by test harness
- **WHEN** user opens `/inventory`
- **THEN** UI SHALL render `[data-testid="inventory-loading"]` while the request is pending
- **AND** `[data-testid="inventory-loading"]` SHALL be removed after successful response
- **AND** the inventory table SHALL render without global `500` fallback

