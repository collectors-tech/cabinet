## 1. Contract

- [x] 1.1 Define the governed natural-language Chat planner contract, provider
  input boundary, skill exposure rules, confirmation boundary, and evidence
  requirements for #1933.
- [ ] 1.2 Add OpenSpec traceability rows for the #1933 planner requirements and
  the #1701 Agent coverage matrix.

## 2. Planner and provider boundary

- [x] 2.1 Add a deterministic fake-provider planner test path for structured
  skill selections without network calls.
- [x] 2.2 Invoke the active healthy assistant provider through the #1481
  provider-neutral runtime boundary when Chat needs natural-language planning.
- [ ] 2.3 Supply only enabled/available skill metadata, JSON schemas, safety
  declarations, and the canonical Agent context envelope to the planner.
- [ ] 2.4 Reject or clarify ambiguous, unsupported, disabled, or unavailable
  selections without fabricating completed work.

## 3. Governed execution

- [ ] 3.1 Execute policy-approved read-only selections directly with grounded
  Cabinet results and profile isolation.
- [ ] 3.2 Convert local-write selections into previews only, then apply exactly
  once through the existing confirmation endpoint.
- [ ] 3.3 Preserve selected-record context for update/rename requests and ask for
  clarification when the selected target is absent or ambiguous.
- [ ] 3.4 Keep provider/tool failures recoverable with redacted next-action
  guidance in both main Chat and side-panel Chat.

## 4. Evidence

- [ ] 4.1 Record provider, entry point, selected skill, context, preview/apply
  token state, decision, and final outcome in workflow/action evidence without
  secrets or raw provider payloads.
- [ ] 4.2 Add Go/API tests for tool selection, clarification, read-only
  execution, preview/confirm/apply, denial, and replay/idempotency.
- [ ] 4.3 Add side-panel/main Chat coverage for the shared planner contract.
- [ ] 4.4 Add packaged Windows smoke evidence for a provider-backed
  conversational read and a confirmed local write.
- [ ] 4.5 Run strict OpenSpec validation, focused Go/API/UI tests, packaged smoke
  where required, and `git diff --check`.
