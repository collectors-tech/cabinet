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

## Browser Companion pairing and recovery

The optional Browser Companion connects only to Cabinet on your own computer. Start pairing in the extension, then open Integrations in an unlocked Cabinet window. Compare the extension name and six-digit pairing code before selecting **Approve**. Reject any device, origin or code you do not recognise. Approval is required before the extension receives a usable session.

Cabinet lists paired extensions under **Browser Companion access** without displaying their credentials. Use **Revoke** for a lost, replaced or suspicious extension, or **Revoke all** if the computer or browser profile may be compromised. Reinstalling the extension or clearing browser storage removes its credential; revoke the old session in Cabinet and pair again. Credential rotation revokes the previous value automatically.

Chrome and Edge development builds, unpacked builds and store releases can have different extension origins. Cabinet treats each origin as a separate device boundary, so pair them separately. Never put a Browser Companion credential in a URL, log, screenshot, support request or chat.

### Security and privacy boundary

Cabinet accepts companion traffic only over loopback and binds a random credential to one Cabinet instance, active profile, extension origin, device identity, protocol version, expiry and capability set. Pairing requests expire, exchanges work once, credentials are stored as verifiers in Cabinet, and origin, Host, remote-address, rate, concurrency and body-size checks fail closed. Module discovery includes only enabled integration instances for the paired profile and removes secret, token, password, cookie and API-key configuration.

The companion captures only supported page data that the collector is legitimately viewing. It must not export browser cookies, solve access challenges, crawl invisibly or perform provider writes. Cabinet validates the versioned envelope, commits it to the active profile's durable capture inbox and then sends typed observations to Market Watch, Discoveries or Purchase Inbox review. It does not silently create inventory.

### Install and use the Chrome or Edge companion

Cabinet uses one modular Manifest V3 companion for Chrome and Edge. A paired extension loads the current profile's enabled browser integrations from Cabinet; installing a separate extension per provider is not required. During development, open `chrome://extensions` or `edge://extensions`, enable developer mode, select **Load unpacked**, and choose the repository's `browser-extension` directory. Store-signed packages and checksums are handled by the extension release process.

The popup always shows a Cabinet row and then zero, one or many enabled browser modules. Choose **Open Cabinet** to focus the local app. For a provider, grant its exact optional site permission, open its page and choose **Check session**. An open tab alone is not treated as proof of login. The companion reports **Site access required**, **Browser required**, **Signed out**, **Action required**, **Ready to sync**, or **Page not supported** from bounded module evidence. A challenge always requires normal user action in the provider tab.

Enabled integrations, display names, start pages, exact origins, URL patterns, capture schemas, workflows, redaction rules, fixture version, item/media policy, review destination, cadence, help path and readiness evidence come from Cabinet's versioned module registry. Chrome and Edge therefore use the same host code, and later provider modules can be added without provider branches in the popup or background worker. The extension persists idempotent pending jobs across service-worker, browser and Cabinet restarts, uses bounded retry backoff, and keeps its pending/error state visible. A job leaves the extension queue only after Cabinet confirms durable terminal processing. Partial captures stay visible for review and never imply that previously seen provider items disappeared.

Image submissions accept only verified JPEG or PNG bytes tied to a durable capture field. Cabinet checks the SHA-256 digest and decoded dimensions, rejects login/challenge HTML masquerading as an image, and writes accepted files through canonical media storage. Identical content is stored once per profile but can retain several provenance links. The extension and API never return Cabinet's local media path.

This boundary cannot protect a credential from malware or another person who already controls the unlocked operating-system account, Cabinet process or browser profile. If local compromise is suspected, close Cabinet, secure the operating-system account, revoke all companion sessions after recovery, remove unknown extensions, reinstall the trusted extension and pair again. Restore Cabinet from a known-good backup if its local database may have been altered.

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
