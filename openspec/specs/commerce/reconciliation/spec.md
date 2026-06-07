## Purpose
Define the first deliverable slice for unified commerce reconciliation across intent states, purchases, and expected arrivals.

## Requirements
### Requirement COMMERCE-RECONCILIATION-001: Commerce lifecycle SHALL persist profile-scoped intent and purchase states
Cabinet SHALL persist profile-scoped lifecycle entries for `wishlist`, `watchlist`, `cart`, `offer`, and `purchase` states linked to canonical items.

#### Scenario: Record lifecycle state
- **GIVEN** an authenticated user has an active profile and a canonical item exists
- **WHEN** the user submits `POST /api/commerce/lifecycle` with `item_id`, `state`, and optional source/order metadata
- **THEN** Cabinet MUST return `201` with a persisted lifecycle entry linked to the active profile.

### Requirement COMMERCE-RECONCILIATION-002: Purchases SHALL create expected arrivals instead of inventory items directly
Cabinet SHALL model purchases as expected-arrival work first, not immediate inventory reconciliation.

#### Scenario: Purchase creates expected arrival
- **GIVEN** an authenticated user records a `purchase` lifecycle entry for a canonical item
- **WHEN** Cabinet persists the purchase
- **THEN** Cabinet MUST also create an `expected_arrival` record linked to that purchase with status `expected`.

### Requirement COMMERCE-RECONCILIATION-003: Expected arrivals SHALL support reconciliation state transitions
Cabinet SHALL support `expected`, `delivered`, `reconciled`, and `cancelled` expected-arrival states.

#### Scenario: Reconcile delivered purchase
- **GIVEN** an expected arrival exists for the active profile
- **WHEN** the user updates the arrival status to `reconciled` with a delivered date and optional instance reference
- **THEN** Cabinet MUST persist the updated reconciliation state and expose it through `GET /api/commerce/arrivals`.

### Requirement COMMERCE-RECONCILIATION-004: Purchases SHALL preserve source-backed forwarder provenance
Cabinet SHALL treat forwarder imports as source/provenance evidence for purchase reconciliation rather than as a disconnected primary package workflow.

#### Scenario: Preserve forwarder source evidence for purchase matching
- **GIVEN** an authenticated user has an active profile and Cabinet imports a Stackry or freight-forwarder record with package, sender, shipment, tracking, import-time, and raw-source evidence
- **WHEN** Cabinet stores the imported evidence for reconciliation
- **THEN** Cabinet MUST keep the forwarder source/provenance metadata traceable to purchase and expected-arrival candidates, including source/provider, package or reference id, sender/source text, import timing, raw/source payload reference when available, and the current match-review state.

### Requirement COMMERCE-RECONCILIATION-005: Purchases SHALL expose forwarder match review states
Cabinet SHALL expose forwarder-backed purchase match state from the Purchases review surface so users can inspect unmatched, suggested, confirmed, and rejected or ignored source matches without losing provenance.

#### Scenario: Review and decide a source-backed purchase match
- **GIVEN** a profile has a purchase or expected-arrival candidate and forwarder source evidence that is unmatched, suggested, confirmed, or rejected
- **WHEN** the user reviews the Purchases surface
- **THEN** Cabinet MUST show the match state, enough source evidence to inspect the suggested match, and controls or follow-through paths to confirm a good match or reject/ignore a bad match while preserving both purchase evidence and forwarder provenance.
