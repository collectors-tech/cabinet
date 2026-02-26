## Purpose
Define deterministic UI scale/performance requirements, data profiles, and success thresholds so large datasets remain usable and testable.

## Requirements
### Requirement: UI scalability validation SHALL use deterministic dataset profiles
Cabinet SHALL validate UI behavior using deterministic dataset profiles S0, S1, S2, and S3 with seed-driven reproducibility.

#### Scenario: Reproduce S2 benchmark
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** a benchmark run uses profile `S2` with a fixed seed
- **THEN** generated dataset and performance outputs SHALL be reproducible for the same seed inputs

### Requirement: Data profiles SHALL define sample and bulk coverage for all table-heavy screens
Scalability profiles SHALL include realistic distribution for inventory, discovery, pricing, media, and wishlist data.

#### Scenario: Run table-heavy screen checks on S3
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** UI tests execute on S3
- **THEN** inventory/discovery/pricing/reports workflows SHALL remain operational without route-crashing

### Requirement: Screen-level interaction targets SHALL be measurable and enforceable
Key UI interactions SHALL have defined threshold targets for S2 and non-crash constraints for S3.

#### Scenario: Validate S2 interaction thresholds
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** S2 interaction benchmark executes
- **THEN** initial render, navigation, search, sort, and details-open medians SHALL remain within target limits

### Requirement: Scale validation SHALL include high-volume action workflows
Scale testing SHALL include rapid search/filter loops, discovery action throughput, and report export behavior under high data volume.

#### Scenario: Discovery action throughput on S3
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user performs repeated discovery actions under S3 load
- **THEN** actions SHALL complete within bounded latency and SHALL not cause unrecoverable UI stalls

### Requirement: Scale validation SHALL enforce UI resilience states
Large data operations SHALL still expose deterministic loading, empty, error, and ready states.

#### Scenario: S3 delayed response handling
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** large list API response is delayed
- **THEN** UI SHALL show loading state and remain responsive until transition to ready or error state

## Acceptance Criteria
1. Dataset profile definitions exist for S0/S1/S2/S3 with deterministic generation inputs.
2. Sample profile and bulk profile requirements are explicit and screen-mapped.
3. Interaction thresholds are defined and testable for S2.
4. S3 constraints include no-crash and bounded-latency requirements for critical actions.
5. Scalability scenarios include inventory, discover, dashboard refresh, and reports export flows.

## Success Criteria
1. S2 passes all target metrics:
   - initial render <= 1000ms
   - navigation median <= 150ms
   - search median <= 300ms
   - sort <= 250ms
   - details open <= 120ms
2. S3 runs complete without crashes or unrecoverable stalls.
3. Memory growth remains bounded during sustained navigation/interaction soak.

## Data Profiles
### S0 Empty
- 0 items
- 0 candidates
- 0 photos
Use for first-run empty-state and onboarding behavior validation.

### S1 Starter (sample profile)
- 100 items
- 200 instances
- 300 photos
- 150 barcodes
- 50 discovery candidates

### S2 Growth (primary performance profile)
- 5,000 items
- 15,000 instances
- 20,000 photos
- 8,000 barcodes
- 2,000 discovery candidates
- 1,000 wishlist entries
- 12 months daily pricing history

### S3 Stress (bulk profile)
- 25,000 items
- 80,000 instances
- 150,000 photo metadata rows
- 40,000 barcodes
- 10,000 discovery candidates
- 5,000 wishlist entries
- 24 months pricing history

## Required Distribution Rules
- Brand cardinality >= 50
- Category cardinality >= 30
- Tag cardinality >= 200
- Part number uniqueness > 99%
- Realistic status distribution across sealed/blister/loose/custom/on_track

## E2E and Benchmark Mapping Requirements
Required scenario IDs:
- `SCAL-001` inventory rapid search loop on S2
- `SCAL-002` filter/sort/details loop on S3
- `SCAL-003` home attention refresh under high event volume
- `SCAL-004` discovery triage actions on 10k candidates
- `SCAL-005` reports load/export on 24-month history

Each scenario SHALL map to:
- benchmark/test harness command
- dataset profile
- pass/fail threshold criteria

## Source Mapping
- Scale and performance constraints are canonically captured in this spec.
