## 1. Reproduce and isolate

- [x] 1.1 Reproduce the dashboard home copy failures on a fresh 17882 branch build.
- [x] 1.2 Confirm the UI is leaking raw translation keys because the English dashboard locale entries are missing.

## 2. Product fix

- [x] 2.1 Restore the missing English dashboard locale keys used by the current home screen.

## 3. Validation

- [x] 3.1 Re-run `dashboard/ui-screen-home/spec.cy.ts` on a fresh 17882 branch build.
- [ ] 3.2 Feed the result back into the broader regression gate.