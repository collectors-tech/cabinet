## Why

The broader regression chain surfaced dashboard home-screen failures that turned out to be a real product-side locale bug. The English pages bundle was missing the dashboard keys used by the current home screen, causing raw `dashboard.*` translation keys to leak into the UI and break the dashboard shell copy contract.

## What Changes

- Restore the missing English dashboard pages translations.
- Keep the dashboard actionable/loading/error/recent-items shell aligned with the current UI contract.

## Capabilities

### Modified Capabilities

- `dashboard-home-copy`: the English dashboard home shell SHALL render the expected actionable/loading/error/recent-items copy instead of raw translation keys.

## Impact

- Affected product file: `ui.web/src/locales/en/pages.json`
- Affected test: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts`
- Related issues: `#457`