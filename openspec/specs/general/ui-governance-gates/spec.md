## Purpose
Define mandatory UI/UX governance gates that determine when a screen is truly done for release.

## Requirements
### Requirement UI-GOVERNANCE-GATES-001: Information hierarchy gate SHALL be enforced on every screen
Each screen SHALL present one primary above-the-fold outcome and SHALL de-prioritize secondary or technical controls.

#### Scenario: Hierarchy review
- **GIVEN** an authenticated user opens the Dashboard route (`/`) with successful `GET /api/dashboard`
- **WHEN** dashboard content is rendered
- **THEN** page title `Home` and supporting summary text SHALL appear above the fold
- **AND** the primary outcome panel `What needs attention now` SHALL be visible without opening diagnostics/admin panels

### Requirement UI-GOVERNANCE-GATES-002: Action clarity gate SHALL ensure operationally clear controls
Primary action controls SHALL be visible without scrolling on desktop and SHALL use task language labels.

#### Scenario: Action row validation
- **GIVEN** an authenticated user opens the Dashboard route (`/`)
- **WHEN** top-level actions are displayed
- **THEN** a primary action button with task-language label `Refresh Dashboard` SHALL be visible without scrolling

### Requirement UI-GOVERNANCE-GATES-003: Layout behavior gate SHALL enforce shell stability
Left navigation SHALL remain fixed, page header SHALL remain sticky, and page body SHALL own vertical scroll.

#### Scenario: Scroll ownership validation
- **GIVEN** an authenticated user opens a top-level workspace route with overflowing content
- **WHEN** the page scroll position changes
- **THEN** the header SHALL keep sticky positioning
- **AND** sidebar navigation SHALL remain visible
- **AND** body content SHALL continue to scroll independently beneath the shell

### Requirement UI-GOVERNANCE-GATES-004: Progressive disclosure gate SHALL protect first-run usability
Diagnostic and admin-heavy controls SHALL not dominate first-run surfaces and SHALL be behind dedicated screens or expansion boundaries.

#### Scenario: First-run usability validation
- **GIVEN** an authenticated user opens the Dashboard route (`/`)
- **WHEN** evaluating the first-run action path
- **THEN** the dashboard primary surface SHALL NOT present diagnostics controls as primary actions
- **AND** diagnostics-heavy guidance SHALL remain in dedicated support/settings surfaces

### Requirement UI-GOVERNANCE-GATES-005: Test gate SHALL be mandatory for screen completion
Each screen SHALL have structure, primary action, and state coverage before issue closure.

#### Scenario: Screen closure readiness
- **GIVEN** dashboard API responses include success and failure paths
- **WHEN** UI is tested for structure, primary action, and state transitions
- **THEN** passing evidence SHALL include:
  - structure proof (header/sidebar and core content)
  - primary action proof (`Refresh Dashboard` + `Retry`)
  - state proof (loading, error, ready transitions)

### Requirement UI-GOVERNANCE-GATES-006: Governance gate evidence SHALL be attached to remediation waves
Each remediation wave SHALL include before/after references, gap mapping, and unresolved follow-up issues.

#### Scenario: Remediation wave review
- **GIVEN** a remediation wave is completed for governance gates
- **WHEN** evidence is reviewed in-app and in migration artifacts
- **THEN** sidebar runtime metadata (`sidebar-runtime-meta`) SHALL expose build/version context
- **AND** support surface copy SHALL reference diagnostics workflows
- **AND** migration wave summary SHALL record commands, test results, and follow-up status

### Requirement UI-GOVERNANCE-GATES-007: UI lint and format gates SHALL pass before baseline blocker closure
UI baseline blocker issues SHALL only be closed when the UI package lint and format gates run cleanly against the current tree.

#### Scenario: Lint and format baseline closure
- **GIVEN** a UI lint or format baseline blocker is being closed
- **WHEN** closure evidence is captured
- **THEN** `npm run lint` from `ui.web` SHALL exit successfully
- **AND** `npm run format:check` from `ui.web` SHALL exit successfully
- **AND** any remaining lint output SHALL be warnings only, with no errors

## Acceptance Criteria
1. All gate requirements are normative (`SHALL`) and scenario-testable.
2. PR and issue closure criteria include mandatory gate evidence and in-session test results.
3. Screen-level remediation cannot be marked complete without gate compliance.

## Success Criteria
1. UI regressions reduce due to explicit gate enforcement in issue closure.
2. First-run user workflows remain focused and non-overwhelming.
3. Navigation/header/layout behavior remains stable across screens.

## Source Mapping
- Governance gates are fully defined in this canonical spec.
