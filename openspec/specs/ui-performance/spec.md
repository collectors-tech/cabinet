## Purpose
Define measurable UI performance thresholds and resilience-state contracts.

## Requirements
### Requirement UI-PERFORMANCE-001: Screen-level interaction targets SHALL be measurable and enforceable
Key UI interactions SHALL have defined threshold targets for S2 and non-crash constraints for S3.

#### Scenario: Validate S2 interaction thresholds
- **GIVEN** S2 benchmark executes
- **WHEN** render/navigation/search/sort/details timings are measured
- **THEN** medians SHALL remain within target limits

### Requirement UI-PERFORMANCE-002: Scale validation SHALL enforce UI resilience states
Large-data operations SHALL expose deterministic loading, empty, error, and ready states.

#### Scenario: S3 delayed response handling
- **GIVEN** large list API response is delayed
- **WHEN** UI waits for response
- **THEN** UI SHALL remain responsive with loading state until ready/error transition

