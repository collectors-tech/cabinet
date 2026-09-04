# Design: Cabinet ZITADEL application authentication

## Boundary

Cabinet consumes the shared ZITADEL issuer and Login V2 service. The identity
owner provisions one Cabinet project/application boundary per environment and
applies Cabinet branding. Cabinet never deploys ZITADEL or shares another
product's application credentials.

## Flow

1. `/sign-in` loads provider capability without receiving a secret.
2. The backend creates one-time state, nonce and PKCE verifier data and sends
   only the challenge in the authorisation redirect.
3. The shared Cabinet-branded login authenticates the user.
4. The backend consumes state once, exchanges the code and validates the ID
   token signature, issuer, client and project audiences, authorised party,
   expiry, nonce and required Cabinet role.
5. Tokens remain in the backend. The browser receives a random HTTP-only,
   SameSite=Lax application session cookie, Secure when the public origin is
   HTTPS.
6. Backend API middleware validates that application session and requires
   `cabinet.admin` for user-administration endpoints.
7. Refresh rotates server-side token state. Logout deletes the Cabinet session,
   expires the cookie and hands off to the discovered provider logout endpoint
   without putting an ID token in the URL.

## Roles

- `cabinet.user`: normal production application access
- `cabinet.demo`: demo-environment access
- `cabinet.admin`: application access plus user administration

Audience alone is never authorisation. At least one configured environment
role is mandatory and privileged APIs check `cabinet.admin` independently.

## Failure behaviour

Discovery mismatches, non-HTTPS remote configuration, unknown signing keys,
wrong algorithms or claims, state replay, missing roles, expired sessions and
provider errors fail closed. Browser errors are stable codes and do not contain
provider responses or tokens.

## Custom login decision

Cabinet keeps its branded entry page but does not embed or handle the identity
password. The shared owner configures the self-hostable ZITADEL Login V2 UI and
Session API integration on the Cabinet identity domain. Each environment
enables **Use new login UI**, pins its custom base URL and trusted domain, and
applies the Cabinet organisation branding/private-labelling settings. This
preserves OIDC at the application boundary and avoids implementing a partial
password/MFA flow inside Cabinet.
