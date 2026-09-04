## Why

#1940 tracks route metadata drift across the authenticated Cabinet shell. Page
headers, sidebar navigation, command/search navigation, and browser document
titles currently keep overlapping names and icons. That has already left
`/purchases` without a specific browser title, `/scanner` using `Scanner` while
the UI says `Market Watch`, and Settings child routes with generic titles.

## What Changes

- Define one typed route metadata registry for authenticated routes.
- Store canonical path or pattern, title, description, icon, navigation group,
  browser-title eligibility, and test IDs in the registry.
- Use the registry for sidebar/search navigation, page `HeaderTitle` metadata,
  and document-title resolution where practical.
- Cover Dashboard, Inventory, Collections, Wishlist, Media, Purchases,
  Integrations, Chats, Inbox, Discoveries, Reports, Market Watch, Settings and
  children, Users, Help, and known error/fallback routes.
- Correct the `/purchases` document title gap and the `/scanner` versus Market
  Watch naming mismatch.
- Add table-driven coverage that fails when an authenticated route lacks
  metadata.

## Capabilities

### Modified Capabilities

- `ui-foundation-components`: adds a canonical authenticated route metadata
  contract for visible shell headers and icons.
- `ui-foundation-interactions`: adds route-title and responsive-header
  consistency requirements for browser titles, navigation labels, and fallback
  behaviour.

## Impact

- Affected code: `ui.web/src/lib/document-title.ts`,
  `ui.web/src/components/layout`, authenticated route/page feature headers, and
  command/search navigation sources.
- Affected tests: route metadata unit tests, header component/route tests, and
  focused Cypress coverage for responsive headers.
- Related issues: `#1940`, `#1941`, `#1938`, `#1939`.
