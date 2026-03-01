## Purpose
Define documentation source-of-truth, migration lock-in policy, and legacy artifact handling.

## Requirements

### Requirement DOCUMENTATION-GOVERNANCE-001: OpenSpec Is The Normative Documentation Source
Cabinet SHALL treat `openspec/specs/*/spec.md` as the normative source for product, UI, API behavior, and acceptance criteria.

#### Scenario: Normative policy is explicit
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** contributors reference requirements
- **THEN** they SHALL implement from OpenSpec specs, not legacy markdown notes

### Requirement DOCUMENTATION-GOVERNANCE-002: Legacy Docs Directory Contains No Markdown Sources
The `docs/` directory SHALL contain no `*.md` files. Legacy markdown content SHALL be migrated into canonical OpenSpec specs and then removed.

#### Scenario: Legacy markdown is removed after canonical migration
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** documentation migration runs
- **THEN** markdown files from `docs/**/*.md` are absorbed into OpenSpec canonical specs
- **AND** `docs/` has zero markdown files

### Requirement DOCUMENTATION-GOVERNANCE-003: OpenAPI Contract Remains in docs/api/openapi.yaml
The runtime API contract SHALL remain at `docs/api/openapi.yaml` for server and docs tooling compatibility.

#### Scenario: OpenAPI contract remains stable
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** API parity tests execute
- **THEN** they SHALL read `docs/api/openapi.yaml`
- API outcome MUST be explicit: `200` on success, `4xx` for validation/auth conflicts, and `5xx` for unexpected failures
- **AND** documentation migration SHALL NOT relocate that file

### Requirement DOCUMENTATION-GOVERNANCE-004: Legacy Migration Mapping Is Preserved
A migration mapping SHALL preserve file-by-file lineage and requirement-marker counts for every moved markdown file.

#### Scenario: Key legacy files are traceable
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** reviewing migration lineage
- **THEN** the following source paths are traceable in archived form:
  - `docs/FULL_FEATURE_LIST.md`
  - `docs/SPEC.md`
  - `docs/USE_CASES_AND_SCENARIOS.md`
  - `docs/ui-spec/02-SCREEN-SPECS.md`
  - `docs/ui-spec/05-TEST-MATRIX-UI.md`

### Requirement DOCUMENTATION-GOVERNANCE-005: File-By-File Migration Audit Is Maintained
Cabinet SHALL maintain a strict per-file migration audit at `openspec/migrations/legacy-docs-file-audit.yaml` covering every legacy docs markdown file from baseline commit `82294546bf0b715fe49394e1c5a885d3045294d2`.

#### Scenario: Migration audit covers entire baseline
- **GIVEN** the baseline commit and docs markdown inventory are known
- **WHEN** migration completeness is validated
- **THEN** each baseline markdown source SHALL exist in the audit with:
  - source path
  - migration status (`migrated` or `reference_only`)
  - requirement marker count
  - one or more target canonical locations

## Migration Inventory (Source -> Canonical)
- `docs/FULL_FEATURE_LIST.md` -> domain capability specs under `openspec/specs/*`
- `docs/SPEC.md` -> domain capability specs under `openspec/specs/*`
- `docs/USE_CASES_AND_SCENARIOS.md` -> UC mappings in `openspec/specs/README.md` + screen specs
- `docs/UI_ENDPOINT_PARITY.md` -> `openspec/specs/general/ui-data-contract-parity/spec.md`
- `docs/ui-spec/*.md` -> `openspec/specs/ui-*/spec.md`
- `docs/auth/CLERK_BILLING_SETUP.md` -> `openspec/specs/general/cloud-auth-billing/spec.md`


