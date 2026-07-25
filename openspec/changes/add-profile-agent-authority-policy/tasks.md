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
- [ ] 2.4 Apply the shared guard to Chat, Assistant side panel, MCP, and
  Telegram/API dispatch paths.

## 3. Settings and audit

- [ ] 3.1 Add Settings > Skills controls for the effective Agent authority
  mode and external-write approval.
- [ ] 3.2 Record policy changes and allowed/blocked/applied skill decisions
  with entry point, skill, decision, outcome, and redacted payload references.

## 4. Evidence

- [x] 4.1 Add focused unit coverage for the shared authority guard.
- [ ] 4.2 Add API and end-to-end coverage for allowed, blocked, and bypass
  attempt paths across Chat, side panel, direct API, MCP, and Telegram.
- [ ] 4.3 Add restart plus backup/restore persistence coverage.
- [ ] 4.4 Run strict OpenSpec validation and record results on #1932.
