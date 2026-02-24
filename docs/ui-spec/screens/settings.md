# Settings Screen Spec

## Use Cases
1. User runs diagnostics and maintenance safely.
2. User manages license and backups.
3. User configures profile settings and secrets.

## UI Sections
1. Diagnostics
2. Maintenance
3. License management
4. Backup and restore
5. Settings and secrets forms

## State Behavior
- Loading: diagnostics/license/backups loading.
- Empty: no backups/no license state messages.
- Error: explicit operation-specific errors with retry.
- Success: updated status and confirmation text.

## Acceptance Criteria
- [ ] Destructive operations require explicit confirmation.
- [ ] Runtime diagnostics and recovery status are visible.
- [ ] License import errors are clear and recoverable.
- [ ] Backup restore cannot run unless confirmed.

## Test Cases
- `SET-001` load and save settings/secrets.
- `SET-002` diagnostics toggle and refresh.
- `SET-003` license import + status.
- `SET-004` guarded restore flow.

