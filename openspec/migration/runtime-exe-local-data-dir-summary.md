# Runtime Exe-Local Data Directory Summary (#244)

Date: 2026-03-04
Issue: #244

## OpenSpec ID
- RUNTIME-CORE-010

## Delivered behavior
- Default data directory resolution now prefers executable-local path first:
  - `<exe_dir>/data`
- Added deterministic resolution helper:
  - `resolveExecutableLocalDataDir()`
  - supports `CABINET_EXE_DIR` override for deterministic test/runtime harness usage.
- Existing explicit override precedence is preserved:
  - `CABINET_DATA_DIR` still wins via config load override path.

## Test evidence
- fail-first:
  - `go test ./internal/config -count=1` failed on `TestLoadDefaultDataDirUsesExecutableLocalPathFirst`
- passing gates:
  - `go test ./internal/config -count=1`
  - `go test ./internal/app -count=1`
  - `go test ./tests -count=1`
  - `openspec validate --all`
  - `pwsh -NoLogo -NoProfile -File .\\scripts\\build-cabinet.ps1`
