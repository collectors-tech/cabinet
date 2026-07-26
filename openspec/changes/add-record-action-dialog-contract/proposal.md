## Why

#1938 tracks inconsistent table row actions and CRUD dialogs across Cabinet
surfaces. Current pages mix kebab menus, standalone icon buttons, inline edit
actions, bespoke destructive confirmations, and page-specific dialog behavior.
That makes keyboard behavior, focus return, destructive copy, loading states,
and capability-gated actions hard to verify or migrate consistently.

## What Changes

- Define one shared, capability-driven record action menu contract for Cabinet
  table rows.
- Standardize action order, labels, destructive grouping, disabled/omitted
  unsupported operations, row-click isolation, and accessible trigger behavior.
- Define a shared create/edit dialog contract for titles, descriptions, icons,
  field validation, server errors, cancel/close handling, dirty-state
  protection, loading state, double-submit prevention, focus trap, initial
  focus, and focus return.
- Define a shared destructive confirmation contract that names the record and
  distinguishes soft delete/archive from permanent deletion.
- Publish `surface-action-matrix.md` as the baseline for current Inventory,
  Wishlist, Users, Collections, Media, Purchases, Market Watch, Discoveries,
  Integrations, and Settings row-action behavior before broad page migration
  starts.

## Capabilities

### Modified Capabilities

- `ui-foundation-components`: adds the reusable record action menu and CRUD
  dialog contracts.
- `ui-foundation-interactions`: adds keyboard, pointer, focus, row-navigation,
  disabled-action, loading, error, and destructive-confirmation interaction
  requirements.

## Impact

- Affected code: `ui.web/src/components`, `ui.web/src/components/data-table`,
  existing feature row action/dialog copies under `ui.web/src/features/*`.
- Affected tests: component tests for record action menu and dialog states,
  plus focused Cypress migration tests when individual pages adopt the shared
  contract.
- Affected documentation: OpenSpec specs, traceability, and the #1938
  surface/action matrix.
- Related issues: `#1938`, `#1939`, `#1940`, `#1941`.
