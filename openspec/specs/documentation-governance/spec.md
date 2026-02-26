## Purpose
Define documentation source-of-truth, migration lock-in policy, and legacy artifact handling.

## Requirements

### Requirement: OpenSpec Is The Normative Documentation Source
Cabinet SHALL treat `openspec/specs/*/spec.md` as the normative source for product, UI, API behavior, and acceptance criteria.

#### Scenario: Normative policy is explicit
- **WHEN** contributors reference requirements
- **THEN** they SHALL implement from OpenSpec specs, not legacy markdown notes

### Requirement: Legacy Docs Directory Contains No Markdown Sources
The `docs/` directory SHALL contain no `*.md` files. Legacy markdown content SHALL be migrated into canonical OpenSpec specs and then removed.

#### Scenario: Legacy markdown is removed after canonical migration
- **WHEN** documentation migration runs
- **THEN** markdown files from `docs/**/*.md` are absorbed into OpenSpec canonical specs
- **AND** `docs/` has zero markdown files

### Requirement: OpenAPI Contract Remains in docs/api/openapi.yaml
The runtime API contract SHALL remain at `docs/api/openapi.yaml` for server and docs tooling compatibility.

#### Scenario: OpenAPI contract remains stable
- **WHEN** API parity tests execute
- **THEN** they SHALL read `docs/api/openapi.yaml`
- **AND** documentation migration SHALL NOT relocate that file

### Requirement: Legacy Migration Mapping Is Preserved
A migration mapping SHALL preserve the source-to-archive path lineage for every moved markdown file.

#### Scenario: Key legacy files are traceable
- **WHEN** reviewing migration lineage
- **THEN** the following source paths are traceable in archived form:
  - `docs/FULL_FEATURE_LIST.md`
  - `docs/SPEC.md`
  - `docs/USE_CASES_AND_SCENARIOS.md`
  - `docs/ui-spec/02-SCREEN-SPECS.md`
  - `docs/ui-spec/05-TEST-MATRIX-UI.md`

## Migration Inventory (Source -> Canonical)
- `docs/FULL_FEATURE_LIST.md` -> domain capability specs under `openspec/specs/*`
- `docs/SPEC.md` -> domain capability specs under `openspec/specs/*`
- `docs/USE_CASES_AND_SCENARIOS.md` -> UC mappings in `openspec/specs/README.md` + screen specs
- `docs/UI_ENDPOINT_PARITY.md` -> `openspec/specs/ui-data-contract-parity/spec.md`
- `docs/ui-spec/*.md` -> `openspec/specs/ui-foundation-*` and `openspec/specs/ui-screen-*`
- `docs/auth/CLERK_BILLING_SETUP.md` -> `openspec/specs/cloud-auth-billing/spec.md`
