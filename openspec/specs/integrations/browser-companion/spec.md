## Purpose
Define the Cabinet Browser Companion trust boundary, versioned host/module contract, passive capture envelopes, local sync and user-controlled recovery.

## Requirements

### Requirement INTEGRATION-047: Browser Companion modules MUST submit only authenticated passive capture payloads
Cabinet SHALL expose an authenticated module registry and payload sync endpoint so paired companion modules can capture marketplace evidence without hidden write actions.

#### Scenario: Passive module registry requires a paired session
- **GIVEN** the Browser Companion asks Cabinet for enabled modules
- **WHEN** `GET /api/companion/modules` is requested without a valid paired session credential
- **THEN** Cabinet MUST reject the request as unauthorised
- **AND** the legacy predictable `companion:<profile_id>` bearer value MUST NOT authenticate.

#### Scenario: Companion payload sync rejects hidden write attempts
- **GIVEN** a paired Browser Companion module submits a captured marketplace payload
- **WHEN** the payload is not marked passive or reports an attempted write
- **THEN** Cabinet MUST reject the payload with a deterministic passive-capture error
- **AND** Cabinet MUST NOT create inventory, wishlist, purchase or provider-write records from that payload.

#### Scenario: Companion payload sync accepts validated passive evidence
- **GIVEN** an enabled module submits a passive payload with a valid capture URL, payload type, confidence score and session-bound profile
- **WHEN** Cabinet accepts the payload transport envelope
- **THEN** the response MUST report `sync_mode=passive_capture` and `remote_write=false`
- **AND** the response MUST include non-secret audit evidence naming the module, session, protocol version and passive sync mode.

### Requirement INTEGRATION-067: Pairing MUST require explicit approval inside Cabinet's configured authentication boundary
One extension installation SHALL establish trust through a short-lived request, visible six-digit code, explicit profile-scoped approval and one-time exchange. Credential-free local-device mode SHALL permit same-origin loopback management for an active profile only while no passkey is registered; LAN, ZITADEL and registered-passkey modes SHALL require their configured authenticated or unlocked session.

#### Scenario: Pairing management matches the configured Cabinet boundary
- **GIVEN** a same-origin loopback Cabinet UI with an active profile
- **WHEN** local-device mode is explicitly credential-free and the profile has no registered passkey
- **THEN** the collector MUST be able to list and approve Browser Companion pairing requests without a simulated server credential
- **AND** LAN mode, ZITADEL mode, and a registered-but-locked local profile MUST reject the same management request without a valid session.

#### Scenario: Exchange before approval fails closed
- **GIVEN** an extension-origin-bound pairing request created through `POST /api/companion/pairing/requests`
- **WHEN** the extension attempts `POST /api/companion/pairing/exchanges` before the collector approves the request in Cabinet
- **THEN** Cabinet MUST reject the exchange
- **AND** no session credential MUST be created.

#### Scenario: Approved challenge is exchanged only once
- **GIVEN** the collector recognises the device, origin and pairing code and approves the requested capability subset
- **WHEN** the extension exchanges the matching short-lived secret, device identity and protocol version
- **THEN** Cabinet MUST return at least 256 bits of random credential material once with `Cache-Control: no-store`
- **AND** Cabinet MUST persist only a credential verifier
- **AND** every subsequent exchange attempt for that request MUST fail as replayed.

### Requirement INTEGRATION-068: Companion sessions MUST be isolated and revocable
Every session SHALL be bound to one Cabinet instance, profile, extension origin, device identity, protocol version, expiry and capability set.

#### Scenario: Boundaries reject impersonation
- **GIVEN** a valid Browser Companion credential
- **WHEN** it is replayed from another extension origin or device, submitted for another profile, used for an ungranted capability, sent through a non-loopback Host/remote address, or submitted after expiry/revocation
- **THEN** Cabinet MUST reject the request with a deterministic authentication, binding or capability error.

#### Scenario: Sessions survive restart and remain recoverable
- **GIVEN** a paired extension and a restarted Cabinet runtime
- **WHEN** the same bound credential reconnects before expiry or the collector rotates/revokes access
- **THEN** valid access MUST survive restart
- **AND** rotation MUST atomically revoke the previous credential
- **AND** the Integrations UI MUST list redacted sessions and support revoke-one and revoke-all without showing bearer material.

