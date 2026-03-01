## Purpose
Define runtime API documentation endpoints and OpenAPI source contract.

## Requirements
### Requirement API-DOCS-001: Runtime SHALL expose OpenAPI document endpoint
Cabinet SHALL expose OpenAPI YAML at a stable runtime endpoint.

#### Scenario: Request OpenAPI YAML
- **GIVEN** runtime is healthy
- **WHEN** client requests `/api/openapi.yaml`
- **THEN** endpoint MUST return `200` with content-type `application/yaml` or `text/yaml`

### Requirement API-DOCS-002: Runtime SHALL expose human-readable API docs route
Cabinet SHALL expose a user-facing API docs page at `/apidocs`.

#### Scenario: Open API docs page
- **GIVEN** runtime is healthy and OpenAPI YAML is available
- **WHEN** user opens `/apidocs`
- **THEN** docs UI MUST render without `404` and MUST load schema from `/api/openapi.yaml`
