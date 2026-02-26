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
  - `auth_mode` (`none|oauth|api_key|hybrid`)
  - `capabilities.search` (boolean)
  - `capabilities.stock_observation` (boolean)
  - `capabilities.pricing` (boolean)
  - `capabilities.health` (boolean)
  - `state` (`ready|degraded|disabled`)

### Requirement INTEGRATION-002: Registry MUST include eBay and Amazon providers
Cabinet SHALL define provider entries for `ebay` and `amazon` with explicit capability and credential requirements.

#### Scenario: Core marketplaces available
- **GIVEN** provider registry is active for current runtime
- **WHEN** integrations client loads provider list
- **THEN** entries for `ebay` and `amazon` MUST exist with non-empty `provider_id` and `integration_mode`

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
- **GIVEN** current runtime loads AU webshop provider family
- **WHEN** `GET /api/providers/registry` returns provider entries
- **THEN** all configured domains MUST be represented in `base_domain` field
- API outcome MUST be explicit: `200` on success, `4xx` for validation/auth conflicts, and `5xx` for unexpected failures

### Requirement INTEGRATION-004: Registry entries MUST map to provider capability specs
Each provider entry SHALL map to a provider-specific OpenSpec capability.

#### Scenario: Provider traceability
- **GIVEN** migration review for provider contracts is executed
- **WHEN** provider entry is selected for build/testing
- **THEN** mapped provider specs MUST exist:
  - `provider-ebay`
  - `provider-amazon`
  - `provider-au-webshops`