### Requirement INTEGRATION-069: Module discovery MUST expose only safe enabled profile configuration
Authenticated module discovery SHALL project only browser-capable modules backed by enabled integration instances in the paired profile.

#### Scenario: Discover profile modules
- **GIVEN** a paired session with `modules:read`
- **WHEN** `GET /api/companion/modules` is requested
- **THEN** Cabinet MUST return protocol version, paired profile, module/provider identifiers, integration-instance identifier, passive actions and safe configuration
- **AND** disabled or cross-profile instances MUST be absent
- **AND** secret, token, password, cookie and API-key configuration MUST be removed server-side.

### Requirement INTEGRATION-070: Companion transport MUST be bounded, versioned and auditable
The loopback transport SHALL negotiate protocol v1 capabilities and enforce body, media, rate and concurrency limits before item/media persistence executes.

#### Scenario: Reject unsafe transport requests
- **GIVEN** a pairing, capture or media request
- **WHEN** its JSON is malformed or oversized, media exceeds 8 MiB, required checksum/idempotency/profile metadata is absent, the content type is unsupported, or rate/concurrency limits are exceeded
- **THEN** Cabinet MUST fail closed with a deterministic error and appropriate `4XX` response
- **AND** it MUST NOT persist item or media data.

#### Scenario: Validate media transport before persistence
- **GIVEN** a session with `media:submit` and a JPEG or PNG body no larger than 8 MiB
- **WHEN** the bound profile, parent capture, typed field, filename, SHA-256 and bounded idempotency key are valid
- **THEN** Cabinet MUST verify the declared type against decoded image bytes and dimensions
- **AND** an HTML login/challenge response or invalid image MUST fail before any file or database record is written.

### Requirement INTEGRATION-071: Pairing recovery MUST preserve the local trust boundary
Cabinet SHALL document the browser-origin, local-compromise and recovery boundaries and SHALL keep user approval and revocation controls inside the local Cabinet UI.

#### Scenario: Recover from a lost or suspicious extension
- **GIVEN** an extension is removed, its local storage is lost, a credential is suspected compromised, or the Chrome/Edge extension identity changes
- **WHEN** the collector follows the Browser Companion recovery guidance
- **THEN** they MUST be able to revoke one or all profile sessions and pair again
- **AND** development and production extension origins MUST pair separately
- **AND** no recovery step may ask the collector to copy a credential into a URL, log, screenshot or chat.

### Requirement INTEGRATION-072: One MV3 host MUST project enabled integrations from Cabinet
The Browser Companion SHALL use one Chrome/Edge Manifest V3 host whose popup and background runtime consume versioned profile modules from Cabinet rather than provider-specific host branches.

#### Scenario: Project zero, one or many enabled modules
- **GIVEN** a paired profile with any number of enabled browser-capable integration instances
- **WHEN** the extension refreshes authenticated module discovery
- **THEN** it MUST show a Cabinet row and one accessible row per returned integration-instance identifier
- **AND** disabled, cross-profile or hardcoded provider rows MUST remain absent
- **AND** a new valid provider module MUST NOT require a popup or background-worker code change.

### Requirement INTEGRATION-073: Provider permission and readiness MUST remain truthful
Each browser module SHALL declare bounded exact HTTPS origins, a start URL and selectors for ready, logged-out and challenge states.

#### Scenario: Check a provider session
- **GIVEN** a module row in the companion
- **WHEN** the collector grants or removes optional site access, opens its provider page, or checks session readiness
- **THEN** the extension MUST distinguish permission-required, browser-required, logged-out, action-required, ready and unsupported states
- **AND** an open tab alone MUST NOT prove login
- **AND** challenge evidence MUST take priority and MUST NOT be solved or bypassed by the extension.

#### Scenario: Keep Cabinet authoritative for module behaviour
- **GIVEN** an enabled browser integration
- **WHEN** Cabinet publishes its versioned module definition
- **THEN** the definition MUST include URL patterns, capture schemas, supported workflows, redaction rules, fixture version, capture mode, item/media policy, review destination, cadence and help path
- **AND** `sync_available` MUST remain false until a packaged capture script and durable Cabinet persistence path both exist.

