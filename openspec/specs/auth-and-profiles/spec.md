## Purpose
Define local identity, WebAuthn, session lock, recovery, and profile isolation behavior.

## Requirements
### Requirement: WebAuthn SHALL be required for first-run local identity
Cabinet SHALL require creation of at least one WebAuthn credential before full workspace access.

#### Scenario: First credential registration
- **WHEN** a new profile completes first launch setup
- **THEN** Cabinet SHALL require successful WebAuthn registration

### Requirement: Cabinet SHALL enforce session locking behavior
Cabinet SHALL lock at startup and SHALL support auto-lock after inactivity.

#### Scenario: Startup lock gate
- **WHEN** the app starts with an existing profile
- **THEN** protected routes SHALL remain blocked until unlock succeeds

### Requirement: Cabinet SHALL support profile-isolated storage and secrets
Each profile SHALL have isolated database, settings, API keys, and license state.

#### Scenario: Profile isolation
- **WHEN** user switches from profile A to profile B
- **THEN** profile A records SHALL not appear in profile B views

### Requirement: Recovery SHALL exist as optional fallback
Cabinet SHALL support recovery passphrase reset flow when usable authenticators are unavailable.

#### Scenario: Recovery reset begin
- **WHEN** profile has no valid authenticator and valid passphrase is provided
- **THEN** Cabinet SHALL issue recovery reset state for credential replacement
