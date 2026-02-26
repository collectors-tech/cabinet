## Purpose
Define cloud ownership authentication bootstrap and billing-driven entitlement behavior for Cabinet.

## Requirements
### Requirement CLOUD-AUTH-BILLING-001: Cloud auth billing capability SHALL be in-scope as optional ownership mode
Cabinet SHALL support cloud-account ownership mode alongside local-first authentication model when configured.

#### Scenario: Cloud mode enabled
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** required cloud auth configuration is present
- **THEN** Cabinet SHALL allow cloud session bootstrap and entitlement resolution

### Requirement CLOUD-AUTH-BILLING-002: Frontend cloud session bootstrap SHALL resolve plan and feature entitlements
Frontend SHALL post cloud session token to bootstrap endpoint and SHALL receive plan + feature flags used for UI gating.

#### Scenario: Bootstrap entitlement response
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** frontend calls cloud bootstrap endpoint with valid session token
- **THEN** runtime SHALL return current plan and enabled feature set

### Requirement CLOUD-AUTH-BILLING-003: Billing webhook SHALL update entitlement state with signature verification
Runtime SHALL process billing lifecycle webhook events and SHALL verify webhook signatures before applying entitlement changes.

#### Scenario: Valid billing webhook update
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** signed billing webhook event for subscription state change is received
- **THEN** runtime SHALL update entitlement override for associated user identity

### Requirement CLOUD-AUTH-BILLING-004: Pro-gated features SHALL be controlled by entitlement state
Cabinet SHALL gate pro features (AI assist, price tracking, scanner automation) based on resolved entitlement state.

#### Scenario: Downgrade entitlement effect
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user entitlement changes from pro to free
- **THEN** pro-gated features SHALL be disabled in UI and enforced in runtime checks

### Requirement CLOUD-AUTH-BILLING-005: Production auth claims SHALL be verified
Production cloud auth mode SHALL not rely on unsigned/unchecked token claim parsing for entitlement trust decisions.

#### Scenario: Invalid token presented
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** bootstrap receives invalid or unverifiable cloud token
- **THEN** runtime SHALL reject bootstrap request and return auth failure state

## Acceptance Criteria
1. In-scope decision is explicit: cloud auth billing is optional but supported capability.
2. Bootstrap flow, webhook flow, and gating flow are defined with normative scenarios.
3. Security-critical signature/claim verification requirements are explicit.

## Success Criteria
1. Entitlement changes propagate predictably from billing events to runtime gating behavior.
2. Cloud ownership mode coexists with local-first mode without scope ambiguity.
3. Security posture is explicit and testable for token/webhook verification paths.

## Source Mapping
- Legacy Clerk billing setup runbook has been normalized into this spec.
