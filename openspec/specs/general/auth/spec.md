## Purpose
Define local identity authentication, WebAuthn, locking, and recovery behavior.

## Requirements
### Requirement AUTH-001: WebAuthn SHALL be required for first-run local identity
Cabinet SHALL require creation of at least one WebAuthn credential before full workspace access.

#### Scenario: First credential registration
- **GIVEN** a new profile is in first-run state
- **WHEN** user completes setup
- **THEN** successful WebAuthn registration SHALL be required before workspace unlock

### Requirement AUTH-002: Cabinet SHALL enforce session locking behavior
Cabinet SHALL lock at startup and SHALL support auto-lock after inactivity.

#### Scenario: Startup lock gate
- **GIVEN** an existing profile is present
- **WHEN** app starts
- **THEN** protected routes SHALL remain blocked until unlock succeeds

### Requirement AUTH-003: Recovery SHALL exist as optional fallback
Cabinet SHALL support recovery passphrase reset flow when usable authenticators are unavailable.

#### Scenario: Recovery reset begin
- **GIVEN** profile has no valid authenticator and a valid passphrase is provided
- **WHEN** user begins recovery reset
- **THEN** Cabinet SHALL issue recovery reset state for credential replacement

