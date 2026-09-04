## Why

#1942 is the remaining named Dashboard Agent Skill gap in the #1701 coverage
matrix. Cabinet already exposes grounded dashboard data for the active profile,
but Agent cannot yet answer "what needs my attention?" or "what changed?"
through a read-only, governed skill.

## What Changes

- Add `cabinet.dashboard.summarise_activity` as a built-in, read-only Agent
  Skill.
- Reuse canonical dashboard application data instead of scraping UI text or
  duplicating Dashboard calculations.
- Return structured current totals, attention signals, recent items, record
  identifiers, destination links, and time-window caveats.
- Keep time-window answers truthful when Cabinet only has current snapshots.
- Preserve profile isolation, source context, and non-secret audit evidence.
- Update coverage and traceability so the Dashboard skill is no longer listed
  as a planned gap.

## Capabilities

### Modified Capabilities

- `agent-skills-registry`: add a governed read-only Dashboard summary skill
  with explicit status, safety, permissions, context, bindings, output, and
  evidence behavior.
- `dashboard`: expose Dashboard attention signals and current summary data to a
  read-only Agent Skill without mutating Cabinet state.

## Impact

- Affected code: `internal/agentskills`, `internal/app`, Dashboard service/API
  helpers, Agent Skill preview/apply runtime, and source-context adapters.
- Affected tests: focused registry/unit/API tests for populated, empty,
  partial/unavailable, and profile-isolated Dashboard summaries; main Chat and
  side-panel proof where dispatch is available.
- Affected documentation: OpenSpec specs, traceability, and #1701 skill
  coverage matrix.
- Related issues: `#1942`, `#1701`, `#1714`, `#1933`.
