## Purpose
Define candidate matching and not-in-collection triage behavior.

## Requirements
### Requirement: Matching engine SHALL classify candidates by confidence state
Cabinet SHALL classify candidates into matched, suggested, or not-in-collection states.

#### Scenario: Matching classification
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** matching run processes candidates
- **THEN** each candidate SHALL receive a confidence state classification

### Requirement: Part-number extraction SHALL feed matching decision
Cabinet SHALL extract candidate part numbers from listing metadata for comparison against canonical records.

#### Scenario: Part-number extraction
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** listing title or metadata includes part number signal
- **THEN** matching input SHALL include extracted part number candidate

### Requirement: Not-in-collection panel SHALL support actionable triage
Cabinet SHALL support ignore, add-to-wishlist, track-price, and create-item actions.

#### Scenario: Discovery triage action
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user applies discovery action
- **THEN** Cabinet SHALL persist requested action outcome

### Requirement: Discovery filters SHALL support price/query/date
Cabinet SHALL provide filtering for discovery queue triage.

#### Scenario: Discovery filtered view
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user applies query and date filter
- **THEN** panel SHALL return filtered not-in-collection candidates
