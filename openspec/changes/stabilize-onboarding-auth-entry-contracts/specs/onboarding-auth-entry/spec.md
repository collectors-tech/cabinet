## MODIFIED Requirements

### Requirement: Public onboarding auth entry routes SHALL remain deterministic
Cabinet SHALL keep the public onboarding auth routes (`/sign-in`, `/sign-in-2`, `/sign-up`, `/forgot-password`, `/otp`) behaviorally aligned with the visible UI and navigation affordances they expose.

#### Scenario: Secondary auth links navigate deterministically
- **GIVEN** a user is on a public auth route
- **WHEN** they activate the visible secondary links for sign-in, sign-up, forgot-password, privacy, or terms
- **THEN** Cabinet SHALL navigate to the linked public route and expose the current page copy for that destination

#### Scenario: Passkey and OTP flows match current auth behavior
- **GIVEN** a user uses passkey or OTP auth flows from the public auth surfaces
- **WHEN** the current Cabinet auth implementation completes or falls back
- **THEN** the Cypress contract SHALL assert the real resulting route and fallback guidance rather than outdated auth-shell assumptions