### Requirement INTEGRATION-074: Background sync control MUST survive suspension
The MV3 service worker SHALL persist idempotent queued work and expose bounded retry, pending and error state without claiming provider or Cabinet mutations that have not completed.

#### Scenario: Resume after interruption
- **GIVEN** a queued passive sync job and an interrupted service worker, browser or Cabinet runtime
- **WHEN** the extension resumes
- **THEN** it MUST restore one copy of the job, apply bounded retry and circuit-breaker state, and remove it only after Cabinet accepts it
- **AND** the popup and badge MUST keep pending or error state visible.

### Requirement INTEGRATION-075: Extension permissions and privacy MUST be reviewable
The extension SHALL document its loopback access, credential storage, optional-origin model, passive-only boundary and permission-removal path.

#### Scenario: Review or withdraw browser access
- **GIVEN** a collector evaluating or using the companion
- **WHEN** they review its privacy disclosure or remove a module permission
- **THEN** the disclosure MUST state that cookies, passwords, tokens and challenge answers are prohibited
- **AND** removing the exact site permission MUST stop that module's browser access without revoking unrelated modules.

### Requirement INTEGRATION-076: Companion item and media sync MUST be durable, idempotent and recoverable
Cabinet SHALL commit each validated versioned capture envelope before acknowledgement, dispatch typed records through canonical review pipelines and retain actionable queue provenance across restart, backup and relocation.

#### Scenario: Commit typed item and purchase observations
- **GIVEN** a paired module submits a valid search, item-detail, purchase, readiness or explicit user-intent envelope
- **WHEN** the module, schema, provider, integration instance, source origin, redaction summary, payload digest and idempotency key match the registry contract
- **THEN** Cabinet MUST durably commit the raw envelope before reporting `committed=true`
- **AND** it MUST dispatch provider observations to Market Watch/Discoveries or purchases to the review inbox without directly creating inventory
- **AND** replay of the same digest/key MUST not duplicate observations while reuse of the key with another digest MUST fail.

#### Scenario: Report a successful provider capture in Market Watch run history
- **GIVEN** a complete Browser Companion provider search dispatches one or more normalized candidates
- **WHEN** the collector reads `GET /api/scanner/runs?query_set_id=<captured-query-set>`
- **THEN** the persisted run MUST retain the canonical provider scope and `trigger_type=browser_companion`
- **AND** its terminal status MUST be `succeeded` with a positive `result_count`
- **AND** a historical Browser Companion row persisted as `completed` MUST be reported as `succeeded` after upgrade
- **AND** failed provider runs MUST retain their failure state and recovery evidence.

#### Scenario: Preserve partial and interrupted synchronisation
- **GIVEN** a partial page range or a job interrupted while Cabinet or the MV3 worker restarts
- **WHEN** processing resumes
- **THEN** previously observed candidates MUST remain available and must not be interpreted as deleted
- **AND** the capture inbox MUST expose pending, partial, review, retryable, failed and cancelled state with a bounded checkpoint
- **AND** the extension MUST remove its durable job only after Cabinet returns a committed terminal acknowledgement.

#### Scenario: Persist and deduplicate canonical images
- **GIVEN** a valid media submission tied to an accepted capture and typed media field
- **WHEN** its JPEG/PNG bytes, digest and dimensions pass validation
- **THEN** Cabinet MUST write the immutable original, renditions and provenance manifest through canonical media storage before acknowledgement
- **AND** equal content in one profile MUST reuse one asset while retaining every capture/field link
- **AND** the response MUST not disclose a local filesystem path.

### Requirement INTEGRATION-077: Frontline capture MUST preserve provider identity and fail closed
Cabinet SHALL expose one passive, user-present Frontline Hobbies module for enabled `au-webshop-frontlinehobbies-com-au` instances while persisting Market Watch output under the canonical `frontlinehobbies` scope.

#### Scenario: Capture a supported public Frontline result page
- **GIVEN** the paired collector grants the exact Frontline storefront and CDN origins and opens a supported public result page
- **WHEN** the module proves a ready product-card shape and the collector starts sync
- **THEN** it MUST emit bounded version-1 search results with canonical URL, AUD price, stock, image and field-confidence evidence
- **AND** Cabinet MUST preserve integration, transport, module and schema provenance through Market Watch, Discoveries and confirmed Wishlist or Inventory hand-off
- **AND** the extension MUST perform no provider fetch, click, cookie/token access, challenge solution or remote write.

