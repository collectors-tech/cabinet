## 1. Contract

- [x] 1.1 Define the canonical Agent context envelope fields and supported
  launch surfaces for #1714.
- [ ] 1.2 Add OpenSpec traceability rows mapping the context-envelope
  requirements to implementation evidence.

## 2. Shared context model

- [x] 2.1 Add a shared server/client context-envelope type or adapter for main
  Chat, side-panel Chat, and Agent Skill dispatch.
  - Added `internal/agentcontext` with a canonical envelope normalizer and
    wired `/api/chat/messages` to persist `agent_context` for chat messages.
- [ ] 2.2 Preserve profile, route/surface, thread, selected record, intent,
  attachment/media, source channel, permission/setup, and workflow/audit IDs in
  Agent requests.
- [ ] 2.3 Return clarification/setup guidance when required profile, route,
  selection, provider, permission, or setup context is missing.

## 3. Entry points

- [x] 3.1 Ensure main `/chats` and side-panel Chat use the same envelope model.
  - `TestChatMessagesNormalizeAgentContextEnvelopeForMainAndSidePanel` covers
    main Chat and side-panel Chat requests preserving profile/thread/channel
    fields through the same `agent_context` shape while retaining surface and
    selected-record differences.
- [ ] 3.2 Add at least one supported table/detail surface launch path that
  passes selected-record context.
- [ ] 3.3 Preserve profile/thread/workflow context across governed route changes
  while keeping side-panel Chat state available.
- [ ] 3.4 Record context evidence in workflow/action/audit metadata without
  storing secrets or invented targets.

## 4. Evidence

- [ ] 4.1 Add focused Go/API coverage for main Chat context, side-panel context,
  selected item context, missing context, and route-change continuity.
- [ ] 4.2 Add Cypress coverage for side-panel Agent context from at least one
  table/detail surface.
- [ ] 4.3 Run touched UI validation, focused Go tests, strict OpenSpec
  validation, and `git diff --check`.
