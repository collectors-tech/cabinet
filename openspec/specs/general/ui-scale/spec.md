## Purpose
Define dataset-scale behavior requirements for UI workflows.

## Requirements
### Requirement UI-SCALE-001: UI scalability validation SHALL use deterministic dataset profiles
Cabinet SHALL validate UI behavior with deterministic S0/S1/S2/S3 data profiles.

#### Scenario: Reproduce S2 benchmark
- **GIVEN** profile S2 and fixed seed are selected
- **WHEN** scale benchmark runs
- **THEN** generated dataset and outputs SHALL be reproducible for same seed

### Requirement UI-SCALE-002: Data profiles SHALL define sample and bulk coverage for table-heavy screens
Scale profiles SHALL include realistic distributions for inventory, discovery, pricing, media, and wishlist.

#### Scenario: Run table-heavy checks on S3
- **GIVEN** S3 profile is loaded
- **WHEN** table-heavy workflows execute
- **THEN** workflows SHALL remain operational without route crashes

### Requirement UI-SCALE-003: Scale validation SHALL include high-volume action workflows
Scale testing SHALL include rapid search/filter loops, discovery throughput, and report export under high volume.

#### Scenario: Discovery action throughput on S3
- **GIVEN** S3 load is active
- **WHEN** repeated discovery actions run
- **THEN** actions SHALL complete within bounded latency without unrecoverable stalls

