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
