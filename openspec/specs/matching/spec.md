## Purpose
Define matching engine confidence classification behavior.

## Requirements
### Requirement MATCHING-001: Matching engine SHALL classify candidates by confidence state
Cabinet SHALL classify candidates into matched, suggested, or not-in-collection states.

#### Scenario: Matching classification
- **GIVEN** normalized candidates are available
- **WHEN** matching run executes
- **THEN** each candidate SHALL receive a confidence-state classification

### Requirement MATCHING-002: Part-number extraction SHALL feed matching decision
Cabinet SHALL extract candidate part numbers from listing metadata for canonical comparison.

#### Scenario: Part-number extraction
- **GIVEN** listing metadata includes part-number signal
- **WHEN** matching input is prepared
- **THEN** extracted part number SHALL be included in matching input

