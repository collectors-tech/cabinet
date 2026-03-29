## Purpose
Define onboarding and auth screen behavior for first-run identity and workspace unlock.

## Requirements
### Requirement UI-SCREEN-ONBOARDING-AUTH-001: Onboarding/Auth SHALL enforce WebAuthn-first completion
Onboarding/Auth SHALL block advanced workspace access until required identity steps are complete.

#### Scenario: Required identity gate
- **GIVEN** active profile exists and `GET /api/auth/requirements` reports `requires_registration=true` or missing unlock state
- **WHEN** user has incomplete identity setup
- **THEN** advanced workspace routes SHALL remain locked and UI MUST show onboarding gate CTA until `requires_registration=false`

### Requirement UI-SCREEN-ONBOARDING-AUTH-002: Onboarding/Auth SHALL persist and resume progress
Progress SHALL persist through reload/restart until completion.

#### Scenario: Resume incomplete onboarding
- **GIVEN** onboarding progress for active profile is persisted with last incomplete step index
- **WHEN** user returns after restart
- **THEN** onboarding SHALL resume at last incomplete step and step state MUST match persisted profile-scoped progress

### Requirement UI-SCREEN-ONBOARDING-AUTH-003: Onboarding/Auth SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states for profile and auth checks.

#### Scenario: Auth check error
- **GIVEN** onboarding screen requests auth/profile requirements and API call returns failure
- **WHEN** auth requirements fetch fails
- **THEN** screen SHALL show actionable error with retry and MUST NOT advance wizard state until successful fetch

### Requirement UI-SCREEN-ONBOARDING-AUTH-006: Sign-in screen SHALL support Google, Apple, and Microsoft providers
Social/enterprise sign-in options SHALL include Google, Apple, and Microsoft in addition to any existing providers.

#### Scenario: Render provider buttons
- **GIVEN** sign-in screen is rendered
- **WHEN** provider actions are displayed
- **THEN** UI MUST include buttons for `Google`, `Apple`, and `Microsoft` with deterministic enabled/disabled state based on configuration

### Requirement UI-SCREEN-ONBOARDING-AUTH-007: Identity provider platform decision SHALL be explicit and configurable
Authentication implementation SHALL define whether Clerk is the source-of-truth identity platform and expose provider configuration deterministically.

#### Scenario: Resolve identity platform
- **GIVEN** runtime auth configuration is loaded
- **WHEN** sign-in screen initializes
- **THEN** app MUST resolve identity platform mode (for example Clerk-based vs local provider stack) and render only configured provider actions

### Requirement UI-SCREEN-ONBOARDING-AUTH-008: Sign-in SHALL support passkeys (WebAuthn) for passwordless login
Sign-in flow SHALL support passkey authentication compatible with platform authenticators and password managers (for example 1Password passkeys).

#### Scenario: Sign in with passkey
- **GIVEN** user account has enrolled passkey credential
- **WHEN** user selects passkey sign-in and completes WebAuthn prompt
- **THEN** authentication MUST succeed without password and redirect to authenticated shell

#### Scenario: Passkey fallback behavior
- **GIVEN** passkey is unavailable or challenge fails
- **WHEN** passkey auth attempt fails
- **THEN** UI MUST provide deterministic fallback to other enabled methods (password/social) with actionable error message

### Requirement UI-SCREEN-ONBOARDING-AUTH-009: Setup wizard SHALL gate sign-in when runtime setup config is missing
Sign-in route SHALL present a full-screen setup wizard before auth controls when runtime setup config file is missing.

#### Scenario: Setup required branch
- **GIVEN** `GET /api/runtime/setup-status` returns `{"setup_required":true}`
- **WHEN** user opens sign-in route
- **THEN** UI MUST render setup wizard, MUST hide sign-in form, and MUST expose deterministic `Complete Setup` action

#### Scenario: Setup completion branch
- **GIVEN** setup wizard is visible and user triggers `Complete Setup`
- **WHEN** `POST /api/runtime/setup-complete` returns 200 with `{"ok":true,"setup_required":false}`
- **THEN** setup wizard MUST dismiss and sign-in form MUST render without route change

### Requirement UI-SCREEN-ONBOARDING-AUTH-010: Sign-in SHALL expose a visible Create account path
Sign-in screen SHALL include a first-time-user CTA that routes deterministically to sign-up.

