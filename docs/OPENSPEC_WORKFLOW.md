# OpenSpec Workflow for Cabinet

## Purpose
Define how OpenSpec is used in this repository without replacing existing governance.

OpenSpec is used as:
- spec-first planning and review
- structured change artifacts (`proposal`, `design`, `specs`, `tasks`)

GitHub Issues remain:
- source of truth for execution tracking
- closure and evidence record

## Rule Alignment
This repo still follows:
- issue-first tracking in `collectors-tech/cabinet`
- test-first implementation
- in-session test execution and evidence before closure

OpenSpec adds planning rigor before implementation.

## Repository Layout
- `openspec/changes/<change-name>/`
  - `proposal.md`
  - `design.md`
  - `specs/<capability>/spec.md`
  - `tasks.md`
- `openspec/specs/`
  - baseline/specified capabilities (after archiving completed changes)

## Required Flow (Per Feature/Issue)
1. Create or reuse GitHub issue.
2. Create OpenSpec change:
   - `openspec new change <kebab-name>`
3. Author artifacts:
   - `proposal.md` (why + what)
   - `design.md` (how)
   - `specs/*` (requirements/scenarios)
   - `tasks.md` (implementation checklist)
4. Validate:
   - `openspec validate --changes --strict --no-interactive`
5. Implement with test-first policy.
6. Close GitHub issue only with passing evidence.
7. Archive completed OpenSpec change:
   - `openspec archive <change-name>`

## Commands
Initialize:
```powershell
openspec init . --tools "codex,cursor"
```

Create change:
```powershell
openspec new change <change-name>
```

Show status:
```powershell
openspec list
openspec status --change <change-name>
```

Validate:
```powershell
openspec validate --changes --strict --no-interactive
```

## Current Seeded Changes
- `stabilize-inventory-runtime-regressions`
- `complete-screen-api-parity-audits`
- `finalize-onboarding-and-collector-ux`

## Migration Control Docs
- Spec migration catalog: `docs/OPENSPEC_MIGRATION_CATALOG.md`
- Strict migration TODO: `docs/OPENSPEC_MIGRATION_TODO.md`

## Mapping to Existing Backlog
- Stability and non-500 reliability: `#151`, `#149`, `#147`, `#154`
- Screen/API parity audits: `#143`, `#144`, `#145`, `#152`
- Onboarding and core collector UX: follow-up issues from `docs/APP_COMPLETION_ANALYSIS.md`

## Acceptance Standard
No implementation issue is complete unless:
- OpenSpec change artifacts exist and validate
- tests were written and run in-session
- commit/push complete
- GitHub issue checklist is complete
