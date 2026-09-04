# Cabinet Browser Companion privacy and permissions

The companion talks to Cabinet only through loopback addresses on this computer. Pairing creates a profile-bound credential stored in extension-local storage; supported Chromium runtimes restrict that storage to trusted extension contexts. Website content scripts cannot read it.

Provider access is optional site access. The manifest declares the ability to request HTTPS origins so new Cabinet modules do not require a new extension build, but the companion requests only the exact, non-wildcard origin published by an enabled Cabinet module. The collector can decline or remove each permission from the popup or browser settings.

The companion checks bounded selectors and returns matching selector identifiers, not page HTML. A provider module may passively synchronise supported item and image observations that the collector chooses to capture. It must never read or export cookies, passwords, session tokens or challenge answers; solve or bypass a site challenge; perform provider writes; or crawl hidden pages. A challenge is reported as **Action required** for the collector to handle in the normal browser tab.

Removing an integration's permission stops browser access for that module. Revoking its session in Cabinet stops all local API access. Removing the extension deletes its local state; revoke the old session in Cabinet before pairing a replacement.
