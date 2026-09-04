## 1. Spec + Reproduction

- [x] 1.1 Capture the failing fresh-runtime startup evidence under current full validation flow.
- [x] 1.2 Define the fresh-runtime startup reliability contract in OpenSpec.

## 2. Runtime Startup Remediation

- [x] 2.1 Reduce fresh database migration/startup overhead in the real runtime path.
- [x] 2.2 Keep migration semantics safe for both fresh and existing databases.
- [x] 2.3 Verify targeted startup-bound tests pass without ad hoc per-run relaxations.

## 3. Validation + Unblock

- [x] 3.1 Re-run `go test ./internal/nfr ./tests -count=1 -v`.
- [x] 3.2 Re-run `go test ./... -count=1`.
- [x] 3.3 Re-run the blocked #446 validation chain once startup gates are green.
  - Reconciled in #1872 on 2026-07-11: #446 was closed after PR #465 merged to `develop` and demo2 was redeployed from merge commit `02aa7b4`.
- [x] 3.4 Update issue/PR evidence and close #448 only when the blocker is actually cleared.
  - Reconciled in #1872 on 2026-07-11: #448 was closed after PR #466 merged to `develop` and demo2 was redeployed from merge commit `05d0bd4`.
