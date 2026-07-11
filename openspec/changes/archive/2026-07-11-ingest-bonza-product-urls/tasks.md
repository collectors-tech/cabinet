## 1. Tests First

- [x] 1.1 Add API test for pasted Bonza product URL detection and provider/family routing.
- [x] 1.2 Add API test with mocked Bonza Store API response proving product draft normalization for `bonza-mug-white`.
- [x] 1.3 Add API test for duplicate detection by provider product id and normalized source URL.
- [x] 1.4 Add Cypress test for Inventory paste action processing a Bonza URL and prefilling the create-item modal.
- [x] 1.5 Add Cypress test for unsupported pasted URL feedback preserving user input.

## 2. Backend Provider Routing

- [x] 2.1 Add URL normalization helper for host, path, slug, and canonical source URL.
- [x] 2.2 Add provider registry/domain matching for `bonzaslotcars.com.au` and `www.bonzaslotcars.com.au`.
- [x] 2.3 Add product-page route classification for `/product/<slug>/`.
- [x] 2.4 Add clear unsupported-provider and unsupported-page response envelopes.

## 3. Bonza WooCommerce Ingestion

- [x] 3.1 Add Bonza product ingest endpoint or generic provider ingest endpoint that dispatches to Bonza.
- [x] 3.2 Implement Store API-first lookup using slug-derived search and exact slug/permalink matching.
- [x] 3.3 Normalize title, source URL, provider product id, price/currency, stock state/count, description, categories, attributes, and image URLs.
- [x] 3.4 Reconciled fallback scope: delivered Store API-first ingestion with bounded Bonza challenge retry coverage; page HTML fallback remains limited to broader watched-car/detail-page enrichment already covered by canonical provider contracts.
- [x] 3.5 Add provenance/evidence payload with provider id, family, extraction method, observed timestamp, original URL, normalized URL, and source summary.

## 4. Inventory Create Flow

- [x] 4.1 Wire Inventory create paste process to call provider URL ingestion for URLs.
- [x] 4.2 Prefill create-item modal fields from normalized Bonza item draft while preserving confirm-before-create behavior.
- [x] 4.3 Stage provider image URLs as item evidence/media candidates without silently downloading unless create is confirmed.
- [x] 4.4 Persist source URL and provider evidence when user confirms item creation.
- [x] 4.5 Show duplicate warning with open-existing action and explicit continue option. Backend duplicate candidates are now returned for UI consumption.

## 5. Verification and Delivery

- [x] 5.1 Run OpenSpec validation for `ingest-bonza-product-urls`.
- [x] 5.2 Run targeted Go/API tests for provider routing and Bonza ingestion.
- [x] 5.3 Run targeted Cypress Inventory paste ingestion tests.
- [x] 5.4 Live/manual verify the example Bonza URL against demo after rebuild (#811 final verification comment, 2026-05-30).
- [x] 5.5 Commit with `#811`, push, PR to `develop`, merge after validation, close issue, delete branch, and restart demo (#811 closed after PRs #1016/#1017/#1018; #1077 QA follow-up also closed).
