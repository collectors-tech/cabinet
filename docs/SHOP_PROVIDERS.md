# Shop Provider Catalog

Purpose:
- Track target store providers for scanner/pricing integrations.
- Define integration mode per provider (official API vs. web ingestion).

## Provider List (Requested)

| Provider | URL | Region | Integration Mode | Notes |
|---|---|---|---|---|
| Bonza Slot Cars | https://bonzaslotcars.com.au/ | AU | Web ingestion candidate | Product pages appear to expose availability text/stock hints. |
| Frontline Hobbies | https://www.frontlinehobbies.com.au/ | AU | Web ingestion candidate | Validate stock signal consistency by category/template. |
| Hobby Tech Toys | https://hobbytechtoys.com.au/ | AU | Web ingestion candidate | Confirm robots/terms and anti-bot restrictions. |
| Andrews Hobbies | https://andrewshobbies.com.au/ | AU | Web ingestion candidate | Confirm listing schema and stock signal location. |
| Voglers | https://voglers.com.au/ | AU | Web ingestion candidate | Confirm product page stock and price parse stability. |
| ACERC Models | https://www.acercmodels.com/ | AU | Web ingestion candidate | Validate pagination/search route behavior. |
| Mr Toys | https://www.mrtoys.com.au/ | AU | Web ingestion candidate | Confirm category coverage for slot car inventory segments. |

## API Feasibility (Asked: eBay and Amazon)

### eBay
- Official API available: Yes.
- Recommended path for Cabinet scanner/pricing:
  - eBay Buy APIs (Browse API) for listing discovery/search.
  - OAuth app credentials required.
- Notes:
  - Rate limits and policy compliance required.
  - Keep provider adapter isolated in `internal/scanner/providers/ebay`-style module boundaries.

Official docs:
- https://developer.ebay.com/api-docs/buy/browse/overview.html
- https://developer.ebay.com/develop/apis

### Amazon
- Official API available: Yes, but split by use case.
- Product Advertising API (PA-API):
  - For Amazon Associates use cases.
  - Access tied to Associates account requirements.
- Selling Partner API (SP-API):
  - For sellers managing their own catalog/orders/inventory.
  - Not a general public retail listing API for third-party buyer tooling.
- Practical implication for Cabinet:
  - Amazon integration is possible, but constrained by program eligibility and policy scope.
  - Treat as future provider with explicit legal/compliance review before implementation.

Official docs:
- https://webservices.amazon.com/paapi5/documentation/
- https://developer-docs.amazon.com/sp-api/

## Integration Policy

1. Prefer official APIs where available and policy-compatible.
2. For web ingestion providers, require:
- robots/terms review
- deterministic parser tests
- retry/backoff and provider-health monitoring
3. Normalize all provider outputs to common candidate/pricing/stock fields.

