## Purpose
Define eBay provider contract for scanner/search integration.

## Requirements
### Requirement INTEGRATION-005: eBay provider MUST support authenticated listing search
Cabinet SHALL execute eBay listing queries using profile-scoped credentials and query-set criteria.

#### Scenario: Search eBay listings
- **GIVEN** active profile stores valid OAuth bearer token and query set contains keywords/region/max price
- **WHEN** scanner calls eBay provider adapter for run request
- **THEN** provider call MUST return `200` and normalized candidates with fields:
  - `listing_id`
  - `title`
  - `price.amount`
  - `price.currency`
  - `url`
  - `seller`
  - `first_seen`
  - `last_seen`

### Requirement INTEGRATION-006: eBay provider MUST expose health state
Cabinet SHALL report eBay provider health and recent failure telemetry via provider health endpoints.

#### Scenario: eBay health check
- **GIVEN** provider health service is enabled
- **WHEN** `GET /api/provider/health?provider=ebay` is requested
- **THEN** response MUST be `200` with:
  - `provider: "ebay"`
  - `state: ready|degraded|disabled`
  - `last_error` (nullable)
  - `retry_after_seconds` (nullable integer)

### Requirement INTEGRATION-007: eBay provider MUST capture stock observations when available
Cabinet SHALL persist stock/availability observations from eBay listing payloads when present.

#### Scenario: Persist eBay stock observation
- **GIVEN** normalized eBay candidate includes `availability` or `quantity` signals
- **WHEN** candidate persistence runs
- **THEN** candidate state MUST transition to `normalized_stock_observed=true` and store:
  - `stock_signal.raw`
  - `stock_signal.normalized_state`
  - `stock_signal.observed_at`
