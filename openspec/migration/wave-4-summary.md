# Wave 4 Summary

- wave number: 4
- scope: provider/scanner API contracts
- issue: #189
- status: completed

## IDs moved to implemented
- `CANDIDATES-001`
- `CANDIDATES-002`
- `INTEGRATION-008`
- `INTEGRATION-009`
- `INTEGRATION-012`
- `INTEGRATION-013`
- `INTEGRATION-014`
- `INTEGRATION-015`
- `SCANNER-002`

## IDs still partial/planned in this scope
- `SCANNER-001`: query-set create contract still only spec-structure coverage
- `SCANNER-003`: retry/failure workflow not yet bound to deterministic contract test for this ID

## Runtime/API behavior implemented
- Added provider registry endpoint `GET /api/providers/registry`.
- Added Amazon contract run endpoint `POST /api/providers/amazon/run` with deterministic disabled envelope.
- Added AU webshop stock parser endpoint `POST /api/providers/au-webshops/parse-stock`.
- Expanded scheduled scanner run summary envelope at `POST /api/scanner/run/scheduled`.

## Test commands and results
- `go test ./internal/app -run TestWave4 -count=1` -> pass
- `go test ./tests -count=1` -> pass
- `go test ./internal/app -count=1` -> pass
- `openspec validate --all` -> pass (`57 passed, 0 failed`)

## Net counts
- partial before -> after: `136 -> 127`
- implemented before -> after: `30 -> 39`
- reduction: `9`

## Commit
- commit: `<pending>`
