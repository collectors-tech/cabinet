## ADDED Requirements

### Requirement: SETUP-WIZ-022 Initial setup mutations SHALL use a temporary bootstrap boundary
Cabinet SHALL allow only the setup completion, config import and storage
validation mutations to cross the authentication boundary without a session
while initial runtime setup is required. The normal request and session
boundaries MUST apply again after setup exists.

#### Scenario: Incomplete remote identity configuration does not dead-end first-run completion
- **GIVEN** `cabinet.json` is missing, setup status requires the wizard and remote identity mode is selected but not fully configured
- **WHEN** a trusted Cabinet client sends `POST /api/runtime/setup-complete`
- **THEN** the request MUST reach setup payload validation without requiring an authenticated session
- **AND** a valid payload MUST be able to create the initial configuration

#### Scenario: Config import and storage validation remain usable before sign-in
- **GIVEN** `cabinet.json` is missing and the setup wizard is active before sign-in
- **WHEN** a trusted Cabinet client sends `POST /api/runtime/setup-import` or `POST /api/runtime/setup-storage-validate`
- **THEN** the request MUST reach the corresponding setup handler without requiring an authenticated session
- **AND** the handler MUST preserve its normal validation and deterministic error contract

#### Scenario: Bootstrap exception closes after setup
- **GIVEN** a valid `cabinet.json` exists
- **WHEN** a client sends one of the setup mutation requests
- **THEN** Cabinet MUST NOT apply the initial-setup authentication exception
- **AND** the request MUST follow the normal configured local or remote session policy

#### Scenario: Bootstrap allowlist remains method and path constrained
- **GIVEN** initial setup is required
- **WHEN** a request uses another API path or a non-POST method for a setup mutation path
- **THEN** Cabinet MUST NOT apply the initial-setup authentication exception

#### Scenario: Request boundary remains enforced during bootstrap
- **GIVEN** initial setup is required
- **WHEN** a setup mutation arrives from an untrusted host or a cross-site origin
- **THEN** Cabinet MUST reject the request before setup handling
- **AND** no configuration mutation may be persisted
