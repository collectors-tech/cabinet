## Why

The public auth-entry Cypress suite drifted away from the current sign-in/sign-up/forgot-password/OTP surfaces. After the home-shell/setup route fixes in #446, the remaining failures were all stale auth-entry expectations rather than product regressions.

## What Changes

- Align the onboarding auth-entry contracts with the current public auth routes and copy.
- Keep keyboard/click/link tests deterministic against the actual auth UI behavior.
- Preserve regression coverage for passkey fallback, password toggles, forgot-password routing, legal links, and OTP handoff.

## Capabilities

### Modified Capabilities

- `onboarding-auth-entry`: public auth entry routes SHALL expose deterministic navigation and state transitions matching the current Cabinet auth UI.

## Impact

- Affected tests: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`
- Related issues: `#451`