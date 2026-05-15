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

### Requirement PROVIDER-OPENAI-UX-005: OpenAI config SHALL use a clean card and method-aware dialog
Cabinet MUST adapt the SCHA OpenAI setup pattern into a compact provider card plus a dialog-owned configuration flow. The provider card MUST avoid setup clutter, while the dialog MUST separately present Browser Auth, API key, and Test OpenAI sections.

#### Scenario: Clean OpenAI card with dialog-owned setup
- **GIVEN** user opens `/integrations`
- **WHEN** the OpenAI / ChatGPT provider card renders
- **THEN** card-level setup controls MUST be limited to status and a primary connect/manage action
- **AND** card-level Validate/Sync/Test/configuration clutter MUST be absent
- **WHEN** user opens the OpenAI config dialog
- **THEN** Browser Auth, API key, and Test OpenAI sections MUST be visible inside the dialog
- **AND** duplicate method narration such as `OpenAI is using: Browser Auth` MUST NOT render

### Requirement PROVIDER-OPENAI-UX-006: Browser Auth SHALL require verifiable proof before connected readiness
Cabinet MUST NOT mark OpenAI Browser Auth connected from navigation, a user return, or a provider tab launch alone. Connected readiness requires a verifiable callback/artifact/proof recorded by Cabinet.

#### Scenario: Browser Auth setup-needed until proof exists
- **GIVEN** OpenAI Browser Auth has no verified callback/artifact/proof
- **WHEN** the OpenAI dialog renders
- **THEN** Browser Auth MUST show setup-needed or unavailable state
- **AND** OpenAI MUST NOT be considered connected through Browser Auth
- **AND** the UI MUST explain that navigation alone is not connected proof

### Requirement PROVIDER-OPENAI-UX-007: API key setup SHALL save secrets separately from profile settings
Cabinet MUST keep OpenAI API-key entry write-only and store the key through the profile secrets API while storing non-secret OpenAI defaults in profile settings.

#### Scenario: API key connect writes secret and non-secret defaults
- **GIVEN** user enters an OpenAI API key and default model in the OpenAI dialog
- **WHEN** user connects or saves OpenAI
- **THEN** Cabinet MUST write `openai_api_key` through `/api/profiles/:profileId/secrets`
- **AND** Cabinet MUST write non-secret defaults such as `assistant_default_provider`, `assistant_default_model`, and active method through `/api/profiles/:profileId/settings`
