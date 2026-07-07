# Agent Live Telegram Channel Checklist

Issue: #1773
Parent: #1716
PR context: #1772
Traceability source: `openspec/traceability/agent-live-telegram-channel-checklist.md`

This is the human-readable validation target for the live Telegram-channel gate. Record only non-secret evidence here. Do not record bot tokens, raw private message content, personal identifiers, full chat transcripts, or unredacted sender/chat identifiers.

## Current Status

Status: blocked - live sender/chat setup is not configured on the active demo2 profile.

Last checked: 2026-07-08 03:20 AEST

Evidence:

- Demo runtime: `.work-agent/logs/issue-1773/demo2-runtime-precondition.json`
- Demo health: `.work-agent/logs/issue-1773/demo2-healthz-precondition.txt`
- Telegram provider health: `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json`

Non-secret provider-health result:

- `provider=telegram`
- `status=needs_config`
- `state=disabled`
- `code=TELEGRAM_SENDER_CHAT_REQUIRED`
- `sender_chat_authorized=false`
- `bot_token_present=false`
- `webhook_configured=false`
- `credential_returned=false`
- `next_action=authorize_sender_chat`

Because the live channel is not configured, no live authorized text, authorized media, unauthorized sender, or Bot API delivery-failure scenarios were executed in this run. Fixture/proof-packet evidence remains separate from this live-channel checklist.

## Preconditions

- Cabinet runtime branch and commit: `develop`, `rev-0b629fc54569`
- Runtime URL and `/api/runtime` evidence: `http://127.0.0.1:17882`, `.work-agent/logs/issue-1773/demo2-runtime-precondition.json`
- Telegram channel setup state: blocked by `TELEGRAM_SENDER_CHAT_REQUIRED`
- Authorized sender/chat configured: no
- Unauthorized sender/chat available: not verified because authorized setup is missing
- Operator approval for live-channel validation: not verified in repo/runtime state

## Authorized Text Intake

- Source message id: not run
- Cabinet thread id: not run
- Cabinet message id: not run
- Workflow run / preview id: not run
- Response or deep-link state: not run
- Review state: not run
- Mutation state before confirmation: not run
- Runtime/log evidence path: `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json`
- Result: blocked by missing live Telegram sender/chat setup

Expected result: the authorized text request creates auditable Agent thread, message, workflow, response/deep-link, and review evidence without applying a mutation before confirmation.

## Authorized Media Intake

- Source message id: not run
- Cabinet thread id: not run
- Cabinet message id: not run
- Attachment/media evidence: not run
- Workflow run / preview id: not run
- Response or deep-link state: not run
- Review state: not run
- Mutation state before confirmation: not run
- Runtime/log evidence path: `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json`
- Result: blocked by missing live Telegram sender/chat setup

Expected result: the authorized media request creates auditable Agent thread, message, attachment/media context, workflow, response/deep-link, and review evidence without applying a mutation before confirmation.

## Unauthorized Sender Rejection

- Source message id: not run
- Rejection response evidence: not run
- Record absence check: not run
- Mutation absence check: not run
- Runtime/log evidence path: `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json`
- Result: blocked by missing live Telegram sender/chat setup

Expected result: the unauthorized request is rejected before Cabinet creates Agent thread, message, attachment, Inbox, preview, workflow-run, or mutation records.

## Bot API Delivery Failure

- Source message id: not run
- Cabinet workflow state: not run
- Telegram delivery method/status/body evidence: not run
- Secret-return check: `credential_returned=false` in provider-health evidence
- Runtime/log evidence path: `.work-agent/logs/issue-1773/telegram-provider-health-precondition.json`
- Result: blocked by missing live Telegram sender/chat, bot token, and webhook setup

Expected result: outbound delivery failure is recorded without losing Cabinet workflow state and without exposing secrets.

## Closure Notes

- Completed by: pending live-channel setup
- Completion date: pending live-channel setup
- linked issue/PR comment: pending this branch/PR handoff
- Residual blockers: live authorized sender/chat setup, bot token presence, webhook setup, and operator-approved non-secret source-message capture
