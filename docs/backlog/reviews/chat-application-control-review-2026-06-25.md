# Chat application control review — 2026-06-25

## Purpose

Max asked for an urgent review of how Chats should work and whether current behaviour is good enough for the requirement that Chat must be able to operate Cabinet.

## Product and spec baseline

Cabinet product docs say Chats/assistant must provide contextual assistant/copilot support, preserve thread history, operate with app context where appropriate, and support workspace-aware assistant panels/defaults.

OpenSpec already defines the important contracts:

- `CHAT-COPILOT-*`: persistent chat, profile-scoped threads/messages, attachments, preview-before-apply mutations.
- `CHATS-WORKSPACE-*`: real `/chats` workspace, stable send flow, clear Assistant versus Chats boundary, two-pane assistant-ui-inspired layout.
- `ASSISTANT-WORKSPACE-*`: route/profile/selection context, thread continuity, provider/model metadata, compact side-panel, governed actions.
- `ASSISTANT-EXECUTION-*`: capability registry, preview-before-apply, explicit confirmation, workflow-run audit, app-control tools such as `navigate.open_surface` and `update_open_item_title`.

## Current implementation findings

Current implementation is not blank. It already has:

- profile-scoped chat threads and messages
- `/chats` route with thread rail, message canvas, composer, attachment area, and manual action preview/apply controls
- shell Assistant workspace side panel with route/profile/selection context, provider/model state, thread selector, composer, and action preview/apply controls
- `/api/chat/capabilities` capability registry
- `/api/chat/workflow-runs` audit/lifecycle records
- `/api/chat/actions/preview`, `/api/chat/actions/apply`, and `/api/chat/actions/cancel`
- hardcoded action apply support for item creation, item update, wishlist entry creation, and collection assignment

## Critical gap

The current Chat/Assistant implementation does not yet truly operate the application from normal chat instructions.

The main gap is that sending a user message usually persists the message and creates an Inbox handoff, but does not route the message through a real app-control planner/dispatcher that can create a governed route action, data preview, workflow run, or app mutation proposal.

The side panel has a tiny hardcoded navigation inference for layout/settings only, but the broader capability registry is not wired to a planner or tool executor.

The `/chats` route action controls are still form-driven developer controls, not assistant-driven operation from chat text.

The capability registry advertises `navigate.open_surface`, but the preview/apply service does not support that action today. It only supports a small set of hardcoded mutation actions.

## What must change urgently

Cabinet needs a P0 app-control path:

1. Chat receives user message plus route/profile/selection/provider/model context.
2. Runtime builds an intent/app-control plan using deterministic rules first and AI/provider planning only where verified.
3. Plan is matched against the capability registry.
4. Read-only route/app-control actions can surface as preview-only route cards and execute safely.
5. Mutating actions must create previews and require explicit confirmation before apply.
6. Every action creates a durable workflow-run/action timeline record.
7. Results link back to the changed/opened Cabinet surface.
8. Inbox is used for durable handoffs/errors/background work, not as the default response for every message.

## Urgent backlog created from this review

- P0: Wire Chat to app-control planner and dispatcher.
- Implement `navigate.open_surface` as a governed app-control capability.
- Bind capability registry entries to preview/apply execution.
- Replace default assistant handoff behaviour with real response/action routing.
- Render durable Action Timeline from workflow runs.
- Add E2E coverage proving Chat operates Cabinet.

## Non-negotiable guardrails

- Chat must not silently mutate data.
- Mutations must remain preview-before-apply and confirm-required.
- Route/open actions are read-only and should not require mutation confirmation.
- The capability registry must be the source of truth for what Chat can do.
- Provider-backed capabilities must truthfully report setup-needed/unavailable instead of pretending readiness.
- Results must be auditable and linked to route/profile/thread/workflow context.
