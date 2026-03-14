## Purpose
Define how Cabinet presents assistant proposals, confirmation boundaries, execution states, and result summaries for governed action-taking workflows.

## Requirements
### Requirement ASSISTANT-EXECUTION-001: Assistant mutations SHALL render preview-before-apply execution surfaces
Assistant-proposed create/update/delete/import actions MUST render a preview surface before execution.

#### Scenario: Preview assistant mutation
- **GIVEN** assistant proposes a structured mutation
- **WHEN** proposal is rendered to user
- **THEN** UI MUST show a preview/summary of intended action and affected entities before apply becomes possible

### Requirement ASSISTANT-EXECUTION-002: Assistant SHALL require explicit confirmation for state-changing actions
Assistant MUST require explicit confirmation for actions that mutate records, import files, or change important application state.

#### Scenario: Confirm before apply
- **GIVEN** assistant action would change Cabinet state
- **WHEN** user has not explicitly confirmed
- **THEN** runtime MUST NOT execute the action

### Requirement ASSISTANT-EXECUTION-003: Assistant SHALL expose deterministic execution lifecycle states
Assistant MUST expose deterministic queued/running/success/failure states for governed action execution.

#### Scenario: Show execution lifecycle
- **GIVEN** assistant action is accepted for execution
- **WHEN** execution lifecycle progresses
- **THEN** UI MUST show deterministic lifecycle states and outcome summary instead of vague loading behavior

### Requirement ASSISTANT-EXECUTION-004: Assistant SHALL make tool permission boundaries visible
Cabinet MUST expose which classes of assistant actions are read-only, preview-only, confirm-required, or unavailable under the active policy.

#### Scenario: Show permission boundary on blocked action
- **GIVEN** assistant proposes an action outside allowed permission scope
- **WHEN** UI renders the proposal or rejection
- **THEN** user MUST receive explicit permission-state guidance instead of silent failure or hidden omission
