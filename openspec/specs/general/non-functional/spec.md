## Purpose
Define measurable runtime performance and reliability constraints for release readiness.

## Requirements
### Requirement NON-FUNCTIONAL-001: Runtime performance SHALL meet v1 thresholds
Cabinet SHALL meet startup, search, and scanner runtime targets defined for v1 with deterministic local validation that does not require live provider credentials.

#### Scenario: Startup benchmark
- **GIVEN** the runtime NFR gate starts Cabinet against a fresh temporary data directory
- **WHEN** startup fast-exit, indexed inventory search, and local scanner-provider execution are measured
- **THEN** startup SHALL remain within the configured fast-exit threshold
- **AND** indexed search SHALL stay within the target latency budget for a seeded 5,000 item dataset
- **AND** scanner execution SHALL complete through a deterministic local provider without live marketplace credentials

### Requirement NON-FUNCTIONAL-002: Reliability SHALL meet beta crash-free objective
Cabinet SHALL target crash-free session rate above 99 percent for beta and keep local diagnostics capable of failing the gate when core startup/search/scanner paths regress.

#### Scenario: Beta reliability assessment
- **GIVEN** beta telemetry is not available in local validation
- **WHEN** the runtime NFR gate exercises startup, search, and scanner diagnostics
- **THEN** the gate SHALL provide deterministic crash/regression evidence for the core beta readiness paths
- **AND** strict startup mode SHALL fail the gate when startup exceeds the configured threshold

### Requirement NON-FUNCTIONAL-003: Develop and release-candidate gates SHALL preserve beta release evidence
Cabinet SHALL run a required quality gate for pull requests targeting `develop` and provide a manually triggered release-candidate gate for an exact commit SHA before beta promotion.

#### Scenario: Develop pull request quality gate
- **GIVEN** a pull request targets `develop`
- **WHEN** GitHub Actions evaluates the repository workflows
- **THEN** the Develop Quality Gate workflow SHALL run strict OpenSpec validation, UI production build, the complete Go repository test suite including root contract tests, OpenAPI parity/lint/build, and the login/profile/runtime Cypress smoke gate
- **AND** failing steps SHALL preserve useful workflow artifacts or command logs for review
- **AND** the workflow SHALL not merge `develop` into `main`

#### Scenario: Exact release-candidate commit gate
- **GIVEN** a full 40-character commit SHA is supplied to the beta release-candidate workflow
- **WHEN** the workflow checks out the requested commit
- **THEN** the checked-out `HEAD` SHALL equal the supplied SHA
- **AND** the working tree SHALL be clean before validation starts
- **AND** strict OpenSpec validation, `go test ./...`, UI production build, OpenAPI lint/build, and the configured Cypress release pack SHALL run against that exact checkout
- **AND** the workflow SHALL upload logs and a summary that identify the commit SHA and workflow run without merging `develop` into `main`
