## Purpose
Define onboarding and auth screen behavior for first-run identity and workspace unlock.

## Requirements
### Requirement: Onboarding/Auth SHALL enforce WebAuthn-first completion
Onboarding/Auth SHALL block advanced workspace access until required identity steps are complete.

#### Scenario: Required identity gate
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user has incomplete identity setup
- **THEN** advanced workspace SHALL remain locked

### Requirement: Onboarding/Auth SHALL persist and resume progress
Progress SHALL persist through reload/restart until completion.

#### Scenario: Resume incomplete onboarding
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user returns after restart
- **THEN** onboarding SHALL resume at last incomplete step

### Requirement: Onboarding/Auth SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states for profile and auth checks.

#### Scenario: Auth check error
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** auth requirements fetch fails
- **THEN** screen SHALL show actionable error with retry

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
