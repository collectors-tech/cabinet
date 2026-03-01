## Purpose
Define tracked pricing snapshot, trend, and export behavior.

## Requirements
### Requirement WISHLIST-PRICING-DASHBOARD-002: Pricing tracking SHALL record source-aware daily snapshots
Cabinet SHALL persist min/median/latest pricing and per-source stock transitions.

#### Scenario: Snapshot run
- **GIVEN** one or more tracked items exist and at least one provider is enabled for pricing
- **WHEN** scheduled pricing snapshot job runs for the active profile
- **THEN** API/runtime MUST append one history point per tracked item/provider with:
  - `captured_at` (timestamp)
  - `price_min` (number)
  - `price_median` (number)
  - `price_latest` (number)
  - `in_stock` (boolean)

### Requirement WISHLIST-PRICING-DASHBOARD-003: Pricing outputs SHALL include graph and export views
Cabinet SHALL support graph, source breakdown, stats/trend, and history export.

#### Scenario: Export price history
- **GIVEN** an authenticated user selects tracked items in Pricing workspace
- **WHEN** the user requests export for a date range
- **THEN** export endpoint MUST return `200` with a file payload containing:
  - `item_id`
  - `provider_id`
  - `captured_at`
  - `price_latest`
  - `in_stock`
