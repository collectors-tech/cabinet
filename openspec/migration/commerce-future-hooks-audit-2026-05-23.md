# Commerce Reconciliation and Future Hooks Audit

Issue: #844
Date: 2026-05-23

## Scope

This audit checks whether the documented commerce reconciliation and future-hooks contracts still represent hidden backlog, completed implementation, or deferred implementation scope.

## Evidence Checked

- `openspec/specs/commerce/reconciliation/spec.md`
- `openspec/specs/general/future-hooks/spec.md`
- `openspec/traceability.md`
- `internal/app/commerce_reconciliation_api_test.go`
- `internal/app/future_hooks_api_test.go`
- `internal/db/future_hooks_test.go`
- `internal/app/future_hooks.go`
- Closed future-hooks issue #195
- Open follow-up issues #838, #839, #840, and #843

## Findings

### Commerce reconciliation

`COMMERCE-RECONCILIATION-001`, `COMMERCE-RECONCILIATION-002`, and `COMMERCE-RECONCILIATION-003` are implemented for the first backend/API slice:

- `/api/commerce/lifecycle` persists profile-scoped lifecycle entries.
- Purchase lifecycle entries create expected-arrival records instead of inventory records directly.
- `/api/commerce/arrivals` exposes expected-arrival listing and reconciliation state transitions.
- Targeted proof exists in `internal/app/commerce_reconciliation_api_test.go`.

The current OpenSpec did not have direct traceability rows for these requirement IDs, so this audit adds them as implemented.

### Future hooks

`FUTURE-HOOKS-001` and `FUTURE-HOOKS-002` are implemented, not partial:

- `/api/future-hooks` lists disabled, non-operative scaffold hooks.
- `/api/future-hooks/invoke` returns an explicit `hook_not_active` response.
- `canonical_items` keeps disabled default fields for `for_sale` and structured offers.
- Targeted proof exists in `internal/app/future_hooks_api_test.go` and `internal/db/future_hooks_test.go`.
- Closed issue #195 records prior delivery and merge evidence for this same runtime/API closure unit.

The traceability matrix still marked both future-hooks requirements partial, so this audit updates those rows to implemented.

## Remaining Backlog Decision

No new implementation issue is needed for the audited first-slice requirements themselves.

Broader commerce work remains intentionally tracked in existing focused issues:

- #838: freight-forwarder package inbox
- #839: purchase-to-forwarder reconciliation matching
- #840: landed-cost allocation and consolidation planner
- #843: eBay buyer watchlist/saved/liked/cart-like interest state sync

Structured offers and for-sale behavior remain disabled future hooks. They should become implementation issues only when product scope requires active offer/sale workflows, because the current requirement is explicitly for disabled scaffolding and non-operative behavior.

## Result

#844 can be closed after validation because the confirmed hidden gaps were traceability drift, not missing implementation for the audited first-slice requirements.
