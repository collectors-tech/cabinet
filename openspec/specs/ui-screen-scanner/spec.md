## Purpose
Define Scanner screen behavior and scan operations use cases.

## Requirements
### Requirement: Scanner screen SHALL support query set management and execution
The screen SHALL provide query set create/list/update and run-now/scheduled execution controls.

#### Scenario: Use case - run query set now
- **WHEN** user runs selected query set
- **THEN** scanner execution SHALL start and return status/candidate availability

### Requirement: Scanner screen SHALL expose failures, retry, and provider health
The screen SHALL provide diagnosable failures and retry entry points.

#### Scenario: Use case - recover failed query set
- **WHEN** user retries failed query set
- **THEN** scanner SHALL attempt retry and update failure status
