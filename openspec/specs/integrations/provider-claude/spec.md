## Purpose
Define Claude provider adapter behavior through the AI gateway for reasoning-heavy classification and summarization.

## Requirements
### Requirement PROVIDER-CLAUDE-001: Claude adapter SHALL implement gateway canonical contract
Cabinet SHALL normalize Claude responses to the same gateway envelope to preserve caller compatibility.

#### Scenario: Reasoning task response
- **GIVEN** task type requires long-context reasoning and routes to Claude
- **WHEN** provider response is returned
- **THEN** adapter MUST emit canonical `result` and `meta` with `provider="claude"` and stable confidence semantics

### Requirement PROVIDER-CLAUDE-002: Claude adapter SHALL support explainable recommendations
Cabinet SHALL preserve explainability fields for recommendation and classification tasks.

#### Scenario: Explainable inventory recommendation
- **GIVEN** task requests recommended assignment/classification
- **WHEN** Claude adapter returns result
- **THEN** normalized response MUST include recommendation plus explanation fields suitable for user review and approval

### Requirement PROVIDER-CLAUDE-003: Claude adapter SHALL enforce policy-safe output mapping
Cabinet SHALL reject unsupported output shapes and map to deterministic gateway error codes.

#### Scenario: Invalid structured output
- **GIVEN** provider response cannot be normalized to required schema
- **WHEN** adapter validates output
- **THEN** runtime MUST return stable validation error code and SHALL NOT perform downstream state mutation
