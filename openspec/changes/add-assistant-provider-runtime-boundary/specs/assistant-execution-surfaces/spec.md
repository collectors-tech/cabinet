## MODIFIED Requirements

### Requirement: OpenAI-backed assistant capabilities SHALL expose truthful setup and readiness states
Cabinet MUST treat OpenAI/API-key/Browser Auth readiness as provider evidence that gates capabilities, not as copy or navigation state.

#### Scenario: Consume assistant-provider output without delegating Cabinet tools
- **GIVEN** OpenAI-backed assistant provider readiness is verified for the active profile
- **WHEN** governed Chat requests a normal assistant turn
- **THEN** Cabinet SHALL call the assistant-provider runtime through a provider-neutral turn interface
- **AND** Cabinet SHALL store non-secret provider/model/error-class metadata with the Chat or workflow evidence
- **AND** Cabinet SHALL keep Cabinet skill routing, preview-before-apply, confirmation, database mutation, filesystem access, and app-control commands inside Cabinet-governed execution surfaces
- **AND** provider output SHALL be treated as suggestion/text input to the governed planner rather than permission to mutate Cabinet state
