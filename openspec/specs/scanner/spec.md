## Purpose
Define scanner query-set lifecycle, execution controls, and failure recovery behavior.

## Requirements
### Requirement SCANNER-001: Scanner query sets SHALL support user-defined market criteria
Cabinet SHALL support query set criteria including keywords, exclusions, max price, region, condition, and scheduling metadata.

#### Scenario: Create query set
- **GIVEN** valid query set input is provided
- **WHEN** user submits query set form
- **THEN** Cabinet SHALL persist query set definition

### Requirement SCANNER-002: Scanner execution SHALL support manual and scheduled runs with rate limits
Cabinet SHALL support run-now and scheduled execution under rate-limited controls.

#### Scenario: Scheduled scanner run
- **GIVEN** enabled scheduled query set exists
- **WHEN** scheduler triggers execution
- **THEN** scan SHALL execute with rate-limit policy

### Requirement SCANNER-003: Scanner failures SHALL be diagnosable and retryable
Cabinet SHALL log failures and support retry by query set.

#### Scenario: Retry failed scan
- **GIVEN** query set has a failed scanner run
- **WHEN** user requests retry
- **THEN** Cabinet SHALL schedule immediate retry and log outcome

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
