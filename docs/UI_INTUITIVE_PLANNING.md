# Cabinet UI Intuitive Planning

## Purpose
Define a simple, intuitive UI direction for Cabinet and a concrete dashboard model for "What Needs Attention Now".

## Design Goal
- New users should know what to do in under 30 seconds.
- Daily users should see urgent changes without hunting across screens.
- Power features should remain available but not overwhelm first-run flows.

## Proposed Information Architecture
- Home
- Inventory
- Discover
- Scanner
- Reports
- Settings

Notes:
- Keep `Photos`, `Barcodes`, and `AI Assist` as tabs/sections inside `Inventory` to reduce top-nav clutter.
- Keep onboarding/auth as full-screen guided flow before full workspace unlock.

## Home Dashboard Model

### What Needs Attention Now (Primary Panel)
This section should always appear at top of Home and be sorted by urgency.

Include:
1. Watchlist Hits (new items from wishlist/query matches)
- Count of new hits since last visit
- Top rows: item, source, current price, target price, delta
- Actions: `Open Listing`, `Add to Collection`, `Ignore`

2. Price Changes
- Items with significant change since last snapshot (drop/rise)
- Top rows: item, latest, previous, change %, 7d trend
- Actions: `Track`, `Set Alert`, `Open Price History`

3. New Discoveries (Not In My Collection)
- Newly detected candidates not owned
- Top rows: title, source, price, first seen
- Actions: `Ignore`, `Wishlist`, `Track`, `Create Item`

4. Scanner/Provider Failures
- Failed runs and provider health warnings
- Top rows: query set, error reason, last failure
- Actions: `Retry`, `Open Scanner`, `View Logs`

5. Recovery/Security Alerts
- Recovery required, session/auth warnings
- Actions: `Open Settings Diagnostics`, `Re-authenticate`

### Secondary Home Panels
- Collection Snapshot: total items, total instances, estimated value
- Recent Activity: latest adds/edits/imports
- Quick Actions: Add Item, Run Scanner, Export, Backup Now

## Prioritization

### v1 (Must Have)
- Watchlist Hits
- Price Changes (latest vs previous + %)
- New Discoveries
- Scanner Failures
- Quick Actions

### v1.1 (Next)
- Trend sparklines in cards
- User-configurable thresholds (e.g., >=10% change)
- Snooze/acknowledge attention cards
- Personalization of card ordering

## Data Mapping (Current APIs)
- Watchlist hits: `/api/wishlist/hits`
- Pricing changes/history: `/api/pricing/*`
- Discoveries not in collection: `/api/discovery/not-in-collection`
- Scanner failures/health: `/api/scanner/failures`, `/api/scanner/provider-health`
- Dashboard summary: `/api/dashboard`
- Recovery alerts: `/api/runtime/recovery`

## UX Rules for Attention Panel
- Show only actionable signals, not raw logs.
- Each row/card must have at least one clear action button.
- No more than 5 primary cards on first viewport.
- Use plain language status labels: `Needs action`, `Review`, `Healthy`.

## Implementation Notes
- Start with server-computed aggregates for speed and consistency.
- Add optimistic UI updates for simple actions (ignore, wishlist, track).
- Log all attention actions in activity log for auditability.

