# Add Agent Context Entry Points

## Why

#1714 needs a durable contract for opening Cabinet Agent from multiple app
surfaces without losing profile, route, thread, selection, channel, permission,
or workflow context. Existing Agent Skills and authority-policy work governs
execution, but the launch/context model still needs a shared envelope so main
Chat, side-panel Chat, selected table/detail surfaces, Inbox review, and future
external channels do not each invent their own shape.

## What Changes

- Define a canonical Agent context envelope for launch and dispatch requests.
- Require main Chat and side-panel Chat to preserve the same envelope fields.
- Require supported table/detail surfaces to include selected-record context.
- Require missing context to produce clarification/setup guidance, not guessed
  targets.
- Require route-change continuity to preserve profile/thread/workflow state.
- Define evidence tasks for API tests, Cypress side-panel coverage, OpenSpec,
  and traceability updates.

## Impact

- Affects `agent-universal-channels` requirements.
- Expected implementation areas include Chat APIs, assistant side-panel state,
  route/surface context helpers, Agent Skill dispatch adapters, Action
  Timeline/workflow metadata, and focused E2E coverage.
