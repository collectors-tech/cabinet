## Purpose
Define dashboard attention and priority signal behavior.

## Requirements
### Requirement WISHLIST-PRICING-DASHBOARD-004: Dashboard SHALL prioritize actionable collector signals
Cabinet SHALL show discoveries, wishlist hits, price drops, stock alerts, restock alerts, and collection stats.

#### Scenario: Dashboard attention refresh
- **GIVEN** an authenticated user opens Dashboard with an active profile
- **WHEN** dashboard data refresh executes
- **THEN** response/rendered view MUST include:
  - `new_discoveries_count`
  - `wishlist_hits_count`
  - `price_drop_count`
  - `stock_alert_count`
  - `restock_alert_count`
  - `collection_stats.total_items`
