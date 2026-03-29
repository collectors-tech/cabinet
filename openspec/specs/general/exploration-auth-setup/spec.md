## Purpose
Define deterministic exploratory auth setup guidance for Cabinet so route audits and UI review do not stall on avoidable local-vs-Clerk ambiguity.

## Requirements
### Requirement EXPLORATION-AUTH-SETUP-001: Exploratory auth guidance SHALL define local mode as the default route-audit path
Cabinet documentation SHALL define a deterministic default exploratory auth path that prefers local mode for routine route audits and general UI review.

#### Scenario: Choose default exploratory auth mode
- **GIVEN** a developer or reviewer is starting exploratory work not explicitly scoped to Clerk behavior
- **WHEN** they consult the exploratory auth setup guide
- **THEN** the guide MUST direct them to use local auth mode by default
- **AND** the guide MUST include a runnable startup path and example local sign-in expectations
- **AND** the guide MUST identify the preferred repo-level local exploration launcher

### Requirement EXPLORATION-AUTH-SETUP-002: Exploratory auth guidance SHALL distinguish Clerk-specific prerequisites and blockers
Cabinet documentation SHALL distinguish Clerk-only auth prerequisites from the default local exploratory path so Clerk configuration gaps are not confused with general product failures.

#### Scenario: Resolve Clerk exploratory prerequisites
- **GIVEN** exploratory work explicitly targets Clerk flows or permissions behavior
- **WHEN** the reviewer follows the exploratory auth setup guide
- **THEN** the guide MUST enumerate Clerk publishable-key and auth-mode prerequisites
- **AND** the guide MUST identify common Clerk/domain/origin blockers with actionable next steps
- **AND** the guide MUST include a concrete repo-level Clerk startup example and verification target

### Requirement EXPLORATION-AUTH-SETUP-003: Exploratory auth guidance SHALL document sample-data/bootstrap paths
Cabinet documentation SHALL identify how exploratory sessions obtain authenticated sample data for route traversal.

#### Scenario: Choose authenticated sample-data path
- **GIVEN** a reviewer needs representative authenticated data after sign-in
- **WHEN** they follow the exploratory auth setup guide
- **THEN** the guide MUST identify the starter-data path and Showcase DB profile path
- **AND** the guide MUST state when each path should be preferred
- **AND** the guide MUST identify the concrete first-sign-in starter-data path for local exploratory sessions

### Requirement EXPLORATION-AUTH-SETUP-005: Exploratory auth guidance SHALL define local account bootstrap expectations
Cabinet documentation SHALL define the expected local exploratory account behavior so reviewers do not wait for a separately provisioned account before route testing.

#### Scenario: Bootstrap local exploratory account on first sign-in
- **GIVEN** the reviewer is using local auth mode for exploratory work
- **WHEN** they reach the sign-in form on a fresh local runtime
- **THEN** the guide MUST state that no pre-seeded local account is required
- **AND** the guide MUST state that first successful local sign-in is the expected bootstrap path
- **AND** the guide MUST provide example local exploratory credentials that satisfy current validation constraints

### Requirement EXPLORATION-AUTH-SETUP-006: Exploratory auth guidance SHALL provide a repo-level Clerk launcher contract
Cabinet documentation and scripts SHALL provide a deterministic Clerk exploration launcher so reviewers can start a Clerk-oriented runtime with explicit env expectations and verification targets.

#### Scenario: Start Clerk exploration runtime from repo launcher
- **GIVEN** the reviewer needs Clerk-specific exploratory auth coverage
- **WHEN** they run the documented Clerk exploration launcher with a publishable key
- **THEN** the launcher MUST set Clerk-specific env expectations for the child build/runtime process
- **AND** the guide MUST identify the expected runtime URL and verification endpoints
- **AND** the launcher MUST fail fast with actionable guidance when the Clerk publishable key is missing

### Requirement EXPLORATION-AUTH-SETUP-004: Exploratory auth guidance SHALL normalize passkey domain mismatch diagnosis
Cabinet documentation SHALL describe passkey domain/origin mismatch as an auth-environment setup problem and SHALL direct exploratory users toward deterministic fallback behavior.

#### Scenario: Diagnose passkey invalid-domain failure during exploration
- **GIVEN** passkey sign-in fails with an invalid-domain, origin-mismatch, or relying-party mismatch symptom
- **WHEN** the reviewer consults the exploratory auth setup guide
- **THEN** the guide MUST describe the failure as an environment/domain setup issue rather than a generic route failure
- **AND** the guide MUST direct the reviewer to continue with password/provider auth unless the task explicitly requires passkey validation
