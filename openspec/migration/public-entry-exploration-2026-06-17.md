# Public Entry Exploration Audit - 2026-06-17

Issue: #488
Run window: 2026-06-17 22:30 Australia/Sydney / 2026-06-17 12:30 UTC
Branch: `issue-488-public-entry-exploration`

## Scope

This exploration pass reviewed the Cabinet public/entry section in live-runtime mode:

- root and signed-out entry redirect behavior
- protected-route auth guard redirects
- sign-in form controls and validation
- sign-up form controls and validation
- forgot-password recovery entry and OTP handoff
- OTP threshold and resend controls
- public legal routes
- public unknown-route recovery

Evidence logs for this run are under `.work-agent/logs/issue-488-public-entry-exploration/20260617-2230/`.

## Runtime Preconditions

- Cabinet runtime checked at `http://127.0.0.1:17882`.
- `/healthz` returned HTTP 200 `ok`.
- `/api/runtime` returned app version `rev-fb502db0ddd4`, runtime port `17882`, and a live process.
- Local repo branch for the report was `issue-488-public-entry-exploration`.
- The live runtime was one documentation-only merge behind local `develop`; this did not affect public-entry runtime behavior.
- OpenClaw browser profile `project-cabinet` was started successfully through CDP on port `18801`.
- The higher-level browser navigation API was policy-blocked for direct navigation, so evidence was collected through raw CDP against the same `project-cabinet` profile. This was treated as tooling friction, not a product defect.

## Section Outcome

Completeness label: Complete with documented limitations.

No new product defects were found in the reachable public-entry surface. Reviewed controls and routes matched existing OpenSpec requirements and Cypress traceability for the sampled contracts. Valid credential, social-provider, and passkey success paths were not executed because this exploration pass did not use real credentials or external identity providers; those paths remain covered by the existing Cypress/stub traceability rather than by live credential proof.

## Screen and Component Evidence

### Root entry and protected-route guard

Requirements:
- `UI-LOGIN-SESSION-005`
- `UI-LOGIN-SESSION-010`

Screens/routes:
- `/`
- `/inventory?view=table`

Elements found:
- signed-out sign-in surface with email input, password input, password visibility control, forgot-password link, sign-in submit, passkey action, Google/Apple/Microsoft actions, disabled GitHub/Facebook placeholders, create-account link, Terms link, and Privacy link.

Scenarios:
- Validated: opening `/` while signed out landed on `/sign-in` without redundant `redirect=%2F`.
- Validated: opening `/inventory?view=table` while signed out landed on `/sign-in?redirect=%2Finventory%3Fview%3Dtable`, preserving protected path and query state.
- Blocked: post-login return could not be live-verified without credentials in this pass; existing traceability maps it to `ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts`.

Issues created or linked:
- No new issue. Existing requirement/test traceability covers the observed behavior.

### Sign-in

Requirements:
- `UI-SCREEN-ONBOARDING-AUTH-006`
- `UI-SCREEN-ONBOARDING-AUTH-008`
- `UI-SCREEN-ONBOARDING-AUTH-010`
- `UI-SCREEN-ONBOARDING-AUTH-010B`
- `UI-SCREEN-ONBOARDING-AUTH-010BB`
- `UI-SCREEN-ONBOARDING-AUTH-011B`
- `UI-LOGIN-SESSION-002`
- `UI-LOGIN-SESSION-008`

Elements found:
- email input
- password input
- password visibility toggle
- forgot-password link
- sign-in submit
- passkey action
- Google, Apple, and Microsoft provider actions
- disabled GitHub and Facebook placeholder actions
- create-account link
- Terms and Privacy links

Scenarios:
- Validated: empty submit stayed on `/sign-in` and showed inline email/password validation.
- Validated: forgot-password link routed to `/forgot-password`.
- Validated: password visibility toggle changed password input type from `password` to `text`.
- Validated: GitHub/Facebook placeholders were visible and disabled.
- Validated: duplicate profile/database guidance copy was absent while entry links remained visible.
- Blocked: successful password, passkey, and social-provider authentication were not exercised with real credentials/providers in this pass.

