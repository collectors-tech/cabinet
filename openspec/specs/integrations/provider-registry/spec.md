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
