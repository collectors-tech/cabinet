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

### Requirement INTEGRATION-013: Bonza search ingestion MUST support 36-target paging with dynamic fallback
Bonza provider ingestion SHALL attempt configured page-size target (36 where supported) and MUST fall back to detected site pagination while still traversing all result pages and aggregating normalized candidates.

#### Scenario: Bonza paginated search ingestion
- **GIVEN** Market Watch executes provider-scoped query against Bonza
- **WHEN** listing results span multiple pages
- **THEN** ingestion MUST navigate all result pages until terminal page regardless of effective per-page count
- **AND** candidates from all pages MUST be included exactly once in normalized output
- **AND** run output MUST report effective observed page size and page count

### Requirement INTEGRATION-014: Bonza detail-page enrichment MUST fetch watched-car stock level
For watched Bonza cars, ingestion SHALL fetch detail page data to capture stock level signals not present on listing cards.

#### Scenario: Watched car stock-level enrichment
- **GIVEN** watched car candidate exists for Bonza source
- **WHEN** listing-level stock signal is missing/insufficient
- **THEN** runtime MUST request Bonza detail page for that candidate
- **AND** extracted stock level MUST update integration stock observation for purchase prompting logic

#### Scenario: Domain throttle applied
- **GIVEN** domain policy declares `crawl_delay_ms` and `max_requests_per_minute`
- **WHEN** scanner executes multiple AU webshop requests in one run window
- **THEN** scheduler MUST enforce throttle policy and transition provider state to `degraded` after repeated policy violations

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-AU-01 | Bonza paginated search | All 36-item pages traversed and aggregated without duplicates | planned: `ui.web/cypress/e2e/integrations/provider-bonza/spec.cy.ts` `bonza-paginated-search-all-pages` |
| UC-AU-02 | Bonza watched stock enrichment | Watched cars trigger detail-page stock extraction and normalized update | planned: `ui.web/cypress/e2e/integrations/provider-bonza/spec.cy.ts` `bonza-watched-detail-stock-level` |
