# Exploration Auth Setup

## Purpose
This guide makes Cabinet exploratory auth deterministic for local development, review, and route-audit work.

## Default exploratory path: use local auth
For exploratory testing, prefer **local auth mode** unless the specific task is explicitly about ZITADEL application login, cloud entitlements, or passkey/domain behavior.

Why:
- no external auth dependency
- no publishable-key/domain setup required
- fastest path to authenticated route coverage
- avoids identity-provider origin and passkey-domain blockers during routine UI review

## Local auth flow
### Startup
From the repo root, preferred exploratory launcher:

```powershell
pwsh -NoLogo -NoProfile -File .\scripts\runtime\start-exploration-local.ps1 -Background
```

Manual fallback:

```powershell
.\bin\cabinet.exe
```

Default local runtime assumptions:
- app URL: `http://127.0.0.1:17880`
- WebAuthn RP ID: `127.0.0.1`
- WebAuthn origin: `http://127.0.0.1:17880`

### First-run setup wizard
If Cabinet is not configured yet:
1. Complete the setup wizard.
2. Choose **Auth Mode = local**.
3. Finish config creation and launch.

### Local sign-in behavior
For exploratory local auth, Cabinet does **not** require a pre-seeded local account.
The current sign-in form accepts:
- any syntactically valid email address
- any password with length `>= 7`

Example exploratory credentials:
- email: `explorer@cabinet.local`
- password: `password123`

Account-creation expectation:
- treat the first successful local sign-in as the exploratory local account bootstrap path
- do **not** block route audits waiting for a separately provisioned sample account
- these credentials are local-only exploratory credentials, not a ZITADEL account requirement

## Getting authenticated sample data
After local sign-in, use one of these paths:

### Option A: starter/onboarding sample flow
Use the built-in onboarding sample data path after first auth completion when the task needs a freshly bootstrapped dataset.

Concrete path:
1. complete first local sign-in
2. continue through the onboarding starter-data choice
3. choose the starter/sample-data path so Cabinet seeds the profile via `POST /api/onboarding/sample-data`

Verification target:
- targeted Cypress proof: `ui.web/cypress/e2e/general/onboarding-starter-data/spec.cy.ts`
- run with: `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/onboarding-starter-data/spec.cy.ts -Browser chrome -RequireE2EHooks`

### Option B: Showcase DB profile
When the profile switcher is available, choose **Showcase DB** for deterministic demo content across inventory, wishlist, and related routes.

Use Showcase DB when you need:
- repeatable route traversal
- visible seeded content for demos
- stable exploratory screenshots/evidence

Verification target:
- targeted Cypress proof: `ui.web/cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts` (`UI-FOUNDATION-SHELL-NAVIGATION-010 provides Showcase DB profile with seeded demo context`)
- run with: `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks`

## When to use ZITADEL instead
Use **ZITADEL auth mode** only when the task explicitly needs:
- ZITADEL application login or callback coverage (`/api/auth/zitadel/*`)
- verified server-managed cloud session and effective entitlement behavior
- billing/plan/permissions verification
- provider-specific auth UX, role, cookie, or error handling

## ZITADEL exploratory prerequisites
Before attempting ZITADEL exploration, confirm all of the following:

1. `CABINET_AUTH_IDENTITY_MODE=zitadel` is set.
2. `CABINET_ZITADEL_ISSUER`, `CABINET_ZITADEL_CLIENT_ID`, and `CABINET_ZITADEL_AUDIENCE` are set.
3. The browser origin you are using matches the configured Cabinet public origin and ZITADEL callback URL.
4. If passkeys are involved, the active domain/origin matches the configured passkey relying-party expectations.

### ZITADEL startup example
From the repo root, set the environment for the current shell, then start Cabinet normally:

```powershell
$env:CABINET_AUTH_IDENTITY_MODE = 'zitadel'
$env:CABINET_ZITADEL_ISSUER = 'https://identity.example.com'
$env:CABINET_ZITADEL_CLIENT_ID = 'cabinet-client'
$env:CABINET_ZITADEL_AUDIENCE = 'cabinet-project'
pwsh -NoLogo -NoProfile -File .\scripts\runtime\start-exploration-local.ps1 -Background
```

Expected verification after launch:
- `http://127.0.0.1:17880/healthz` returns `200 ok`
- `http://127.0.0.1:17880/api/auth/provider-options` returns `identity_mode = "zitadel"` and `zitadel_configured = true`
- first-run flows should be completed with **Auth Mode = zitadel**

## Troubleshooting matrix
| Symptom | Likely cause | What to do |
| --- | --- | --- |
| `Auth mode must be local or zitadel.` | Retired or misspelled auth mode was submitted | Switch setup/startup config to `local` for routine exploration or `zitadel` for provider-specific validation |
| `/api/auth/provider-options` reports `zitadel_configured=false` | Required `CABINET_ZITADEL_*` values are incomplete | Set issuer, client id, and audience, then restart and re-check provider options |
| `This is an invalid domain.` during passkey sign-in | Current origin/domain is not passkey-enabled for the auth setup | Prefer local password/provider sign-in for exploration; if validating ZITADEL/passkeys specifically, align the active domain/origin with the configured relying-party/domain settings |
| Passkey guidance says `Passkey sign-in is not available on this domain yet...` | Cabinet normalized a domain/origin mismatch into deterministic fallback guidance | Treat this as an auth-environment setup issue, not a generic UI failure; continue with password/provider sign-in or fix the domain/origin setup |
| ZITADEL sign-in redirects but session/bootstrap fails | Incomplete ZITADEL env, callback origin, role, or token/bootstrap configuration | Re-check issuer, client id, audience, callback URL, required roles, and the exact runtime URL being used |
| Exploratory route audit is blocked on auth ambiguity | Wrong auth mode chosen for the task | Reset to the default rule: local mode for routine exploration, ZITADEL only for provider-specific work |

## Recommended exploratory decision rule
- **Routine route audits / UI review:** use **local** auth mode
- **Cloud/plan/entitlement verification:** use **ZITADEL**
- **Passkey-specific validation:** confirm domain/origin first; otherwise do not treat passkey domain mismatch as a product-route blocker

## Evidence to capture in exploratory reports
Whenever auth affects a route review, record:
- auth mode used (`local` or `zitadel`)
- runtime URL
- startup path used (`start-exploration-local.ps1`, direct `bin\cabinet.exe`, or other explicit launcher)
- whether setup wizard was completed or bypassed
- whether the session used first-sign-in local bootstrap or an existing local account
- active profile/sample-data path used (`starter` vs `Showcase DB`)
- exact auth blocker text if any
