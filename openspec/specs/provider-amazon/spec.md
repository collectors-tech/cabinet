## Purpose
Define Amazon provider contract for discovery and pricing ingestion.

## Requirements
### Requirement: Amazon provider SHALL declare integration mode
Cabinet SHALL classify Amazon integration mode (official API, affiliate feed, or web ingestion) with explicit availability status and constraints.

#### Scenario: Resolve Amazon integration mode
- **GIVEN** provider registry is loaded
- **WHEN** Amazon provider metadata is requested
- **THEN** Cabinet SHALL return the declared integration mode and constraint flags

### Requirement: Amazon provider SHALL normalize listing candidates when enabled
Cabinet SHALL normalize Amazon listing payloads into candidate schema when provider mode is enabled for scanning.

#### Scenario: Ingest Amazon candidates
- **GIVEN** Amazon provider mode is enabled and credentials/config are valid
- **WHEN** scanner executes an Amazon-backed query set
- **THEN** Amazon listing payloads SHALL be normalized into candidate records

### Requirement: Amazon provider SHALL expose unsupported-state diagnostics when disabled
Cabinet SHALL return explicit unsupported/disabled diagnostics when Amazon integration mode is unavailable.

#### Scenario: Amazon mode unavailable
- **GIVEN** Amazon integration mode is not configured for active profile
- **WHEN** user attempts Amazon scan execution
- **THEN** Cabinet SHALL return a clear disabled/unsupported provider diagnostic state

