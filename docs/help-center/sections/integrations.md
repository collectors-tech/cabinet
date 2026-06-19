# Integrations

## Use Integrations for
- Connecting providers
- Validating tokens
- Running sync and status checks
- Reviewing Market Watch saved searches

## Common actions
- Connect/disconnect provider
- Validate token
- Run provider sync

## eBay setup
Use the eBay integration setup when you want Cabinet to run authenticated eBay listing searches from Market Watch.

Before saving the provider, prepare the eBay bearer token that Cabinet will use for Browse API requests. Cabinet only displays token presence after save; it does not show the bearer token back in the setup panel.

Set the marketplace to the eBay region you expect saved searches to query. The default production marketplace is usually enough for normal runs, but the setup panel also accepts a base URL override for controlled environments. Treat the base URL override as an advanced routing setting and keep it blank unless you are deliberately pointing Cabinet at a non-default eBay Browse endpoint.

After saving credentials and marketplace settings, use Validate in the integration dialog. The setup status panel shows token state, marketplace, validation status, provider health, and the next action. If the provider is ready, run eBay query sets from Market Watch; the setup dialog validates configuration but does not execute saved searches.

Market Watch eBay query sets keep keywords, exclusions, max price, schedule, rate-limit settings, provider scope, and page size together. Create or edit an eBay-scoped saved search, then use Run now or scheduled refreshes to collect candidates. Output details preserve source URL, price, shipping, stock, seller, query, and provider provenance before handoff to Discoveries, Wishlist, or Inventory.

If a run fails with `PROVIDER_AUTH_MISSING` or `PROVIDER_AUTH_INVALID`, review the saved bearer token and provider health before retrying. If a run fails with `PROVIDER_SEARCH_FAILED`, check provider health, upstream eBay guidance, and any retry timing before running the query again. Live credential and marketplace capability evidence is still required before treating a real eBay account as production-ready.

## Market Watch saved searches
Use Market Watch when you want Cabinet to watch provider listings without re-entering the same search each time.

Saved searches keep the provider scope, keywords, filters, schedule, and rate-limit settings together. You can create, edit, delete, run now, or run scheduled refreshes from the Market Watch screen.

## Reviewing run output
Run results show provider attribution, latest run status, last run time, and candidate counts. Table view is useful when you have several saved searches and need to compare recent output quickly.

Open output details before handing results off. The detail view keeps enough source context for follow-up actions to remain traceable.

## Failures and retries
Provider failures stay visible instead of silently dropping candidates. Use the retry action after checking the guidance shown on the failed run. When the retry recovers, Market Watch reloads the saved-search state and clears stale failure entries.

## Discoveries and Wishlist handoff
Send useful saved-search output to Discoveries for review, or add a candidate to Wishlist when it is worth tracking. Cabinet preserves saved-search provenance such as provider, query set, query name, and provider scope so the Wishlist entry can still be understood after reload.
