## Purpose
Define not-in-collection discovery triage behavior.

## Requirements
### Requirement: Not-in-collection panel SHALL support actionable triage
Cabinet SHALL support ignore, add-to-wishlist, track-price, and create-item actions.

#### Scenario: Discovery triage action
- **GIVEN** a candidate is in not-in-collection state
- **WHEN** user applies a triage action
- **THEN** Cabinet SHALL persist the requested action outcome

### Requirement: Discovery filters SHALL support price/query/date
Cabinet SHALL provide filtering for discovery queue triage.

#### Scenario: Discovery filtered view
- **GIVEN** discovery queue has candidates
- **WHEN** user applies query and date filters
- **THEN** panel SHALL return filtered not-in-collection candidates

