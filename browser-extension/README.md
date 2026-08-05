# Cabinet Browser Companion

This directory is the source for one Manifest V3 extension that runs unpacked in Chrome and Edge. It loads enabled, profile-scoped browser modules from local Cabinet after explicit pairing. Provider definitions, exact origins and page patterns, capture schemas, workflows, redaction rules, fixture version, supported passive actions, review destination, cadence and bounded readiness selectors are data; the background host and popup contain no provider-specific branches.

## Development install

1. Start Cabinet on its normal loopback address and unlock the intended profile.
2. Open `chrome://extensions` or `edge://extensions`, enable developer mode and choose **Load unpacked**.
3. Select this `browser-extension` directory.
4. Open the companion, choose **Connect**, then approve the displayed request and pairing code in Cabinet's Integrations screen.
5. Finish pairing in the companion. Enabled browser-capable integrations then appear automatically.

Unpacked Chrome and Edge installs can receive different extension origins. Pair each browser separately. Store packaging, signing, checksums and rollback are owned by #2034.

## Truthful states

- **Site access required** — exact provider-origin permission has not been granted.
- **Browser required** — permission exists but no matching page is open.
- **Signed out** — a bounded module readiness marker shows the user is logged out.
- **Action required** — the normal page requires user action, including a challenge.
- **Ready to sync** — a supported signed-in readiness marker is visible.
- **Page not supported** — a matching page is open but no declared state can be proven.

The extension does not infer login from an open tab alone. It retains bounded idempotent jobs through service-worker and browser restarts, applies retry backoff, and exposes pending/error state through the popup and badge.

The popup shows setup, sync, pause/resume and review actions only when the Cabinet module contract proves they are available. Cabinet advertises `sync_available: true` only when a module has a packaged capture script and the durable Cabinet capture pipeline is available. The core loads that separately packaged script, accepts only declared bounded typed fields, strips query/fragment data from source URLs, rejects raw page/session fields, and keeps the passive envelope queued until Cabinet returns a committed terminal acknowledgement. Cabinet-side pending, failed and review counts remain visible and open the module's configured review surface.

## Frontline Hobbies module

The versioned `frontlinehobbies-search-capture` module is projected only for an enabled `au-webshop-frontlinehobbies-com-au` integration instance. It requests exact Frontline storefront/CDN origins, recognises rendered public product cards, and submits search batches under the `frontlinehobbies` Market Watch scope. Ready, partial, signed-out, challenge and selector-drift fixtures are deterministic; challenges and unknown page shapes fail closed. It is manual and user-present, capped at six attempts per minute, and has no cookie, token, network-fetch, click, cart, checkout or other provider-write mechanism.

Fixture and integration tests are source evidence, not live acceptance. A normal user-present Frontline search must still be attached to #1944, and the same journey must be repeated from the exact package in #1869.

## Bonza Slot Cars module

The versioned `bonzaslotcars-search-capture` module is projected only for an enabled `au-webshop-bonzaslotcars-com-au` integration instance. It requests the two exact Bonza storefront origins, recognises rendered WooCommerce product cards, and submits search batches under the `bonzaslotcars` Market Watch scope. Ready, partial, signed-out, Sucuri-challenge and selector-drift fixtures are deterministic; challenges and unknown page shapes fail closed. It is manual and user-present, capped at six attempts per minute, and has no cookie, token, challenge-decoding, network-fetch, click, cart, checkout or other provider-write mechanism.

The Sucuri readiness signal is a fixed, bounded script-marker identifier. The bridge returns only that identifier, never script contents, raw page data or challenge material. The collector completes any normal site interaction in the Bonza tab before checking readiness and choosing sync.

Fixture and integration tests are source evidence, not live acceptance. A normal user-present Bonza search must still be attached to #1945, and the same journey must be repeated from the exact package in #1869.
