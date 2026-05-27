## 1. Tests First

- [ ] 1.1 Add API test for pasted Bonza product URL detection and provider/family routing.
- [ ] 1.2 Add API test with mocked Bonza Store API response proving product draft normalization for `bonza-mug-white`.
- [ ] 1.3 Add API test for duplicate detection by provider product id and normalized source URL.
- [ ] 1.4 Add Cypress test for Inventory paste action processing a Bonza URL and prefilling the create-item modal.
- [ ] 1.5 Add Cypress test for unsupported pasted URL feedback preserving user input.

## 2. Backend Provider Routing

- [ ] 2.1 Add URL normalization helper for host, path, slug, and canonical source URL.
- [ ] 2.2 Add provider registry/domain matching for `bonzaslotcars.com.au` and `www.bonzaslotcars.com.au`.
- [ ] 2.3 Add product-page route classification for `/product/<slug>/`.
- [ ] 2.4 Add clear unsupported-provider and unsupported-page response envelopes.

## 3. Bonza WooCommerce Ingestion

- [ ] 3.1 Add Bonza product ingest endpoint or generic provider ingest endpoint that dispatches to Bonza.
- [ ] 3.2 Implement Store API-first lookup using slug-derived search and exact slug/permalink matching.
- [ ] 3.3 Normalize title, source URL, provider product id, price/currency, stock state/count, description, categories, attributes, and image URLs.
- [ ] 3.4 Add limited product page metadata/HTML fallback for missing Store API fields.
- [ ] 3.5 Add provenance/evidence payload with provider id, family, extraction method, observed timestamp, original URL, normalized URL, and source summary.

## 4. Inventory Create Flow

- [ ] 4.1 Wire Inventory create paste process to call provider URL ingestion for URLs.
- [ ] 4.2 Prefill create-item modal fields from normalized Bonza item draft while preserving confirm-before-create behavior.
- [ ] 4.3 Stage provider image URLs as item evidence/media candidates without silently downloading unless create is confirmed.
- [ ] 4.4 Persist source URL and provider evidence when user confirms item creation.
- [ ] 4.5 Show duplicate warning with open-existing action and explicit continue option.

## 5. Verification and Delivery

- [ ] 5.1 Run OpenSpec validation for `ingest-bonza-product-urls`.
- [ ] 5.2 Run targeted Go/API tests for provider routing and Bonza ingestion.
- [ ] 5.3 Run targeted Cypress Inventory paste ingestion tests.
- [ ] 5.4 Live/manual verify the example Bonza URL against demo after rebuild.
- [ ] 5.5 Commit with `#811`, push, PR to `develop`, merge after validation, close issue, delete branch, and restart demo.
