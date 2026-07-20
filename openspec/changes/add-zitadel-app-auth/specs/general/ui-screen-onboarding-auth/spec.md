## ADDED Requirements

### Requirement: Cabinet SHALL use a server-owned ZITADEL OIDC session for remote identity

When remote identity mode is `zitadel`, Cabinet SHALL use Authorization Code
with PKCE through the backend and SHALL keep access, ID and refresh tokens out
of browser persistence and application URLs.

#### Scenario: Remote sign-in succeeds

- **GIVEN** the environment's isolated Cabinet application and role grant exist
- **WHEN** a user begins sign-in from `/sign-in` and completes the branded login
- **THEN** the backend MUST consume state once and validate nonce, signature,
  issuer, client and environment audience, authorised party and expiry
- **AND** the user MUST have an allowed Cabinet role
- **AND** the browser MUST receive only an opaque HTTP-only SameSite application
  session cookie

#### Scenario: Remote identity is denied

- **WHEN** discovery, state, nonce, signature, issuer, audience, authorised
  party, expiry or role validation fails
- **THEN** Cabinet MUST NOT create an application session
- **AND** the sign-in surface MUST show a stable error without provider secrets

### Requirement: Cabinet SHALL retain a branded identity experience

Cabinet SHALL keep `/sign-in` as the product entry surface and the shared
identity owner SHALL configure ZITADEL Login V2 on the Cabinet identity domain
with Cabinet organisation branding, private labelling and a verified custom
Login V2 base URL. Cabinet SHALL NOT collect the ZITADEL password itself.

#### Scenario: User enters remote sign-in

- **WHEN** a user opens `/sign-in` in a ZITADEL environment
- **THEN** the page MUST identify Cabinet rather than a generic identity vendor
- **AND** it MUST explain that account creation and recovery continue in the
  secure identity step
- **AND** it MUST NOT persist a password or provider token

#### Scenario: Login V2 is provisioned for an environment

- **WHEN** the identity owner provisions the environment's Cabinet application
- **THEN** **Use new login UI** MUST be enabled with the exact configured Login
  V2 base URL and trusted domain
- **AND** the Cabinet logo, colours, private-labelling policy and hidden vendor
  watermark MUST be applied and verified before cutover

### Requirement: Remote Cabinet APIs SHALL enforce session and role authority in the backend

Remote APIs SHALL reject a missing, invalid or expired Cabinet application
session. User-administration APIs SHALL additionally require `cabinet.admin`.

#### Scenario: Normal role calls an admin endpoint

- **GIVEN** a valid remote session with only `cabinet.user` or `cabinet.demo`
- **WHEN** it calls `/api/users` or a child endpoint
- **THEN** the backend MUST return forbidden

### Requirement: Identity configuration SHALL be isolated by environment

Local, demo and production SHALL use distinct ZITADEL applications, exact
redirect and post-logout URLs, audiences, secrets and role requirements while
consuming the same governed shared identity foundation.

#### Scenario: Demo identity is presented to production

- **WHEN** a valid demo token or application session is presented to production
- **THEN** production MUST deny it because its audience, authorised party or
  role contract does not match
