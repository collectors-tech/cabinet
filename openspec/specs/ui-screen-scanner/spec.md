## Purpose
Define Scanner screen behavior for query set management, execution, and diagnostics.

## Requirements
### Requirement: Scanner SHALL support query set CRUD and run controls
Scanner SHALL allow creating/loading query sets and triggering manual/scheduled runs.

#### Scenario: Run query set
- **WHEN** user runs a selected query set
- **THEN** scanner execution status and outputs SHALL update

### Requirement: Scanner SHALL expose provider health and failure retry
Scanner SHALL expose health diagnostics and retry controls for failed runs.

#### Scenario: Retry failed query set
- **WHEN** user retries failed scanner run
- **THEN** retry SHALL execute and status SHALL update

### Requirement: Scanner SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states for query sets and candidates.

#### Scenario: Scanner empty state
- **WHEN** no query sets exist
- **THEN** screen SHALL provide create-first guidance

## Acceptance Criteria
- UC IDs cover query management, execution, and failure handling.
- E2E mapping includes run-now and retry paths.

## Success Criteria
- Scanner workflows are operable without hidden dependency knowledge.
- Failures are diagnosable and retryable in-screen.

## Data Profiles
- Sample: 2 query sets, 50 candidates
- Bulk: 50 query sets, 200,000 candidates

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SCN-01 | Create/load query set | Query set list updates | planned: `cypress/e2e/ui/scanner.cy.ts` `scanner-queryset-crud` |
| UC-SCN-02 | Run query set now | Execution starts and candidates load | planned: `cypress/e2e/ui/scanner.cy.ts` `scanner-run-now` |
| UC-SCN-03 | Retry failed run | Retry attempt updates status | planned: `cypress/e2e/ui/scanner.cy.ts` `scanner-retry-failure` |
| UC-SCN-04 | No query sets | Empty guidance appears | planned: `cypress/e2e/ui/scanner.cy.ts` `scanner-empty-state` |
| UC-SCN-05 | Scanner API failure | Error + retry shown | planned: `cypress/e2e/ui/scanner.cy.ts` `scanner-error-state` |
