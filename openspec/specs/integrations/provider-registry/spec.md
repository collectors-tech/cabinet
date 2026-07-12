## Purpose
Define the canonical integrations/provider registry used by scanner, pricing, and integrations UI.

## Requirements
### Requirement INTEGRATION-001: Provider registry MUST define provider identity and capabilities
Cabinet SHALL maintain a provider registry with stable provider IDs, display metadata, auth mode, and capability flags (search, stock, pricing, media, health).

#### Scenario: Registry entry load
- **GIVEN** an authenticated user with `admin` role opens Integrations and runtime registry is loaded
- **WHEN** `GET /api/providers/registry` is called
- **THEN** response MUST be `200` with payload fields per provider:
  - `provider_id` (string, stable)
  - `display_name` (string)
  - `base_domain` (string)
  - `integration_mode` (`official_api|web_ingestion|program_api`)
  - `api_family` (`woo_store_api|boost_shopify|algolia|custom`)
  - `api_support_profile` (string, e.g. `v1`, `store_v1`, `boost_v2`)
  - `auth_mode` (`none|oauth|api_key|hybrid`)
  - `capabilities.search` (boolean)
  - `capabilities.stock_observation` (boolean)
  - `capabilities.pricing` (boolean)
  - `capabilities.health` (boolean)
  - `state` (`ready|degraded|disabled`)
  - `has_token` (boolean, write-only credential presence signal)

### Requirement INTEGRATION-002: Registry MUST include eBay and Amazon providers
Cabinet SHALL define provider entries for `ebay` and `amazon` with explicit capability and credential requirements.

#### Scenario: Core marketplaces available
- **GIVEN** provider registry is active for current runtime
- **WHEN** integrations client loads provider list
- **THEN** response MUST be `200` and entries for `ebay` and `amazon` MUST exist with non-empty `provider_id`, `integration_mode`, and `state`

### Requirement INTEGRATION-003: Registry MUST include configured AU webshop providers
Cabinet SHALL include AU webshop providers from product scope:
- bonzaslotcars.com.au
- frontlinehobbies.com.au
- hobbytechtoys.com.au
- andrewshobbies.com.au
- voglers.com.au
- acercmodels.com
- mrtoys.com.au

#### Scenario: AU webshop catalog rendered
- **GIVEN** AU webshop providers are configured in runtime provider catalog
- **WHEN** `GET /api/providers/registry` returns provider entries
- **THEN** all configured domains MUST be represented in `base_domain` field
  - with `200` for successful response
  - with `4xx` for validation/auth conflicts
  - with `5xx` for unexpected runtime failures

### Requirement INTEGRATION-004: Registry entries MUST map to provider capability specs
Each provider entry SHALL map to a provider-specific OpenSpec capability.

#### Scenario: Provider traceability
- **GIVEN** migration review for provider contracts is executed
- **WHEN** provider entry is selected for build/testing
- **THEN** mapped provider specs MUST exist:
  - `provider-ebay`
  - `provider-amazon`
  - `provider-au-webshops`

### Requirement INTEGRATION-021: Provider registry MUST include operational health snapshot fields
Cabinet SHALL expose provider health and last-run metadata required by integrations cards and detail panels.

#### Scenario: Provider operational snapshot load
- **GIVEN** runtime provider services have health telemetry for configured providers
- **WHEN** `GET /api/providers/registry` is requested
- **THEN** each provider entry MUST include:
  - `health.status` (`ok|degraded|down|unknown`)
  - `health.last_checked_at` (timestamp or null)
  - `health.message` (string)
  - `last_run.status` (`idle|running|success|failed|never`)
  - `last_run.finished_at` (timestamp or null)

### Requirement INTEGRATION-023: Provider registry MUST expose setup guidance and credential-presence signal
Cabinet SHALL expose registry fields needed for safe credential UX and guided setup.

### Requirement INTEGRATION-024: Provider registry MUST expose provider-to-API-spec support mapping
Registry entries SHALL declare API family mapping so Integrations UI can display how each provider is implemented (Woo/Boost/Algolia/custom).

#### Scenario: Provider API mapping available in registry
- **GIVEN** integrations UI requests provider registry
- **WHEN** payload is returned
- **THEN** each provider entry MUST include `api_family` and `api_support_profile`
- **AND** mapping MUST correspond to published provider API family contracts

#### Scenario: Registry payload supports credential-safe integrations UI
- **GIVEN** active profile settings and provider registry are loaded
- **WHEN** `GET /api/providers/registry` returns provider entries
- **THEN** each provider entry MUST include:
  - `setup_instructions` (string)
  - `has_token` (boolean presence signal only)
