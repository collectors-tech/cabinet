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

### Requirement DEFAULT-SITE-SEARCH-004: Provider-bound saved searches MUST support create/edit/delete with persisted filter payloads
Cabinet MUST support provider-scoped saved searches for Bonza/Amazon/eBay with persisted filter payloads.

#### Scenario: Manage saved searches with provider filters
- **GIVEN** user is in Market Watch with active profile and provider selector controls
- **WHEN** user creates, edits, and deletes a saved search with provider scope and keyword filters
- **THEN** runtime MUST persist updated saved search payload via scanner query-set APIs
- **AND** UI MUST reflect deterministic state transitions for create/edit/delete outcomes

### Requirement DEFAULT-SITE-SEARCH-005: Saved searches MUST support run-now and scheduled refresh execution
Cabinet MUST execute saved searches immediately and through scheduled refresh flow with deterministic summaries.

#### Scenario: Run-now and scheduled refresh execution
- **GIVEN** one or more saved searches exist with provider scope and schedule metadata
- **WHEN** user runs `Run Now` or triggers scheduled refresh
- **THEN** runtime MUST execute matching query sets and return deterministic run summary payloads
- **AND** UI MUST surface execution status and summary data for user verification

### Requirement DEFAULT-SITE-SEARCH-006: Saved-search output MUST support Discoveries and Wishlist handoff
Cabinet MUST provide workflow handoff from saved-search output into Discoveries and Wishlist actions.

#### Scenario: Handoff from output detail
- **GIVEN** saved-search output detail is open with candidate results
- **WHEN** user triggers Discoveries handoff or Wishlist handoff action
- **THEN** runtime MUST call discovery/wishlist APIs with deterministic payload contract
- **AND** UI MUST surface action feedback indicating handoff result
- **AND** Wishlist handoff metadata MUST preserve the source provider, query-set id, query name, and saved provider scope that produced the discovery candidate
