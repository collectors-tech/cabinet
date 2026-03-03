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
