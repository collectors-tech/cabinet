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

#### Scenario: Spec gap discovered
- **GIVEN** action exists without explicit requirement coverage
- **WHEN** validation run reconciles action-to-spec mapping
- **THEN** spec IDs SHALL be added append-only and committed with linked issue reference
