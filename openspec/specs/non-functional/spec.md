## Purpose
Define measurable runtime performance and reliability constraints for release readiness.

## Requirements
### Requirement NON-FUNCTIONAL-001: Runtime performance SHALL meet v1 thresholds
Cabinet SHALL meet startup, search, and scanner runtime targets defined for v1.

#### Scenario: Startup benchmark
- **GIVEN** benchmark executes on baseline hardware
- **WHEN** startup benchmark run is measured
- **THEN** startup SHALL complete within target threshold

### Requirement NON-FUNCTIONAL-002: Reliability SHALL meet beta crash-free objective
Cabinet SHALL target crash-free session rate above 99 percent for beta.

#### Scenario: Beta reliability assessment
- **GIVEN** beta telemetry and diagnostics are available
- **WHEN** reliability assessment runs
- **THEN** crash-free sessions SHALL meet objective
