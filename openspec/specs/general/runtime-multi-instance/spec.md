## Purpose
Define high-density runtime orchestration requirements for launching and managing up to 100 concurrent Cabinet instances for stress and compatibility testing.

## Requirements
### Requirement RUNTIME-MULTI-001: Runtime orchestrator MUST support launching 100 parallel instances in test mode
Cabinet tooling SHALL provide automated launch orchestration for up to 100 concurrent instances using isolated runtime contexts.

#### Scenario: Launch 100 instances
- **GIVEN** stress harness is executed with `Count=100`
- **WHEN** instance launch completes
- **THEN** exactly 100 instance records MUST be emitted with startup health status

### Requirement RUNTIME-MULTI-002: Port allocation MUST be deterministic and collision-safe
Each launched instance SHALL receive a unique host/port combination derived deterministically from orchestrator inputs.

#### Scenario: Deterministic port assignment
- **GIVEN** base port and count are fixed inputs
- **WHEN** orchestrator assigns ports
- **THEN** each instance MUST have a unique reachable URL

### Requirement RUNTIME-MULTI-003: Per-instance metadata MUST include identity and runtime context
Each instance record SHALL include instance name, PID, URL, port, data directory, and runtime metadata snapshot.

#### Scenario: Metadata manifest output
- **GIVEN** stress harness completes launch
- **WHEN** instance manifest is written
- **THEN** each entry MUST include `instance_name`, `pid`, `url`, `port`, `data_dir`, and `healthy`

### Requirement RUNTIME-MULTI-004: Data/profile isolation MUST be enforced per instance
Parallel instances SHALL run with distinct runtime data directories and lock files to prevent cross-instance overwrite.

#### Scenario: Isolated data roots
- **GIVEN** orchestrator launches parallel instances
- **WHEN** each process starts
- **THEN** each instance MUST use a unique `CABINET_DATA_DIR`

### Requirement RUNTIME-MULTI-005: Orchestrator MUST expose deterministic start/stop/status behavior
Stress harness SHALL provide deterministic startup, health verification, and shutdown semantics for full instance set.

#### Scenario: Full lifecycle orchestration
- **GIVEN** launch sequence begins
- **WHEN** process lifecycle completes
- **THEN** harness MUST stop all spawned processes in teardown and emit lifecycle summary

### Requirement RUNTIME-MULTI-006: Health/discovery view MUST be emitted for all active instances
Harness output SHALL include aggregate health/discovery data for all launched instances.

#### Scenario: Health discovery output
- **GIVEN** instances are launched
- **WHEN** health checks run
- **THEN** report MUST include healthy/failed counts and failed instance identifiers

### Requirement RUNTIME-MULTI-007: Resource guardrails MUST apply adaptive backoff
Orchestration SHALL enforce memory/CPU guardrails and apply bounded backoff during bulk launch when thresholds are exceeded.

#### Scenario: Guardrail backoff
- **GIVEN** aggregate process resource metrics exceed configured thresholds
- **WHEN** next launch cycle executes
- **THEN** orchestrator MUST apply configured backoff delay before additional launches

### Requirement RUNTIME-MULTI-008: Scale reports MUST persist reproducible evidence artifacts
Stress execution SHALL persist machine-readable manifest + markdown report with acceptance snapshot.

#### Scenario: Scale report generation
- **GIVEN** stress run completes
- **WHEN** report artifacts are written
- **THEN** report MUST include requested count, healthy count, failed count, duration, and artifact paths
