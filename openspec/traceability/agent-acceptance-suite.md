# Agent Acceptance Suite Evidence Map

Issue: #1716

This map defines the first acceptance-suite slice for universal Agent work. It deliberately separates fixture/proof-packet evidence from live production Telegram-channel evidence so validation cannot treat local fixtures as a live-channel claim.

| Scenario group | Requirement IDs | Evidence status | Validation targets | Required proof |
| --- | --- | --- | --- | --- |
| Main Chat read-only app work | AGENT-UNIVERSAL-CHANNELS-001, AGENT-UNIVERSAL-CHANNELS-002 | implemented-fixture | `ui.web/cypress/e2e/chats/chats-workspace/spec.cy.ts` (`CHATS-WORKSPACE-008/#1503 dispatches normal main Chat text to app-control route planning without Inbox noise`) | Route-planning workflow run has `confirmation_state=not_required`, leaves the user on `/chats/` until explicit navigation, and does not create Inbox noise. |
| Side-panel Chat read-only app work | AGENT-UNIVERSAL-CHANNELS-001, AGENT-UNIVERSAL-CHANNELS-002 | implemented-fixture | `ui.web/cypress/e2e/chats/assistant-workspace/spec.cy.ts` (`ASSISTANT-WORKSPACE-009/#1503 dispatches normal side-panel text to app-control route planning without Inbox noise`) | Route/profile/thread context is sent with the message, the action timeline records `navigate.open_surface`, and the app route remains stable until explicit navigation. |
| Preview/cancel/apply mutation | AGENT-UNIVERSAL-CHANNELS-002 | implemented-fixture | `ui.web/cypress/e2e/chats/assistant-execution-surfaces/spec.cy.ts` (`ASSISTANT-EXECUTION-001/002/003/004 renders preview-before-apply with confirm and explicit permission guidance`); `internal/chat/service_test.go` (`TestServiceThreadMessagePreviewApplyLifecycle`) | Preview is created before mutation, cancel leaves state unchanged, confirm applies, and Action Timeline/audit evidence records the result. |
| Attachment success/failure | AGENT-UNIVERSAL-CHANNELS-003 | implemented-fixture | `internal/app/chat_api_test.go` (`TestChatAPIsThreadMessageAttachmentAndPreviewApply` attachment assertions); `ui.web/cypress/e2e/chats/assistant-workspace/spec.cy.ts` (`AGENT-ATTACHMENTS-001 handles side-panel attachments with the same scoped message binding as main Chat`) | Supported attachment is scoped to profile/thread/message with provenance; unsupported local-path input returns deterministic validation guidance without unrelated records; cross-thread attachment reuse is rejected; side-panel Chat uploads explicit user-selected attachments and binds them to the next message with route/profile/model context. |
| Telegram authorized text/media intake | AGENT-UNIVERSAL-CHANNELS-004, AGENT-UNIVERSAL-CHANNELS-005 | implemented-fixture | `internal/app/telegram_capture_api_test.go` (`TestTelegramCatalogCaptureAPIRequiresPersistedSenderAuthorization`, `TestTelegramCatalogCaptureWebhookAPIResolvesProfileAuthorization`) | Authorized sender/chat creates thread/message/workflow/preview records, preserves media provenance, and does not mutate inventory before confirmation. |
| Telegram unauthorized sender rejection | AGENT-UNIVERSAL-CHANNELS-004 | implemented-fixture | `internal/app/telegram_capture_api_test.go` (`TestTelegramCatalogCaptureAPIRequiresPersistedSenderAuthorization`, `TestTelegramCatalogCaptureWebhookAPIResolvesProfileAuthorization`) | Unauthorized sender/chat is rejected before creating Agent thread, message, attachment, Inbox, preview, workflow-run, or mutation records. |
| Telegram proof-packet versus live-channel distinction | AGENT-UNIVERSAL-CHANNELS-004, AGENT-UNIVERSAL-CHANNELS-005 | partial | `internal/app/telegram_capture_api_test.go` (`TestTelegramExternalIntakeProofRequiresAuthorizedProviderEvidence`); `openspec/traceability/agent-live-telegram-channel-checklist.md`; follow-up #1773 | Non-secret proof packet records provider/runtime evidence; live production Telegram-channel validation remains pending until the manually approved live-channel checklist in #1773 is completed. |

## Live-Channel Validation Gate

#1716 is not complete until the final PR or follow-up validation comment includes either:

- a manual live Telegram-channel checklist with non-secret sender/chat setup state, source message id, response/deep-link/review state, and runtime proof; or
- an explicit blocker/follow-up issue explaining why production-channel validation is unavailable.

Current live-channel follow-up: #1773 (`openspec/traceability/agent-live-telegram-channel-checklist.md` and `docs/validation/agent-live-telegram-channel-checklist.md`). #1716 must remain open or explicitly blocked until that checklist is completed or superseded by an approved blocker/follow-up.

Current #1773 precondition result: blocked by missing live Telegram sender/chat setup on demo2. The non-secret provider-health evidence is recorded in `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json` and reports `TELEGRAM_SENDER_CHAT_REQUIRED`, `sender_chat_authorized=false`, `bot_token_present=false`, `webhook_configured=false`, and `credential_returned=false`.
