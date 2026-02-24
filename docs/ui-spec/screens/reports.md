# Reports Screen Spec

## Use Cases
1. User reviews wishlist hit behavior and pricing trends.
2. User inspects source-level price breakdown.
3. User inspects source-level stock availability trends.
4. User exports pricing history for analysis.

## UI Sections
1. Wishlist summary
2. Pricing trend/history summary
3. Source breakdown
4. Export controls
5. Stock trend summary (in-stock transitions, latest stock count by source)

## State Behavior
- Loading: report section loading states.
- Empty: no report data yet with next-step guidance.
- Error: report API error and retry.
- Success: report sections populated.

## Acceptance Criteria
- [ ] Export produces non-empty payload when data exists.
- [ ] No-data states are explicit and non-failing.
- [ ] Trend and source sections are independently loadable.
- [ ] Wishlist hits and pricing views remain consistent with selected item context.
- [ ] Stock trend uses the same selected-item context as pricing trend.

## Test Cases
- `REP-001` load wishlist and hits.
- `REP-002` load trend and stats.
- `REP-003` load source breakdown.
- `REP-004` export history success.
- `REP-005` load stock trend and latest stock per source.
