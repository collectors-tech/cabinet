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
- Current unresolved coverage gap summary: `traceability-partial-ids=90`.

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

## Wave 2 evidence (runtime/diagnostics/security)
- Wave completed with missing runtime behavior implemented and directly tested.
- IDs moved to implemented:
  - `CLOUD-AUTH-BILLING-005`
  - `DIAGNOSTICS-002`, `DIAGNOSTICS-003`, `DIAGNOSTICS-004`
  - `ERRORS-003`
  - `SECURITY-001`
- Net reduction: partial IDs `144 -> 138` (6 reduced).
- Security policy decision for plaintext fallback:
  - API secrets continue to prefer OS keyring.
  - If explicit insecure fallback is enabled, values are now encrypted-at-rest (`enc:v1:` AES-GCM) and never stored plaintext in SQLite.
  - Rationale: preserves local development/test operability while eliminating plaintext persistence risk.

## Wave 3 evidence (cloud gating/runtime update)
- Wave completed with direct runtime API contract proofs for remaining high-priority blockers.
- IDs moved to implemented:
  - `CLOUD-AUTH-BILLING-004`
  - `RUNTIME-CORE-002`
- Net reduction: partial IDs `138 -> 136` (2 reduced).
- Runtime behavior implemented:
  - Pro-gated mutation endpoints now enforce deterministic `403` response with consistent envelope (`error`, `error_code`, `feature`, `message`) when resolved entitlement is not pro.
  - Added runtime signed-update harness endpoint `/api/runtime/update/install` with deterministic signature and channel validation outcomes.

## Wave 4 evidence (provider/scanner contracts)
- Wave completed with runtime provider/scanner contract routes plus direct API tests.
- IDs moved to implemented:
  - `CANDIDATES-001`, `CANDIDATES-002`
  - `INTEGRATION-008`, `INTEGRATION-009`, `INTEGRATION-012`, `INTEGRATION-013`, `INTEGRATION-014`, `INTEGRATION-015`
  - `SCANNER-002`
- Net reduction: partial IDs `136 -> 127` (9 reduced).
- Runtime behavior implemented:
  - Added canonical provider registry endpoint `/api/providers/registry` with Amazon integration-mode metadata and AU webshop domains.
  - Added Amazon provider run endpoint `/api/providers/amazon/run` with deterministic `409 PROVIDER_DISABLED` envelope when disabled and normalized candidate contract when enabled.
  - Added AU webshop stock parser endpoint `/api/providers/au-webshops/parse-stock` returning deterministic `stock_signal` contract.
  - Expanded scheduled scanner run contract to return summary envelope fields: `run_id`, `query_sets_executed`, `candidates_collected`, `failures`.
- Test evidence:
  - `internal/app/traceability_wave4_provider_scanner_test.go`

## Wave 5 evidence (collection/discovery/pricing/matching workflows)
- Wave completed with API-backed workflow contract tests and no runtime code remediation required.
- IDs moved to implemented:
  - `SEARCH-001`, `SEARCH-002`
  - `DISCOVERY-001`, `DISCOVERY-002`
  - `WISHLIST-PRICING-DASHBOARD-001`, `WISHLIST-PRICING-DASHBOARD-002`, `WISHLIST-PRICING-DASHBOARD-003`, `WISHLIST-PRICING-DASHBOARD-004`
  - `MATCHING-001`, `MATCHING-002`
- Net reduction: partial IDs `127 -> 117` (10 reduced).
- Test evidence:
  - `internal/app/traceability_wave5_collection_workflows_test.go`

## Wave 6 evidence (profiles/collection/data/logging/scanner contracts)
- Wave completed with deterministic API tests for profile isolation, collection domain flows, data-management maintenance, scanner retry, and logging export/redaction contracts.
- IDs moved to implemented:
  - `PROFILES-001`
  - `COLLECTION-DOMAIN-001`, `COLLECTION-DOMAIN-002`, `COLLECTION-DOMAIN-003`
  - `SCANNER-001`, `SCANNER-003`
  - `DATA-MANAGEMENT-001`, `DATA-MANAGEMENT-002`
  - `LOGGING-003`, `LOGGING-004`
- Net reduction: partial IDs `117 -> 107` (10 reduced).
- Test evidence:
  - `internal/app/traceability_wave6_ops_contracts_test.go`

## Wave 7 evidence (ai/chat/media/lookup/settings/licensing contracts)
- Wave completed with deterministic API contract coverage for AI availability/toggle, chat persistence + confirmation flow, media upload/reorder derivatives, barcode local+external lookup behavior, entitlement/license gates, and settings credential persistence.
- IDs moved to implemented:
  - `AI-ASSIST-001`, `AI-ASSIST-002`, `AI-ASSIST-004`
  - `CHAT-COPILOT-002`, `CHAT-COPILOT-003`, `CHAT-COPILOT-004`
  - `PHOTOS-MEDIA-001`, `PHOTOS-MEDIA-002`, `PHOTOS-MEDIA-003`
  - `BARCODES-001`, `BARCODES-002`, `LOOKUP-001`
  - `ENTITLEMENTS-001`
  - `LICENSING-001`, `LICENSING-002`
  - `SETTINGS-001`, `SETTINGS-002`
- Net reduction: partial IDs `107 -> 90` (17 reduced).
- Test evidence:
  - `internal/app/traceability_wave7_ai_settings_lookup_test.go`
  - `internal/app/chat_api_test.go`
  - `internal/app/photos_api_test.go`
  - `internal/app/license_api_test.go`
