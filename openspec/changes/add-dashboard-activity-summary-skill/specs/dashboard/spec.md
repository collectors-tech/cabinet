## MODIFIED Requirements

### Requirement: Dashboard SHALL prioritize actionable collector signals
Cabinet SHALL show discoveries, wishlist hits, price drops, stock alerts, restock alerts, and collection stats, and SHALL expose the same grounded data to governed read-only Agent summaries without duplicating calculations.

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

#### Scenario: Agent summarizes Dashboard attention
- **GIVEN** an authenticated Agent Skill invocation is bound to one active profile
- **WHEN** `cabinet.dashboard.summarise_activity` asks what needs attention or what changed for a bounded time window
- **THEN** Cabinet SHALL compute the summary from canonical Dashboard/application data for that profile
- **AND** it SHALL include discoveries, wishlist hits, price drops, low-stock signals, restocks, collection count/value where available, recently added items, usable destination links, and relevant record identifiers
- **AND** it SHALL distinguish evidence-backed time-window changes from current snapshot values
- **AND** empty, partial, missing, or dependency-failure states SHALL be truthful and actionable instead of fabricating changes
- **AND** no response SHALL include data from another profile
