# Market Watch Beta Provider Proof

## Selected Beta Provider

Cabinet 0.1 beta uses Bonza Slot Cars as the first no-secret Market Watch provider proof path.

Provider:
- Registry ID: `au-webshop-bonzaslotcars-com-au`
- Market Watch scope: `bonzaslotcars`
- Public base URL: `https://bonzaslotcars.com.au`
- Access method: WooCommerce Store API, `GET /wp-json/wc/store/v1/products`
- Authentication: none for public product search

## Release Boundaries

Bonza can be marked `available_live_validated` only after a successful Market Watch run records provider health and persists normalized result candidates for the active profile. Until then, it remains `manual_url_capture_only` because the WooCommerce Store API path can be affected by storefront challenge pages, product availability changes, or rate controls outside Cabinet.

eBay remains setup-required until approved credentials and live capability evidence exist. Unsupported providers must keep disabled actions or beta-limited status so beta users are not shown unproven paths as connected production integrations.

## Terms, Rate Limits, and Cache Policy

The Bonza proof path reads public storefront catalogue data and does not perform checkout, account, cart, seller, or write operations. The beta run path should stay manual by default and use conservative page sizes. Automated scheduling must remain opt-in and should pause if the provider health status becomes failed, rate-limited, unavailable, or challenge-gated.

Live evidence artifacts must be non-secret. Acceptable evidence includes:
- HTTP status and response metadata for the Store API request
- normalized result count and provider provenance
- one or more redacted result samples with title, URL host/path, price/currency where available, stock state where available, and observed time
- Cabinet run log paths and registry status after the run

Do not attach cookies, credentials, customer data, or full raw storefront response dumps to GitHub issues.

## Validation Target

The release proof target is:
1. create a Bonza-scoped saved Market Watch query
2. run it through `/api/providers/bonza/run`
3. persist candidates with Bonza provenance
4. record provider health for `bonzaslotcars` and `au-webshop-bonzaslotcars-com-au`
5. show `/api/providers/registry` as `available_live_validated` with `live_evidence_state=validated`
6. attach non-secret probe evidence to #1871 and #1864
