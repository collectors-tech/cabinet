## Purpose
Define AU webshop provider family contract for slot-car collector discovery and stock-aware candidate ingestion.

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
