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

### Requirement PROVIDER-FAMILY-005: Provider onboarding SHALL support URL-based API family auto-detection
Given a provider homepage URL, onboarding SHALL run detection heuristics to infer likely API family and confidence before manual confirmation.

### Requirement PROVIDER-FAMILY-006: BigCommerce family SHALL support storefront-access-first retrieval with token-aware fallback
BigCommerce-backed providers SHALL use available storefront-accessible data paths first and support token-aware GraphQL/management API integration when credentials are provided.

### Requirement PROVIDER-FAMILY-007: Doofinder family SHALL support hashid-based search with origin-aware headers
Doofinder-backed providers SHALL execute search via Doofinder search endpoint using discovered `hashid` and required origin/referrer headers when enforced by endpoint policy.

#### Scenario: Doofinder search execution
- **GIVEN** provider is classified as Doofinder family and hashid is discovered from Doofinder config
- **WHEN** runtime executes query
- **THEN** runtime MUST call Doofinder search endpoint with query/page/rpp params
- **AND** runtime MUST include origin/referrer headers where required to avoid forbidden responses
- **AND** provider-specific run output MUST persist normalized candidates into the shared scanner/Discoveries candidate store and hydrate latest-run snapshot metadata on query-set reload

#### Scenario: Doofinder discovery inputs
- **GIVEN** onboarding detection scans provider assets
- **WHEN** Doofinder scripts/config are present
- **THEN** runtime MUST extract `store` UUID, `zone`, and `hashid` (`search_engines` mapping) for query execution

#### Scenario: BigCommerce public/storefront-access run
- **GIVEN** provider is classified as BigCommerce family and no privileged API token is configured
- **WHEN** query run executes
- **THEN** runtime MUST use storefront-accessible endpoints/content paths and normalize candidates
- **AND** runtime MUST record capability limits when stock/variant depth is unavailable without token

#### Scenario: BigCommerce token-enabled run
- **GIVEN** provider has valid BigCommerce storefront/admin credentials configured
- **WHEN** query run executes
- **THEN** runtime MAY use GraphQL Storefront and/or Management API paths for richer catalog/stock fields
- **AND** run summary MUST declare auth mode and data depth source

#### Scenario: URL-based family detection
- **GIVEN** user enters provider homepage URL
- **WHEN** detection process scans page HTML/assets/scripts/known endpoints
- **THEN** runtime MUST propose `api_family` with confidence score and evidence markers
- **AND** user MUST confirm or override mapping before saving provider

#### Scenario: Detection heuristics evidence
- **GIVEN** detector runs against provider URL
- **WHEN** result is returned
- **THEN** evidence MUST include matched markers such as:
  - WooCommerce: `/wp-json/wc/store/v1`, `woocommerce` markers
  - Boost/Shopify: `services.mybcapps.com/bc-sf-filter`, Boost script signatures
  - Algolia: `algoliasearch(` calls, app/search key/index markers
  - Shopify JSON: `/products.json` or `/collections/*/products.json` endpoint responses
  - Doofinder: `cdn.doofinder.com` loader/config script + `hashid`/`search_engines` markers

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
| UC-PF-05 | URL auto-detection | Entered provider URL returns proposed family + confidence + evidence markers | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `provider-family-url-autodetect` |
| UC-PF-06 | Manual override after detection | User can override detected family before provider save | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `provider-family-autodetect-override` |
| UC-PF-07 | BigCommerce storefront-access mode | Provider run works with storefront-accessible data paths and declares limits | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `bigcommerce-storefront-mode` |
| UC-PF-08 | BigCommerce token-enabled mode | Provider uses token-enabled API paths for deeper stock/catalog fields | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `bigcommerce-token-enabled-mode` |
| UC-PF-09 | Doofinder hashid search | Provider executes Doofinder search with discovered hashid and origin-aware headers | planned: `ui.web/cypress/e2e/integrations/provider-families/spec.cy.ts` `doofinder-hashid-search-contract` |
