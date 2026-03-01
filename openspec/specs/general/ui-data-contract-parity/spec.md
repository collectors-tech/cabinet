## Purpose
Define strict UI-to-API parity requirements so every production screen is fully backed by Cabinet contracts and testable states.

## Requirements
### Requirement UI-DATA-CONTRACT-PARITY-001: Every authenticated screen SHALL map to explicit API contracts
Each top-level authenticated screen SHALL declare the APIs it depends on for read and mutation flows.

#### Scenario: Screen parity review
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** a screen is reviewed for release readiness
- **THEN** its required API routes SHALL be explicitly mapped and validated

### Requirement UI-DATA-CONTRACT-PARITY-002: Screen data flows SHALL support loading, empty, error, and ready states
For each API-backed screen section, UI SHALL define deterministic behavior for loading, empty, error, and ready states.

#### Scenario: API error on screen load
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** required API call fails during screen load
- **THEN** screen SHALL render actionable error state without collapsing into generic fatal route failure

### Requirement UI-DATA-CONTRACT-PARITY-003: Mutating actions SHALL map to endpoint-level success and failure outcomes
Each mutation control SHALL declare request payload contract and expected success/failure rendering behavior.

#### Scenario: Mutation failure handling
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** mutation endpoint returns failure
- **THEN** UI SHALL preserve current context and render inline actionable error feedback

### Requirement UI-DATA-CONTRACT-PARITY-004: Endpoint parity SHALL include E2E verification mapping
Each critical endpoint-screen mapping SHALL map to at least one automated E2E validation case (existing or planned).

#### Scenario: Endpoint parity audit
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** endpoint parity is audited
- **THEN** each critical endpoint SHALL reference a Cypress spec path and test identifier

## Acceptance Criteria
1. All top-level screens (dashboard, collection, scanner, discoveries, ai, barcodes, photos, pricing, reports, settings) have explicit endpoint mapping coverage.
2. Settings, users, integrations, and chat routes do not rely on sample/template data in production paths.
3. Endpoint-level state behavior exists for loading/empty/error/ready.
4. Parity matrix is enforceable in PR review and issue closure.

## Success Criteria
1. No production screen route fails due to missing API wiring.
2. Known parity regressions are caught by automated tests before merge.
3. Engineers can trace each UI action to API contract without external tribal knowledge.

## E2E Mapping Requirements
Minimum required parity suites:
- `ui-routing-and-nav`:
  - validates route accessibility and screen mount behavior
- `ui-endpoint-parity-core`:
  - validates read and mutation paths for each top-level screen
- `ui-settings-users-chat-parity`:
  - validates audits for users/settings/chat API wiring
- `ui-regression-non-500`:
  - validates critical non-500 screen behavior for inventory/wishlist and major routes

Each mapped test must include:
- screen identifier
- endpoint(s) exercised
- expected state assertion(s)

## Data Profiles
### Sample parity profile
- 1 active profile
- 5 items with instances/photos/barcodes
- 2 query sets
- 3 discoveries
- 3 wishlist rows
- 1 tracked pricing row

### Bulk parity profile
- 10,000 items
- 50,000 instances
- 200,000 scanner candidates
- 2,000 tracked pricing rows

Bulk profile SHALL be used for parity behavior under table/grid-heavy screens to ensure API contracts remain stable under scale.

## Source Mapping
- Legacy endpoint parity notes are normalized into this capability.
