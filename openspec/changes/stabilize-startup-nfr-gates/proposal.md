## Why

Fresh-runtime validation is currently unstable on Windows developer machines because Cabinet startup performs a long SQLite migration chain during `app.New()` against empty temp databases. Under full-suite load this causes:
- startup migration timeout failures before the app binds
- non-functional startup gate failures even when the app eventually starts

That leaves unrelated feature PRs blocked behind infrastructure noise and weakens confidence in runtime startup performance.

## What Changes

- Make fresh SQLite startup/migration execution deterministic enough for full local validation.
- Preserve the existing startup timeout contract without requiring feature branches to relax it ad hoc.
- Keep startup NFR evidence meaningful by measuring the real fresh-runtime path rather than papering over flakiness.

## Capabilities

### New Capabilities

- `fresh-runtime-startup`: Fresh tempdir runtime startup + migration path remains reliable under validation load.

### Modified Capabilities

- `non-functional`: sharpen startup reliability expectations for validation startup flows.

## Impact

- Affected code: `internal/db`, `internal/app`, startup-bound tests under `internal/nfr` and `tests`
- Affected tests: `go test ./internal/nfr`, `go test ./tests`, full `go test ./...`
- Related issues: `#448`, unblocks `#446` / PR `#447`