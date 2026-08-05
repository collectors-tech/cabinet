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

### Requirement INTEGRATION-067: Pairing MUST require explicit approval inside unlocked Cabinet
One extension installation SHALL establish trust through a short-lived request, visible six-digit code, explicit profile-scoped approval and one-time exchange.

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
The loopback transport SHALL negotiate protocol v1 capabilities and enforce body, media, rate and concurrency limits before later item/media persistence work executes.

#### Scenario: Reject unsafe transport requests
- **GIVEN** a pairing, capture or media request
- **WHEN** its JSON is malformed or oversized, media exceeds 8 MiB, required checksum/idempotency/profile metadata is absent, the content type is unsupported, or rate/concurrency limits are exceeded
- **THEN** Cabinet MUST fail closed with a deterministic error and appropriate `4XX` response
- **AND** it MUST NOT persist item or media data.

#### Scenario: Validate media transport before persistence
- **GIVEN** a session with `media:submit` and an image or octet-stream body no larger than 8 MiB
- **WHEN** the bound profile, SHA-256 and bounded idempotency key are valid
- **THEN** Cabinet MUST accept the v1 transport checks
- **AND** until #2032 supplies durable media ingestion it MUST return `501 companion_media_persistence_not_implemented` rather than claiming synchronisation.

### Requirement INTEGRATION-071: Pairing recovery MUST preserve the local trust boundary
Cabinet SHALL document the browser-origin, local-compromise and recovery boundaries and SHALL keep user approval and revocation controls inside the local Cabinet UI.

#### Scenario: Recover from a lost or suspicious extension
- **GIVEN** an extension is removed, its local storage is lost, a credential is suspected compromised, or the Chrome/Edge extension identity changes
- **WHEN** the collector follows the Browser Companion recovery guidance
- **THEN** they MUST be able to revoke one or all profile sessions and pair again
- **AND** development and production extension origins MUST pair separately
- **AND** no recovery step may ask the collector to copy a credential into a URL, log, screenshot or chat.
