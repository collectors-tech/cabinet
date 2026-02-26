## ADDED Requirements

### Requirement: Dashboard SHALL prioritize actionable attention signals
The system SHALL present a primary attention panel containing actionable discovery, pricing, and health signals.

#### Scenario: Attention panel render
- **WHEN** a user opens dashboard
- **THEN** the panel SHALL show highest-priority actionable items with direct actions

#### Scenario: Dashboard action routing
- **WHEN** a user selects an attention action
- **THEN** the system SHALL navigate to the correct destination with relevant context

### Requirement: Dashboard SHALL avoid placeholder and redundant content
The system SHALL avoid template filler text and duplicate control strips that do not add operational value.

#### Scenario: Production dashboard content
- **WHEN** dashboard loads in production mode
- **THEN** all visible cards and labels SHALL correspond to real runtime data or explicit empty states