#### Scenario: Create account entry path
- **GIVEN** runtime setup is complete and sign-in route is visible
- **WHEN** user scans sign-in actions for first-time account creation
- **THEN** UI MUST show visible `Create account` link/button
- **AND** control MUST navigate to `/sign-up`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010B: Sign-in forgot-password entry SHALL be visible and deterministic
Sign-in SHALL expose a visible forgot-password recovery entry path that supports deterministic mouse and keyboard navigation to `/forgot-password`.

#### Scenario: Forgot-password entry path from sign-in
- **GIVEN** runtime setup is complete and user is on `/sign-in`
- **WHEN** user scans the password field controls for recovery actions
- **THEN** `Forgot password?` MUST be visible on the sign-in surface
- **AND** activation by mouse or keyboard MUST navigate deterministically to `/forgot-password`
- **AND** focus/route handoff MUST complete without side effects on `/sign-in`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010BB: Sign-in GitHub/Facebook actions SHALL be explicit and deterministic
Sign-in SHALL expose GitHub and Facebook provider actions with explicit visible/disabled behavior until provider sign-in flows are implemented.

#### Scenario: Sign-in GitHub and Facebook actions
- **GIVEN** runtime setup is complete and user is on `/sign-in`
- **WHEN** user inspects alternative provider actions on the sign-in surface
- **THEN** UI MUST show visible `GitHub` and `Facebook` actions
- **AND** those actions MUST remain deterministically disabled when no sign-in provider flow is wired
- **AND** keyboard focus/activation MUST NOT navigate away from `/sign-in`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010BC: Sign-in-2 forgot-password entry SHALL be visible and deterministic
`/sign-in-2` SHALL expose a visible forgot-password recovery entry path that supports deterministic mouse and keyboard navigation to `/forgot-password`.

#### Scenario: Forgot-password entry path from sign-in-2
- **GIVEN** runtime setup is complete and user is on `/sign-in-2`
- **WHEN** user scans the password field controls for recovery actions
- **THEN** `Forgot password?` MUST be visible on the sign-in-2 surface
- **AND** activation by mouse or keyboard MUST navigate deterministically to `/forgot-password`
- **AND** focus/route handoff MUST complete without side effects on `/sign-in-2`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010C: Sign-up secondary links SHALL be visible and deterministic
Sign-up SHALL expose visible return/legal links that support deterministic mouse and keyboard navigation to `/sign-in`, `/terms`, and `/privacy`.

#### Scenario: Sign-up sign-in/legal entry paths
- **GIVEN** runtime setup is complete and user is on `/sign-up`
- **WHEN** user scans the form header/footer for account-return and legal actions
- **THEN** `Sign In`, `Terms of Service`, and `Privacy Policy` MUST be visible on the sign-up surface
- **AND** activation by mouse or keyboard MUST navigate deterministically to `/sign-in`, `/terms`, and `/privacy`
- **AND** route handoff MUST complete without side effects on `/sign-up`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010CC: Sign-in-2 legal links SHALL be visible and deterministic
`/sign-in-2` SHALL expose visible legal links that support deterministic mouse and keyboard navigation to `/terms` and `/privacy`.

#### Scenario: Sign-in-2 legal entry paths
- **GIVEN** runtime setup is complete and user is on `/sign-in-2`
- **WHEN** user scans the footer/legal copy for policy actions
- **THEN** `Terms of Service` and `Privacy Policy` MUST be visible on the sign-in-2 surface
- **AND** activation by mouse or keyboard MUST navigate deterministically to `/terms` and `/privacy`
- **AND** route handoff MUST complete without side effects on `/sign-in-2`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010CCD: Sign-in-2 GitHub/Facebook actions SHALL be explicit and deterministic
`/sign-in-2` SHALL expose GitHub and Facebook provider actions with explicit visible/disabled behavior until provider sign-in flows are implemented.

#### Scenario: Sign-in-2 GitHub and Facebook actions
- **GIVEN** runtime setup is complete and user is on `/sign-in-2`
- **WHEN** user inspects alternative provider actions on the sign-in-2 surface
- **THEN** UI MUST show visible `GitHub` and `Facebook` actions
- **AND** those actions MUST remain deterministically disabled when no provider flow is wired
- **AND** keyboard focus/activation MUST NOT navigate away from `/sign-in-2`

