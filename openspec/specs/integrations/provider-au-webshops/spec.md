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
- Lightspeed storefront catalogue parsing and health-check semantics -> `PROVIDER-FAMILY-005`
- Shopify public catalogue parsing and health-check semantics -> `PROVIDER-FAMILY-010`

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
- hobbyco.com.au
- metrohobbies.com.au

#### Scenario: AU provider catalog list
- **GIVEN** integrations registry endpoint is requested by an authenticated admin user
- **WHEN** AU webshop providers are returned
- **THEN** response MUST include all configured domains with `integration_mode: web_ingestion`

### Requirement PROVIDER-AU-WEBSHOPS-004: AU webshop provider allowlist MUST be deterministic and test-covered
Cabinet SHALL maintain a deterministic approved AU webshop domain allowlist used by provider registry payload generation.

#### Scenario: Approved allowlist contract
- **GIVEN** provider registry payload is generated for active profile
- **WHEN** AU webshop domains are enumerated from registry output
- **THEN** the domain set MUST include exactly:
  - `bonzaslotcars.com.au`
  - `frontlinehobbies.com.au`
  - `hobbytechtoys.com.au`
  - `andrewshobbies.com.au`
  - `voglers.com.au`
  - `acercmodels.com`
  - `mrtoys.com.au`
  - `hobbyco.com.au`
  - `metrohobbies.com.au`
- **AND** allowlist verification MUST be covered by automated runtime tests

### Requirement PROVIDER-AU-WEBSHOPS-005: AU webshop domain source MUST be configuration-driven with deterministic fallback
Cabinet SHALL resolve AU webshop domain catalog from profile/runtime configuration (`integration.au_webshops.domains`) so operators can update provider domains without code changes or recompilation.

#### Scenario: Config-driven domain source
- **GIVEN** active profile settings include `integration.au_webshops.domains` with comma-separated domains
- **WHEN** `/api/providers/registry` is requested
- **THEN** AU webshop provider entries MUST be generated from configured domains
- **AND** provider IDs MUST follow `au-webshop-<normalized-domain>`

#### Scenario: Invalid config fallback
- **GIVEN** `integration.au_webshops.domains` is missing, blank, or parses to zero valid domains
- **WHEN** `/api/providers/registry` is requested
- **THEN** runtime MUST fallback to default approved allowlist domain set
- **AND** fallback behavior MUST be deterministic across repeated requests

### Requirement PROVIDER-AU-WEBSHOPS-006: AU hobby shop registry entries MUST expose adapter-matrix metadata
Cabinet SHALL classify each approved AU hobby shop provider with stable adapter metadata so Add Integration, source matching, Market Watch, and future parser health checks can route provider work without a duplicate hardcoded provider list.

#### Scenario: Adapter matrix is projected through provider registry
- **GIVEN** `/api/providers/registry` returns the approved AU hobby shop providers
- **WHEN** registry consumers inspect provider metadata
- **THEN** Acer RC Models MUST use a `lightspeed-storefront` adapter
- **AND** Andrew's Hobbies and Metro Hobbies MUST use `shopify-storefront` candidate metadata
- **AND** Voglers MUST use `bigcommerce-storefront`
- **AND** Frontline Hobbies MUST use `generic-structured-storefront`
- **AND** Hobbytech Toys MUST use `shopify-boost-storefront`
- **AND** Hobbyco and Mr Toys MUST remain `generic-storefront-crawler` until implementation probes promote them
- **AND** Bonza Slot Cars MUST expose its Woo Store API path as `woocommerce-store-api` while retaining manual/product-URL capture constraints elsewhere
- **AND** every provider entry MUST expose stable `market_watch_scope`, `auth_mode: none`, and source/matching/catalogue capabilities for search, pricing, stock observation, and health

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

### Requirement PROVIDER-AU-WEBSHOPS-BONZA-URL-001: Bonza product URL ingestion SHALL populate Cabinet item draft data
Bonza provider ingestion SHALL support direct product URL ingestion for `bonzaslotcars.com.au/product/<slug>/` and return a normalized Cabinet item draft.

