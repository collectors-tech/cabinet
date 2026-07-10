## MODIFIED Requirements

### Requirement: Data/profile isolation MUST be enforced per instance under explicit parallel mode
Parallel instances SHALL run with distinct runtime data directories and lock files to prevent cross-instance overwrite, and multi-instance behavior SHALL only occur under an explicit parallel/start-orchestration path.

#### Scenario: Isolated data roots for orchestrated parallel launch
- **GIVEN** orchestrator launches parallel instances in explicit multi-instance mode
- **WHEN** each process starts
- **THEN** each instance MUST use a unique `CABINET_DATA_DIR`
- **AND** each instance MUST use distinct runtime lock/lifecycle files under that data root

#### Scenario: Normal desktop launch does not imply multi-instance mode
- **GIVEN** a normal Cabinet desktop launch requests a host/port already serving a healthy Cabinet runtime
- **WHEN** explicit parallel mode is not enabled
- **THEN** the launch MUST follow runtime-core singleton attach behavior rather than creating another parallel runtime

#### Scenario: Explicit restart does not count as multi-instance mode
- **GIVEN** a Cabinet launch requests restart of an already-running healthy Cabinet endpoint
- **WHEN** explicit parallel mode is not enabled
- **THEN** the restart flow MUST replace the running endpoint instance rather than creating a second concurrent runtime
