## Purpose
Define Scanner screen behavior for query set management, execution, and diagnostics.

## Requirements
### Requirement UI-SCREEN-SCANNER-001: Scanner SHALL support query set CRUD and run controls
Scanner SHALL allow creating/loading query sets and triggering manual/scheduled runs.

#### Scenario: Create Query Set action
- **GIVEN** scanner route is loaded
- **WHEN** user enters query set name/keywords and clicks `Create Query Set`
- **THEN** query set MUST be created and displayed in scanner list with deterministic success/error feedback

#### Scenario: Run query set
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user runs a selected query set
- **THEN** scanner execution status and outputs SHALL update

### Requirement UI-SCREEN-SCANNER-002: Scanner SHALL expose provider health and failure retry
Scanner SHALL expose health diagnostics and retry controls for failed runs.

#### Scenario: Retry failed query set
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user retries failed scanner run
- **THEN** retry SHALL execute and status SHALL update

### Requirement UI-SCREEN-SCANNER-003: Scanner SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states for query sets and candidates.

#### Scenario: Scanner empty state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** no query sets exist
- **THEN** screen SHALL provide create-first guidance

### Requirement UI-SCREEN-SCANNER-004: Scanner run failures SHALL surface actionable guidance instead of raw keys
Scanner run and retry failures SHALL map backend taxonomy/status to user-readable guidance with deterministic recovery actions.

#### Scenario: Run Now failure guidance
- **GIVEN** scanner query set exists and `POST /api/scanner/run` returns non-2xx (for example `400`)
- **WHEN** user clicks `Run Now`
- **THEN** UI MUST render human-readable failure summary (not raw key such as `run_failed_400`)
- **AND** UI MUST render actionable next steps for query validation/provider health/credentials checks
- **AND** UI MAY expose raw diagnostic code only in secondary details region

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
| UC-SCN-06 | Create query set from form | New query set appears after `Create Query Set` action | planned: `ui.web/cypress/e2e/integrations/ui-screen-scanner/spec.cy.ts` `scanner-create-query-set` |
| UC-SCN-07 | Run Now failure taxonomy mapping | UI shows actionable guidance, secondary diagnostics | planned: `ui.web/cypress/e2e/integrations/ui-screen-scanner/spec.cy.ts` `UI-SCREEN-SCANNER-004 maps run failures to actionable guidance` |
