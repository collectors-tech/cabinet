## 1. Contract

- [x] 1.1 Define the Dashboard activity summary Agent Skill contract,
  read-only safety boundary, Dashboard data source, truthful time-window
  behavior, and validation plan for #1942.
- [ ] 1.2 Update OpenSpec traceability and the #1701 Agent skill coverage matrix
  with implemented Dashboard summary evidence.

## 2. Skill registration and runtime

- [ ] 2.1 Register `cabinet.dashboard.summarise_activity` as a built-in
  read-only skill with explicit profile context and no mutation permissions.
- [ ] 2.2 Add a Dashboard summary service/runtime adapter that reuses canonical
  Dashboard data and returns structured totals, attention signals, recent
  items, destination links, and record identifiers.
- [ ] 2.3 Distinguish current snapshot values from evidence-backed time-window
  changes without implying unavailable history.
- [ ] 2.4 Preserve source context and create non-secret audit/workflow evidence
  for direct skill execution.

## 3. Edge states

- [ ] 3.1 Handle empty Dashboard data with truthful "nothing needs attention"
  output and usable destination links.
- [ ] 3.2 Handle partial or unavailable Dashboard dependencies with actionable
  warnings instead of fabricated totals.
- [ ] 3.3 Prove profile isolation so one profile cannot read another profile's
  Dashboard summary.
- [ ] 3.4 Prove the read-only skill never creates mutation previews,
  confirmation tokens, or applied mutations.

## 4. Evidence

- [ ] 4.1 Add focused Go/API tests for populated, empty, partial/unavailable,
  time-window caveat, no-mutation, and profile-isolation paths.
- [ ] 4.2 Add main Chat and side-panel proof when the governed dispatch path can
  invoke the skill.
- [ ] 4.3 Run strict OpenSpec validation, focused Go tests, touched UI/API
  validation where relevant, and `git diff --check`.
