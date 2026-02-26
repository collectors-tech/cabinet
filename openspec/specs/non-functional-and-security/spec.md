## Purpose
Define performance, reliability, and security/privacy constraints.

## Requirements
### Requirement: Runtime performance SHALL meet v1 thresholds
Cabinet SHALL meet startup, search, and scanner runtime targets defined for v1.

#### Scenario: Startup benchmark
- **WHEN** benchmark run executes on baseline hardware
- **THEN** startup SHALL complete within target threshold

### Requirement: Reliability SHALL meet beta crash-free objective
Cabinet SHALL target crash-free session rate above 99 percent for beta.

#### Scenario: Beta reliability assessment
- **WHEN** beta telemetry and diagnostics are evaluated
- **THEN** crash-free sessions SHALL meet reliability objective

### Requirement: Secrets SHALL never be stored in plaintext SQLite records
Cabinet SHALL store sensitive keys in OS-backed secure storage.

#### Scenario: Secret persistence
- **WHEN** API key is saved for profile
- **THEN** plaintext secret SHALL not be persisted in SQLite data tables

### Requirement: License verification SHALL function offline
Cabinet SHALL verify license state without requiring cloud access.

#### Scenario: Offline license check
- **WHEN** runtime is offline
- **THEN** existing license SHALL validate using local verification path
