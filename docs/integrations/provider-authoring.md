# Integration Provider Authoring

Use this guide when adding a new Cabinet integration provider or expanding an existing provider manifest. The provider registry is the source of truth for available integrations; UI surfaces and provider APIs should project from it instead of creating parallel provider catalogs.

## Source Of Truth

Provider manifests live in `internal/app/provider_registry_manifest.go`.

Every provider manifest must define:

- stable `provider_id`
- `display_name`
- `base_domain`
- `provider_category`
- `provider_type`
- `api_family`
- `api_support_profile`
- `active_mode`
- `integration_mode`
- `api_available`
- `auth_requirement`
- `auth_mode`
- `config_schema_ref`
- `workflow_refs`
- `capability_flags`
- `setup_instructions`

Use one manifest entry per distinct provider or platform adapter. Do not collapse marketplace, storefront/source matcher, browser-auth, chat/AI, notification, or workflow/local providers into a generic bucket when separate adapters, auth modes, or evidence exist.

## Add A Provider

1. Add the manifest to `coreIntegrationProviderManifests` or the appropriate family helper.
2. Choose a stable `provider_id` that can safely be referenced by profile-scoped integration instances.
3. Set `provider_category` and `provider_type` to match the provider's actual role.
4. Set `api_family` and `api_support_profile` to the implemented adapter contract. Use `web_ingestion` or a concrete API family such as `algolia`, `boost_shopify`, `bigcommerce`, `woo_store_api`, `official_api`, `ai_provider`, or `messaging_channel`.
5. Set `auth_requirement` and `auth_mode` truthfully. Providers with secrets or external login state must expose only credential-presence and setup state, not stored secrets.
6. Add `config_schema_ref` for provider setup fields. If a full schema-driven form is not implemented yet, point to the intended schema reference and keep the setup state honest.
7. Add `workflow_refs` only for workflows that are implemented or intentionally exposed as setup-needed/blocked.
8. Add `capability_flags` for search, stock observation, pricing, health, assistant, media capture, text capture, or other advertised capabilities.
9. Write `setup_instructions` as operator-safe guidance that names the next action without leaking tokens or private provider data.

## Add Workflow Actions

Workflow action definitions live next to provider manifests in `integrationWorkflowActionDefinitions`. Add or update a definition before adding its ID to a provider's `workflow_refs`.

Every workflow action must define:

- stable `id` / `workflow_ref`
- `label` and operator-safe `description`
- workflow `type`
- `input_schema` and `output_schema`
- `requires_auth` and `requires_secrets`
- capability list
- `side_effect_level` (`read_only`, `preview_only`, `write`, or `destructive`)
- `confirmation_required`
- `schedule_support`
- `inbox_events`
- `health_impact`
- `execution_mode`

Use `read_only` for inspection-only workflows, `preview_only` when Cabinet records reviewable local output before a later apply step, `write` for external or Cabinet state mutation, and `destructive` when an operation can end, remove, overwrite, or publish irreversible provider state. Mutating and destructive workflows must require explicit confirmation unless a narrower issue proves a safer contract.

Workflow failures and required user actions must advertise Inbox event metadata such as `workflow_failed`, `required_action`, `confirmation_pending`, or `result_inbox_updated` so UI, assistant, and automation consumers can route failures consistently. Do not add a provider `workflow_refs` entry for a workflow that has no registry definition.

## Add Setup Schemas

Setup schema definitions live next to provider manifests in `integrationConfigSchemaDefinitions`. Add or update a schema before pointing a provider manifest at its `config_schema_ref`.

Every setup schema must define:

- stable `schema_ref`
- `persistence_scope`
- `submit_target`
- optional `secret_target`
- optional `validate_action`
- ordered fields with `key`, `label`, `type`, `required`, `write_only`, and `persistence`

