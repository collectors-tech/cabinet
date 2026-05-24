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
  - and scanner run APIs MUST return `401` with `error_code="PROVIDER_AUTH_INVALID"` when bearer token is expired or rejected

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
- **GIVEN** eBay Browse item summary contains `estimatedAvailabilities` status and quantity signals
- **WHEN** provider normalization and candidate persistence run
- **THEN** candidate state MUST preserve normalized stock fields including:
  - `stock_state`
  - `stock_count`
  - `last_seen`

### Requirement INTEGRATION-025: eBay buyer-interest sync MUST preserve state and provenance
Cabinet SHALL import eBay watched, saved, liked, and cart-like buyer-interest states without collapsing them into owned inventory.

#### Scenario: Import eBay buyer-interest state
- **GIVEN** an eBay account sync returns listing interest state `watched`, `saved`, `liked`, or `cart_like`
- **WHEN** Cabinet maps the listing into its buyer-interest intake model
- **THEN** the mapped record MUST include:
  - source provider `ebay`
  - source account identifier when available
  - source listing id
  - normalized interest state
  - deterministic provenance key
  - owned inventory flag set to false
- **AND** watched/saved states MUST target Wishlist while liked/cart-like states MUST target Discoveries unless a later user action promotes the item.

### Requirement INTEGRATION-026: eBay buyer-interest write-back MUST be capability gated
Cabinet SHALL only offer add/remove/watch-state write-back when the exact eBay API capability has been verified for the active account and marketplace.

#### Scenario: Unsupported eBay write-back stays blocked
- **GIVEN** Cabinet has imported eBay buyer-interest state but has no verified write-back capability
- **WHEN** a write-back action is evaluated
- **THEN** Cabinet MUST report write-back as blocked with a capability-not-verified reason
- **AND** it MUST NOT imply the remote eBay watch, saved, liked, or cart-like state was changed.
