## Purpose
Define eBay provider contract for scanner/search integration.

## Requirements
### Requirement INTEGRATION-005: eBay provider MUST support authenticated listing search
Cabinet SHALL execute eBay listing queries using profile-scoped credentials and query-set criteria.

#### Scenario: Search eBay listings
- **GIVEN** active profile stores non-empty `ebay_bearer_token`, query set `q1` contains at least one keyword, and optional max-price filter
- **WHEN** scanner calls eBay provider adapter during run for `q1`
- **THEN** provider call MUST return `200` and normalized candidates with fields:
  - `listing_id`
  - `title`
  - `price.amount`
  - `price.currency`
  - `url`
  - `seller`
  - `first_seen`
  - `last_seen`
  - and MUST return `401` with `error_code="PROVIDER_AUTH_INVALID"` when bearer token is expired or rejected

### Requirement INTEGRATION-006: eBay provider MUST expose health state
Cabinet SHALL report eBay provider health and recent failure telemetry via provider health endpoints.

#### Scenario: eBay health check
- **GIVEN** provider health table has latest status entry for provider `ebay`
- **WHEN** `GET /api/provider/health?provider=ebay` is requested
- **THEN** response MUST be `200` with:
  - `provider: "ebay"`
  - `state: ready|degraded|disabled`
  - `last_error` (nullable)
  - `retry_after_seconds` (nullable integer)

### Requirement INTEGRATION-007: eBay provider MUST capture stock observations when available
Cabinet SHALL persist stock/availability observations from eBay listing payloads when present.

#### Scenario: Persist eBay stock observation
- **GIVEN** normalized eBay candidate contains `availability` text or numeric quantity signal
- **WHEN** candidate persistence runs
- **THEN** candidate state MUST transition to `normalized_stock_observed=true` and store:
  - `stock_signal.raw`
  - `stock_signal.normalized_state`
  - `stock_signal.observed_at`

### Requirement INTEGRATION-028: eBay seller listing lifecycle commands MUST be safety gated
Cabinet SHALL model eBay seller listing draft, publish, revise, end, and relist commands separately so local draft creation is not confused with external marketplace writes.

#### Scenario: Listing draft creation stays local-only
- **GIVEN** Cabinet is creating an eBay listing draft from a Cabinet item with a title
- **WHEN** the seller listing lifecycle command is previewed or executed with draft-only capability
- **THEN** the draft command MAY be allowed as a local-only action
- **AND** the command MUST report `remote_write=false`.

#### Scenario: Publish, revise, end, and relist require confirmed API capability
- **GIVEN** Cabinet previews an eBay seller listing publish, revise, end, or relist command
- **WHEN** the active account has no verified confirmed eBay API write capability
- **THEN** the command MUST be blocked with a write-capability-not-verified reason
- **AND** Cabinet MUST NOT call the eBay lifecycle adapter.

#### Scenario: Confirmed lifecycle writes use mocked eBay responses in tests
- **GIVEN** Cabinet has verified confirmed eBay API lifecycle capability in the command contract
- **WHEN** publish, revise, end, or relist is executed with explicit confirmation
- **THEN** Cabinet MAY call the lifecycle client
- **AND** backend tests MUST prove the command consumes mocked eBay responses instead of relying on live marketplace writes.

#### Scenario: Unconfirmed lifecycle writes remain blocked
- **GIVEN** Cabinet has verified confirmed eBay API lifecycle capability
- **WHEN** publish, revise, end, or relist is requested without explicit confirmation
- **THEN** the command MUST remain blocked with a confirmation-required reason
- **AND** the eBay lifecycle adapter MUST NOT be called.
