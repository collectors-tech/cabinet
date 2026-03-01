## Purpose
Define wishlist item lifecycle behavior.

## Requirements
### Requirement WISHLIST-PRICING-DASHBOARD-001: Wishlist entries SHALL link to canonical items and target pricing
Cabinet SHALL persist wishlist records with target price, priority, and notes.

#### Scenario: Create wishlist entry
- **GIVEN** an authenticated user has an active profile and a canonical item exists with `item_id`
- **WHEN** the user submits `POST /api/wishlist` with `item_id`, `target_price`, `priority`, and `notes`
- **THEN** API MUST return `201` and persisted record MUST include:
  - `wishlist_id` (string)
  - `item_id` matching request
  - `target_price` (number)
  - `priority` (`low|medium|high`)
  - `notes` (string, optional)