### Requirement UI-SCREEN-ONBOARDING-AUTH-010D: Sign-up GitHub/Facebook actions SHALL be explicit and deterministic
Sign-up SHALL expose GitHub and Facebook provider actions with explicit visible/disabled behavior until provider sign-up flows are implemented.

#### Scenario: Sign-up GitHub and Facebook actions
- **GIVEN** runtime setup is complete and user is on `/sign-up`
- **WHEN** user inspects alternative provider actions on the sign-up surface
- **THEN** UI MUST show visible `GitHub` and `Facebook` actions
- **AND** those actions MUST remain deterministically disabled when no sign-up provider flow is wired
- **AND** keyboard focus/activation MUST NOT navigate away from `/sign-up`

### Requirement UI-SCREEN-ONBOARDING-AUTH-011: Sign-up submit SHALL provide deterministic completion feedback
Sign-up flow SHALL provide a deterministic outcome after valid submission so first-time users are never left on a dead-end state.

### Requirement UI-SCREEN-ONBOARDING-AUTH-011B: Sign-in password visibility toggle SHALL be deterministic and keyboard-accessible
Sign-in password field SHALL expose a visibility toggle that switches the input type deterministically, updates accessible state/label, and remains keyboard-activatable.

#### Scenario: Sign-in password visibility toggle
- **GIVEN** runtime setup is complete and user is on `/sign-in`
- **WHEN** user activates the password visibility toggle by mouse or keyboard
- **THEN** password input MUST switch deterministically between masked and text-visible modes
- **AND** toggle accessible label/state MUST update to reflect the current mode
- **AND** focus/activation MUST remain on the sign-in route without side effects

### Requirement UI-SCREEN-ONBOARDING-AUTH-011C: Sign-up password visibility toggles SHALL be deterministic and keyboard-accessible
Sign-up password and confirm-password fields SHALL expose visibility toggles that switch deterministically, update accessible state/label, and remain keyboard-activatable.

#### Scenario: Sign-up password visibility toggles
- **GIVEN** runtime setup is complete and user is on `/sign-up`
- **WHEN** user activates the password or confirm-password visibility toggle by mouse or keyboard
- **THEN** each field MUST switch deterministically between masked and text-visible modes
- **AND** each toggle accessible label/state MUST update to reflect the current mode
- **AND** focus/activation MUST remain on the sign-up route without side effects

### Requirement UI-SCREEN-ONBOARDING-AUTH-011D: Sign-in-2 password visibility toggle SHALL be deterministic and keyboard-accessible
The `/sign-in-2` password field SHALL expose a visibility toggle that switches deterministically, updates accessible state/label, and remains keyboard-activatable.

#### Scenario: Sign-in-2 password visibility toggle
- **GIVEN** runtime setup is complete and user is on `/sign-in-2`
- **WHEN** user activates the password visibility toggle by mouse or keyboard
- **THEN** password input MUST switch deterministically between masked and text-visible modes
- **AND** toggle accessible label/state MUST update to reflect the current mode
- **AND** focus/activation MUST remain on the `/sign-in-2` route without side effects

#### Scenario: Successful sign-up completion
- **GIVEN** user is on `/sign-up` with valid email/password/confirm password values
- **WHEN** user activates `Create Account`
- **THEN** UI MUST show in-progress state while submitting
- **AND** on success MUST authenticate the new user session and navigate to the canonical authenticated dashboard destination (`/dashboard`)

#### Scenario: Failed sign-up completion
- **GIVEN** user submits valid sign-up payload and backend returns failure
- **WHEN** submission fails
- **THEN** UI MUST show actionable error feedback and remain recoverable on `/sign-up`

### Requirement UI-SCREEN-ONBOARDING-AUTH-012: Forgot-password submit SHALL provide deterministic recovery handoff
Forgot-password flow SHALL provide a deterministic next-step outcome after valid email submission so users are never left on a cleared form with no recovery guidance.

#### Scenario: Successful forgot-password handoff
- **GIVEN** user is on `/forgot-password` with a valid email value
- **WHEN** user activates `Continue`
- **THEN** UI MUST show in-progress state while submitting
- **AND** on success MUST navigate to `/otp` as the recovery verification step
- **AND** MUST NOT replace the successful submit outcome with a fresh empty-field validation error on `/forgot-password`

