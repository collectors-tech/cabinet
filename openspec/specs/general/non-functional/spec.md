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
