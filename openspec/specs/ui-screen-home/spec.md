## Purpose
Define Home dashboard screen behavior and actionable use cases.

## Requirements
### Requirement: Home screen SHALL show actionable command-center signals
Home SHALL prioritize what needs immediate action for collector workflows.

#### Scenario: Use case - review urgent actions
- **WHEN** user opens Home screen
- **THEN** Home SHALL show actionable items for discoveries, pricing changes, and health alerts

### Requirement: Home screen SHALL provide quick action entry points
Home SHALL expose quick actions for common flows such as add item, run scanner, and open collection workspace.

#### Scenario: Use case - start action from Home
- **WHEN** user clicks a quick action
- **THEN** app SHALL navigate to destination workflow with context
