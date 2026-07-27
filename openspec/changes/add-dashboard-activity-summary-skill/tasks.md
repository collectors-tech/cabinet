## 1. Contract

- [x] 1.1 Define the Dashboard activity summary Agent Skill contract,
  read-only safety boundary, Dashboard data source, truthful time-window
  behavior, and validation plan for #1942.
- [x] 1.2 Update OpenSpec traceability and the #1701 Agent skill coverage matrix
  with implemented Dashboard summary evidence.
  - Added traceability row `AGENT-SKILLS-REGISTRY-012` and updated the
    Dashboard row in `openspec/traceability/agent-skill-coverage.md` with the
    direct read-only Dashboard summary registry/API evidence.

## 2. Skill registration and runtime

- [x] 2.1 Register `cabinet.dashboard.summarise_activity` as a built-in
  read-only skill with explicit profile context and no mutation permissions.
  - Added focused registry coverage in `internal/agentskills/registry_test.go`
    proving built-in source metadata, read-only/no-confirm permissions,
    profile/workspace context, Dashboard capability/workflow bindings,
    output schema refs, missing-context blocker, and read-only profile
    authority allowance.
- [x] 2.2 Add a Dashboard summary service/runtime adapter that reuses canonical
  Dashboard data and returns structured totals, attention signals, recent
  items, destination links, and record identifiers.
- [x] 2.3 Distinguish current snapshot values from evidence-backed time-window
  changes without implying unavailable history.
- [x] 2.4 Preserve source context and create non-secret audit/workflow evidence
  for direct skill execution.
  - Direct apply coverage asserts source context retention and a non-secret
    applied authority audit row for `cabinet.dashboard.summarise_activity`.

## 3. Edge states

- [x] 3.1 Handle empty Dashboard data with truthful "nothing needs attention"
  output and usable destination links.
- [x] 3.2 Handle partial or unavailable Dashboard dependencies with actionable
  warnings instead of fabricated totals.
  - Added focused API coverage proving partial recent-item dependency failures
    return current Dashboard totals plus sanitized warnings, and unavailable
    Dashboard dependencies return no inferred totals with fallback destination
    links and no raw storage error leakage.
- [x] 3.3 Prove profile isolation so one profile cannot read another profile's
  Dashboard summary.
- [x] 3.4 Prove the read-only skill never creates mutation previews,
  confirmation tokens, or applied mutations.

## 4. Evidence

- [x] 4.1 Add focused Go/API tests for populated, empty, partial/unavailable,
  time-window caveat, no-mutation, and profile-isolation paths.
- [x] 4.2 Add main Chat and side-panel proof when the governed dispatch path can
  invoke the skill.
  - Added side-panel Agent Skill dispatcher proof in
    `ui.web/cypress/e2e/chats/assistant-workspace-dashboard-summary/spec.cy.ts`.
    Main Chat natural-language routing remains sequenced behind #1933, so this
    slice records available side-panel dispatch evidence without claiming #1933.
- [ ] 4.3 Run strict OpenSpec validation, focused Go tests, touched UI/API
  validation where relevant, and `git diff --check`.
