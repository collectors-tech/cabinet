# OpenSpec Hardening Review Notes

## What was fixed
- Added deterministic IDs to every requirement heading missing IDs across all spec files.
- Replaced vague placeholder GIVEN preconditions with concrete actor/config/data preconditions.
- Added explicit API status-code acceptance bullets where API-trigger scenarios previously lacked status expectations.
- Regenerated traceability matrix to cover all requirement IDs with implemented test links or explicit planned TODO mappings.

## ID namespaces used
- Namespace format derived from capability folder names (e.g., `AI-ASSIST-001`, `UI-SCREEN-HOME-001`, `RUNTIME-CORE-001`).
- Existing IDs were preserved unchanged (e.g., `INTEGRATION-*`, `OPS-001`).

## Any deprecations
- No requirement IDs were deprecated in this pass.

## Remaining planned tests
- Traceability entries marked `planned` or `partial` represent requirement IDs without direct runtime/E2E assertion coverage yet.
- Priority planned gaps remain around extended provider contracts and UI workflow runtime validations.
