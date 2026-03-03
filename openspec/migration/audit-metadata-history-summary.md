# Audit Metadata + History Summary (#232)

## Scope
- Issue: #232
- Spec IDs: AUDIT-METADATA-001, AUDIT-METADATA-002, AUDIT-HISTORY-001, AUDIT-HISTORY-002, AUDIT-HISTORY-003
- Spec path: `openspec/specs/general/audit-metadata-history/spec.md`

## Changes
- Added item metadata fields: `created_by`, `updated_by`, `deleted_at`, `deleted_by`.
- Added append-only `audit_events` storage and entity timeline query support in collection repository.
- Logged audit events for item create, update, soft-delete/status transitions, restore, and permanent delete.
- Captured deterministic before/after payloads for tracked fields in update/status mutations.
- Added repository test proving metadata persistence, append-only history, diff capture, and timeline ordering.

## Commands Run
- `go test ./internal/collection -count=1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/inventory/folder-tree-control/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- `go test ./internal/collection -count=1`: **pass**
- Cypress inventory tree suite: **9 passing / 0 failing**
- `go test ./internal/app -count=1`: **pass**
- `go test ./tests -count=1`: **pass**
- `openspec validate --all`: **pass**

## Outcome
- Core item lifecycle now persists actor/timestamp metadata and writes immutable audit timeline events queryable by entity.
