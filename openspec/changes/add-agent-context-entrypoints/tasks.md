## 1. Contract

- [x] 1.1 Define the canonical Agent context envelope fields and supported
  launch surfaces for #1714.
- [x] 1.2 Add OpenSpec traceability rows mapping the context-envelope
  requirements to implementation evidence.
  - Added `AGENT-UNIVERSAL-CHANNELS-007` to `openspec/traceability.md`
    and updated `openspec/traceability/agent-skill-coverage.md` to bind
    main Chat, side-panel Chat, selected-record launches, route continuity,
    and workflow evidence to #1714 tests.

## 2. Shared context model

- [x] 2.1 Add a shared server/client context-envelope type or adapter for main
  Chat, side-panel Chat, and Agent Skill dispatch.
  - Added `internal/agentcontext` with a canonical envelope normalizer and
    wired `/api/chat/messages` to persist `agent_context` for chat messages.
- [x] 2.2 Preserve profile, route/surface, thread, selected record, intent,
  attachment/media, source channel, permission/setup, and workflow/audit IDs in
  Agent requests.
  - Added Agent Skill preview/apply request normalization from
    `agent_context` so skill dispatch inherits profile, source
    surface/channel/thread, selected-record IDs, route/setup/workflow context,
    media IDs, and attachment IDs without exposing audit-only IDs in preview
    responses.
- [x] 2.3 Return clarification/setup guidance when required profile, route,
  selection, provider, permission, or setup context is missing.
  - Added a pre-authority clarification response for Agent Skill preview/apply
    requests launched with `agent_context`; missing route, explicit selected
    target, provider, and setup readiness now return actionable
    `missing_context` guidance instead of invented direct API placeholders.

## 3. Entry points

- [x] 3.1 Ensure main `/chats` and side-panel Chat use the same envelope model.
  - `TestChatMessagesNormalizeAgentContextEnvelopeForMainAndSidePanel` covers
    main Chat and side-panel Chat requests preserving profile/thread/channel
    fields through the same `agent_context` shape while retaining surface and
    selected-record differences.
- [x] 3.2 Add at least one supported table/detail surface launch path that
  passes selected-record context.
  - Inventory row selection now persists a canonical selected-record bridge for
    side-panel Agent launches and sends `inventory.item.detail` context to
    Agent Skill preview/apply requests.
- [x] 3.3 Preserve profile/thread/workflow context across governed route changes
  while keeping side-panel Chat state available.
  - Side-panel chat messages now send the canonical `agent_context` envelope,
    and `AGENT-CONTEXT-004/#1714` Cypress coverage proves the same thread
    remains active after governed navigation while the next message records the
    changed route.
- [x] 3.4 Record context evidence in workflow/action/audit metadata without
  storing secrets or invented targets.
  - App-control workflow runs now store sanitized `agent_context` evidence in
    workflow input for Action Timeline review, including profile/source
    surface/channel/route/thread/workflow/audit fields while omitting arbitrary
    secret-looking request fields.

## 4. Evidence

- [x] 4.1 Add focused Go/API coverage for main Chat context, side-panel context,
  selected item context, missing context, and route-change continuity.
  - `TestChatMessagesNormalizeAgentContextEnvelopeForMainAndSidePanel` now
    covers explicit top-level chat `agent_context` normalization for side-panel
    selected-record context.
  - `TestChatMessageAppControlPlannerDispatchesDeterministicActions` covers
    sanitized context evidence in app-control workflow runs.
  - `TestAgentSkillPreviewNormalizesAgentContextEnvelope` covers selected item,
    media, attachment, source channel, permission/setup, workflow, and audit
    context propagation into Agent Skill preview/apply dispatch.
  - `TestAgentSkillPreviewClarifiesMissingAgentContext` covers deterministic
    missing-route, selection, provider, and setup guidance before authority or
    mutation preview.
  - `AGENT-CONTEXT-004/#1714` covers route-change continuity for side-panel
    Chat at the UI workflow level.
- [x] 4.2 Add Cypress coverage for side-panel Agent context from at least one
  table/detail surface.
  - `AGENT-CONTEXT-003/#1714` covers an inventory row launch into side-panel
    Agent Skill preview with route, surface, thread, source channel, and
    selected inventory item context.
- [x] 4.3 Run touched UI validation, focused Go tests, strict OpenSpec
  validation, and `git diff --check`.
  - Final validation evidence:
    `npm run build` from `ui.web`,
    `go test ./internal/app -run 'Test(ChatMessagesNormalizeAgentContextEnvelopeForMainAndSidePanel|ChatMessageAppControlPlannerDispatchesDeterministicActions|AgentSkillPreview(NormalizesAgentContextEnvelope|ClarifiesMissingAgentContext))' -count=1`,
    `openspec validate add-agent-context-entrypoints --strict --no-interactive`,
    and `git diff --check`.
