## Purpose
Define Cabinet product-level OpenAI / ChatGPT integration UX for provider setup, capability scoping, and assistant consumption.

## Requirements
### Requirement PROVIDER-OPENAI-UX-001: OpenAI / ChatGPT SHALL be configurable through Integrations as a visible provider capability
Cabinet MUST expose OpenAI / ChatGPT as a first-class integration option with visible connection/configuration state rather than treating it as hidden implementation detail only.

#### Scenario: Render provider card in integrations
- **GIVEN** user opens `/integrations`
- **WHEN** OpenAI / ChatGPT capability is available in provider registry
- **THEN** integrations UI MUST show provider card or equivalent setup surface with visible status and action controls

### Requirement PROVIDER-OPENAI-UX-002: OpenAI / ChatGPT integration SHALL support explicit auth-mode UX
Cabinet MUST define and expose the supported auth model for OpenAI / ChatGPT integration, including operator-facing API-key workflows and any supported account-connect flow.

#### Scenario: Configure auth method
- **GIVEN** user opens OpenAI / ChatGPT integration details
- **WHEN** auth/setup controls render
- **THEN** UI MUST show deterministic supported auth method(s), safe credential handling, and validation feedback

### Requirement PROVIDER-OPENAI-UX-003: OpenAI / ChatGPT integration SHALL expose capability scoping for assistant workflows
Cabinet MUST define which assistant capabilities use OpenAI / ChatGPT integration (for example chat help, photo analysis, tagging/classification, generation) and expose that scope to operators/users where appropriate.

#### Scenario: Inspect provider capability scope
- **GIVEN** OpenAI / ChatGPT provider is configured
- **WHEN** provider details are viewed
- **THEN** UI MUST show which Cabinet assistant workflows are enabled or disabled for that provider configuration

### Requirement PROVIDER-OPENAI-UX-004: Provider/model defaults SHALL integrate with Assistant workspace selection behavior
Cabinet MUST define how integrations-level provider defaults interact with Assistant workspace provider/model selection and thread metadata.

#### Scenario: Use integrations default in assistant
- **GIVEN** Integrations defines default OpenAI / ChatGPT provider/model settings
- **WHEN** user opens Assistant workspace
- **THEN** assistant selection UI MUST reflect those defaults deterministically
- **AND** thread metadata MUST record the actual provider/model used for messages and executions
