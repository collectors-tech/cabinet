# Agent Live Telegram Channel Checklist

Issue: #1773
Parent: #1716
PR context: #1772

This checklist is the live-channel validation gate for the #1716 Agent acceptance suite. It must be completed only with non-secret evidence. Do not record bot tokens, raw private message content, personal identifiers, or full chat transcripts.

## Preconditions

- Cabinet runtime branch and commit:
- Runtime URL and `/api/runtime` evidence:
- Telegram channel setup state:
- Authorized sender/chat configured:
- Unauthorized sender/chat available:
- Operator approval for live-channel validation:

## Authorized Text Intake

- Source message id:
- Cabinet thread id:
- Cabinet message id:
- Workflow run / preview id:
- Response or deep-link state:
- Review state:
- Mutation state before confirmation:
- Runtime/log evidence path:
- Result: pending

Expected result: the authorized text request creates auditable Agent thread, message, workflow, response/deep-link, and review evidence without applying a mutation before confirmation.

## Authorized Media Intake

- Source message id:
- Cabinet thread id:
- Cabinet message id:
- Attachment/media evidence:
- Workflow run / preview id:
- Response or deep-link state:
- Review state:
- Mutation state before confirmation:
- Runtime/log evidence path:
- Result: pending

Expected result: the authorized media request creates auditable Agent thread, message, attachment/media context, workflow, response/deep-link, and review evidence without applying a mutation before confirmation.

## Unauthorized Sender Rejection

- Source message id:
- Rejection response evidence:
- Record absence check:
- Mutation absence check:
- Runtime/log evidence path:
- Result: pending

Expected result: the unauthorized request is rejected before Cabinet creates Agent thread, message, attachment, Inbox, preview, workflow-run, or mutation records.

## Closure Notes

- Completed by:
- Completion date:
- linked issue/PR comment:
- Residual blockers:
