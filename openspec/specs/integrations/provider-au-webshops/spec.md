## Purpose
Define AU webshop provider family contract for slot-car collector discovery and stock-aware candidate ingestion.

## Family Contract References
This provider group reuses shared family-level contracts from:
- `openspec/specs/integrations/provider-api-families/spec.md`

Mapped reusable behavior:
- WooCommerce ingestion and fallback semantics -> `PROVIDER-FAMILY-001`
- Boost/Shopify runtime discovery + session handling semantics -> `PROVIDER-FAMILY-002`
- Algolia runtime discovery + drift-safe fallback semantics -> `PROVIDER-FAMILY-003`
- Shared pagination/stock normalization semantics -> `PROVIDER-FAMILY-004`

## Requirements
### Requirement INTEGRATION-011: AU webshop provider family MUST maintain domain catalog
Cabinet SHALL maintain provider entries for:
- bonzaslotcars.com.au
- frontlinehobbies.com.au
- hobbytechtoys.com.au
- andrewshobbies.com.au
- voglers.com.au
- acercmodels.com
- mrtoys.com.au

#### Scenario: AU provider catalog list
- **GIVEN** integrations registry endpoint is requested by an authenticated admin user
- **WHEN** AU webshop providers are returned
- **THEN** response MUST include all configured domains with `integration_mode: web_ingestion`

### Requirement INTEGRATION-012: AU webshop ingestion MUST extract stock observations
Cabinet SHALL parse stock/availability from webshop listing pages where available and persist normalized stock observations.

#### Scenario: Webshop stock extraction
- **GIVEN** listing HTML contains explicit stock/availability signal text
- **WHEN** parser normalization runs for candidate ingestion
- **THEN** candidate record MUST include:
  - `stock_signal.raw`
  - `stock_signal.normalized_state` (`in_stock|low_stock|out_of_stock|unknown`)
  - `stock_signal.source_domain`
  - `stock_signal.observed_at`

### Requirement OPS-001: AU webshop providers MUST enforce robots/terms policy and throttling
Cabinet MUST store per-domain crawling policy metadata including robots/terms review status, crawl delay/rate limit, and failure backoff behavior.

### Requirement OPS-002: Provider ingestion MUST support configurable items-per-page with safe limits
Provider query execution SHALL support per-provider `items_per_page` configuration to control request volume and avoid overscraping behavior.

#### Scenario: Use configured items-per-page
- **GIVEN** provider config defines `items_per_page`
- **WHEN** ingestion query is executed
- **THEN** request pagination MUST use configured value for API/listing fetches where supported
- **AND** runtime MUST report effective observed page size in run summary

#### Scenario: Safe cap enforcement
- **GIVEN** configured `items_per_page` exceeds provider-safe cap
- **WHEN** ingestion starts
- **THEN** runtime MUST clamp to provider-safe cap and emit warning/telemetry note
- **AND** default value MUST be conservative when unset

### Requirement INTEGRATION-013: Bonza search ingestion MUST support 36-target paging with dynamic fallback
Bonza provider ingestion SHALL attempt configured page-size target (36 where supported) and MUST fall back to detected site pagination while still traversing all result pages and aggregating normalized candidates.

#### Scenario: Bonza paginated search ingestion
- **GIVEN** Market Watch executes provider-scoped query against Bonza
- **WHEN** listing results span multiple pages
- **THEN** ingestion MUST navigate all result pages until terminal page regardless of effective per-page count
- **AND** candidates from all pages MUST be included exactly once in normalized output
- **AND** run output MUST report effective observed page size and page count

#### Scenario: WooCommerce per-page cookie hint
- **GIVEN** Bonza search endpoint is called via `/?post_type=product&s=<query>`
- **WHEN** provider runtime sets/receives `woocommerce_products_per_page=36`
- **THEN** ingestion SHOULD request 36-target listing pages using cookie/session hint
- **AND** runtime MUST fall back to detected effective page size if server response differs

### Requirement INTEGRATION-014: Bonza watched-car stock enrichment MUST be API-first with detail-page fallback
For watched Bonza cars, ingestion SHALL prefer WooCommerce Store API stock fields and use detail-page parsing only when API stock data is missing/insufficient.

### Requirement INTEGRATION-015: Frontline provider config discovery SHALL be runtime-resolved from site assets with safe fallback
Frontline integration SHALL discover Algolia runtime config (application ID, search key, index names) from maintained site assets (e.g., `pd-search.js`) and use cached last-known-good config when discovery fails.

