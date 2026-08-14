# Cabinet Agent and Chats

Cabinet Agent is the conversational way to understand and drive Cabinet. Use
the compact Agent beside the screen you are working on, then expand the same
conversation into Chats when you need more room or want to review its history.

## What stays with the conversation

- Active profile, Cabinet route, and selected-record context
- Provider and model assigned to the thread
- Messages, attachments, pending previews, and Action Timeline evidence
- The same confirmation boundary when moving between compact and full views

## Common actions

- Open **Cabinet Agent** from the shell by pointer or keyboard.
- Ask what the Agent can do to see registry-derived availability, setup, and
  confirmation states for the active profile.
- Use **Open this thread in full Cabinet Agent** to continue the exact thread in
  Chats; returning to the compact Agent does not create a new thread.
- Review setup guidance or a proposed change before continuing.
- Confirm a local mutation only after Cabinet shows its target and impact.
- Use Action Timeline to review what Cabinet planned and what actually ran.

## How Agent actions work

Ask for an outcome in ordinary language. Cabinet selects the governed
capability and preserves the active profile, thread, route, and selected-record
context. The primary Chat experience does not ask you to choose internal skill
IDs or carry provider, setup, secret, or mutation parameters through a form.

Read-only answers can return immediately. Changes use a server-owned review
card and are not applied until you explicitly confirm them. When provider or
permission setup is required, Cabinet links to the owning **Settings** or
**Integrations** screen; credentials stay on that screen and are never echoed
in Chat.

Cabinet Agent does not give an AI provider direct database, shell, secret, or
provider-write authority. A provider can propose a governed skill; Cabinet
validates the active profile policy and context, controls dispatch, requires
confirmation where applicable, and records the outcome.
