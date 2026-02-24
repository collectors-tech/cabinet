# 03 Dashboard "What Needs Attention Now" Strict Spec

## Purpose
Home must answer: "What needs action now?" with prioritized, actionable signals.

## Card Set (v1 mandatory)
1. Recovery and Security Alerts
2. Scanner and Provider Failures
3. Watchlist Hits (new and below target)
4. Significant Price Changes
5. New Discoveries Not Owned

## Card Schema (uniform)
- `title`
- `severity`: critical | action | review | info
- `count`
- `last_updated_at`
- `top_rows` (max 5)
- `primary_actions` (1-3)

## Ranking Order
1. Recovery/security
2. Scanner/provider failures
3. Watchlist hits below target
4. Significant price changes
5. New discoveries

## Thresholds and Logic
1. Significant price change:
- Trigger if absolute percent change >= 10% since previous snapshot.
- Percent = `(latest - previous) / previous * 100`.

2. Watchlist hit priority:
- Rank by `(target_price - current_price) / target_price`.
- Higher positive score first.

3. Discovery freshness:
- Rank by first seen descending, then confidence descending.

4. Failure urgency:
- Rank by attempts descending, then last_error_at descending.

## Card Actions (strict)
1. Recovery/security:
- `Open Diagnostics`
- `Re-authenticate`

2. Scanner failures:
- `Retry Now`
- `Open Scanner`
- `View Logs`

3. Watchlist hits:
- `Open Listing`
- `Add to Collection`
- `Ignore`

4. Price changes:
- `View History`
- `Track`
- `Set Alert`

5. Discoveries:
- `Wishlist`
- `Track`
- `Create Item`
- `Ignore`

## Behavior Rules
1. Max 5 cards above fold.
2. Each card supports `Snooze 24h` and `Dismiss`.
3. Dismissed cards reappear only with new qualifying events.
4. All actions write activity log entries.

## Empty and Error Behavior
- Empty: `No urgent changes right now` + `Run Scanner` CTA.
- Error: `Unable to load attention data` + `Retry` CTA.

## Acceptance Criteria
- [ ] All five card types implemented.
- [ ] Ranking logic and thresholds applied consistently.
- [ ] Snooze/dismiss behavior persisted per profile.
- [ ] Card actions deep-link into correct workspaces.

