## Purpose
Define scanner query-set lifecycle, execution controls, and failure recovery behavior.

## Requirements
### Requirement: Scanner query sets SHALL support user-defined market criteria
Cabinet SHALL support query set criteria including keywords, exclusions, max price, region, condition, and scheduling metadata.

#### Scenario: Create query set
- **GIVEN** valid query set input is provided
- **WHEN** user submits query set form
- **THEN** Cabinet SHALL persist query set definition

### Requirement: Scanner execution SHALL support manual and scheduled runs with rate limits
Cabinet SHALL support run-now and scheduled execution under rate-limited controls.

#### Scenario: Scheduled scanner run
- **GIVEN** enabled scheduled query set exists
- **WHEN** scheduler triggers execution
- **THEN** scan SHALL execute with rate-limit policy

### Requirement: Scanner failures SHALL be diagnosable and retryable
Cabinet SHALL log failures and support retry by query set.

#### Scenario: Retry failed scan
- **GIVEN** query set has a failed scanner run
- **WHEN** user requests retry
- **THEN** Cabinet SHALL schedule immediate retry and log outcome
