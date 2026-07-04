## Why

Cabinet has working Chat, assistant execution, Telegram capture, and Agent Skill Registry contracts, but the user-facing Agent entry model is still spread across lower-level specs. The next implementation wave needs one product contract that says where Cabinet Agent is available, how attachments are handled across in-app and external channels, and how Telegram-originated work enters the same governed preview/confirm/apply boundary.

## What Changes

- Define universal Cabinet Agent access from `/chats`, side-panel Chat, table/detail screens, Inbox review, and external channels.
- Define Agent expectations for capability explanation, skill listing, guided workflows, read-only work, previewed mutations, confirmation, Action Timeline evidence, and unsupported/setup-needed states.
- Define attachment behavior for main Chat, side-panel Chat, Telegram-originated media, supported/unsupported files, profile/thread/message scoping, provenance, and validation errors.
- Define Telegram/external-channel setup, authorization, text/media intake, thread/message creation, skill routing, preview/confirm/apply boundaries, and non-secret production proof.
- Link this parent specification to the Agent implementation and validation issues.

## Capabilities

### New Capabilities

- `agent-universal-channels`: universal Agent entry points, attachment intake, and governed external-channel routing.

## Impact

- Affected specs: `openspec/changes/define-universal-agent-channel-contracts/specs/agent-universal-channels/spec.md`
- Affected traceability: `openspec/traceability.md`
- Related issues: #1701, #1703, #1704, #1705, #1706, #1708, #1709, #1710, #1711, #1712, #1715, #1716
