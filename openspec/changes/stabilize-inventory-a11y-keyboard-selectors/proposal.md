## Why

After clearing chats/auth/parity/components route-contract drift, the remaining broader regression blocker was a focused accessibility test miss in the inventory keyboard-only workflow. The product controls still worked; the test was anchored to one legacy exact placeholder string even though the inventory/tasks filter placeholder now varies by mode.

## What Changes

- Align the inventory keyboard-only accessibility contract with the current filter control selector surface.
- Keep the workflow coverage focused on keyboard execution behavior, not one exact placeholder string variant.

## Capabilities

### Modified Capabilities

- `ui-foundation-accessibility`: inventory keyboard-only workflow coverage SHALL target the stable filter control surface even when the placeholder copy varies by current view mode.

## Impact

- Affected test: `ui.web/cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts`
- Related issues: `#455`