- **AND** registry response MUST NOT expose clear credential/token values

### Requirement INTEGRATION-027: Provider registry MUST publish provider manifests and config schemas
Cabinet SHALL treat each integration provider as a manifest-backed registry entry with a schema-driven setup contract.

#### Scenario: Registry manifest and setup schema load
- **GIVEN** a provider is listed by `GET /api/providers/registry`
- **WHEN** Cabinet builds the registry payload for the active profile
- **THEN** each entry MUST expose manifest metadata for:
  - stable provider ID
  - display name
  - provider category/type
  - integration mode
  - API family/support profile
  - capability flags
  - setup instructions
  - default health and required-action state
- **AND** configurable providers MUST expose a setup schema that identifies each non-secret field, field type, label, validation rule, default value, and persistence key
- **AND** secret fields MUST expose only write-only field metadata and credential-presence state, never the stored secret value
- **AND** OpenAI / ChatGPT MUST publish schema-driven active-profile setup fields for active auth method, default assistant model, write-only API-key secret entry, and Browser Auth proof state so consumers can render setup without hardcoded provider-specific field lists

#### Scenario: Schema-driven Add Integration setup
- **GIVEN** the user opens Add Integration and selects a provider from the registry-driven catalog
- **WHEN** the provider setup form renders
- **THEN** editable fields MUST be generated from the provider setup schema rather than hardcoded provider-specific field lists
- **AND** the form MUST preserve visible or programmatic labels, validation feedback, and write-only handling for secret fields
- **AND** saving setup MUST persist non-secret values through active-profile settings and secret values through the profile secrets path

### Requirement INTEGRATION-028: Provider registry MUST expose workflow/action metadata
Cabinet SHALL expose provider workflow and action metadata so UI surfaces can present only supported operations with clear safety state.

#### Scenario: Provider actions are discoverable from registry metadata
- **GIVEN** a registry entry supports scanner, Market Watch, import/export, media, assistant, notification, seller, or setup workflows
- **WHEN** `GET /api/providers/registry` is requested
- **THEN** the entry MUST expose workflow/action metadata including stable action ID, label, capability category, execution mode, read/write classification, confirmation requirement, availability state, and required next action when unavailable
- **AND** registry metadata MUST distinguish local-only actions, read-only remote actions, and remote writes that require explicit confirmation
- **AND** UI controls MUST be disabled or routed to the correct workflow when action metadata marks an action unavailable, blocked, or handled outside the provider dialog

### Requirement INTEGRATION-064: Provider workflow registry MUST define execution contracts
Cabinet SHALL maintain typed workflow/action registry definitions tied to provider manifest `workflow_refs` so UI, assistant, automation, scanner, import/export, notification, and validation consumers can discover provider operations without hardcoded provider-specific action lists.

#### Scenario: Workflow registry actions expose safety and routing metadata
- **GIVEN** a provider manifest declares workflow references for assistant, notification, Market Watch, provider diagnostics, buyer-interest, seller-operation, or listing-lifecycle workflows
- **WHEN** `GET /api/providers/registry` is requested
- **THEN** each declared workflow with a registry definition MUST appear in the provider `actions` payload with:
  - stable `action_id` / `workflow_ref`
  - operator-safe `label` and `description`
  - workflow `type`
  - `input_schema` and `output_schema`
  - `requires_auth` and `requires_secrets`
  - capability list
  - `side_effect_level`
  - `confirmation_required`
  - `schedule_support`
  - `inbox_events`
  - `health_impact`
  - `execution_mode`
  - `availability_state`
- **AND** `side_effect_level` MUST distinguish `read_only`, `preview_only`, `write`, and `destructive` operations
- **AND** mutation or destructive workflows MUST advertise explicit confirmation requirements
- **AND** workflow failure and required-action outcomes MUST advertise Inbox event metadata so downstream consumers can route errors, warnings, required user action, and result-inbox updates consistently

### Requirement INTEGRATION-065: Provider registry MUST publish typed setup schema definitions
Cabinet SHALL maintain typed setup schema definitions tied to provider manifest `config_schema_ref` values so Add Integration and provider setup consumers can render configuration fields without hardcoded provider-specific forms.

#### Scenario: Setup schemas expose renderer-safe field metadata
- **GIVEN** a provider manifest declares a `config_schema_ref`
- **WHEN** `GET /api/providers/registry` is requested
- **THEN** the provider entry MUST expose `setup_schema` with:
  - stable `schema_ref`
  - `persistence_scope`
  - `submit_target`
  - optional `secret_target`
  - optional `validate_action`
  - ordered field metadata
