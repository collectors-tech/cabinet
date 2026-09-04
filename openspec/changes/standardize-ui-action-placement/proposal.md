## Why

#1941 tracks drift in where Cabinet puts page, table, row, and dialog actions.
Some pages expose whole-page actions in the global header, while Reports and
Market Watch duplicate title blocks and keep important actions in local content
regions. Other surfaces mix filters, bulk actions, row operations, and dialog
confirmation controls in bespoke locations.

The result is harder scanning, duplicated controls, and inconsistent keyboard
and screen-reader expectations.

## What Changes

- Define one authenticated-shell action placement contract for Cabinet pages.
- Publish an action-placement matrix covering page, table/list, bulk, row, and
  dialog action regions for authenticated routes before code migration.
- Keep whole-page actions in the global page header.
- Keep filters, search, sort, view switches, selection, and bulk actions in the
  relevant table/list toolbar.
- Keep single-record operations in the shared row action menu contract from
  #1938/#1939.
- Keep cancel/confirm/apply controls in dialog footers for the active dialog.
- Remove duplicate visible title/description/action blocks once parity is
  proven by tests.
- Preserve global utility actions such as language, theme, configuration, and
  profile as shell utilities instead of page actions.

## Capabilities

### Modified Capabilities

- `ui-foundation-components`: adds canonical action regions and page-action
  header affordance requirements.
- `ui-foundation-interactions`: adds responsive overflow, keyboard,
  screen-reader, focus, and duplicate-action requirements for action movement.

## Impact

- Affected code: authenticated route headers, page feature action blocks,
  table/list toolbars, shared row action components, and dialog footers under
  `ui.web/src`.
- Affected tests: focused Cypress coverage for action regions, representative
  responsive header overflow, and packaged shell contract checks.
- Affected documentation: OpenSpec specs, traceability, and the #1941
  action-placement matrix.
- Related issues: `#1938`, `#1939`, `#1940`, `#1941`.
