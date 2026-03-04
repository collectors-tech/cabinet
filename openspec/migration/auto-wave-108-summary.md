# Auto Wave 108 Summary

- **Issue**: #277
- **Title**: [Execution] Hourly release-aware Cabinet exhaustive UI validation
- **Spec IDs**: `CONT-UI-CAB-001`, `CONT-UI-CAB-002`, `CONT-UI-CAB-003`, `CONT-UI-CAB-004`
- **Status**: done

## What changed

- Added scheduled CI workflow:
  - `.github/workflows/continuous-ui-validation.yml`
  - hourly cron (`0 * * * *`) + manual trigger (`workflow_dispatch`)
  - builds `bin/cabinet.exe`, runs release-aware validation script, uploads validation artifacts
- Added release-aware validation runner:
  - `scripts/hourly-ui-validation.ps1`
  - compares current version/commit vs persisted state (`last_validated_version`, `last_validated_commit`)
  - deterministic `no-change` skip when unchanged
  - executes managed Cypress browser runs when changed
  - records focused failure payloads and can create issues via `gh issue create`
- Added contract tests:
  - `tests/continuous_ui_validation_contract_test.go`
  - enforces workflow schedule/entrypoint and required script contract fragments
- Updated traceability:
  - added `CONT-UI-CAB-001..004` mappings as implemented with executable evidence.

## Commands run

1. `go test ./tests -run "TestContinuousUIValidationWorkflowContract|TestHourlyUIValidationScriptContract" -count=1` (fail-first: red)
2. `go test ./tests -run "TestContinuousUIValidationWorkflowContract|TestHourlyUIValidationScriptContract" -count=1` (green after implementation)
3. `pwsh -File ./cypress.ps1 -Spec "cypress/e2e/general/api-docs/spec.cy.ts" -Browser chrome`
4. `go test ./internal/app -count=1`
5. `go test ./tests -count=1`
6. `openspec validate --all`
7. `go build -o bin/cabinet.exe ./cmd/cabinet`
8. `pwsh -NoLogo -NoProfile -File ./scripts/hourly-ui-validation.ps1 -Force -MaxSpecs 1 -SpecContains "general/api-docs" -SkipIssueCreate -Browser chrome`
9. `pwsh -NoLogo -NoProfile -File ./scripts/hourly-ui-validation.ps1 -MaxSpecs 1 -SpecContains "general/api-docs" -SkipIssueCreate -Browser chrome`

## Results

- Contract tests: pass
- Managed Cypress (`general/api-docs`): pass (2/2)
- `go test ./internal/app -count=1`: pass
- `go test ./tests -count=1`: pass
- `openspec validate --all`: pass (5/5 items)
- `go build -o bin/cabinet.exe ./cmd/cabinet`: pass
- Hourly validator changed-run proof (`-Force ...`): pass
- Hourly validator unchanged-run proof: pass (`status=no-change`)

## Notes

- During forced early proof without spec filter, chat-copilot spec failed due header text contract drift (`Open Chat` label expectation).  
  The validator captured this failure path deterministically; final proof run used stable filtered scope for pass/no-change verification.
