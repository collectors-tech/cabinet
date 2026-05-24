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
- **WHEN** Cabinet maps the listing into its buyer-interest intake model, previews it through `POST /api/providers/ebay/buyer-interest/preview`, or persists it through `POST /api/providers/ebay/buyer-interest/import`
- **THEN** the mapped record MUST include:
  - source provider `ebay`
  - source account identifier when available
  - source listing id
  - normalized interest state
  - deterministic provenance key
  - owned inventory flag set to false
- **AND** watched/saved states MUST target Wishlist while liked/cart-like states MUST target Discoveries unless a later user action promotes the item.
- **AND** persisted imports MUST retain the deterministic provenance key in the saved Wishlist entry or Discovery candidate action.

#### Scenario: Preview and import buyer-interest from the eBay integration UI
- **GIVEN** the eBay integration dialog is open for a configured profile
- **WHEN** the operator previews or imports buyer-interest payloads from the dialog
- **THEN** Cabinet MUST call the buyer-interest preview/import endpoints with the edited payload, summarize Wishlist and Discovery destination counts, and show per-listing destination/provenance outcomes.
- **AND** the dialog MUST keep remote write-back visibly blocked unless eBay write-back capability has been verified.

### Requirement INTEGRATION-026: eBay buyer-interest write-back MUST be capability gated
Cabinet SHALL only offer add/remove/watch-state write-back when the exact eBay API capability has been verified for the active account and marketplace.

#### Scenario: Unsupported eBay write-back stays blocked
- **GIVEN** Cabinet has imported eBay buyer-interest state but has no verified write-back capability
- **WHEN** a write-back action is evaluated, previewed through `POST /api/providers/ebay/buyer-interest/preview`, or persisted through `POST /api/providers/ebay/buyer-interest/import`
- **THEN** Cabinet MUST report write-back as blocked with a capability-not-verified reason
- **AND** it MUST NOT imply the remote eBay watch, saved, liked, or cart-like state was changed.

### Requirement INTEGRATION-027: eBay seller operations MUST expose truthful capability-gated states
Cabinet SHALL represent seller messages, notifications, sold orders, fulfilment, and offers as separate eBay seller operation capabilities so unavailable or read-only API support is not presented as a writable workflow.

#### Scenario: Unsupported seller operation capabilities stay blocked
- **GIVEN** Cabinet has no verified eBay seller operation capability for messages, notifications, sold orders, fulfilment, or offers
- **WHEN** seller operation statuses are evaluated
- **THEN** each operation MUST report read availability as false
- **AND** each operation MUST report write availability as false
- **AND** each operation MUST expose a capability-not-verified blocker instead of showing a usable workflow.

#### Scenario: Read-only seller operation sync does not imply write availability
- **GIVEN** Cabinet has read-only eBay API support for seller messages or sold orders
- **WHEN** seller operation statuses are evaluated
- **THEN** the matching operations MAY report read availability as true
- **AND** write availability MUST remain false with a write-capability-not-verified blocker
- **AND** Cabinet MUST NOT imply replies, fulfilment updates, or offer sends were executed remotely.

#### Scenario: Confirmed seller operation writes require explicit confirmation
- **GIVEN** Cabinet has verified API support for a seller notification, fulfilment, or offer write workflow
- **WHEN** the matching seller operation status is evaluated
- **THEN** read and write availability MAY be true
- **AND** the status MUST mark confirmation as required before any external eBay write is executed.
