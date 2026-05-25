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

### Requirement API-DOCS-003: OpenAPI operations SHALL declare client error contracts
Every documented OpenAPI operation SHALL include at least one explicit `4XX` response contract so generated clients, validation gates, and API readers can distinguish success payloads from deterministic client-error envelopes.

#### Scenario: Validate operation error coverage
- **GIVEN** `docs/api/openapi.yaml` is the API documentation source of truth
- **WHEN** the OpenAPI validation gate inspects operation responses
- **THEN** every operation MUST declare at least one `4XX` response
- **AND** shared client-error responses MUST use the canonical error response schema where applicable
