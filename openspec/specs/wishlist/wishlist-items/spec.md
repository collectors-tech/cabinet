## Purpose
Define wishlist item lifecycle behavior.

## Requirements
### Requirement WISHLIST-PRICING-DASHBOARD-001: Wishlist entries SHALL link to canonical items and target pricing
Cabinet SHALL treat the canonical item record as the source-of-truth for wishlist membership, using item lifecycle status plus attached wishlist metadata for target price, priority, and notes.

#### Scenario: Create wishlist entry
- **GIVEN** an authenticated user has an active profile and a canonical item exists with `item_id`
- **WHEN** the user submits `POST /api/wishlist` with `item_id`, `target_price`, `priority`, and `notes`
- **THEN** API MUST return `201`, persist wishlist metadata, and mark the canonical item lifecycle status as `wishlist`
- **AND** the persisted record MUST include:
  - `wishlist_id` (string)
  - `item_id` matching request
  - `target_price` (number)
  - `priority` (`low|medium|high`)
  - `notes` (string, optional)

### Requirement WISHLIST-ITEMS-002: Wishlist entries SHALL convert explicitly to owned inventory
Cabinet SHALL provide an explicit active-profile scoped conversion action that moves a wanted item into owned Inventory without relying on generic deletion semantics.

#### Scenario: Convert wishlist entry to owned
- **GIVEN** an authenticated user has an active profile with a wishlist entry for a canonical item
- **WHEN** the user submits `POST /api/wishlist/convert-owned` with the wishlist entry `id`
- **THEN** API MUST return `200`
- **AND** the wishlist entry MUST be removed from `/api/wishlist`
- **AND** the canonical item MUST no longer appear in `/api/items?status=wishlist`
- **AND** the canonical item MUST appear in `/api/items` with lifecycle status `active`

#### Scenario: Reject cross-profile wishlist conversion
- **GIVEN** a wishlist entry belongs to a different profile than the active profile
- **WHEN** the user submits `POST /api/wishlist/convert-owned` with that wishlist entry `id`
- **THEN** API MUST reject the conversion
- **AND** the original wishlist entry and canonical item status MUST remain unchanged

### Requirement WISHLIST-ITEMS-003: Wishlist purchase and delivery state SHALL synchronize downstream records
Cabinet SHALL treat wishlist `owned`/Purchased state as purchase intent evidence and explicit `delivered` state as inventory receipt evidence while preserving category on the canonical item.

#### Scenario: Purchased wishlist entry creates purchase lifecycle evidence
- **GIVEN** an authenticated user has an active profile with a wishlist entry for a canonical item
- **WHEN** the user creates or updates the wishlist entry with `owned=true`
- **THEN** Cabinet MUST persist the wishlist entry as Purchased
- **AND** Cabinet MUST create or update one `purchase` commerce lifecycle entry with `source=wishlist`, `external_ref` equal to the wishlist entry id, purchase amount, quantity, and notes
- **AND** Cabinet MUST create or update the linked expected-arrival record for that purchase lifecycle entry
- **AND** Cabinet MUST create or update one wishlist-linked inventory instance for the canonical item
- **AND** if an inventory instance for that canonical item already exists, Cabinet MUST link and increment that instance instead of creating a duplicate
- **AND** repeating the same Purchased save MUST update the existing wishlist-linked inventory instance without double-counting quantity
- **AND** the canonical item category MUST remain unchanged for downstream Inventory visibility

#### Scenario: Delivered wishlist entry creates inventory receipt evidence
- **GIVEN** an authenticated user has an active profile with a wishlist entry for a canonical item
- **WHEN** the user creates or updates the wishlist entry with `delivered=true`
- **THEN** Cabinet MUST persist the wishlist entry as both Delivered and Purchased
- **AND** Cabinet MUST mark the linked purchase arrival as `delivered`
- **AND** Cabinet MUST create or update one inventory instance with purchase condition, quantity, acquisition price, and acquisition date from wishlist purchase details
- **AND** Cabinet MUST mark the canonical item lifecycle status as `active`
- **AND** the canonical item category MUST remain unchanged for Inventory visibility
