## Purpose
Define mandatory UI/UX governance gates that determine when a screen is truly done for release.

## Requirements
### Requirement: Information hierarchy gate SHALL be enforced on every screen
Each screen SHALL present one primary above-the-fold outcome and SHALL de-prioritize secondary or technical controls.

#### Scenario: Hierarchy review
- **WHEN** a screen is reviewed for completion
- **THEN** one primary outcome SHALL be clearly dominant above the fold

### Requirement: Action clarity gate SHALL ensure operationally clear controls
Primary action controls SHALL be visible without scrolling on desktop and SHALL use task language labels.

#### Scenario: Action row validation
- **WHEN** user opens a top-level screen
- **THEN** primary action controls SHALL be visible and semantically task-oriented

### Requirement: Layout behavior gate SHALL enforce shell stability
Left navigation SHALL remain fixed, page header SHALL remain sticky, and page body SHALL own vertical scroll.

#### Scenario: Scroll ownership validation
- **WHEN** content exceeds viewport height
- **THEN** side navigation and header SHALL remain fixed while body content scrolls

### Requirement: Progressive disclosure gate SHALL protect first-run usability
Diagnostic and admin-heavy controls SHALL not dominate first-run surfaces and SHALL be behind dedicated screens or expansion boundaries.

#### Scenario: First-run usability validation
- **WHEN** a first-run user enters starter workflows
- **THEN** advanced diagnostics controls SHALL be de-prioritized from primary action path

### Requirement: Test gate SHALL be mandatory for screen completion
Each screen SHALL have structure, primary action, and state coverage before issue closure.

#### Scenario: Screen closure readiness
- **WHEN** a screen issue is marked ready for close
- **THEN** at least one structure test, one primary action test, and one state test SHALL have passing evidence in-session

### Requirement: Governance gate evidence SHALL be attached to remediation waves
Each remediation wave SHALL include before/after references, gap mapping, and unresolved follow-up issues.

#### Scenario: Remediation wave review
- **WHEN** a remediation wave is completed
- **THEN** review evidence SHALL include screenshots, gate gap mapping, test evidence, and follow-up issues

## Acceptance Criteria
1. All gate requirements are normative (`SHALL`) and scenario-testable.
2. PR and issue closure criteria include mandatory gate evidence and in-session test results.
3. Screen-level remediation cannot be marked complete without gate compliance.

## Success Criteria
1. UI regressions reduce due to explicit gate enforcement in issue closure.
2. First-run user workflows remain focused and non-overwhelming.
3. Navigation/header/layout behavior remains stable across screens.

## Source Mapping
- `docs/ui-spec/13-UI-UX-STRATEGY-GATE.md`
