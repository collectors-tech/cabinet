## 1. Contract

- [x] 1.1 Define the profile-scoped Agent authority modes and shared decision
  guard contract.
- [x] 1.2 Add OpenSpec traceability rows for the implemented policy and
  cross-entry-point enforcement evidence.
  - `AGENT-SKILLS-REGISTRY-011` and `AGENT-UNIVERSAL-CHANNELS-006` in
    `openspec/traceability.md` map #1932 policy, Settings, audit, direct API,
    Chat, Assistant side-panel, MCP, and Telegram evidence to the modified
    OpenSpec requirements.

## 2. Shared enforcement

- [x] 2.1 Add the first shared Agent Skill authority guard covering read-only,
  default local-change, external-write, destructive, and profile-mismatch
  decisions.
- [x] 2.2 Persist the profile policy with the default
  `ask_before_local_changes` mode for existing and new profiles.
- [x] 2.3 Apply the shared guard to direct Agent Skill preview/apply API calls.
- [x] 2.4 Apply the shared guard to Chat, Assistant side panel, MCP, and
  Telegram/API dispatch paths.

## 3. Settings and audit

- [x] 3.1 Add Settings > Skills controls for the effective Agent authority
  mode and external-write approval.
- [x] 3.2 Record policy changes and allowed/blocked/applied skill decisions
  with entry point, skill, decision, outcome, and redacted payload references.
  - Current evidence records Settings > Skills profile policy changes in
    `audit_events` as `profile_agent_authority_policy` rows with non-secret
    before/after policy fields. App-side Agent Skill authority reviews now
    persist allowed/blocked decision rows with entry point, skill, outcome,
    blocker, source context, and redacted parameter-key references; successful
    direct apply calls also append an `applied` outcome row. MCP runtime
    `tools/call` authority reviews emit redacted receipt rows for allowed and
    blocked decisions without storing tool argument values.

## 4. Evidence

- [x] 4.1 Add focused unit coverage for the shared authority guard.
- [x] 4.2 Add API and end-to-end coverage for allowed, blocked, and bypass
  attempt paths across Chat, side panel, direct API, MCP, and Telegram.
  - Current evidence includes direct Skill API, Chat action API, Telegram Agent
    text API, MCP `tools/call`, and side-panel Agent Skill read-only blocker
    coverage, plus Settings > Skills policy save coverage. Side-panel E2E
    evidence now also covers a late-apply bypass attempt: preview is prepared in
    `ask_before_local_changes`, the profile is changed to `read_only`, the user
    confirms apply, and the direct API apply path returns
    `agent_authority_read_only` without creating the inventory item.
- [x] 4.3 Add restart plus backup/restore persistence coverage.
  - `TestAgentAuthorityPolicySurvivesRestartAndBackupRestore` proves the
    profile authority policy and redacted policy/decision audit rows survive a
    database restart and a backup/restore round trip after an intervening
    policy mutation.
- [ ] 4.4 Run strict OpenSpec validation and record results on #1932.
