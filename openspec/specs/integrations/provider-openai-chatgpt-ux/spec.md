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
- **AND** `/api/providers/registry` MUST expose non-secret schema metadata for the OpenAI active method, default assistant model, write-only API-key secret, and Browser Auth proof state so setup consumers do not need hardcoded provider fields

### Requirement PROVIDER-OPENAI-UX-003: OpenAI / ChatGPT integration SHALL expose capability scoping for assistant workflows
Cabinet MUST define which assistant capabilities use OpenAI / ChatGPT integration (for example chat help, photo analysis, tagging/classification, generation) and expose that scope to operators/users where appropriate.

#### Scenario: Inspect provider capability scope
- **GIVEN** OpenAI / ChatGPT provider is configured
- **WHEN** provider details are viewed
- **THEN** UI MUST show which Cabinet assistant workflows are enabled or disabled for that provider configuration
- **AND** provider registry metadata MUST expose stable assistant action ids, workflow refs, read/write classification, confirmation requirements, availability state, and setup next action for OpenAI-backed assistant workflows

### Requirement PROVIDER-OPENAI-UX-004: Provider/model defaults SHALL integrate with Assistant workspace selection behavior
Cabinet MUST define how integrations-level provider defaults interact with Assistant workspace provider/model selection and thread metadata.

#### Scenario: Use integrations default in assistant
- **GIVEN** Integrations defines default OpenAI / ChatGPT provider/model settings
- **WHEN** user opens Assistant workspace
- **THEN** assistant selection UI MUST reflect those defaults deterministically
- **AND** thread metadata MUST record the actual provider/model used for messages and executions

### Requirement PROVIDER-OPENAI-UX-005: OpenAI config SHALL use a clean card and method-aware dialog
Cabinet MUST adapt the SCHA OpenAI setup pattern into a compact provider card plus a dialog-owned configuration flow. The provider card MUST avoid setup clutter, while the dialog MUST lead with friendly ChatGPT browser sign-in and keep API-key entry secondary under an Advanced disclosure.

#### Scenario: Clean OpenAI card with dialog-owned setup
- **GIVEN** user opens `/integrations`
- **WHEN** the OpenAI / ChatGPT provider card renders
- **THEN** card-level setup controls MUST be limited to status and a primary connect/manage action
- **AND** card-level Validate/Sync/Test/configuration clutter MUST be absent
- **WHEN** user opens the OpenAI config dialog
- **THEN** ChatGPT browser sign-in MUST be the prominent recommended action and state that no API key is required
- **AND** API-key controls MUST remain hidden under an explicit Advanced disclosure until requested
- **AND** duplicate method narration such as `OpenAI is using: Browser Auth` MUST NOT render
- **AND** generic operational Sync controls MUST NOT render inside the setup-needed OpenAI dialog
- **AND** API-key and test controls MUST have durable visible or programmatic labels rather than placeholder-only setup fields

### Requirement PROVIDER-OPENAI-UX-006: Browser Auth SHALL use supported ChatGPT sign-in and require verifiable proof before connected readiness
Cabinet MUST use the supported local Codex ChatGPT login flow rather than reading browser cookies, extension storage, or API keys. Cabinet MUST NOT mark OpenAI Browser Auth connected from navigation or a user return alone. Connected readiness requires authenticated runtime status plus a successful no-action provider test recorded for the active profile.

#### Scenario: Browser Auth setup-needed until proof exists
- **GIVEN** OpenAI Browser Auth has no verified callback/artifact/proof
- **WHEN** the OpenAI dialog renders
- **THEN** Browser Auth MUST show setup-needed or unavailable state
- **AND** OpenAI MUST NOT be considered connected through Browser Auth
- **AND** the UI MUST let the user continue with ChatGPT without entering an API key

#### Scenario: Connect an existing ChatGPT login
- **GIVEN** the supported local Codex runtime reports an authenticated ChatGPT session
- **WHEN** the user chooses Continue with ChatGPT
- **THEN** Cabinet MUST run a bounded no-action provider verification turn
- **AND** Cabinet MUST bind only non-secret readiness evidence to the selected profile
- **AND** Cabinet Chat MUST use the verified browser-auth runtime without reading the profile API-key secret
- **AND** disconnecting the Cabinet profile MUST preserve the user's global Codex login

### Requirement PROVIDER-OPENAI-UX-011: Browser-authenticated Chat SHALL preserve Cabinet authority boundaries
Cabinet MUST use ChatGPT browser authentication only as the language-provider transport. Provider execution MUST be isolated from Cabinet tools, files, browser control, integrations, plugins, and mutation authority.

