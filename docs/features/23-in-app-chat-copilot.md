# 23 In-App Chat Copilot

## Feature
- Name: In-App Chat Copilot
- ID: 23
- Status: draft

## Goal
- Provide a universal in-app chat assistant that helps users understand their collection and execute guided actions quickly.

## Scope
- In scope:
  - Persistent chat UI docked on right rail.
  - Open/close toggle and keyboard open/focus behavior.
  - Context-aware assistant (active profile/screen/item/candidate/filter).
  - Suggested actions with confirm-before-apply for mutations.
  - Per-profile local thread history.
- Out of scope:
  - Fully autonomous mutation without confirmation.
  - Cloud account dependency.

## User Stories
- As a collector, I want to ask the app what needs attention and what to buy next.
- As a collector, I want to open and close chat at any time without losing my current screen state.
- As a collector, I want suggested add/update actions I can confirm in one step.

## Functional Requirements
- FR-1: Chat panel can be opened and closed from any main screen.
- FR-2: Chat panel state (open/closed + width) persists per profile.
- FR-3: Chat composer supports free text and contextual prompts.
- FR-4: Chat suggestions can launch app actions (`create item`, `update`, `wishlist`, `track`) only after explicit confirmation.
- FR-5: Chat context sharing is user-configurable and least-privilege by default.
- FR-6: Every assistant-initiated action is logged in diagnostics.

## API and Integration Touchpoints
- Endpoints/services:
  - `POST /api/chat/message`
  - `POST /api/chat/action/preview`
  - `POST /api/chat/action/apply`
  - `GET /api/chat/threads`
  - `POST /api/chat/threads`
- External providers:
  - OpenAI model endpoint (user-provided key)

## Data Model Touchpoints
- Tables/entities:
  - `chat_threads`
  - `chat_messages`
  - `chat_action_logs`
  - `chat_panel_preferences`
- Settings/secrets:
  - OpenAI key in secure storage
  - chat context-scope settings in profile settings

## UX Flow
- Entry point:
  - Header chat button, keyboard shortcut, optional quick-action button from dashboard.
- Primary path:
  - Open panel -> ask question -> receive answer with suggested actions -> preview -> confirm apply.
- Failure/recovery path:
  - Provider/API error shown inline with retry and fall back to manual action links.

## Interaction Pattern Reference (VS Code/Codex style)
- Persistent sidebar chat with optional right-sidebar placement.
- Context add controls for files/selection equivalent in Cabinet (active item/candidate/filter).
- Slash-command style shortcuts for common intents.
- Approval modes for actions before execution.

References:
- https://openai.com/index/introducing-codex/
- https://developers.openai.com/codex/ide
- https://developers.openai.com/codex/ide/extensions

## Acceptance Criteria
- [ ] AC-1: Chat can open/close globally without losing current workspace context.
- [ ] AC-2: Right-rail layout is responsive and does not break main workflow.
- [ ] AC-3: Action preview/confirm is required for mutations.
- [ ] AC-4: Thread history persists per profile and can be resumed.
- [ ] AC-5: Chat failure states are actionable and do not block manual workflows.

## Test Strategy
- Unit: prompt/context building, action preview validation, permission checks.
- API: chat/thread/action endpoints success + error + guardrails.
- E2E:
  - Open/close chat from multiple screens.
  - Ask collection question with context.
  - Preview and confirm `add item` action.
  - Verify no mutation occurs without explicit confirm.

## Dependencies
- Internal: profile context, item/discovery/pricing services, logging, UI shell layout.
- External: OpenAI service availability and key validity.

## Risks
- Risk: assistant responses look authoritative without enough confidence.
- Mitigation: explicit confidence and preview-only mutations.

## Telemetry and Diagnostics
- Events/logs:
  - chat opened/closed
  - message sent/received
  - action preview requested
  - action apply success/failure
- Error signals:
  - provider unavailable
  - invalid key
  - policy/permission denied

## Open Questions
- Q1: Which actions are allowed in v1 apply mode vs preview-only?
- Q2: Should chat be free-tier limited by message count?
