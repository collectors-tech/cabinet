## 1. Contract first

- [x] 1.1 Create issue #1952 and record the security acceptance checklist.
- [x] 1.2 Add the auth contract test and capture its four failing assertions.
- [x] 1.3 Define the OpenSpec identity and role boundary.

## 2. Backend application auth

- [x] 2.1 Add discovery, Authorization Code + PKCE, state and nonce.
- [x] 2.2 Validate JWKS signatures and strict token claims.
- [x] 2.3 Add opaque HTTP-only application sessions, refresh and logout.
- [x] 2.4 Enforce remote API sessions and backend admin roles.

## 3. Product and deployment integration

- [x] 3.1 Add the Cabinet-branded login, signed-out and error states.
- [x] 3.2 Add per-environment ZITADEL values and shared Login V2 contract.
- [x] 3.3 Preserve truthful local-device mode.

## 4. Evidence

- [x] 4.1 Run focused Go, UI, contract and OpenSpec validation.
- [x] 4.2 Run the production UI build and CI regression gates.
- [x] 4.3 Commit, push and open the issue-linked stacked PR.
