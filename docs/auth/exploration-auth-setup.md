# Exploration Auth Setup

## Purpose
This guide makes Cabinet exploratory auth deterministic for local development, review, and route-audit work.

## Default exploratory path: use local auth
For exploratory testing, prefer **local auth mode** unless the specific task is explicitly about Clerk.

Why:
- no external auth dependency
- no publishable-key/domain setup required
- fastest path to authenticated route coverage
- avoids Clerk-origin and passkey-domain blockers during routine UI review

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
- these credentials are local-only exploratory credentials, not a Clerk account requirement

## Getting authenticated sample data
After local sign-in, use one of these paths:

### Option A: starter/onboarding sample flow
Use the built-in onboarding sample data path after first auth completion when the task needs a freshly bootstrapped dataset.

Concrete path:
1. complete first local sign-in
2. continue through the onboarding starter-data choice
3. choose the starter/sample-data path so Cabinet seeds the profile via `POST /api/onboarding/sample-data`

### Option B: Showcase DB profile
When the profile switcher is available, choose **Showcase DB** for deterministic demo content across inventory, wishlist, and related routes.

Use Showcase DB when you need:
- repeatable route traversal
- visible seeded content for demos
- stable exploratory screenshots/evidence

## When to use Clerk instead
Use **Clerk auth mode** only when the task explicitly needs:
- Clerk route coverage (`/clerk/*`)
- Clerk token/session bootstrap behavior
- Clerk billing/plan/permissions verification
- Clerk-specific auth UX or error handling

## Clerk exploratory prerequisites
Before attempting Clerk exploration, confirm all of the following:

1. `VITE_CLERK_PUBLISHABLE_KEY` is set.
2. The runtime/setup flow is configured for `clerk` auth mode.
3. The browser origin you are using is allowed by Clerk.
4. If passkeys are involved, the active domain/origin matches the configured passkey relying-party expectations.

## Troubleshooting matrix
| Symptom | Likely cause | What to do |
| --- | --- | --- |
| `Missing Clerk key` in setup wizard | Clerk mode selected without publishable key | Switch back to `local` for exploratory work, or provide `VITE_CLERK_PUBLISHABLE_KEY` before continuing |
| `/clerk` route shows "No Publishable Key Found!" | `VITE_CLERK_PUBLISHABLE_KEY` not loaded | Create/update `.env`, restart the app, and confirm the key is visible to the UI runtime |
| `This is an invalid domain.` during passkey sign-in | Current origin/domain is not passkey-enabled for the auth setup | Prefer local password/provider sign-in for exploration; if validating Clerk/passkeys specifically, align the active domain/origin with the configured relying-party/domain settings |
| Passkey guidance says `Passkey sign-in is not available on this domain yet...` | Cabinet normalized a domain/origin mismatch into deterministic fallback guidance | Treat this as an auth-environment setup issue, not a generic UI failure; continue with password/provider sign-in or fix the domain/origin setup |
| Clerk sign-in page loads but session/bootstrap fails | Incomplete Clerk env, origin, or token/bootstrap configuration | Re-check publishable key, allowed origins, and the exact runtime URL being used |
| Exploratory route audit is blocked on auth ambiguity | Wrong auth mode chosen for the task | Reset to the default rule: local mode for routine exploration, Clerk only for Clerk-specific work |

## Recommended exploratory decision rule
- **Routine route audits / UI review:** use **local** auth mode
- **Cloud/plan/entitlement verification:** use **Clerk**
- **Passkey-specific validation:** confirm domain/origin first; otherwise do not treat passkey domain mismatch as a product-route blocker

## Evidence to capture in exploratory reports
Whenever auth affects a route review, record:
- auth mode used (`local` or `clerk`)
- runtime URL
- startup path used (`start-exploration-local.ps1`, direct `bin\cabinet.exe`, or other explicit launcher)
- whether setup wizard was completed or bypassed
- whether the session used first-sign-in local bootstrap or an existing local account
- active profile/sample-data path used (`starter` vs `Showcase DB`)
- exact auth blocker text if any