### Requirement INTEGRATION-016: Hobbytech provider SHALL support Shopify search via Boost/mybcapps endpoint with runtime config discovery
Hobbytech integration SHALL support Shopify-backed search using Boost Commerce endpoint (`services.mybcapps.com/bc-sf-filter/search`) with runtime discovery of required query/session parameters and resilient fallback behavior.

#### Scenario: Hobbytech search execution
- **GIVEN** provider query executes for Hobbytech
- **WHEN** runtime calls Boost/mybcapps search endpoint
- **THEN** response parsing MUST extract product candidates from structured fields and/or returned HTML payload blocks
- **AND** pagination MUST honor provider limit/page parameters until terminal page

#### Scenario: Hobbytech config/session drift fallback
- **GIVEN** required session/template parameters drift or expire
- **WHEN** search request fails or returns unusable payload
- **THEN** runtime MUST refresh discovery inputs from site assets/pages and retry with bounded attempts
- **AND** emit drift warning with fallback status for operator review

#### Scenario: Frontline config discovery from asset
- **GIVEN** Frontline provider run starts
- **WHEN** runtime fetches configured discovery asset path(s)
- **THEN** parser MUST extract Algolia application ID, public search key, and index names from script content
- **AND** discovery output MUST be versioned and timestamped in provider metadata

#### Scenario: Config drift/fallback handling
- **GIVEN** asset parsing fails or values change unexpectedly
- **WHEN** provider query executes
- **THEN** runtime MUST fallback to last-known-good config if available
- **AND** emit drift warning/event for operator review without hard-crashing entire provider lane

#### Scenario: Watched car stock-level enrichment (API-first)
- **GIVEN** watched car candidate exists for Bonza source
- **WHEN** product data is fetched from `wp-json/wc/store/v1/products`
- **THEN** runtime MUST extract stock signal from API fields first (e.g., `is_in_stock`, `low_stock_remaining`, `add_to_cart` availability)
- **AND** normalized stock state MUST update integration stock observation for purchase prompting logic

#### Scenario: Detail-page fallback enrichment
- **GIVEN** API stock signal is missing/insufficient for watched car candidate
- **WHEN** enrichment fallback runs
- **THEN** runtime MUST request Bonza product detail page for that candidate
- **AND** extracted fallback stock signal MUST be merged with source attribution (`api|detail_page`)

#### Scenario: Domain throttle applied
- **GIVEN** domain policy declares `crawl_delay_ms` and `max_requests_per_minute`
- **WHEN** scanner executes multiple AU webshop requests in one run window
- **THEN** scheduler MUST enforce throttle policy and transition provider state to `degraded` after repeated policy violations

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-AU-01 | Bonza paginated search | All 36-item pages traversed and aggregated without duplicates | planned: `ui.web/cypress/e2e/integrations/provider-bonza/spec.cy.ts` `bonza-paginated-search-all-pages` |
| UC-AU-02 | Bonza watched stock enrichment | Watched cars trigger detail-page stock extraction and normalized update | planned: `ui.web/cypress/e2e/integrations/provider-bonza/spec.cy.ts` `bonza-watched-detail-stock-level` |
| UC-AU-03 | Provider items-per-page config | Query run uses configured page-size with safe-cap fallback + telemetry | planned: `ui.web/cypress/e2e/integrations/provider-bonza/spec.cy.ts` `bonza-items-per-page-config-safe-cap` |
| UC-AU-04 | Frontline Algolia config discovery | Provider extracts app/key/index from site asset and records discovery metadata | planned: `ui.web/cypress/e2e/integrations/provider-frontline/spec.cy.ts` `frontline-algolia-config-discovery` |
| UC-AU-05 | Frontline config drift fallback | Provider uses last-known-good config on parse drift and emits warning event | planned: `ui.web/cypress/e2e/integrations/provider-frontline/spec.cy.ts` `frontline-config-drift-fallback` |
| UC-AU-06 | Hobbytech Shopify/Boost search | Provider executes mybcapps endpoint and parses candidate products deterministically | planned: `ui.web/cypress/e2e/integrations/provider-hobbytech/spec.cy.ts` `hobbytech-mybcapps-search` |
| UC-AU-07 | Hobbytech session drift recovery | Provider refreshes discovery/session params and retries boundedly on drift | planned: `ui.web/cypress/e2e/integrations/provider-hobbytech/spec.cy.ts` `hobbytech-session-drift-recovery` |
