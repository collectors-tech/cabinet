## Purpose
Define continuous, scheduled, executable UI validation for Cabinet including release-aware runtime checks and exhaustive user-action verification.

## Requirements
### Requirement CONT-UI-CAB-001: Cabinet hourly validation SHALL be release-aware
Hourly validation run SHALL compare active build/version against latest available build and skip full run when unchanged.

#### Scenario: No new build available
- **GIVEN** hourly validation starts
- **WHEN** latest build/version matches current validated build
- **THEN** runner SHALL record `no-change` status and wait for next schedule

### Requirement CONT-UI-CAB-002: Cabinet hourly validation SHALL execute real browser interactions
Validation run MUST open live app and execute real user actions (no synthetic-only claims).

#### Scenario: Exhaustive action execution
- **GIVEN** a runnable Cabinet instance is available
- **WHEN** validation run executes
- **THEN** every user-clickable action on audited screens MUST be interacted with and recorded pass/fail

### Requirement CONT-UI-CAB-003: Validation failures SHALL create/update issues with reproducible evidence
Failed actions MUST be linked to focused issues with expected vs actual behavior and reproduction details.

#### Scenario: Action failure recorded
- **GIVEN** any action fails during run
- **WHEN** run completes
- **THEN** issue(s) SHALL include screen, action, error text, and evidence path

### Requirement CONT-UI-CAB-004: Scheduled validation SHALL maintain OpenSpec + commit traceability
Validation outputs MUST update OpenSpec/traceability when contracts are missing and commit changes with issue-linked messages.

### Requirement CONT-UI-CAB-005: Validation SHALL verify control intent outcomes, not click-only execution
For each interactive control, validation SHALL assert intended outcome (route/state/data/error feedback), not merely that click action executed.

#### Scenario: Intent verification per control
- **GIVEN** an interactive control is discovered in audit inventory
- **WHEN** control is executed
- **THEN** run output MUST record expected vs actual intent outcome and pass/fail status

### Requirement CONT-UI-CAB-006: Validation SHALL include full form-field behavior contract checks
Validation SHALL enumerate and test form fields (required/optional, valid/invalid input handling, error messages, submit/save behavior, keyboard accessibility).

#### Scenario: Form field validation contract
- **GIVEN** a form is present on audited screen
- **WHEN** field interactions and submissions are executed
- **THEN** run output MUST include field-level validation outcomes and any failures/backlog issue links

### Requirement CONT-UI-CAB-007: Validator exploration runtime SHALL preflight required Node dependencies
Before running exploratory UI scripts, validator runtime SHALL ensure required Node dependencies (including `playwright`) are installed and importable.

#### Scenario: Exploration preflight checks dependencies
- **GIVEN** validator is about to run `node scripts/cabinet_ui_cycle.mjs`
- **WHEN** dependency preflight executes
- **THEN** missing dependencies SHALL be installed or reported with actionable remediation prior to exploration execution

#### Scenario: Spec gap discovered
- **GIVEN** action exists without explicit requirement coverage
- **WHEN** validation run reconciles action-to-spec mapping
- **THEN** spec IDs SHALL be added append-only and committed with linked issue reference

### Requirement CONT-UI-CAB-008: Cypress validation SHALL have a project-local container image path
Cabinet SHALL provide a repo-local container image definition for isolated Cypress runtime lanes so browser validation can start from a known build artifact instead of a stale shared desktop runtime.

#### Scenario: Build image from source
- **GIVEN** the Cabinet repository is checked out at a known commit
- **WHEN** the container image is built from the repository root
- **THEN** the image build MUST install UI dependencies, build the static UI bundle, compile the Go Cabinet runtime, and copy the UI bundle into the Go build context before producing the runtime image.

#### Scenario: Run isolated validation runtime
- **GIVEN** a Cypress lane starts the Cabinet container image
- **WHEN** the container runtime launches
- **THEN** it MUST disable browser auto-open, expose a deterministic HTTP port, use a mounted `/data` directory, set an E2E profile and instance name, and allow parallel runtime execution for lane isolation.

#### Scenario: Report runtime health to lane orchestration
- **GIVEN** a Cypress lane starts the Cabinet container image
- **WHEN** the container runtime is waiting for readiness
- **THEN** the image MUST expose a health check against `/healthz` so orchestration can distinguish app readiness from Cypress assertion failures.
