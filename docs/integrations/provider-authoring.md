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
- UI template or Cypress coverage when Add Integration, provider details, or Market Watch projection behavior changes.
- Provider-specific API/fixture tests when a new adapter, parser, or live-provider workflow is added.
- `openspec validate --all --strict --no-interactive`.
- `git diff --check`.

## Reference Tests

Current #1463 provider registry guards include:

- `internal/app/openspec_provider_specs_test.go` (`TestIntegrationRegistryOpenSpecCoversIssue1463ConsumerContract`, `TestIntegrationProviderAuthoringGuideCoversIssue1463Workflow`)
- `internal/app/integration_migration_regression_test.go` (`TestProviderRegistryProjectsCanonicalManifestCategories`, `TestProviderRegistryProjectsMarketWatchProviderScopes`)
- `internal/app/ui_template_contract_test.go` (`TestMarketWatchProviderControlsUseProviderRegistryContract`, `TestIntegrationsProviderDetailActionsUseProviderRegistryContract`)
