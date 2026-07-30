## Why

#1933 is now unblocked because the provider runtime boundary (#1481), profile
authority policy (#1932), and Agent context envelope (#1714) are merged. Cabinet
still answers ordinary Chat prompts with deterministic phrase matching or
hard-coded fallback text, so natural-language Chat cannot safely choose and run
registered Agent Skills.

## What Changes

- Add a governed Chat planner/dispatcher contract that consumes assistant
  provider output as structured planning input, not as permission to mutate
  Cabinet state.
- Expose only enabled and available Agent Skills, their schemas, safety metadata,
  and the active Agent context envelope to the planner.
- Route read-only skill selections through the same policy, validation, audit,
  and profile-isolation guards used by direct Agent Skill execution.
- Route mutations to preview-only results and require the existing explicit
  confirmation/apply path before any write.
- Add deterministic fake-provider tests for skill selection, clarification,
  read-only execution, preview/confirm, denial, and replay/idempotency.
- Preserve provider, entry point, context, selected skill, preview token, apply
  outcome, and error evidence without leaking secrets or raw provider payloads.

## Capabilities

### Modified Capabilities

- `assistant-execution-surfaces`: Chat becomes the natural-language planner entry
  point while Cabinet keeps tool selection, preview, confirmation, and mutation
  authority inside governed execution services.
- `chat-copilot`: main Chat and side-panel Chat use the same planner contract and
  context envelope for provider-backed skill selection and recoverable failures.
- `agent-universal-channels`: natural-language planning reuses the canonical
  profile, route, selection, thread, attachment, permission, setup, and audit
  envelope.

## Impact

- Affected code: `internal/ai`, `internal/app`, `internal/chat`,
  `internal/agentskills`, and Agent context helpers.
- Affected tests: deterministic fake-provider planner Go/API tests, Chat and
  side-panel Cypress coverage, policy-denial and replay/idempotency tests, and
  packaged Windows smoke for one provider-backed read plus one confirmed local
  write.
- Affected documentation: OpenSpec specs, traceability, and #1701 Agent coverage
  matrix.
- Related issues: `#1933`, `#1701`, `#1481`, `#1932`, `#1714`, `#1707`.