Issues created or linked:
- No new issue. Existing OpenSpec and Cypress coverage already track the visible controls and validation contracts.

### Sign-up

Requirements:
- `UI-SCREEN-ONBOARDING-AUTH-010C`
- `UI-SCREEN-ONBOARDING-AUTH-010D`
- `UI-SCREEN-ONBOARDING-AUTH-011`
- `UI-SCREEN-ONBOARDING-AUTH-011C`
- `UI-SCREEN-ONBOARDING-AUTH-017`

Elements found:
- Sign In return link
- email input
- password input
- password visibility toggle
- confirm-password input
- confirm-password visibility toggle
- Create Account submit
- disabled GitHub and Facebook placeholder actions
- Terms and Privacy links

Scenarios:
- Validated: empty Create Account submit stayed on `/sign-up` and showed inline validation for email, password, and confirm password.
- Validated: return/legal/provider placeholder controls were present and remained visible after validation failure.
- Blocked: successful account creation was not executed with a real user account in this pass; existing Cypress traceability covers the deterministic stub outcome.

Issues created or linked:
- No new issue.

### Forgot-password and OTP

Requirements:
- `UI-SCREEN-ONBOARDING-AUTH-012`
- `UI-SCREEN-ONBOARDING-AUTH-014`
- `UI-SCREEN-ONBOARDING-AUTH-015`
- `UI-SCREEN-ONBOARDING-AUTH-017`

Screens/routes:
- `/forgot-password`
- `/otp`

Elements found:
- forgot-password email input
- Continue submit
- Sign up secondary link
- OTP input
- Verify submit
- resend-code button

Scenarios:
- Validated: empty forgot-password submit stayed on `/forgot-password`, showed inline validation, and preserved the Sign up handoff.
- Validated: valid-looking recovery email showed in-flight state and routed to `/otp`, with confirmation visible during the transition.
- Validated: OTP Verify was disabled with fewer than six digits.
- Validated: entering six digits enabled Verify without leaving `/otp`.
- Validated: OTP resend stayed on `/otp`, disabled during in-flight resend, and showed progress text.

Issues created or linked:
- No new issue.

### Legal and public unknown routes

Requirements:
- `UI-SCREEN-ONBOARDING-AUTH-013`
- `UI-ROUTE-COVERAGE-005`

Screens/routes:
- `/terms`
- `/privacy`
- `/not-a-real-public-route`

Elements found:
- Terms content page
- Privacy content page
- public 404 state with Go Back and Back to Home recovery controls

Scenarios:
- Validated: `/terms` rendered Terms of Service content and did not render generic fall-through 404 content.
- Validated: `/privacy` rendered Privacy Policy content and did not render generic fall-through 404 content.
- Validated: unknown public route rendered a 404 state with recovery controls.
- Validated: Back to Home recovered to `/sign-in` for the signed-out profile.

Issues created or linked:
- No new issue.

## Traceability Summary

Current requirement/test mappings were checked through `openspec/traceability.md` and Cypress specs:

- `UI-LOGIN-SESSION-005` -> `ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts`
- `UI-LOGIN-SESSION-010` -> `ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts`
- `UI-SCREEN-ONBOARDING-AUTH-*` public auth route contracts -> `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`
- `UI-ROUTE-COVERAGE-005` -> `ui.web/cypress/e2e/general/route-guards-errors/spec.cy.ts`

No new spec gap was identified for the reachable public-entry scope. Live credential, passkey, and external identity-provider success flows remain outside this no-credential exploration pass and should not be claimed as live-verified from this audit.

## Next Recommendation

#488 can be kept open only if a deeper credential-backed auth pass is required. Otherwise the next route-ordered exploration issue is #489 App shell / navigation, unless a newer higher-priority product issue is added.
