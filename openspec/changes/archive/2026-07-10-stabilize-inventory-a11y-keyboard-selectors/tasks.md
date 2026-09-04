## 1. Reproduce and isolate

- [x] 1.1 Reproduce the accessibility keyboard-flow selector miss on a fresh 17882 branch build.
- [x] 1.2 Confirm the failure is placeholder-selector drift rather than a real accessibility regression.

## 2. Contract alignment

- [x] 2.1 Align the inventory filter selector with the stable placeholder prefix used across current view modes.

## 3. Validation

- [x] 3.1 Re-run `ui-foundation-accessibility/spec.cy.ts` on a fresh 17882 branch build.
- [x] 3.2 Feed the result back into the broader regression gate.
  - Reconciled in #1872 on 2026-07-11: the targeted selector fix was already merged and subsequent Develop Quality Gate runs for #1882, #1883, and #1884 completed successfully; packaged end-to-end release acceptance remains tracked by #1869.
