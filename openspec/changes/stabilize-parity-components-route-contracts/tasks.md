## 1. Reproduce and isolate

- [x] 1.1 Reproduce the parity/components route-contract failures on a fresh 17882 branch build.
- [x] 1.2 Confirm the failures are stale route/surface expectations rather than product regressions.

## 2. Contract alignment

- [x] 2.1 Align parity specs with canonical `/dashboard` and `/settings/profile` targets.
- [x] 2.2 Align foundation components settings assertions with stable visible profile-form surfaces.
- [x] 2.3 Align degraded/recovery assertions with the current dashboard recovery shell.

## 3. Validation

- [x] 3.1 Re-run `ui-data-contract-parity/spec.cy.ts` on a fresh 17882 branch build.
- [x] 3.2 Re-run `ui-foundation-components/spec.cy.ts` on a fresh 17882 branch build.
- [ ] 3.3 Feed the result back into the broader regression gate.