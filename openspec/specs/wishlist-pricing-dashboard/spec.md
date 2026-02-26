## Purpose
Define wishlist, pricing, and dashboard action-signal behavior.

## Requirements
### Requirement WISHLIST-PRICING-DASHBOARD-001: Wishlist entries SHALL link to canonical items and target pricing
Cabinet SHALL persist wishlist records with target price, priority, and notes.

#### Scenario: Create wishlist entry
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user adds item to wishlist with target
- **THEN** Cabinet SHALL persist wishlist-linked record

### Requirement WISHLIST-PRICING-DASHBOARD-002: Pricing tracking SHALL record source-aware daily snapshots
Cabinet SHALL persist min/median/latest pricing and per-source stock transitions.

#### Scenario: Snapshot run
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** pricing snapshot executes for tracked items
- **THEN** Cabinet SHALL append historical data points per item/source

### Requirement WISHLIST-PRICING-DASHBOARD-003: Pricing outputs SHALL include graph and export views
Cabinet SHALL support graph, source breakdown, stats/trend, and history export.

#### Scenario: Export price history
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user requests pricing export
- **THEN** Cabinet SHALL return export payload for selected scope

### Requirement WISHLIST-PRICING-DASHBOARD-004: Dashboard SHALL prioritize actionable collector signals
Cabinet SHALL show discoveries, wishlist hits, price drops, stock alerts, restock alerts, and collection stats.

#### Scenario: Dashboard attention refresh
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user refreshes dashboard
- **THEN** dashboard SHALL render current actionable signals from runtime data
