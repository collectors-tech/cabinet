## Purpose
Define wishlist, pricing, and dashboard action-signal behavior.

## Requirements
### Requirement: Wishlist entries SHALL link to canonical items and target pricing
Cabinet SHALL persist wishlist records with target price, priority, and notes.

#### Scenario: Create wishlist entry
- **WHEN** user adds item to wishlist with target
- **THEN** Cabinet SHALL persist wishlist-linked record

### Requirement: Pricing tracking SHALL record source-aware daily snapshots
Cabinet SHALL persist min/median/latest pricing and per-source stock transitions.

#### Scenario: Snapshot run
- **WHEN** pricing snapshot executes for tracked items
- **THEN** Cabinet SHALL append historical data points per item/source

### Requirement: Pricing outputs SHALL include graph and export views
Cabinet SHALL support graph, source breakdown, stats/trend, and history export.

#### Scenario: Export price history
- **WHEN** user requests pricing export
- **THEN** Cabinet SHALL return export payload for selected scope

### Requirement: Dashboard SHALL prioritize actionable collector signals
Cabinet SHALL show discoveries, wishlist hits, price drops, stock alerts, restock alerts, and collection stats.

#### Scenario: Dashboard attention refresh
- **WHEN** user refreshes dashboard
- **THEN** dashboard SHALL render current actionable signals from runtime data
