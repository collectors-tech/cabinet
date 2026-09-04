## Context

Cabinet already has an AU webshop provider family and a Bonza provider path for query-based Market Watch ingestion. The missing path is user-directed product ingestion from a pasted URL inside Inventory. The example URL, `https://bonzaslotcars.com.au/product/bonza-mug-white/`, is a WooCommerce product page and the public Store API returns the product detail needed to prefill a Cabinet item.

Current Inventory paste handling can keep raw pasted text and prefill simple local drafts, but it does not yet route known provider URLs to provider-specific extractors. Bonza detection can be deterministic because the provider registry already includes `bonzaslotcars.com.au`.

## Goals / Non-Goals

**Goals:**
- Detect known provider product URLs without AI or guessing.
- Route Bonza product URLs through the Bonza/WooCommerce provider path.
- Use WooCommerce Store API first and page metadata/HTML only as fallback.
- Return a normalized Cabinet item draft suitable for the Inventory create modal.
- Preserve source provenance for later evidence/history review.
- Prevent accidental duplicate item creation for the same provider product.
- Add deterministic API and Cypress coverage.

**Non-Goals:**
- Build a generic scraper for all unknown websites in this slice.
- Automatically create items without user confirmation.
- Add privileged WooCommerce credentials or admin API usage.
- Download and persist remote images in the first request unless the user confirms item creation.
- Replace existing Market Watch query-set ingestion.

## Decisions

### 1. URL router before provider ingestion
Cabinet will parse and normalize pasted URLs before attempting ingestion:

```text
raw URL -> normalized host/path -> provider registry match -> provider handler
```

For Bonza:

```text
host: bonzaslotcars.com.au or www.bonzaslotcars.com.au
path: /product/<slug>/
provider: bonzaslotcars
family: woocommerce
action: ingest_product_url
```

Alternative considered: send pasted URLs to AI or broad webpage scraping first. Rejected because known provider URLs are deterministic, faster, more testable, and safer.

### 2. WooCommerce Store API is source of truth
The Bonza handler will resolve product detail through the public Store API. For a product slug, it will query the Store API using a slug-derived search term, then select the product where `slug` or normalized `permalink` matches the pasted URL.

Fallback parsing may read product page metadata such as Open Graph title/image, price HTML, and stock text only when Store API fields are missing.

Alternative considered: scrape the product page directly. Rejected because the Store API returns structured title, price, currency, description, categories, attributes, images, and stock fields.

### 3. Normalize into item draft, not immediate item mutation
The backend response will return a normalized draft object. The Inventory modal will prefill the form and require the user to press Create/Save before mutation.

Expected normalized fields include:
- `provider_id`
- `provider_family`
- `provider_product_id`
- `source_url`
- `title`
- `description`
- `categories`
- `item_type`
- `attributes`
- `price`
- `currency`
- `stock_state`
- `stock_count`
- `image_urls`
- `evidence`

Alternative considered: directly create the item from paste. Rejected because Cabinet workflows should remain confirm-before-apply when external data writes into user inventory.

### 4. Provenance is first-class
The created item should keep enough source evidence to explain where fields came from. At minimum this includes original pasted URL, normalized source URL, provider id, provider product id, ingestion timestamp, extraction method, and a compact raw payload/evidence summary.

### 5. Duplicate check uses provider identity and URL
Before create, Cabinet should check whether an existing item has the same provider/source identity. A match should return an actionable duplicate warning/open-existing option instead of silently creating another item.

## Risks / Trade-offs

- [Bonza Store API search returns multiple products] -> Match exact slug/permalink before accepting a product.
- [Provider markup or API changes] -> Keep API-first tests mocked and live verification documented; fallback only covers safe metadata fields.
- [Remote image persistence is slow or unsafe] -> Stage image URLs first and only persist media during confirmed create or later media import.
- [Duplicate detection misses older items without provenance] -> Also compare normalized source URL where provider product id is absent.
- [Unknown URLs create user confusion] -> Return explicit unsupported-provider or unsupported-page messages.

## Migration Plan

1. Add spec deltas and tests for URL detection, Bonza product ingestion, and Inventory paste prefill.
2. Add backend URL router and Bonza ingest endpoint.
3. Wire Inventory create modal paste processing to the endpoint.
4. Preserve source/evidence fields in item create payloads.
5. Validate with mocked API tests, Cypress paste flow, OpenSpec validation, and live manual verification against the example Bonza URL.
6. Roll back by disabling the new paste-ingest call path; existing manual create and Market Watch ingestion remain unaffected.
