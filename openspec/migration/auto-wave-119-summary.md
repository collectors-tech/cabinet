# Auto Wave 119 Summary

- Issue: #241
- Scope: Runtime PID-only lock + metadata-based attach/open behavior with stale PID recovery.
- Spec IDs: RUNTIME-CORE-012

## What Changed
- Added runtime attach resolver and health probe in `cmd/cabinet/runtime_attach.go`.
- Updated startup flow in `cmd/cabinet/main.go` to:
  - check for running instance using `cabinet.pid` + `cabinet.json` URL metadata,
  - attach/open existing runtime URL when healthy,
  - remove stale PID lock and continue fresh startup when stale/unhealthy.
- Added/extended tests:
  - `cmd/cabinet/runtime_attach_test.go`
  - `cmd/cabinet/main_cli_test.go`
- Updated OpenSpec requirement:
  - `openspec/specs/general/runtime-core/spec.md` (added `RUNTIME-CORE-012`)
- Updated traceability mapping:
  - `openspec/traceability.md`

## Commands Run
1. `go test ./cmd/cabinet -run TestResolveRunningRuntimeAttach -count=1` (fail-first repro, expected fail before implementation)
2. `go test ./cmd/cabinet -count=1`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`
6. `pwsh -File .\\scripts\\build-cabinet.ps1`

## Results
- `go test ./cmd/cabinet -run TestResolveRunningRuntimeAttach -count=1` initially failed with undefined symbol (`resolveRunningRuntimeAttach`).
- All post-fix gates passed.

## Notes
- PID file remains PID-only (`cabinet.pid` stores only process id + newline).
- Endpoint metadata is resolved from `cabinet.json` runtime/meta URL fields.
- Attach log line is deterministic: `CABINET_RUNTIME_ATTACH url=... pid=... data_dir=... resolved_port=...`.