#### Scenario: Bonza mug URL extracts structured product data
- **GIVEN** user submits `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **WHEN** Bonza product ingestion runs
- **THEN** normalized output MUST include title `BONZA MUG WHITE`
- **AND** output MUST include provider product id `19603` when returned by the Store API
- **AND** output MUST include source URL `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **AND** output MUST include AUD price `9.95`
- **AND** output MUST include stock count `3` and stock state `in_stock` when the provider reports `3 in stock`
- **AND** output MUST include the product description text
- **AND** output MUST include categories and attributes returned by the provider
- **AND** output MUST include at least one product image URL when provider images are available

#### Scenario: Bonza categories and attributes map to item metadata
- **GIVEN** Bonza Store API returns categories and attributes for a product
- **WHEN** product data is normalized for Cabinet
- **THEN** categories MUST map to Cabinet category draft values
- **AND** Brand, Scale, and Type attributes MUST map to item metadata or evidence fields
- **AND** Type MAY map to Cabinet Item Type when a matching configured item type exists

#### Scenario: Bonza evidence records extraction source
- **GIVEN** Bonza product ingestion succeeds
- **WHEN** normalized output is returned
- **THEN** evidence MUST include provider `bonzaslotcars`, family `woocommerce`, extraction method `store_api`, product id, original pasted URL, normalized source URL, and observed timestamp

### Requirement PROVIDER-AU-WEBSHOPS-BONZA-URL-002: Bonza product URL ingestion SHALL protect against duplicates
Bonza product URL ingestion SHALL check existing item source evidence before allowing a duplicate provider product to be created silently.

#### Scenario: Existing Bonza product source is detected
- **GIVEN** an inventory item already has source evidence for Bonza product id `19603` or the normalized Bonza mug source URL
- **WHEN** the same Bonza product URL is processed again
- **THEN** runtime MUST return duplicate candidate information
- **AND** Inventory UI MUST offer to open the existing item or continue only with explicit user confirmation

### Requirement INTEGRATION-015: Frontline provider config discovery SHALL be runtime-resolved from site assets with safe fallback
Frontline integration SHALL discover Algolia runtime config (application ID, search key, index names) from maintained site assets (e.g., `pd-search.js`) and use cached last-known-good config when discovery fails.

### Requirement INTEGRATION-016: Hobbytech provider SHALL support Shopify search via Boost/mybcapps endpoint with runtime config discovery
Hobbytech integration SHALL support Shopify-backed search using Boost Commerce endpoint (`services.mybcapps.com/bc-sf-filter/search`) with runtime discovery of required query/session parameters and resilient fallback behavior.

#### Scenario: Hobbytech search execution
- **GIVEN** provider query executes for Hobbytech
- **WHEN** runtime calls Boost/mybcapps search endpoint
- **THEN** response parsing MUST extract product candidates from structured fields and/or returned HTML payload blocks
- **AND** pagination MUST honor provider limit/page parameters until terminal page
- **AND** saved-search run output MUST persist normalized `source="hobbytechtoys"` candidates into the shared scanner/Discoveries candidate store and hydrate latest-run query-set snapshot metadata after reload

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
- **AND** saved-search run output MUST persist normalized `source="frontlinehobbies"` candidates into the shared scanner/Discoveries candidate store and hydrate latest-run query-set snapshot metadata after reload

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

### Requirement PROVIDER-AU-WEBSHOPS-LIGHTSPEED-001: Acer provider SHALL use a Lightspeed storefront adapter
Cabinet SHALL register Acer RC Models (`acercmodels.com`) as a `lightspeed-storefront` catalogue/source-matching provider instead of a generic webshop fallback.

#### Scenario: Acer Lightspeed registry metadata
- **GIVEN** `/api/providers/registry` returns AU webshop providers
- **WHEN** the Acer RC Models provider entry is projected
- **THEN** it MUST expose `adapter_type: lightspeed-storefront`
- **AND** it MUST expose `api_family: lightspeed`, `api_support_profile: lightspeed_storefront_v1`, `active_mode: lightspeed_catalog`, and `market_watch_scope: acercmodels`
- **AND** it MUST expose source/matching/catalogue capabilities for search, pricing, stock observation, and health while keeping auth mode `none`

#### Scenario: Lightspeed product fixture parser health
- **GIVEN** a Lightspeed storefront product/category fixture contains product id, title, canonical URL, price, SKU/part number, category, image, availability, and description fields
- **WHEN** parser normalization runs for Acer source matching
- **THEN** Cabinet MUST emit normalized candidates with listing id, title, source URL, AUD price, seller domain, source scope `acercmodels`, stock state/count, and image URL
- **AND** if a fixture has no parseable title, URL, id, or price for every product, parser health MUST fail rather than silently reporting an empty healthy result

