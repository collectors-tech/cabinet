## Purpose
Define candidate persistence and deduplication behavior.

## Requirements
### Requirement CANDIDATES-001: Candidate records SHALL preserve listing and stock context
Cabinet SHALL store listing id, pricing, seller, URL, media, first/last seen, status, and stock observations.

#### Scenario: Candidate persistence
- **GIVEN** provider returns normalized listing candidates
- **WHEN** ingestion persists candidates
- **THEN** Cabinet SHALL store candidate records with stock context fields

### Requirement CANDIDATES-002: Candidate ingestion SHALL deduplicate via fingerprint
Cabinet SHALL compute deterministic candidate fingerprint and prevent duplicate record creation for same provider/query-set scope.

#### Scenario: Duplicate candidate returned across runs
- **GIVEN** an existing candidate with same provider/query-set fingerprint exists
- **WHEN** later scan returns the same fingerprint candidate
- **THEN** Cabinet SHALL update existing record instead of creating a duplicate

