## Purpose
Define scanner query set, execution, candidate persistence, and failure handling behavior.

## Requirements
### Requirement: Scanner query sets SHALL support user-defined market criteria
Cabinet SHALL support query set criteria including keywords, exclusions, max price, region, condition, and scheduling metadata.

#### Scenario: Create query set
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user submits valid query set form
- **THEN** Cabinet SHALL persist query set definition

### Requirement: Scanner execution SHALL support manual and scheduled runs with rate limits
Cabinet SHALL support run-now and scheduled execution under rate-limited controls.

#### Scenario: Scheduled scanner run
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** scheduler triggers enabled query set
- **THEN** Cabinet SHALL execute scan with rate-limit policy

### Requirement: Candidate records SHALL preserve listing and stock context
Cabinet SHALL store listing id, pricing, seller, URL, media, first/last seen, status, and stock observations.

#### Scenario: Candidate persistence
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** provider returns listing candidates
- **THEN** Cabinet SHALL persist normalized candidate records with stock state fields

### Requirement: Candidate ingestion SHALL deduplicate via fingerprint
Cabinet SHALL compute a deterministic candidate fingerprint during normalization and SHALL prevent duplicate candidate record creation for the same fingerprint within the same provider/query-set scope.

#### Scenario: Duplicate candidate returned across runs
- **GIVEN** an existing candidate record already exists for the same provider/query-set fingerprint
- **WHEN** scanner ingestion receives a candidate with the same fingerprint in a later run
- **THEN** Cabinet SHALL update `last_seen` and mutable fields on the existing record instead of inserting a new record

### Requirement: Scanner failures SHALL be diagnosable and retryable
Cabinet SHALL log failures and support retry by query set.

#### Scenario: Retry failed scan
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user selects retry for failed query set
- **THEN** Cabinet SHALL schedule immediate retry attempt and log outcome
