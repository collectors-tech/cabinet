# 06 Per-Screen Detailed Specs (Strict)

## Scope
Defines strict implementation detail per screen: user use cases, UX behavior, acceptance criteria, and test cases.

## Template (applies to every section)
- Use Cases
- Required UI Sections
- Required States (`loading`, `empty`, `success`, `error`)
- Acceptance Criteria
- Test Cases

## Home
### Use Cases
- User opens app and needs immediate triage of urgent events.
- User jumps from attention card to action workflow with one click.

### Required UI Sections
- What Needs Attention Now
- Collection Snapshot
- Recent Activity
- Quick Actions

### Acceptance Criteria
- [ ] Attention cards are ordered by strict priority from `03-DASHBOARD-ATTENTION-STRICT.md`.
- [ ] Each attention card includes at least one primary action button.
- [ ] Empty home state includes CTA: `Run Scanner` and `Add First Item`.
- [ ] Home refresh does not block navigation and displays progress state.

### Test Cases
- `HOME-001`: Renders ordered cards with sample mixed-priority data.
- `HOME-002`: Clicking card action deep-links to correct screen.
- `HOME-003`: Empty state appears when all counts are zero.
- `HOME-004`: Error state provides retry and recovers after success.

## Inventory: Items tab
### Use Cases
- User searches and filters collection quickly.
- User creates/edits canonical item and instance.

### Required UI Sections
- Search/filter/sort bar
- Item table/list
- Item details panel
- Quick add + advanced metadata form
- Instances panel

### Acceptance Criteria
- [ ] Search and filters apply within 300ms for local dataset sizes up to v1 target.
- [ ] No hidden required fields in quick add path.
- [ ] User can create item without navigating away.
- [ ] Details panel retains selection while filters change unless item removed from result.

### Test Cases
- `INV-I-001`: Search by part number and title.
- `INV-I-002`: Filter by brand/category/tag and clear filters.
- `INV-I-003`: Create item and verify immediate list visibility.
- `INV-I-004`: Instance add/update displays in instance panel.

## Inventory: Photos tab
### Use Cases
- User uploads item photos and sets primary image.
- User captures image from camera and previews full screen.

### Required UI Sections
- Item selection input
- Upload controls (file + drag/drop)
- Camera controls
- Photo list with primary/delete actions
- Fullscreen preview modal

### Acceptance Criteria
- [ ] Upload supports at least JPG/PNG and shows deterministic errors.
- [ ] Primary image switch updates immediately in list state.
- [ ] Camera permission denial is non-fatal and user-guided.
- [ ] Fullscreen preview is keyboard-closeable (`Esc`).

### Test Cases
- `INV-P-001`: Upload valid file and list refreshes.
- `INV-P-002`: Set primary then verify primary marker.
- `INV-P-003`: Delete photo then verify removal.
- `INV-P-004`: Camera denied path shows actionable message.

## Inventory: Barcodes tab
### Use Cases
- User adds barcode manually and runs lookup.
- User launches external search for unmatched codes.

### Required UI Sections
- Barcode input and action buttons
- Barcode list
- Local match summary
- External link output

### Acceptance Criteria
- [ ] Barcode add validates non-empty and normalized format.
- [ ] Lookup result clearly distinguishes local match vs no match.
- [ ] External search link is generated safely and encoded.

### Test Cases
- `INV-B-001`: Add and load barcodes for selected item.
- `INV-B-002`: Local lookup returns expected match count.
- `INV-B-003`: External search text/link rendered on no local match.

## Inventory: AI Assist tab
### Use Cases
- User asks AI to normalize listing title.
- User identifies item from photo URL.
- User explicitly confirms before AI suggestions are applied.

### Required UI Sections
- AI enable/disable controls
- Title normalize input
- Photo identify input
- Suggestion preview
- Confirm apply action

### Acceptance Criteria
- [ ] AI must never auto-write without user confirmation.
- [ ] Suggestion confidence is visible when available.
- [ ] AI errors are scoped to AI panel and do not break other tabs.

### Test Cases
- `INV-AI-001`: Toggle AI enable/disable.
- `INV-AI-002`: Title normalize returns suggestion and confirm apply works.
- `INV-AI-003`: Photo identify error and retry path.

## Discover
### Use Cases
- User triages not-owned candidates from scanner results.

### Required UI Sections
- Filter bar (query, max price, date)
- Candidate list
- Candidate row actions

### Acceptance Criteria
- [ ] Filter controls update request query params correctly.
- [ ] Row actions update state without requiring full page reload.
- [ ] Candidate list supports empty and stale-data messaging.

### Test Cases
- `DISC-001`: Apply filters and verify expected result subset.
- `DISC-002`: Ignore action removes or marks row.
- `DISC-003`: Wishlist/Track/Create Item action success path.

## Scanner
### Use Cases
- User manages query sets and runs scans manually/scheduled.
- User reviews failures and retries specific query set.

### Required UI Sections
- Query set editor/list
- Run controls
- Failures panel
- Provider health panel
- Matching summary

### Acceptance Criteria
- [ ] Run scheduled/manual paths show explicit status feedback.
- [ ] Failure rows include query set id and reason.
- [ ] Retry action available only for retryable failures.

### Test Cases
- `SCAN-001`: Create/edit query set and persist values.
- `SCAN-002`: Run now and show completion status.
- `SCAN-003`: Load failures and retry single failure.
- `SCAN-004`: Provider health load with error fallback.

## Reports
### Use Cases
- User reviews pricing changes and exports history.
- User tracks wishlist performance and source breakdown.

### Required UI Sections
- Wishlist summary
- Pricing trend/points section
- Source breakdown list/chart
- Export controls

### Acceptance Criteria
- [ ] Export action returns non-empty payload when data exists.
- [ ] Trend and source sections handle no-data gracefully.
- [ ] Reports are filterable by selected item/source/date where supported.

### Test Cases
- `REP-001`: Load wishlist and hits.
- `REP-002`: Load trend and stats sections.
- `REP-003`: Export history and verify byte count > 0.

## Settings
### Use Cases
- User configures profile settings/secrets.
- User runs maintenance tasks and diagnostics.
- User manages license and backups.

### Required UI Sections
- Diagnostics controls and status
- Maintenance actions (reindex/repair/backup)
- License import/status
- Backup list and guarded restore
- Profile settings/secrets forms

### Acceptance Criteria
- [ ] Destructive actions require explicit confirmation.
- [ ] Runtime and recovery diagnostics are visible and refreshable.
- [ ] License import errors are explicit and recoverable.
- [ ] Backup restore path requires confirmation checkbox.

### Test Cases
- `SET-001`: Load and save settings and secrets.
- `SET-002`: Toggle debug mode and verify status update.
- `SET-003`: Import license and load status.
- `SET-004`: Backup restore blocked without confirmation.