### Requirement PROVIDER-AU-WEBSHOPS-SHOPIFY-001: Andrew's Hobbies and Metro Hobbies SHALL use public Shopify storefront parsing
Cabinet SHALL support Shopify public catalogue parsing for approved Shopify-backed AU hobby shop providers without requiring credentials or using private/admin Shopify APIs.

#### Scenario: Shopify product fixture parser health
- **GIVEN** a Shopify `/products.json` or collection-products fixture contains product id, title, handle, vendor, product type, variants, price, SKU, availability, image, tags, and description fields
- **WHEN** parser normalization runs for Andrew's Hobbies or Metro Hobbies source matching
- **THEN** Cabinet MUST emit normalized candidates with listing id, title, canonical source URL, AUD price, seller domain, source scope, SKU, brand, category, stock state/count, extraction method, and image URL
- **AND** if a fixture has no parseable title, handle, id, or price for every product, parser health MUST fail rather than silently reporting an empty healthy result
- **AND** the adapter MUST remain public-storefront only: no customer login, cart, checkout, payment, private API, or admin API behavior

#### Scenario: Shopify provider run persists candidate and registry proof
- **GIVEN** a saved Market Watch query is scoped to Andrew's Hobbies or Metro Hobbies
- **WHEN** Cabinet runs the Shopify provider API against public `/products.json` catalogue data
- **THEN** normalized candidates MUST persist to the shared scanner candidate store with the provider market-watch scope
- **AND** reloading query sets MUST show the latest-run succeeded status and persisted candidate count
- **AND** registry provider health and beta release status MUST reflect the successful public-storefront proof while keeping the no-login/no-cart/no-checkout/no-private/admin-API guardrails visible

### Requirement PROVIDER-AU-WEBSHOPS-GENERIC-STRUCTURED-001: Generic structured storefront providers SHALL parse public product metadata safely
Cabinet SHALL support a generic structured storefront parser for source-matching providers where public product pages expose structured product metadata but no safer platform-specific adapter has been confirmed.

#### Scenario: JSON-LD product fixture parser health
- **GIVEN** a public product page fixture contains schema.org Product JSON-LD with name, SKU, brand, category, image, canonical URL, offer price, offer currency, and availability
- **WHEN** Cabinet parses the product page through the generic structured storefront adapter
- **THEN** Cabinet MUST normalize the product into the shared source-matching candidate shape with listing ID, title, SKU, brand, category, price, currency, stock state, image, canonical URL, source, and seller
- **AND** the adapter MUST remain public storefront/source-matching only: no customer login, cart, checkout, payment, private API, or admin API behavior

#### Scenario: Missing core fields fail provider health
- **GIVEN** a structured product fixture omits core product fields such as title, canonical URL, or price
- **WHEN** Cabinet parses the fixture through the generic structured storefront adapter
- **THEN** the adapter MUST return no normalized scanner candidates
- **AND** it MUST surface a parser-health failure that can be routed to provider health/status output instead of silently accepting partial data

#### Scenario: Provider-shaped JSON-LD variants stay deterministic
- **GIVEN** Frontline Hobbies and Mr Toys product fixtures expose Product JSON-LD through different public page shapes such as `@graph` nodes and top-level JSON-LD arrays
- **WHEN** Cabinet parses those fixtures through the generic structured storefront adapter
- **THEN** each provider-shaped fixture MUST normalize to one shared source-matching candidate with provider-specific source, seller, listing ID, SKU, stock state, price, image, and canonical URL metadata
- **AND** unsupported public pages without Product JSON-LD MUST fail provider health clearly and remain manual-review/manual-URL-capture candidates

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
| UC-AU-08 | Generic structured product parse | Provider parses public Product JSON-LD into shared source-matching candidates and fails health on missing core fields or unsupported pages | implemented: `internal/app/shopping_provider_fixture_contract_test.go` `TestShoppingProviderFixturesNormalizeSharedCandidateShape`, `TestGenericStructuredStorefrontFixtureDetectsMissingCoreFields`, `TestGenericStructuredStorefrontProviderFixturesCoverProviderSpecificShapes`, `TestGenericStructuredStorefrontFixtureRejectsUnsupportedPageForManualReview` |
