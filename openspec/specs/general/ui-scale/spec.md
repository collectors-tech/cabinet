## Purpose
Define dataset-scale behavior requirements for UI workflows.

## Requirements
### Requirement UI-SCALE-001: UI scalability validation SHALL use deterministic dataset profiles
Cabinet SHALL validate UI behavior with deterministic S0/S1/S2/S3 data profiles.

#### Scenario: Reproduce S2 benchmark
- **GIVEN** E2E scale bootstrap request body includes `profile="S2"` and fixed `seed`
- **WHEN** client calls `POST /api/test/scale/bootstrap` twice with identical payload after reset
- **THEN** both responses SHALL return `200`
- **AND** both responses SHALL include identical `dataset_hash` and `counts` payloads
- **AND** response payload SHALL include `profile`, `seed`, `profile_id`, `query_set_id`, `dataset_hash`, and `counts`

### Requirement UI-SCALE-002: Data profiles SHALL define sample and bulk coverage for table-heavy screens
Scale profiles SHALL include realistic distributions for inventory, discovery, pricing, media, and wishlist.

#### Scenario: Run table-heavy checks on S3
- **GIVEN** `profile="S3"` is loaded through `POST /api/test/scale/bootstrap`
- **WHEN** client executes table-heavy API workflows:
  - `GET /api/items?status=active`
  - `GET /api/search/items?q=scale&limit=25`
  - `GET /api/scanner/candidates?query_set_id=<id>`
- **THEN** each workflow call SHALL return `200`
- **AND** `api/items` response SHALL contain non-empty `items` array
- **AND** no unrecoverable route/API crash SHALL occur during the sequence

### Requirement UI-SCALE-003: Scale validation SHALL include high-volume action workflows
Scale testing SHALL include rapid search/filter loops, discovery throughput, and report export under high volume.

#### Scenario: Discovery action throughput on S3
- **GIVEN** `profile="S3"` scale load is active and `query_set_id` is available
- **WHEN** client runs repeated high-volume loops of:
  - `GET /api/search/items?q=scale&limit=20`
  - `GET /api/scanner/candidates?query_set_id=<id>`
- **THEN** all loop iterations SHALL return `200`
- **AND** the run SHALL complete without unrecoverable stall or server termination

