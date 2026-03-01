# Next Steps — E2E Foundation (Cabinet)

Use this once current in-flight wave finishes.

## Goal
Switch Cabinet to a durable E2E-first delivery model with persistent documentation.

## Create / maintain these files
1. `openspec/migration/E2E_JOURNEY_TEMPLATE.md`
2. `openspec/migration/E2E_COVERAGE_MATRIX.csv`
3. `openspec/migration/E2E_WORKING_RULES.md`

## Required policy
- User-facing requirements require Cypress E2E proof.
- API/unit tests are supporting layers only.
- Do not mark `implemented` in traceability without executable E2E evidence.

## Suggested Wave 8 focus
1. Onboarding/auth UI journeys
2. Shell/navigation journeys
3. Chat copilot journeys
4. Media/inventory UI journeys

## `E2E_COVERAGE_MATRIX.csv` columns
`JourneyID,JourneyName,RequirementIDs,SpecFile,TestNames,Status,LastVerifiedAt,Wave,Blocker`

Status values: `planned`, `in_progress`, `implemented`, `blocked`

## Minimum per-wave deliverables
- `openspec/migration/wave-<n>-summary.md`
- `openspec/migration/wave-<n>-changed-files.txt`
- updates to `openspec/traceability.md`
- updates to `openspec/migration/E2E_COVERAGE_MATRIX.csv`

## Validation gates
- Run targeted Cypress specs for wave IDs
- Run supporting API/unit tests
- Run `openspec validate --all`
- Keep unrelated dirty files untouched

## Definition of done (per wave)
- IDs moved to implemented have Cypress proof mapped in traceability
- Remaining partial IDs have explicit blockers
- Validation output is captured in wave summary
