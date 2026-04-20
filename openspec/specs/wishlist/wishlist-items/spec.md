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
