## Purpose
Define scanner query-set lifecycle, execution controls, and failure recovery behavior.

## Requirements
### Requirement SCANNER-001: Scanner query sets SHALL support user-defined market criteria
Cabinet SHALL support query set criteria including keywords, exclusions, max price, region, condition, and scheduling metadata.

#### Scenario: Create query set
- **GIVEN** active profile exists and payload includes `name`, at least one `keyword`, and optional filters (`region`, `max_price`, `condition`)
- **WHEN** client submits scanner query set creation request
- **THEN** runtime MUST persist query set and return `201` with stable `id` and normalized filter fields

### Requirement SCANNER-002: Scanner execution SHALL support manual and scheduled runs with rate limits
Cabinet SHALL support run-now and scheduled execution under rate-limited controls.

#### Scenario: Scheduled scanner run
- **GIVEN** at least one query set is `enabled=true`, has non-empty `schedule_cron`, and `rate_limit_rps` is configured
- **WHEN** scheduler executes `POST /api/scanner/run/scheduled`
- **THEN** runtime MUST execute only eligible query sets, enforce configured rate limits/backoff, and return `200` run summary with `run_id`, `query_sets_executed`, `candidates_collected`, and `failures`

### Requirement SCANNER-003: Scanner failures SHALL be diagnosable and retryable
Cabinet SHALL log failures and support retry by query set.

#### Scenario: Retry failed scan
- **GIVEN** `scanner_failures` contains an entry for query set `q1`
- **WHEN** client calls `POST /api/scanner/failures/retry` with `query_set_id=q1`
- **THEN** runtime MUST return `200`, mark retry requested, and append retry activity to scanner logs

### Requirement INTEGRATION-015: Scanner normalization MUST map provider outputs to common candidate schema
Cabinet MUST normalize all provider outputs (official APIs and web ingestion) to shared candidate/pricing/stock fields before persistence.

#### Scenario: Provider normalization contract
- **GIVEN** scanner run receives heterogeneous provider payloads (API JSON and web-ingestion parse outputs)
- **WHEN** normalization pipeline runs before persistence
- **THEN** persisted candidate contract MUST include:
  - `listing_id`
  - `title`
  - `price.amount`
  - `price.currency`
  - `url`
  - `seller`
  - `stock_signal.normalized_state`
  - `source.provider_id`
  - `first_seen`
  - `last_seen`