#### Scenario: Fail closed on an incomplete or unsupported page
- **GIVEN** a paginated result, sign-in form, challenge page or changed product-card shape
- **WHEN** the module checks readiness or capture state
- **THEN** it MUST report partial, signed-out, action-required or selector-drift state respectively
- **AND** a partial capture MUST NOT remove earlier candidates
- **AND** fixtures MUST remain explicitly separate from the external user-present live evidence and packaged acceptance required by #1944 and #1869.

### Requirement INTEGRATION-078: Bonza capture MUST replace challenge decoding with user-present sync
Cabinet SHALL expose one passive, user-present Bonza Slot Cars module for enabled `au-webshop-bonzaslotcars-com-au` instances while persisting Market Watch output under the canonical `bonzaslotcars` scope.

#### Scenario: Capture a supported Bonza result page after normal browser interaction
- **GIVEN** the paired collector grants the exact Bonza storefront origins, opens a supported public result page and completes any normal Sucuri site check themselves
- **WHEN** the module proves a ready WooCommerce product-card shape and the collector starts sync
- **THEN** it MUST emit bounded version-1 search results with listing/variation identity, canonical URL, AUD price, stock, image and field-confidence evidence
- **AND** Cabinet MUST preserve integration, transport, module and schema provenance through the durable Market Watch/Discoveries pipeline
- **AND** only a complete `browser_companion` capture for scope `bonzaslotcars` MAY satisfy the provider's source live-evidence state.

#### Scenario: Fail closed without challenge or session bypass
- **GIVEN** direct Store API extraction or the open browser page encounters a Sucuri challenge
- **WHEN** Cabinet detects the bounded challenge marker
- **THEN** direct ingestion MUST make no cookie-bearing retry and MUST return `browser_action_required`
- **AND** the extension MUST return action-required without exporting script contents, cookies, tokens, challenge answers or raw page data
- **AND** the extension MUST perform no provider fetch, click, challenge decoding, challenge solution or remote write.

#### Scenario: Preserve incomplete and external evidence boundaries
- **GIVEN** a paginated result, sign-in form or changed product-card shape
- **WHEN** the module checks readiness or capture state
- **THEN** it MUST report partial, signed-out or selector-drift state respectively
- **AND** a partial capture MUST NOT remove earlier candidates
- **AND** fixtures MUST remain explicitly separate from the external user-present live evidence and packaged acceptance required by #1945 and #1869.

### Requirement INTEGRATION-079: Browser Companion packages MUST be exact, verifiable and recoverable
Cabinet SHALL produce separate Chrome and Edge private-beta packages from one exact source commit with deterministic contents, a release manifest, file and archive SHA-256 values, protocol compatibility and truthful manual-distribution controls.

#### Scenario: Build and verify one exact candidate
- **GIVEN** a full source commit, source timestamp and unused extension version
- **WHEN** the release workflow builds the Browser Companion candidate
- **THEN** it MUST test before packaging and produce separate Chrome and Edge ZIPs whose allow-listed files, production manifest, target, version, source commit, protocol range and checksums pass the repository verifier
- **AND** development identity and files, tests, fixtures, source maps, secrets, challenge/session bypass code and unexpected permissions MUST be rejected
- **AND** the same inputs MUST reproduce the same package SHA-256 values.

#### Scenario: Install, update or roll back a private candidate
- **GIVEN** a collector receives an exact target ZIP, checksum and release manifest
- **WHEN** they verify and manually load the extracted package
- **THEN** documentation MUST state that it is not an installer or store release and has no automatic updates
- **AND** only the exact provider origin required by an enabled module may be granted at runtime
- **AND** an unknown version, failed checksum, reused version, missing manifest or incompatible Cabinet protocol MUST fail closed
- **AND** upgrade, rollback, revoke and uninstall guidance MUST preserve visible jobs and remove stale paired sessions.

#### Scenario: Keep candidate creation separate from release acceptance
- **GIVEN** source packaging controls have passed
- **WHEN** a private/internal candidate is created under #1868
- **THEN** clean Chrome and Edge install/pair/sync/recovery proof MUST remain pending until #1869 tests the exact package files
- **AND** no external release or immutable tag MAY be published before #1864 approval.
