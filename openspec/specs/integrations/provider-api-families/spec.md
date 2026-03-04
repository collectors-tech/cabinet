## Purpose
Define reusable integration contracts for common commerce API families so new providers can be onboarded without redefining parsing/pagination/stock semantics each time.

## Requirements

### Requirement PROVIDER-FAMILY-001: WooCommerce family SHALL use Store API-first product ingestion
WooCommerce-backed providers SHALL prefer `wp-json/wc/store/v1/products` (or equivalent public Store API surface) for query and pagination before HTML scraping.

#### Scenario: WooCommerce API-first query
- **GIVEN** provider is classified as WooCommerce family
- **WHEN** query run executes
- **THEN** runtime MUST call Store API endpoint with query/paging params and normalize product candidates
- **AND** HTML listing/detail scraping MUST be fallback-only when required fields are unavailable via API

### Requirement PROVIDER-FAMILY-002: Boost/Shopify family SHALL use Boost search API with runtime config discovery
Boost-backed Shopify providers SHALL execute search via Boost endpoint (`services.mybcapps.com/bc-sf-filter/search` or equivalent) using runtime-discovered config/session parameters.

#### Scenario: Boost API query run
- **GIVEN** provider is classified as Boost family
- **WHEN** query run executes
- **THEN** runtime MUST resolve discovery inputs (shop/template/session/query defaults), call Boost endpoint, and normalize candidates
- **AND** response parsing MUST support structured product fields with HTML fallback parsing when needed

### Requirement PROVIDER-FAMILY-003: Algolia family SHALL support runtime key/index discovery with drift-safe fallback
Algolia-backed providers SHALL discover application/search keys and index names from site assets/runtime config and fallback to last-known-good settings on drift.

#### Scenario: Algolia config drift
- **GIVEN** runtime config discovery fails or discovered values change unexpectedly
- **WHEN** provider query executes
- **THEN** runtime MUST use last-known-good config for bounded continuity and emit drift warning telemetry

### Requirement PROVIDER-FAMILY-004: Family contracts SHALL expose shared pagination/stock normalization semantics
All API families SHALL map to common pagination/stock schema to keep Market Watch behavior consistent.

#### Scenario: Unified normalization output
- **GIVEN** provider run returns family-specific payload
- **WHEN** normalization completes
- **THEN** output MUST include common fields: `provider_id`, `query`, `page`, `effective_page_size`, `candidate_count`, `stock_signal` (with source attribution), and `observed_at`

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-PF-01 | Woo API-first run | Woo provider uses Store API first with fallback behavior | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `woo-api-first-contract` |
| UC-PF-02 | Boost API run | Boost provider resolves config/session and parses candidates deterministically | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `boost-api-contract` |
| UC-PF-03 | Algolia drift fallback | Algolia provider continues with last-known-good config and warning telemetry | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `algolia-drift-fallback-contract` |
| UC-PF-04 | Unified normalization | Family-specific payloads normalize to common run-summary fields | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `provider-family-normalization` |