#### Scenario: Failed forgot-password handoff
- **GIVEN** user submits a valid forgot-password payload and backend returns failure
- **WHEN** submission fails
- **THEN** UI MUST show actionable error feedback and remain recoverable on `/forgot-password`

#### Scenario: Forgot-password keyboard submit and secondary handoff
- **GIVEN** user is on `/forgot-password`
- **WHEN** user submits the form with keyboard activation after entering a valid email
- **THEN** `Continue` MUST show deterministic in-flight handling and complete the same OTP handoff as click submission
- **AND** the visible `Sign up` secondary link MUST remain keyboard-activatable and route deterministically to `/sign-up`

### Requirement UI-SCREEN-ONBOARDING-AUTH-013: Legal auth links SHALL resolve to public policy content
Auth legal links SHALL route to public content pages instead of falling through to the generic not-found screen.

#### Scenario: Privacy policy route
- **GIVEN** runtime setup is complete and an unauthenticated visitor opens `/privacy`
- **WHEN** the route renders from a sign-in or sign-up legal link
- **THEN** UI MUST render `Privacy Policy` content
- **AND** MUST NOT render the generic `404` not-found screen

#### Scenario: Terms of service route
- **GIVEN** runtime setup is complete and an unauthenticated visitor opens `/terms`
- **WHEN** the route renders from a sign-in or sign-up legal link
- **THEN** UI MUST render `Terms of Service` content
- **AND** MUST NOT render the generic `404` not-found screen

### Requirement UI-SCREEN-ONBOARDING-AUTH-014: OTP resend SHALL remain in OTP recovery context
OTP recovery SHALL provide an in-place resend action that confirms resend progress/outcome without redirecting the user back to sign-in.

#### Scenario: Successful OTP resend
- **GIVEN** user is on `/otp` after entering the recovery verification step
- **WHEN** user activates `Resend a new code.`
- **THEN** UI MUST keep the user on `/otp`
- **AND** MUST show a deterministic resend progress/confirmation message
- **AND** MUST NOT redirect to `/sign-in`

### Requirement UI-SCREEN-ONBOARDING-AUTH-015: OTP verify control SHALL enforce deterministic threshold and submit handoff
OTP verification controls SHALL keep the verify action disabled until a full six-digit code is present and SHALL transition deterministically once a valid code is submitted.

#### Scenario: OTP verify enablement threshold
- **GIVEN** user is on `/otp`
- **WHEN** fewer than six digits are present in the OTP input
- **THEN** `Verify` MUST remain disabled
- **AND** once six digits are present `Verify` MUST become enabled without leaving the route

#### Scenario: OTP verify submit handoff
- **GIVEN** six digits are present in the OTP input
- **WHEN** user submits verification from the OTP form
- **THEN** UI MUST show deterministic in-flight handling
- **AND** MUST complete the happy-path handoff without redirecting back to `/sign-in`

## Acceptance Criteria
- Each onboarding critical step has UC ID and deterministic outcome.
- E2E mapping includes first-run completion and resume behavior.

