## Why

The broader regression handoff after #451 surfaced a focused stale-route contract cluster in general Cypress specs. The product routes were correct; the tests were still asserting legacy `/` and `/settings/` targets plus a few stale recovery/profile-surface assumptions.

## What Changes

- Align general parity/components Cypress contracts with the canonical Cabinet routes.
- Keep dashboard degraded/recovery assertions stable against the current degraded-state shell.
- Keep settings profile component assertions anchored to stable visible form controls rather than responsive-only or localization-sensitive surfaces.

## Capabilities

### Modified Capabilities

- `ui-data-contract-parity`: authenticated parity coverage SHALL use canonical home/settings routes.
- `ui-foundation-components`: settings component coverage SHALL assert stable visible profile-form surfaces on the canonical settings profile route.

## Impact

- Affected tests:
  - `ui.web/cypress/e2e/general/ui-data-contract-parity/spec.cy.ts`
  - `ui.web/cypress/e2e/general/ui-foundation-components/spec.cy.ts`
- Related issues: `#453`