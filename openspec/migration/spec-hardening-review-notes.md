# OpenSpec Hardening Review Notes

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
