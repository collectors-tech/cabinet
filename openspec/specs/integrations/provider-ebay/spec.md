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

#### Scenario: Seller operation statuses are visible through the integration surface
- **GIVEN** the eBay provider registry payload is loaded by the Integrations screen
- **WHEN** Cabinet renders the eBay integration dialog
- **THEN** seller messages, notifications, sold orders, fulfilment, and offers MUST each display read availability, write availability, and any blocker reason
- **AND** unavailable seller operation workflows MUST remain visibly blocked rather than appearing as executable actions.

#### Scenario: Seller operation action preview gates remote writes
- **GIVEN** Cabinet previews a seller message reply, notification acknowledgement, sold-order sync, fulfilment update, or offer action
- **WHEN** the selected seller operation has no verified write capability or only read-only sync capability
- **THEN** the preview MUST report the action as not allowed for remote write with a capability blocker.
- **AND** read-only sync actions MAY be allowed only when read availability is true and MUST NOT report a remote write.
- **AND** confirmed API write capabilities MUST still require explicit confirmation before the preview reports a remote write as allowed.

#### Scenario: Seller operation preview actions are available from the integration UI
- **GIVEN** the eBay integration dialog shows seller operation statuses
- **WHEN** the operator previews a supported read-only sync or confirmed seller-operation action from the dialog
- **THEN** Cabinet MUST call the seller operation preview API before showing an allowed action outcome.
- **AND** the UI MUST display whether the preview is allowed and whether it would perform a remote write.
- **AND** unsupported seller operations MUST keep preview controls disabled instead of presenting a runnable workflow.

#### Scenario: Seller operation execute actions only complete safe local sync
- **GIVEN** Cabinet executes a seller operation through the seller-operation execute API or integration dialog
- **WHEN** the selected action is a read-only sync with verified read availability
- **THEN** Cabinet MAY mark the action executed locally and MUST report local_only=true.
- **AND** Cabinet MUST report remote_write=false for that completed read-only sync.
- **AND** the response MUST include a per-operation local read result model for messages, notifications, sold orders, fulfilment, or offers with records and summary counts.
- **AND** confirmed remote-write seller operations MUST remain blocked with an adapter-not-configured blocker until a real eBay write adapter is wired.
- **AND** the integration UI MUST display the execute status separately from preview status so local sync completion is not confused with an external eBay write.

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

#### Scenario: Listing lifecycle APIs expose truthful preview and execution states
- **GIVEN** Cabinet receives an eBay seller listing lifecycle API request
- **WHEN** the request previews or executes draft, publish, revise, end, or relist
- **THEN** the HTTP response MUST expose the normalized command, capability, confirmation, local-only, remote-write, allowed, status, blocker, and response fields.
- **AND** local draft execution MAY return a Cabinet-local draft response.
- **AND** confirmed remote lifecycle writes MUST remain blocked with an adapter-required blocker until a real eBay lifecycle adapter is configured.
