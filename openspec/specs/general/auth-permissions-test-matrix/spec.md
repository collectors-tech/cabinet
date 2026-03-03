## Purpose
Define Clerk identity integration and permissions test matrix for multi-account, multi-plan verification.

## Requirements
### Requirement AUTH-PERM-001: Identity provider mode SHALL support Clerk with explicit enable/config state
Runtime SHALL support explicit Clerk integration mode and deterministic fallback/error behavior when Clerk config is incomplete.

#### Scenario: Clerk mode initialization
- **GIVEN** auth mode is configured to Clerk
- **WHEN** app initializes auth stack
- **THEN** sign-in provider list and session resolution MUST use Clerk context deterministically

### Requirement AUTH-PERM-002: Entitlement resolution SHALL map account plan to capability permissions
Plan/subscription levels MUST resolve to explicit capability permissions consumed by API and UI gates.

#### Scenario: Resolve plan capabilities
- **GIVEN** account has active plan (`MVP`, `Creator`, `Teams`)
- **WHEN** permissions are resolved
- **THEN** runtime MUST return deterministic capability set used by feature gates

### Requirement AUTH-PERM-003: Test environment SHALL support multiple seeded accounts across plan levels
System SHALL provide seeded test accounts for each plan/permission level to validate feature gating end-to-end.

#### Scenario: Seed auth test matrix accounts
- **GIVEN** test seed command runs
- **WHEN** account matrix is provisioned
- **THEN** accounts for at least `MVP`, `Creator`, and `Teams` MUST be created with known credentials and expected capabilities

### Requirement AUTH-PERM-004: Feature gates SHALL be validated with account matrix across UI and API
All plan-gated features MUST be verified using matrix accounts through UI and API assertions.

#### Scenario: Validate gated feature by account level
- **GIVEN** feature is plan-gated
- **WHEN** tests run with each seeded account
- **THEN** ineligible accounts MUST be blocked deterministically and eligible accounts MUST succeed

### Requirement AUTH-PERM-005: Permissions audit diagnostics SHALL expose effective role/plan/capabilities for current session
Runtime SHALL expose inspectable effective permissions to simplify debugging and release checks.

#### Scenario: Inspect effective permissions
- **GIVEN** authenticated session exists
- **WHEN** diagnostics endpoint/panel is opened
- **THEN** effective role, plan, and capability set MUST be viewable for current user/session
