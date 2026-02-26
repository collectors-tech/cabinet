## Why

Recent user-visible regressions include inventory route `500` errors and unstable Cypress startup behavior on Windows. These failures block reliable delivery and allow critical defects to reach runtime without guardrails.

## What Changes

- Establish deterministic non-500 behavior for inventory and wishlist routes across empty, seeded, and large datasets.
- Standardize startup behavior for Cypress on Windows so local/CI validation is reliable.
- Add/upgrade regression coverage for known inventory failure paths.
- Align issue completion criteria with required stability evidence.

## Capabilities

### New Capabilities

- `inventory-runtime-reliability`: Non-500 route behavior and error-state handling requirements for inventory and wishlist workflows.
- `cypress-windows-startup`: Deterministic Cypress startup contract for Windows environments.

### Modified Capabilities

- None.

## Impact

- Affected code: `ui.web` inventory and wishlist routes/components, Cypress scripts/config, related API boundary handling.
- Affected tests: Cypress E2E stability and regression suites.
- Related issues: `#151`, `#149`, `#147`, plus stability aspects of `#154`.
