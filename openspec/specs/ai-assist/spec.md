## Purpose
Define AI assist behavior, controls, and mutation safety model.

## Requirements
### Requirement: AI assist SHALL be OpenAI-backed in v1 with user-provided API key
Cabinet SHALL require profile-level OpenAI credentials for AI operations.

#### Scenario: AI operation without key
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** AI feature is invoked without configured key
- **THEN** Cabinet SHALL return actionable configuration error

### Requirement: AI suggestions SHALL include confidence and review state
Cabinet SHALL provide suggestion output with confidence signals.

#### Scenario: Suggest from title
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user requests title normalization
- **THEN** response SHALL include structured suggestions and confidence data

### Requirement: AI mutations SHALL require explicit confirmation
Cabinet SHALL not auto-create or auto-update collection data from AI output.

#### Scenario: Apply AI suggestion
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user attempts to apply suggested mutation
- **THEN** Cabinet SHALL require explicit confirm-before-apply action

### Requirement: AI control toggle SHALL be profile-scoped
Cabinet SHALL support per-profile enable/disable state for AI features.

#### Scenario: Disable AI in profile
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** AI is toggled off
- **THEN** AI endpoints SHALL reject profile requests until re-enabled