## Success Criteria
- New users complete onboarding without dead-end states.
- Identity completion state is consistent after restart.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-ONB-01 | First run identity setup | WebAuthn completion required before unlock | planned: `cypress/e2e/auth/onboarding.cy.ts` `first-run-webauthn-gate` |
| UC-ONB-02 | Restart mid-onboarding | Flow resumes last incomplete step | planned: `cypress/e2e/auth/onboarding.cy.ts` `resume-onboarding` |
| UC-ONB-03 | Onboarding data empty | Guided default state appears | planned: `cypress/e2e/auth/onboarding.cy.ts` `onboarding-empty-state` |
| UC-ONB-04 | Auth requirement error | Error + retry shown, no crash | planned: `cypress/e2e/auth/onboarding.cy.ts` `auth-error-retry` |
| UC-ONB-05 | Setup config missing | Full-screen setup wizard shown before auth | planned: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-009 shows full-screen setup wizard before auth when setup config is missing` |
| UC-ONB-06 | Provider actions render | Google, Apple, and Microsoft provider buttons appear deterministically | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-006 renders Google, Apple, and Microsoft provider actions deterministically` |
| UC-ONB-07 | Identity mode resolution | Identity mode and provider enablement resolve from runtime config | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-007 resolves identity mode and provider enablement from runtime config` |
| UC-ONB-08 | Passkey sign-in | Enrolled passkey auth redirects to authenticated shell without password prompt | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-008 signs in with passkey and redirects without password prompt` |
| UC-ONB-09 | Passkey fallback | Unavailable passkey shows deterministic guidance and keeps alternate methods visible | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-008 shows deterministic fallback guidance when passkey is unavailable` |
| UC-ONB-10 | Sign-up completion | Valid sign-up shows submit progress and navigates to authenticated shell on success | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-011 completes sign-up and redirects to authenticated shell` |
| UC-ONB-10A | Sign-in forgot-password entry | Sign-in shows visible forgot-password recovery entry with deterministic keyboard/mouse navigation to `/forgot-password` | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010B exposes deterministic forgot-password entry from sign-in` |
| UC-ONB-10AA | Sign-in GitHub/Facebook actions | Sign-in shows visible GitHub/Facebook actions with deterministic disabled state until provider flows are wired | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010BB renders deterministic sign-in GitHub and Facebook actions` |
| UC-ONB-10AB | Sign-in-2 forgot-password entry | `/sign-in-2` shows visible forgot-password recovery entry with deterministic keyboard/mouse navigation to `/forgot-password` | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010BC exposes deterministic forgot-password entry from sign-in-2` |
| UC-ONB-10AC | Sign-in-2 legal links | `/sign-in-2` shows visible `Terms of Service` and `Privacy Policy` links with deterministic keyboard/mouse navigation to `/terms` and `/privacy` | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010CC exposes deterministic legal links from sign-in-2` |
| UC-ONB-10AD | Sign-in-2 GitHub/Facebook actions | `/sign-in-2` shows visible GitHub/Facebook actions with deterministic disabled state until provider flows are wired | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010CCD renders deterministic sign-in-2 GitHub and Facebook actions` |
| UC-ONB-10B | Sign-in password visibility toggle | Sign-in password field toggles deterministically between masked/text modes with updated accessible state and keyboard activation | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-011B toggles sign-in password visibility deterministically` |
| UC-ONB-10C | Sign-up secondary links | Sign-up shows visible sign-in/legal links with deterministic keyboard/mouse navigation to `/sign-in`, `/terms`, and `/privacy` | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010C exposes deterministic sign-up secondary links` |
| UC-ONB-10D | Sign-up GitHub/Facebook actions | Sign-up shows visible GitHub/Facebook actions with deterministic disabled state until provider flows are wired | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-010D renders deterministic sign-up GitHub and Facebook actions` |
| UC-ONB-11 | Forgot-password completion | Valid forgot-password submit shows progress and navigates to OTP recovery without fallback validation regression | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-012 completes forgot-password submit and routes to OTP recovery` |
| UC-ONB-11C | Sign-up password visibility toggles | Sign-up password + confirm-password fields toggle deterministically between masked/text modes with updated accessible state and keyboard activation | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-011C toggles sign-up password fields deterministically` |
| UC-ONB-11D | Sign-in-2 password visibility toggle | `/sign-in-2` password field toggles deterministically between masked/text modes with updated accessible state and keyboard activation | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-011D toggles sign-in-2 password visibility deterministically` |
| UC-ONB-11B | Forgot-password controls | `/forgot-password` supports keyboard submit with deterministic loading state and keyboard-activatable `/sign-up` secondary handoff | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-012 supports forgot-password keyboard submit and sign-up handoff` |
| UC-ONB-12 | Privacy legal route | `/privacy` renders public Privacy Policy content instead of the 404 screen | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-013 renders Privacy Policy content on the public privacy route` |
| UC-ONB-13 | Terms legal route | `/terms` renders public Terms of Service content instead of the 404 screen | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-013 renders Terms of Service content on the public terms route` |
| UC-ONB-14 | OTP resend flow | `/otp` resend keeps the user in OTP recovery flow with visible resend confirmation instead of redirecting to sign-in | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-014 keeps resend in OTP context with deterministic feedback` |
| UC-ONB-15 | OTP verify control | `/otp` keeps Verify disabled until six digits are present, then submits deterministically from the OTP route | implemented: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-015 enforces OTP verify threshold and submit handoff` |
