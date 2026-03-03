# Scanner / Market Watch Split (Spec Decision)

## Decision
- Rename provider query-run module from **Scanner** to **Market Watch**.
- Reserve **Scanner** for camera/photo card recognition workflows.

## Contract split
- `integrations/ui-screen-market-watch` => provider-scoped query sets, run-now, provider-attributed results/errors.
- `scanner/ui-screen-card-scanner` => capture/upload recognition flow, confidence/candidates, confirm-before-write.

## Issue bindings
- #270 Rename Scanner to Market Watch
- #271 Provider selector + provider-scoped query sets
- #269 Run Now raw error messaging

## Notes
Keep backward-compatible route behavior documented during transition (temporary aliases allowed).
