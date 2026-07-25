## 1. Contract

- [x] 1.1 Define the profile-scoped Agent authority modes and shared decision
  guard contract.
- [ ] 1.2 Add OpenSpec traceability rows for the implemented policy and
  cross-entry-point enforcement evidence.

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
- [ ] 3.2 Record policy changes and allowed/blocked/applied skill decisions
  with entry point, skill, decision, outcome, and redacted payload references.
  - Current evidence records Settings > Skills profile policy changes in
    `audit_events` as `profile_agent_authority_policy` rows with non-secret
    before/after policy fields. App-side Agent Skill authority reviews now
    persist allowed/blocked decision rows with entry point, skill, outcome,
    blocker, source context, and redacted parameter-key references. Remaining
    follow-up should extend equivalent decision receipts to MCP runtime calls
    and applied-result outcomes where they happen outside the app review helper.

## 4. Evidence

- [x] 4.1 Add focused unit coverage for the shared authority guard.
- [ ] 4.2 Add API and end-to-end coverage for allowed, blocked, and bypass
  attempt paths across Chat, side panel, direct API, MCP, and Telegram.
  - Current evidence includes direct Skill API, Chat action API, Telegram Agent
    text API, MCP `tools/call`, and side-panel Agent Skill read-only blocker
    coverage, plus Settings > Skills policy save coverage. Remaining follow-up
    should fill any missing allowed/bypass matrix cells before closing the item.
- [ ] 4.3 Add restart plus backup/restore persistence coverage.
- [ ] 4.4 Run strict OpenSpec validation and record results on #1932.