Use field types from the shared renderer vocabulary: `text`, `secret`, `url`, `number`, `select`, `multiselect`, `checkbox`, `textarea`, `file`, `oauth-connect`, and `browser-auth-status`. Fields may add placeholder text, helper text, default values, options, validation rules, documentation URLs, read-only state, and conditional rendering metadata.

Persist non-secret fields through `profile_settings` or provider instance state. Persist secrets through `profile_secrets`; write-only fields must expose only the field metadata, secret key alias, and credential-presence state. Never return a saved secret value in `setup_schema`, `setup_status`, provider health, or registry payloads.

At minimum, schema changes should keep these provider shapes covered:

- API key provider: a `secret` field with `profile_secrets` persistence and a validate/test action.
- Browser Auth provider: a read-only `browser-auth-status` field until Cabinet has verified callback/artifact and provider-test proof.
- No-auth/static source provider: non-secret fields plus validate or provider-family detect action, without requiring a secret target.

## Complete Example

Use this shape as the minimum reviewable example for a new provider. Replace the IDs and scopes with the real provider contract, then add targeted tests for the changed manifest, schema, action, UI projection, and persistence path.

```go
integrationProviderManifest{
	ProviderID:        "example-market",
	DisplayName:       "Example Market",
	BaseDomain:        "example.test",
	MarketWatchScope:  "example-market",
	ProviderCategory:  "marketplace",
	ProviderType:      "marketplace",
	APIFamily:         "official_api",
	APISupportProfile: "rest_v1",
	ActiveMode:        "official_api",
	IntegrationMode:   "official_api",
	APIAvailable:      true,
	AuthRequirement:   "api_key",
	AuthMode:          "api_key",
	ConfigSchemaRef:   "integrations/example-market/setup",
	WorkflowRefs:      []string{"market_watch.run", "example.preview_listing"},
	CapabilityFlags: map[string]bool{
		"search": true,
		"pricing": true,
		"health": true,
	},
	SetupInstructions: "Add an Example Market API key, validate health, then run Market Watch scans.",
}
```

```go
integrationConfigSchemaDefinition{
	SchemaRef:        "integrations/example-market/setup",
	PersistenceScope: "active_profile",
	SubmitTarget:     "/api/profiles/:profileId/settings",
	SecretTarget:     "/api/profiles/:profileId/secrets",
	ValidateAction:   "provider.test",
	Fields: []integrationConfigSchemaField{
		{Key: "example_market_region", Label: "Region", Type: "select", Required: true, Persistence: "profile_settings", Default: "AU", Options: []integrationConfigSchemaOption{{Value: "AU", Label: "Australia"}, {Value: "US", Label: "United States"}}},
		{Key: "example_market_base_url", Label: "API base URL", Type: "url", Required: true, Persistence: "profile_settings", Placeholder: "https://api.example.test", ValidationRules: []string{"url"}},
		{Key: "example_market_api_key", Label: "API key", Type: "secret", Required: true, WriteOnly: true, Persistence: "profile_secrets", SecretKey: "example_market_api_key"},
	},
}
```

```go
integrationWorkflowActionDefinition{
	ID:                   "example.preview_listing",
	Label:                "Preview listing",
	Description:          "Build a reviewable listing preview before any marketplace write.",
	Type:                 "marketplace_listing_preview",
	InputSchema:          "example.preview_listing.request.v1",
	OutputSchema:         "example.preview_listing.preview.v1",
	RequiresAuth:         true,
	RequiresSecrets:      true,
	Capabilities:         []string{"pricing", "listing_lifecycle"},
	SideEffectLevel:      "preview_only",
	ConfirmationRequired: true,
	ScheduleSupport:      "manual",
	InboxEvents:          []string{"workflow_failed", "required_action", "confirmation_pending"},
	HealthImpact:         "requires_ready_provider",
	ExecutionMode:        "provider_workflow",
}
```

## Consumer Contracts

These consumers must derive provider identity, category, auth/setup state, capabilities, and workflow/action state from the registry manifest:

