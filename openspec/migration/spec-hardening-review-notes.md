# OpenSpec Hardening Review Notes

- canonical hardening commit: `9ca1215`
- compared-against baseline: `d0fd84c`
- validation snapshot: `openspec validate --all` => `57 passed, 0 failed`
- phase-2 consistency refresh: `2026-02-27T01:24:40.8642210Z`

## What was fixed
- Added deterministic IDs to every requirement heading previously missing IDs.
- Replaced vague placeholder GIVENs with concrete preconditions (actor, profile/config, fixture data).
- Added explicit API status semantics to API-trigger scenarios lacking status-code outcomes.
- Regenerated global traceability mapping for all requirement IDs with implemented or planned test links.

## ID namespaces used
- Capability-derived prefixes from each spec folder (e.g., AI-ASSIST-*, RUNTIME-CORE-*, UI-SCREEN-HOME-*).
- Existing IDs kept unchanged (notably INTEGRATION-*, OPS-001).

## Any deprecations
- None in this pass.

## Remaining planned tests
- IDs marked planned/partial in openspec/traceability.md still require direct runtime/API/E2E test proof.
- Provider and selected UI workflow IDs have explicit TODO test mappings pending implementation.
- Current unresolved coverage gap summary: `traceability-partial-ids=144`.

## Wave 1 evidence (auth/security/error)
- Wave completed with direct runtime/API/UI contract tests and traceability status updates.
- IDs moved to implemented:
  - `AUTH-001`, `AUTH-002`, `AUTH-003`
  - `CLOUD-AUTH-BILLING-001`, `CLOUD-AUTH-BILLING-002`, `CLOUD-AUTH-BILLING-003`
  - `DIAGNOSTICS-001`
  - `ERRORS-001`, `ERRORS-002`
  - `RUNTIME-CORE-001`, `RUNTIME-CORE-003`
  - `SECURITY-002`
- Net reduction: partial IDs `156 -> 144` (12 reduced).
- Remaining notable blockers in this wave scope:
  - `CLOUD-AUTH-BILLING-005` still partial pending strict 401/403 plus explicit non-mutation proof path.
  - `DIAGNOSTICS-002/003/004` remain partial due missing remote telemetry feature-test harness.
  - `ERRORS-003` remains partial pending deterministic taxonomy-to-guidance UI/API proof.
