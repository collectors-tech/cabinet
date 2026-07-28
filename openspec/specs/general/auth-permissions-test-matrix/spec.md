## Purpose
Define ZITADEL identity integration and permissions test matrix for multi-account, multi-plan verification.

## Requirements
### Requirement AUTH-PERM-001: Identity provider mode SHALL support local and ZITADEL with retired Clerk ignored
Runtime SHALL support explicit local and ZITADEL integration modes and deterministic fallback/error behavior when retired Clerk config is present.

#### Scenario: ZITADEL mode initialization
- **GIVEN** auth mode is configured to ZITADEL
- **WHEN** app initializes auth stack
- **THEN** sign-in provider list and session resolution MUST use ZITADEL context deterministically
- **AND** retired Clerk environment values MUST NOT select or report an active provider

### Requirement AUTH-PERM-002: Entitlement resolution SHALL map account plan to capability permissions
Plan/subscription levels MUST resolve to explicit capability permissions consumed by API and UI gates.

#### Scenario: Resolve plan capabilities
- **GIVEN** cloud bootstrap is called with a ZITADEL token containing `plan` (`mvp`, `creator`, or `teams`)
- **WHEN** `/api/auth/cloud/session/bootstrap` succeeds with `200`
- **THEN** response `features` MUST resolve deterministically as:
  - `mvp` -> `["collection_core"]`
  - `creator` -> `["collection_core","ai_assist","scanner_automation"]`
  - `teams` -> `["collection_core","ai_assist","price_tracking","scanner_automation"]`

### Requirement AUTH-PERM-003: Test environment SHALL support multiple seeded accounts across plan levels
System SHALL provide seeded test accounts for each plan/permission level to validate feature gating end-to-end.

#### Scenario: Seed auth test matrix accounts
- **GIVEN** test seed command runs
- **WHEN** account matrix is provisioned
- **THEN** accounts for at least `MVP`, `Creator`, and `Teams` MUST be created with known credentials and expected capabilities

### Requirement AUTH-PERM-004: Feature gates SHALL be validated with account matrix across UI and API
All plan-gated features MUST be verified using matrix accounts through UI and API assertions.

#### Scenario: Validate gated feature by account level
- **GIVEN** `cloud.plan` is resolved for the active session
- **WHEN** feature gate checks run for `scanner_automation`, `price_tracking`, and `ai_assist`
- **THEN** ineligible plans MUST be denied and eligible plans MUST be allowed deterministically
- **AND** API gates MUST evaluate the same capability mapping used by UI permission payloads

### Requirement AUTH-PERM-005: Permissions audit diagnostics SHALL expose effective role/plan/capabilities for current session
Runtime SHALL expose inspectable effective permissions to simplify debugging and release checks.

#### Scenario: Inspect effective permissions
- **GIVEN** cloud session bootstrap has persisted current user auth state
- **WHEN** `/api/auth/cloud/session/effective` is requested with `GET`
- **THEN** response `200` MUST include `provider`, `user_id`, `role`, `plan`, and `features`