- `GET /api/providers/registry`
- `GET /api/providers/:id/*`
- Add Integration provider list and details
- Market Watch provider projection
- scanner/provider family projections that advertise registry-backed providers

Consumer-specific payloads may project only the fields they need, but they must not invent provider identity, category, setup, or capability state outside the canonical manifest-backed registry.

## Config, Health, And Workflow Boundaries

The base registry manifest can name `config_schema_ref`, `workflow_refs`, capability flags, and setup guidance before the full implementation exists. The consumer must show unavailable, setup-needed, disabled, beta, degraded, or blocked states truthfully until the backing feature is implemented and validated.

Use focused child issues for the larger implementation areas:

- schema-driven setup form rendering and field persistence
- profile-scoped integration instances and secret storage
- provider validation and health snapshots
- workflow/action execution contracts
- Inbox-visible failure and required-action events
- provider-specific live credential or external write evidence

Do not claim live provider readiness from route navigation, a visible dialog, a toast, or stored text alone. Mutating provider work needs resulting state or provider/API evidence and explicit confirmation boundaries where required.

## Required Evidence

Each provider-authoring change should include the smallest useful set of evidence for the touched surface:

- OpenSpec requirement or traceability update for new provider contracts.
- Go/API contract coverage for registry payload, provider categories, capabilities, config schema references, workflow references, and secret redaction.
- Go/API setup schema coverage for API-key, Browser Auth, and No-auth/static provider shapes; current guard: `TestProviderRegistryProjectsConfigSchemaShapes`.
- UI template or Cypress coverage when Add Integration, provider details, or Market Watch projection behavior changes.
- Provider-specific API/fixture tests when a new adapter, parser, or live-provider workflow is added.
- `openspec validate --all --strict --no-interactive`.
- `git diff --check`.

## Security Checklist

- Secret setup fields use `Type: "secret"`, `WriteOnly: true`, `Persistence: "profile_secrets"`, and a stable `SecretKey`.
- Registry, setup status, health, last-run, Inbox event, and UI payloads expose credential presence only, never clear secret values.
- Mutating or destructive workflow actions set `ConfirmationRequired: true` and use `SideEffectLevel: "write"` or `"destructive"`.
- Provider setup instructions explain the next safe operator action without including private account, token, or customer data.
- Disabled, beta, unavailable, or credential-blocked providers keep actions unavailable until validated evidence exists.

## Add Integration UI Checklist

- The provider appears in the Add Integration selector from `/api/providers/registry`, including unconfigured providers.
- Provider rows/cards show display name, domain, category/type, auth/setup type, status, description or setup instructions, and key capabilities.
- Search and filters can match provider name, domain, category/type, auth type, capabilities, and status.
- Selecting the provider opens a schema-driven setup form; provider-specific fields must not render before explicit selection.
- Labels are visible and programmatically associated with setup inputs, including secret fields and read-only status fields.
- Disabled, deprecated, beta, setup-needed, and repair-needed states are visually distinct and do not expose unsafe actions.
- Save, validate, disable, repair, and workflow controls verify resulting state or API outcome rather than relying on toasts.

## Reference Tests

Current #1463 provider registry guards include:

- `internal/app/openspec_provider_specs_test.go` (`TestIntegrationRegistryOpenSpecCoversIssue1463ConsumerContract`, `TestIntegrationProviderAuthoringGuideCoversIssue1463Workflow`)
- `internal/app/integration_migration_regression_test.go` (`TestProviderRegistryProjectsCanonicalManifestCategories`, `TestProviderRegistryProjectsMarketWatchProviderScopes`)
- `internal/app/ui_template_contract_test.go` (`TestMarketWatchProviderControlsUseProviderRegistryContract`, `TestIntegrationsProviderDetailActionsUseProviderRegistryContract`)

Current #1468 authoring-guide closure guard:

- `internal/app/openspec_provider_specs_test.go` (`TestIntegrationProviderAuthoringGuideCoversIssue1468AcceptanceChecklist`)
