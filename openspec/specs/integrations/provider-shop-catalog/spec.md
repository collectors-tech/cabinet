## Purpose
Define supported shop-provider catalog and provider API capability classification.

## Requirements
### Requirement INTEGRATION-018: Provider catalog SHALL include configured collector shop domains
Cabinet SHALL include configured shop providers in registry/catalog for discovery and pricing workflows.

#### Scenario: Load shop provider catalog
- **GIVEN** provider registry is initialized for active profile
- **WHEN** user opens Integrations provider catalog
- **THEN** catalog MUST include configured domains, including:
  - `bonzaslotcars.com.au`
  - `frontlinehobbies.com.au`
  - `hobbytechtoys.com.au`
  - `andrewshobbies.com.au`
  - `voglers.com.au`
  - `acercmodels.com`
  - `mrtoys.com.au`
  - `hobbyco.com.au`
  - `metrohobbies.com.au`

### Requirement INTEGRATION-019: Provider entries SHALL declare acquisition mode and API availability
Cabinet SHALL classify each provider entry with integration method and official API availability status.

#### Scenario: Provider capability classification
- **GIVEN** registry provider record exists
- **WHEN** provider detail is requested
- **THEN** response MUST include:
  - `integration_mode` (`official_api|web_ingestion|manual`)
  - `api_available` (boolean)
  - `auth_requirement` (`none|api_key|oauth|unknown`)
