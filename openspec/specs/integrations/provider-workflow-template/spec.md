## Purpose
Define a repeatable development workflow contract used for each integration provider.

## Requirements
### Requirement PROVIDER-WORKFLOW-001: Each provider SHALL implement the standard workflow stages
Each provider implementation MUST cover: connect/auth, validate token, search/query, candidate mapping, apply/import.

#### Scenario: Provider workflow completeness check
- **GIVEN** provider is marked active for delivery
- **WHEN** workflow checklist is reviewed
- **THEN** all five workflow stages MUST be implemented and test-mapped

### Requirement PROVIDER-WORKFLOW-002: Each provider SHALL have deterministic failure handling
Provider workflows MUST define stable error behavior for auth failure, empty results, rate limiting, and upstream failures.

#### Scenario: Provider error handling
- **GIVEN** provider request fails
- **WHEN** runtime handles failure
- **THEN** response MUST return deterministic error code and actionable remediation guidance

### Requirement PROVIDER-WORKFLOW-003: Each provider SHALL support mock and live validation lanes
Provider workflows MUST include mock-mode validation for CI/dev and live-mode verification for real integration confidence.

#### Scenario: Mock and live parity
- **GIVEN** provider workflow test suite runs in mock mode and live mode
- **WHEN** contract outputs are compared
- **THEN** normalized candidate/output schema MUST remain consistent across lanes

### Requirement PROVIDER-WORKFLOW-004: Provider setup MUST have step-by-step token/setup documentation
Token-based providers MUST include user setup docs for app creation, scopes, token retrieval, validation, and troubleshooting.

#### Scenario: User setup from docs
- **GIVEN** user follows provider setup guide
- **WHEN** user configures provider and runs validate-token
- **THEN** setup MUST complete without undocumented steps
