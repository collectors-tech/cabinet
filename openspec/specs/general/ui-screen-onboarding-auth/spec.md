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
