## 1. Reproduce and isolate

- [x] 1.1 Reproduce the auth-entry failure cluster on a fresh 17882 branch build.
- [x] 1.2 Isolate stale expectations vs real product regressions.

## 2. Contract alignment

- [x] 2.1 Align sign-in/sign-up/forgot-password link navigation expectations with the current routes.
- [x] 2.2 Align passkey success/fallback expectations with the current auth behavior and copy.
- [x] 2.3 Align password-toggle and OTP handoff expectations with the current auth surfaces.

## 3. Validation

- [x] 3.1 Re-run `cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` on a fresh 17882 branch build.
- [ ] 3.2 Feed the result back into the broader regression gate.