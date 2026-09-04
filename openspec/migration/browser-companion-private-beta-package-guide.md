# Browser Companion private-beta package guide

The Cabinet Browser Companion private beta is distributed as separate verified
Chrome and Edge ZIP files. A ZIP is **not an installer**, is not available in the
Chrome Web Store or Microsoft Edge Add-ons, and has no automatic updates. Use
only the exact package named by the Cabinet release-candidate manifest.

## Verify the candidate

Record the Cabinet candidate commit and these Browser Companion files before
installation:

- `browser-companion-release-manifest.json`;
- the target-specific Chrome or Edge ZIP;
- the matching `.zip.sha256` file;
- the versioned release notes; and
- `browser-companion-candidate-summary.md`.

The release manifest binds the extension version, immutable candidate tag,
source commit, protocol compatibility range, package file list and SHA-256 for
both target ZIPs. Each candidate version includes a `g<source-commit-prefix>`
suffix. Never substitute a file called `latest`, reuse an extension version for
different bytes, or accept a checksum from a different candidate.

On Windows, verify the downloaded ZIP and compare the result with its
`.sha256` file:

```powershell
Get-FileHash -Algorithm SHA256 .\cabinet-browser-companion-<version>-chrome.zip
Get-Content .\cabinet-browser-companion-<version>-chrome.zip.sha256
```

Stop if the values, filename, source commit, target, version or protocol range
do not match the release manifest. The verifier command used by CI is:

```text
node scripts/verify-browser-companion-package.mjs --manifest <path> --expected-source-commit <40-character-commit>
```

## Install Chrome or Edge

1. Choose the package for the browser being tested. Chrome and Edge packages
   are separate even when their file contents are otherwise equivalent.
2. Extract the verified ZIP into a stable, versioned directory. Do not select
   the ZIP itself and do not select the repository's development source.
3. Open `chrome://extensions` or `edge://extensions`.
4. Enable developer mode, choose **Load unpacked**, and select the extracted
   directory containing the production `manifest.json`.
5. Confirm the extension is named **Cabinet Browser Companion** and that its
   displayed version matches the release manifest. A source checkout is named
   **Cabinet Browser Companion (Development)** and is not packaged evidence.
6. Start the matching Cabinet candidate, open the extension and pair it through
   Cabinet's Integrations screen. Compare the displayed code before approval.

The package requires only Cabinet loopback access. Its optional HTTPS host
permission is a browser-declared superset; Cabinet supplies enabled modules and
the extension asks the collector for the **exact provider origin** at runtime.
Reject an unexpected permission request. The extension must not export cookies
or tokens, solve a challenge, crawl invisibly or perform provider writes.

For #1869, install and pair both target packages from clean browser profiles.
Record the browser/version, package filename, SHA-256, extension version,
source commit, release-manifest path and Cabinet protocol version.

## Upgrade and rollback

There are no automatic updates in the private-beta channel. Before a manual
upgrade, pause sync, allow or cancel visible pending jobs, verify the new ZIP,
and keep the previous verified package until acceptance succeeds. Extract the
new version into a different versioned directory and use **Load unpacked** or
the browser's reload control as documented in the acceptance record.

Keeping the same unpacked directory normally preserves the extension origin and
storage. Loading a different path can create a different extension origin; pair
the new origin, then revoke the old session in Cabinet. Do not copy credentials
between browser profiles or directories.

A rollback is allowed only to a previously verified package whose SHA-256 and
release manifest are retained and whose protocol compatibility range includes
the running Cabinet protocol. Protocol mismatch, unknown version, missing
manifest or failed checksum must fail closed. After rollback, repeat pairing,
reconnect and pending-job recovery checks; do not claim the failed candidate as
accepted.

## Troubleshoot and recover

- **Checksum or manifest mismatch:** delete the untrusted copy and recover the
  exact candidate from the controlled artefact store. Do not install it.
- **Protocol mismatch:** use a Cabinet and companion pair whose manifest ranges
  overlap. Do not bypass the compatibility check.
- **Pairing expired or rejected:** start a new request in an unlocked Cabinet
  window and compare the new code.
- **Site access required:** grant only the module's exact provider origin.
- **Action required:** complete the provider's normal browser interaction
  yourself, then check readiness again. The companion never solves it.
- **Extension disconnected:** confirm Cabinet is running on loopback, then
  reconnect or revoke the stale session and pair again.
- **Selector drift:** stop sync and attach the non-secret readiness/error
  evidence to the provider issue. Do not broaden capture or permissions.

To uninstall, first pause sync and resolve visible pending jobs. Use Cabinet to
**revoke** the companion session, remove the extension from the browser, and
only then delete the extracted package directory. Revoke all sessions if the
browser profile or computer may be compromised.

## Release boundary

Package generation produces a private candidate only. #1869 must prove clean
Chrome and Edge installation, pairing, provider sync, recovery and item/image
persistence from the exact files. #1864 approval is required before creating an
external release, publishing an immutable tag or advertising wider
availability. No package in this guide is claimed for Firefox, Safari or Brave.
