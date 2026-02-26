## Purpose
Define accessibility and keyboard contracts for modal/drawer and core workflows.

## Requirements
### Requirement: Modal and drawer components SHALL satisfy focus and keyboard contracts
Cabinet SHALL trap and restore focus, support escape close, and provide keyboard-accessible primary actions.

#### Scenario: Dialog keyboard behavior
- **GIVEN** modal is open and focus is trapped
- **WHEN** user presses Escape
- **THEN** modal SHALL close and focus SHALL return to trigger

### Requirement: Accessibility semantics SHALL be non-optional for core workflows
Cabinet SHALL provide explicit labels, landmark roles, and non-color-only status indicators.

#### Scenario: Keyboard-only completion
- **GIVEN** user navigates without mouse
- **WHEN** user executes a core workflow by keyboard
- **THEN** controls SHALL remain reachable and actionable

