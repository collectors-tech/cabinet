## Purpose
Define accessibility and keyboard contracts for modal/drawer and core workflows.

## Requirements
### Requirement UI-FOUNDATION-ACCESSIBILITY-001: Modal and drawer components SHALL satisfy focus and keyboard contracts
Cabinet SHALL trap and restore focus, support escape close, and provide keyboard-accessible primary actions.

#### Scenario: Dialog keyboard behavior
- **GIVEN** modal is open and focus is trapped
- **WHEN** user presses Escape
- **THEN** modal SHALL close and focus SHALL return to trigger

### Requirement UI-FOUNDATION-ACCESSIBILITY-002: Accessibility semantics SHALL be non-optional for core workflows
Cabinet SHALL provide explicit labels, landmark roles, and non-color-only status indicators.

#### Scenario: Keyboard-only completion
- **GIVEN** user navigates without mouse
- **WHEN** user executes a core workflow by keyboard
- **THEN** controls SHALL remain reachable and actionable

### Requirement UI-FOUNDATION-ACCESSIBILITY-003: Action controls SHALL provide accessible names
All action buttons, including icon-only controls, SHALL expose accessible names (`aria-label` or equivalent accessible text).

#### Scenario: Action button accessible naming
- **GIVEN** user-facing screen renders action controls (row actions, header actions, drawer/modal actions)
- **WHEN** accessibility inspection runs
- **THEN** each actionable control MUST have a non-empty accessible name and keyboard focus visibility

### Requirement UI-FOUNDATION-ACCESSIBILITY-004: Authenticated shell SHALL preserve landmarks and headings across responsive widths
Cabinet SHALL render a single visible main landmark, stable route heading, visible shell header chrome, and no document-level horizontal overflow for authenticated core routes at desktop and mobile widths.

#### Scenario: Responsive shell landmarks
- **GIVEN** authenticated Inventory shell is opened at desktop and mobile viewport widths
- **WHEN** the page finishes loading
- **THEN** exactly one visible main landmark SHALL be present, the route heading SHALL remain visible and accessible, header chrome SHALL remain available, and document horizontal overflow SHALL not exceed the viewport

