## Purpose
Define default provider search behavior across supported marketplaces/shops.

## Requirements
### Requirement INTEGRATION-016: Default provider search set MUST include configured core sources
Cabinet SHALL default discovery/search runs to enabled core providers unless user overrides query scope.

#### Scenario: Run search with default provider scope
- **GIVEN** provider registry has enabled providers for `ebay`, `amazon`, and configured AU shops
- **WHEN** user runs scanner/discovery search without selecting explicit providers
- **THEN** runtime MUST execute search against all enabled default providers and return `provider_id` in each candidate row

### Requirement INTEGRATION-017: Default provider scope MUST remain user-configurable per query set
Cabinet SHALL allow users to override default provider scope and persist that selection per query set.

#### Scenario: Save custom provider scope
- **GIVEN** user edits a query set and deselects one or more providers
- **WHEN** query set is saved and executed later
- **THEN** runtime MUST use saved provider scope, not global default, and return `200` with results only from selected providers
