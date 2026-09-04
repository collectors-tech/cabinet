# Market Watch Beta Provider Proof

## Selected Beta Provider

Cabinet 0.1 beta uses Bonza Slot Cars as a user-present Browser Companion Market Watch proof path. Voglers remains a direct public BigCommerce provider proof path.

Provider:
- Registry ID: `au-webshop-bonzaslotcars-com-au`
- Market Watch scope: `bonzaslotcars`
- Public base URL: `https://www.bonzaslotcars.com.au`
- Access method: paired `bonzaslotcars-search-capture` module over exact Bonza origins after normal user interaction
- Authentication: no credentials enter Cabinet; any site session remains inside the user's browser

Fallback provider:
- Registry ID: `au-webshop-voglers-com-au`
- Market Watch scope: `voglers`
- Public base URL: `https://www.voglers.com.au`
- Access method: public BigCommerce storefront search page, `GET /search.php?search_query=<query>`
- Authentication: none for public product search

## Release Boundaries

Bonza can be marked `available_live_validated` only after a complete user-present Browser Companion search records `browser_companion` provider health and persists normalized result candidates for the active profile. Direct Store API fixtures or runs cannot satisfy this gate. Until external proof exists it remains `browser_companion_live_evidence_required`, because the public Store API is affected by a Sucuri challenge outside Cabinet. Direct product extraction is best-effort and must return `browser_action_required` after one challenged request without decoding scripts or sending cookies. Under #2054 / `PROVIDER-FAMILY-011`, stalled or partial provider responses must also fail within the shared bounded timeout, persist no false success, and leave a later provider run usable.

Voglers can be marked `available_live_validated` only after a successful public BigCommerce storefront run records provider health and persists normalized result candidates for the active profile. It must parse public search result product cards only, because the observed `/products/search` JSON-style path can return 404 on the live storefront while the public search HTML remains available.

eBay remains setup-required until approved credentials and live capability evidence exist. Unsupported providers must keep disabled actions or beta-limited status so beta users are not shown unproven paths as connected production integrations.

## Terms, Rate Limits, and Cache Policy

The Bonza and Voglers proof paths read public storefront catalogue data and do not perform checkout, account, cart, seller, or write operations. Bonza is manual and user-present, capped at six sync attempts per minute, and cannot run unattended. It must never export cookies/session data, decode or solve a challenge, click, or make provider writes. Voglers scheduling remains opt-in and should pause if provider health becomes failed, rate-limited or unavailable.

Live evidence artifacts must be non-secret. Acceptable evidence includes:
- exact source commit, module/fixture version and granted exact origins
- a redacted screenshot or recording showing the user-opened Bonza result page and Ready-to-sync state
- normalized result count and provider provenance
- the persisted Market Watch run ID, query-set ID, canonical provider scope, `trigger_type=browser_companion`, `status=succeeded` and positive result count
- one or more redacted result samples with title, URL host/path, price/currency where available, stock state where available, and observed time
- Cabinet capture/provider-health evidence and registry status after the run

Do not attach cookies, credentials, customer data, challenge material, script contents or full raw storefront response dumps to GitHub issues.

## Validation Target

The release proof target is:
1. create a Bonza-scoped saved Market Watch query
2. enable `au-webshop-bonzaslotcars-com-au`, pair the companion and grant only the exact Bonza origins
3. open a real Bonza search as the user, complete any normal site interaction and sync a complete rendered result page
4. persist candidates with Bonza integration/module/schema and `bonzaslotcars` scope provenance
5. show the persisted Market Watch run as provider `bonzaslotcars`, trigger `browser_companion`, status `succeeded` and a positive result count
6. record `browser_companion` provider health for `bonzaslotcars`
7. show `/api/providers/registry` as `available_live_validated` with `live_evidence_state=validated`
8. attach non-secret external proof to #1945 and #1929; repeat the exact packaged journey in #1869

The fallback Voglers release proof target is:
1. create a Voglers-scoped saved Market Watch query
2. run it through `/api/providers/bigcommerce/run` using the public `search.php` storefront path
3. persist candidates with Voglers provenance
4. record provider health for `voglers.com.au` and `au-webshop-voglers-com-au`
5. show `/api/providers/registry` as `available_live_validated` with `live_evidence_state=validated`
6. attach non-secret probe evidence to #1871 and #1864