- **AND** setup schema fields MUST support `text`, `secret`, `url`, `number`, `select`, `multiselect`, `checkbox`, `textarea`, `file`, `oauth-connect`, and `browser-auth-status`
- **AND** each field MUST expose `key`, `label`, `type`, `required`, `write_only`, and `persistence`
- **AND** fields MAY expose placeholder, helper text, default value, options, validation rules, documentation URL, read-only state, and conditional rendering metadata
- **AND** write-only fields MUST expose only field metadata, secret key alias, and credential-presence state, never stored secret values

#### Scenario: Required provider shapes are represented
- **GIVEN** Add Integration renders schema-driven setup for common Cabinet provider shapes
- **WHEN** provider registry payloads are loaded
- **THEN** an API key provider MUST expose a secret setup field persisted through `profile_secrets`
- **AND** a Browser Auth provider MUST expose a `browser-auth-status` field that stays read-only until verified proof exists
- **AND** a No-auth/static source provider MUST expose non-secret setup fields and a validate/detect action without requiring a secret target
- **AND** these shapes MUST be covered by focused registry contract tests before UI consumers claim schema-driven setup support

### Requirement INTEGRATION-029: Integration instances MUST persist separately from provider definitions
Cabinet SHALL persist per-profile integration instances and status independently from immutable provider manifests.

#### Scenario: Profile-scoped instance persistence
- **GIVEN** an active profile configures a registry provider
- **WHEN** setup is saved, validated, disabled, updated, listed, or deleted through `/api/profiles/:profileId/integration-instances`
- **THEN** Cabinet MUST persist a profile-scoped integration instance containing provider ID, enabled state, non-secret configuration, credential-presence signals, validation status, health state, last-run summary, and required-action state
- **AND** the stored instance MUST reference the provider manifest instead of duplicating mutable manifest fields
- **AND** secret setup values MUST be written through the profile secrets path and represented in the instance payload only as secret reference keys
- **AND** listing integrations MUST merge provider manifest data with the active profile instance state deterministically

#### Scenario: Required-action state is preserved
- **GIVEN** provider health, validation, or workflow execution detects a missing credential, schema validation failure, provider outage, retry backoff, or manual review requirement
- **WHEN** the registry or integration list is reloaded
- **THEN** the integration instance MUST preserve a required-action code, operator-safe message, affected workflow, and retry/repair guidance until a later validation clears it

### Requirement INTEGRATION-030: Integration failures MUST create inbox-visible events
Cabinet SHALL surface integration failures and required user actions through durable inbox events as well as provider status fields.

#### Scenario: Failure and required-action events are promoted to Inbox
- **GIVEN** a registry-backed provider validation, scheduled run, Market Watch run, scanner run, import/export workflow, assistant workflow, or notification workflow fails or needs user action
- **WHEN** Cabinet records provider status for the active profile
- **THEN** Cabinet MUST create or update a Notification Inbox event with provider ID, provider display name, workflow/action ID, severity, required-action code, status message, target route, and timestamp
- **AND** repeated failures for the same provider/action/root cause SHOULD coalesce into an updated event rather than flooding duplicate notifications
- **AND** resolving the provider issue MUST allow the inbox event to be marked resolved/read without deleting the durable provider-status history

#### Scenario: Telegram setup validation creates required-action Inbox evidence
- **GIVEN** Telegram is represented as a registry-backed messaging provider for the active profile
- **WHEN** provider validation finds missing sender/chat authorization, bot token, or webhook proof
- **THEN** Cabinet MUST return non-secret Telegram setup status and create or update a provider-workflow Inbox event for `telegram.provider_test`
- **AND** the Inbox event MUST carry the required setup action, target `/integrations`, and no bot-token value
- **AND** repeated setup-required validations MUST coalesce into one durable Inbox event for the same provider/action/root cause

#### Scenario: Telegram setup validation resolves prior Inbox evidence
- **GIVEN** a Telegram provider-test setup-required Inbox event exists for the active profile
- **WHEN** sender/chat authorization, bot-token presence, and webhook setup proof are all present and provider-test validation succeeds
- **THEN** Cabinet MUST mark the previous `telegram.provider_test` provider-workflow Inbox event read/resolved
- **AND** the resolved Inbox event MUST retain provider/action/root-cause metadata and add non-secret resolution evidence without returning bot-token material

#### Scenario: Unauthorized Telegram Agent text creates required-action Inbox evidence
- **GIVEN** Telegram Agent text is called for a known profile by a sender/chat that has not been authorized
- **WHEN** Cabinet rejects the message before routing an Agent workflow
- **THEN** Cabinet MUST create or update a provider-workflow Inbox event for `telegram.agent_text`
- **AND** the Inbox event MUST carry the `authorize_sender_chat` required action, target `/integrations`, and source message metadata without storing bot-token values

