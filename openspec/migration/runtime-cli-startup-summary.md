# Runtime CLI Startup Parameters Summary (#245)

Date: 2026-03-04
Issue: #245

## OpenSpec IDs
- RUNTIME-CORE-008
- RUNTIME-CORE-009

## Delivered behavior
- Added deterministic CLI startup parameter parsing in `cmd/cabinet` for:
  - `--port`
  - `--listen`
  - `--data-dir`
  - `--profile`
  - `--instance-name` (alias for profile when profile missing)
  - `--auth-mode` (`local|clerk`)
  - `--base-url`
  - `--allow-parallel`
  - `--log-level` (`debug|info|warn|error`)
  - `--no-open-browser` recognized in parser for compatibility with startup command line.
- CLI overrides are applied as process env before `config.Load()` so they override existing env/defaults deterministically.
- Added deterministic fail-fast validation:
  - conflicting `--port` and `--listen`
  - invalid auth mode
  - invalid port ranges and invalid base URL
  - `--allow-parallel` requires profile/instance context
- Added startup effective configuration line:
  - `CABINET_EFFECTIVE_CONFIG ...`
  - includes `addr`, `host`, `port`, `data_dir`, `profile`, `auth_mode`, `base_url`, `allow_parallel`, `log_level`

## Test evidence
- Failing-first proof:
  - `go test ./cmd/cabinet -count=1` initially failed (undefined startup flag parser symbols)
- Passing gates:
  - `go test ./cmd/cabinet -count=1` -> pass
  - `go test ./internal/app -count=1` -> pass
  - `go test ./tests -count=1` -> pass
  - `openspec validate --all` -> pass
  - `pwsh -NoLogo -NoProfile -File .\\scripts\\build-cabinet.ps1` -> pass