#### Scenario: Run a browser-authenticated Chat turn
- **GIVEN** ChatGPT Browser Auth is verified for the active profile
- **WHEN** the user sends ordinary non-action text in Cabinet Chat
- **THEN** the provider runtime MUST receive bounded profile/thread conversation context
- **AND** Cabinet Chat MUST persist and render the returned ChatGPT text with OpenAI provider/model provenance
- **AND** the provider runtime MUST run without Cabinet tool or mutation authority
- **AND** Cabinet MUST continue to own preview, confirmation, apply, cancel, and audit behavior
- **AND** provider errors MUST remain classified and redacted
- **AND** Cabinet MUST NOT silently replace a selected-provider failure with deterministic success copy

### Requirement PROVIDER-OPENAI-UX-007: API key setup SHALL save secrets separately from profile settings
Cabinet MUST keep OpenAI API-key entry write-only and store the key through the profile secrets API while storing non-secret OpenAI defaults in profile settings.

#### Scenario: API key connect writes secret and non-secret defaults
- **GIVEN** user enters an OpenAI API key and default model in the OpenAI dialog
- **WHEN** user connects or saves OpenAI
- **THEN** Cabinet MUST write `openai_api_key` through `/api/profiles/:profileId/secrets`
- **AND** Cabinet MUST write non-secret defaults such as `assistant_default_provider`, `assistant_default_model`, and active method through `/api/profiles/:profileId/settings`
- **AND** registry setup schema metadata MUST mark `openai_api_key` as write-only `profile_secrets` persistence while keeping default model and active method under `profile_settings`

### Requirement PROVIDER-OPENAI-UX-008: Empty API-key actions SHALL bind validation to the token field
Cabinet MUST present missing OpenAI API-key validation as field-level feedback on the token input, with an accessible correction path, before any settings, secret save, or provider health validation request is attempted.

#### Scenario: Empty API-key connect targets the token input
- **GIVEN** user opens the OpenAI / ChatGPT API-key setup dialog without an existing token
- **WHEN** user attempts to connect or save with an empty API-key field
- **THEN** Cabinet MUST keep the dialog open and focus the token input
- **AND** the token input MUST expose invalid state and be described by visible field-level validation copy
- **AND** Cabinet MUST NOT call profile settings or profile secrets save endpoints

#### Scenario: Empty API-key validate targets the token input
- **GIVEN** user opens the OpenAI / ChatGPT API-key setup dialog without an existing token
- **WHEN** user attempts to validate with an empty API-key field
- **THEN** Cabinet MUST keep the dialog open and focus the token input
- **AND** the token input MUST expose invalid state and be described by visible field-level validation copy
- **AND** Cabinet MUST NOT call the provider health endpoint

### Requirement PROVIDER-OPENAI-UX-009: API key disconnect SHALL clear only API-key readiness
Cabinet MUST let a user explicitly disconnect the OpenAI API-key method by deleting the stored API-key secret and clearing API-key active-method readiness without destroying unrelated Browser Auth state.

#### Scenario: Disconnect API-key method
- **GIVEN** user opens the OpenAI / ChatGPT API-key setup dialog
- **WHEN** user disconnects the API-key method
- **THEN** Cabinet MUST delete openai_api_key through /api/profiles/:profileId/secrets
- **AND** Cabinet MUST clear API-key active method and integration enabled readiness when no other connected method is active
- **AND** Browser Auth state MUST remain present for a future verified Browser Auth proof
- **AND** Test OpenAI MUST return to setup-needed/disabled until a verified active method exists

### Requirement PROVIDER-OPENAI-UX-010: Provider health SHALL report profile-scoped OpenAI readiness without exposing secrets
Cabinet MUST return deterministic OpenAI health/readiness feedback for the active profile so operators can distinguish missing auth, missing API-key secret, Browser Auth proof gaps, and ready-to-test states without exposing credential material.

#### Scenario: Validate API-key readiness
- **GIVEN** OpenAI API-key mode is selected for the active profile
- **WHEN** provider health is requested for OpenAI
- **THEN** Cabinet MUST report setup-needed health when the profile has no `openai_api_key` secret
- **AND** Cabinet MUST report ready health when the profile has an API-key secret
- **AND** the provider registry MUST keep the OpenAI provider state setup-needed until the active API-key method has a stored secret
- **AND** the provider registry health projection MUST mirror the profile-scoped OpenAI setup/readiness state instead of reporting generic scanner health
- **AND** OpenAI assistant action metadata MUST remain setup-needed with required next action until the active profile has verified readiness
- **AND** the health payload MUST NOT include the secret value

#### Scenario: Validate Browser Auth proof readiness
- **GIVEN** OpenAI Browser Auth mode is selected for the active profile
- **WHEN** provider health is requested for OpenAI
- **THEN** Cabinet MUST report setup-needed health until verified Browser Auth proof is present
- **AND** Cabinet MUST report ready health only when Browser Auth state is connected and the verified artifact/proof flag is present
- **AND** the provider registry MUST keep the OpenAI provider state setup-needed until Browser Auth connected state and proof are both present
- **AND** the provider registry health projection MUST keep Browser Auth setup-needed until passed provider-test proof is present