### Requirement INTEGRATION-063: Provider registry MUST be the canonical integration source for app consumers
Cabinet SHALL treat `/api/providers/registry`, `/api/providers/:id/*`, the Add Integration UI list, and the Market Watch provider projection as consumers of one canonical registry definition rather than independent provider catalogs.

#### Scenario: Registry consumers preserve provider category and capability boundaries
- **GIVEN** the canonical registry definition contains providers across marketplace, storefront/source matcher, browser-auth, chat/AI, notification, and workflow/local categories
- **WHEN** Cabinet builds provider registry, provider detail/action, Add Integration, or Market Watch provider projection payloads
- **THEN** each consumer MUST derive provider identity, category, auth mode, setup status, and capability metadata from the same manifest-backed registry source
- **AND** provider categories MUST remain distinct so hobby shop storefront/source matcher providers are not collapsed into one generic provider when platform-specific adapters or discovery evidence exist
- **AND** capability flags MUST preserve config form requirements, health/diagnostics, matching/import/export support, and browser-auth/external-login behavior for each provider
- **AND** consumer-specific payloads MAY project only the fields they need but MUST NOT invent provider identity, category, setup, or capability state outside the canonical registry definition

#### Scenario: Disabled assistant placeholders remain explicit and non-actionable
- **GIVEN** the assistant runtime registry contains provider placeholders that do not yet have supported adapters
- **WHEN** `GET /api/providers/registry` returns chat/AI provider entries
- **THEN** disabled assistant placeholders such as Anthropic and Google MUST appear as `provider_type=assistant` entries with `state=disabled`, `api_available=false`, `active_mode=disabled_placeholder`, and `api_support_profile=placeholder_disabled`
- **AND** disabled assistant placeholders MUST expose no workflow refs or credential setup controls until a supported adapter, setup schema, health check, and workflow mapping exists
- **AND** disabled placeholder capability flags MUST prevent assistant, image-help, and content-generation actions while still making the limitation visible in registry health/setup metadata

#### Scenario: Telegram messaging provider exposes setup and workflow metadata
- **GIVEN** Telegram is used as a Cabinet messaging, notification, catalog-capture, or external Agent channel
- **WHEN** `GET /api/providers/registry` returns the Telegram provider entry
- **THEN** Telegram MUST appear as a `provider_type=messaging` and `provider_category=notification` registry entry with the `integrations/telegram/channel` setup schema
- **AND** the setup schema MUST expose sender ID, chat ID, write-only bot token, and webhook route metadata without returning secret token values
- **AND** registry setup status and health MUST distinguish missing sender/chat authorization, missing bot token, pending webhook proof, and ready-for-live-channel-checklist states
- **AND** Telegram catalog capture and Agent text workflows MUST expose action metadata with setup-needed or available state derived from the same registry readiness projection

### Requirement INTEGRATION-066: Marketplace providers MUST migrate through registry metadata
Cabinet SHALL represent marketplace providers such as eBay and Amazon as manifest-backed registry entries with provider-specific auth, setup, health, workflow, and capability metadata instead of hardcoded Add Integration lists.

#### Scenario: Marketplace providers expose migration metadata
- **GIVEN** Cabinet builds `/api/providers/registry`
- **WHEN** marketplace providers are returned
- **THEN** eBay and Amazon MUST expose stable provider IDs, display names, marketplace category/type, auth mode, config schema refs, setup schema payloads, market-watch scopes, and workflow/action refs from the canonical registry manifest
- **AND** marketplace capability flags MUST cover supported search/import/scanner matching, price checks, purchase/order reconciliation, listing lookup, seller operation, listing lifecycle, and health boundaries where applicable
- **AND** Add Integration and workflow consumers MUST be able to route setup/config and supported marketplace actions from registry metadata without a separate hardcoded marketplace provider catalog

## Validation Coverage Required By Issue #1469
- Provider registry entries and manifest fields: Go registry/OpenAPI contract tests.
- Provider setup schemas and Add Integration form rendering: Go template contract tests plus targeted Cypress for registry-driven field rendering.
- Workflow/action metadata: Go registry contract tests and targeted UI action availability checks.
- Profile-scoped integration instance persistence: Go API/service tests for settings/secrets/status merge behavior.
- Secret redaction and storage: Go API/service tests proving secrets are write-only and never returned in registry payloads.
- Health, status, required-action, and Inbox failure events: Go status/inbox API tests plus targeted Cypress where UI state is rendered.
