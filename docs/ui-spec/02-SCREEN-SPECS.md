# 02 Screen Specs (Strict)

## Scope
Defines each screen sections, required states, and required actions.

## State Vocabulary (mandatory for all screens)
- `loading`: primary skeleton or loading message
- `empty`: no data with guided action
- `success`: populated content
- `error`: actionable error with retry path

## Home
Sections:
1. What Needs Attention Now (priority cards)
2. Collection Snapshot
3. Recent Activity
4. Quick Actions

Required actions:
- Open listing, ignore, wishlist, track, create item
- Retry failed scanner run
- Open diagnostics for recovery/auth issues

## Inventory
### Items tab
- Search + filter row
- List/detail split
- Quick add and advanced edit
- Instance management

### Photos tab
- Load by item
- Upload (file + drag/drop)
- Camera capture permission flow
- Set primary
- Delete
- Fullscreen preview

### Barcodes tab
- Manual add
- Load existing
- Lookup local
- External search launch

### AI Assist tab
- Enable/disable AI
- Suggest from title
- Suggest from photo
- Explicit confirm before applying suggestion

## Discover
Sections:
1. Filter bar (query, max price, date)
2. Not-in-collection list
3. Candidate actions

Required actions:
- Ignore
- Add to wishlist
- Track price
- Create canonical item

## Scanner
Sections:
1. Query set management
2. Run now/scheduled
3. Failures list + retry
4. Provider health
5. Matching run summary

## Reports
Sections:
1. Wishlist and hit summary
2. Pricing trend/points
3. Source breakdown
4. Export controls

## Settings
Sections:
1. Diagnostics
2. Maintenance
3. License
4. Backups and restore
5. Profile settings + secrets

## Strict UX Rules
1. Every empty state has a primary CTA.
2. Every error state has retry and context.
3. No screen should require more than one deep click to reach primary action.
4. Advanced controls are in collapsible sections, not default-expanded.

## Acceptance Criteria
- [ ] Each screen implements all four state types.
- [ ] Required actions are present and testable.
- [ ] Empty and error states are explicit and actionable.

