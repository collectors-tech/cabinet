## Purpose
Define the canonical integrations/provider registry used by scanner, pricing, and integrations UI.

## Requirements
### Requirement: Provider registry SHALL define provider identity and capabilities
Cabinet SHALL maintain a provider registry with stable provider IDs, display metadata, auth mode, and capability flags (search, stock, pricing, media, health).

#### Scenario: Registry entry load
- **GIVEN** provider registry is initialized at runtime
- **WHEN** integrations workspace requests provider definitions
- **THEN** Cabinet SHALL return provider entries with ID, label, auth type, and capability flags

### Requirement: Registry SHALL include eBay and Amazon providers
Cabinet SHALL define provider entries for `ebay` and `amazon` with explicit capability and credential requirements.

#### Scenario: Core marketplaces available
- **GIVEN** provider registry is active
- **WHEN** integrations UI loads core marketplace catalog
- **THEN** entries for `ebay` and `amazon` SHALL be present

### Requirement: Registry SHALL include configured AU webshop providers
Cabinet SHALL include AU webshop providers from product scope:
- bonzaslotcars.com.au
- frontlinehobbies.com.au
- hobbytechtoys.com.au
- andrewshobbies.com.au
- voglers.com.au
- acercmodels.com
- mrtoys.com.au

#### Scenario: AU webshop catalog rendered
- **GIVEN** provider registry includes webshop providers
- **WHEN** user opens integrations provider list
- **THEN** all configured AU webshop domains SHALL be represented in registry metadata

### Requirement: Registry SHALL map provider specs
Each provider entry SHALL map to a provider-specific OpenSpec capability.

#### Scenario: Provider traceability
- **GIVEN** provider registry spec is reviewed
- **WHEN** a provider entry is selected for implementation
- **THEN** a mapped provider spec SHALL exist:
  - `provider-ebay`
  - `provider-amazon`
  - `provider-au-webshops`

