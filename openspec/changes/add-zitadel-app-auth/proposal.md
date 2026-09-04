# Change: add Cabinet ZITADEL application authentication

## Why

The self-hosted demo and production runtimes cannot treat local-device entry or
an unverified browser token as remote authentication. Cabinet needs a branded
application login backed by the existing shared ZITADEL foundation without
deploying another identity service.

## What changes

- Add a provider-neutral backend OIDC boundary with a ZITADEL implementation.
- Use Authorization Code with PKCE, state and nonce, discovery and rotating
  signing keys.
- Store provider tokens only in an opaque server-side application session and
  expose an HTTP-only SameSite cookie to the browser.
- Keep `/sign-in` as the Cabinet product surface and hand credentials to the
  Cabinet-branded shared ZITADEL Login V2 experience.
- Enforce remote API sessions and Cabinet roles in the backend.
- Configure distinct local, demo and production applications and exact URLs.

## Impact

- Affected specs: `general/ui-screen-onboarding-auth`
- Affected code: backend auth routes/middleware, sign-in/sign-out guards,
  deployment identity contracts, operations documentation and CI contracts
- Related issues: #1952, #1951
