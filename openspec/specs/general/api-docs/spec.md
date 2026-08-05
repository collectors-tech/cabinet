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

### Requirement API-DOCS-004: OpenAPI source SHALL pass Redocly merge-gate validation
Cabinet's OpenAPI source SHALL use OpenAPI 3.1-compatible schema forms and YAML syntax that Redocly can parse without structural lint failures.

#### Scenario: Validate OpenAPI source with Redocly
- **GIVEN** `docs/api/openapi.yaml` is the API documentation source of truth
- **WHEN** the Redocly lint and docs-build gates run against the document
- **THEN** nullable response fields MUST use OpenAPI 3.1 schema representations accepted by Redocly
- **AND** inline descriptions MUST be represented so punctuation in prose is parsed as description text rather than object properties

### Requirement API-DOCS-005: Release gates SHALL enforce complete runtime route parity
Every shipped runtime API route and HTTP method SHALL have one matching OpenAPI operation, and every documented operation SHALL map back to a shipped runtime route.

#### Scenario: Validate runtime and OpenAPI inventories
- **GIVEN** Cabinet registers static routes and parameterised route families in the Go runtime
- **WHEN** the OpenAPI parity suite runs in the develop, release-candidate, or main gate
- **THEN** the suite MUST compare runtime paths and HTTP methods against `docs/api/openapi.yaml` in both directions
- **AND** each operation MUST have a unique `operationId` and an explicit `4XX` response
- **AND** E2E-only routes MAY be excluded only through a reviewed exclusion entry with a reason
- **AND** the gate MUST fail when the named parity suite executes zero tests

### Requirement API-DOCS-006: OpenAPI SHALL distinguish Cabinet security boundaries
The API contract SHALL identify routes that are intentionally public, routes that accept an unlocked local Cabinet session or a Cabinet OIDC session, and routes that require a Browser Companion profile credential.

#### Scenario: Inspect protected companion and provider operations
- **GIVEN** a client reads the OpenAPI source
- **WHEN** it inspects companion ingestion, integration-instance, Frontline, Bonza, and Hobbytech operations
- **THEN** companion session, module discovery, capture and media operations MUST declare their paired profile bearer requirement
- **AND** companion pairing request/exchange operations MUST explicitly declare their one-time unauthenticated exchange boundary while approval/session-management operations declare Cabinet-session protection
- **AND** integration-instance and provider run operations MUST declare local-session and OIDC-session alternatives
- **AND** intentionally public health, runtime discovery and OpenAPI operations MUST remain explicitly unauthenticated
