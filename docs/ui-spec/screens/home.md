# Home Screen Spec

## Use Cases
1. User sees urgent events immediately after launch.
2. User takes action from an attention card in one click.
3. User reviews collection summary without leaving Home.

## UI Sections
1. What Needs Attention Now
2. Collection Snapshot
3. Recent Activity
4. Quick Actions

## State Behavior
- Loading: card placeholders and KPI skeleton.
- Empty: "No urgent changes right now" + `Run Scanner` CTA.
- Error: inline error + `Retry`.
- Success: ranked attention cards and snapshot metrics.

## Acceptance Criteria
- [ ] Attention cards follow ranking rules in `03-DASHBOARD-ATTENTION-STRICT.md`.
- [ ] Every card has at least one actionable button.
- [ ] Quick Actions include: Add Item, Run Scanner, Open Discover, Backup Now.
- [ ] Home refresh updates only relevant panels, not full app hard-block.

## Test Cases
- `HOME-001` card rank order with mixed-priority fixtures.
- `HOME-002` card action deep-link routing.
- `HOME-003` empty-state CTA availability.
- `HOME-004` error + retry recovery.

