## Why

Cabinet Agent Skills already declare safety and permission metadata, but profiles
do not yet have one persisted authority policy that every Agent entry point can
enforce before preview or apply. That leaves read-only, local-change,
external-write, and destructive decisions spread across individual surfaces.

## What Changes

- Add a profile-scoped Agent authority policy with read-only,
  ask-before-local-changes, and approved-external-actions modes.
- Route Agent Skill permission decisions through one shared guard before Chat,
  Assistant side panel, direct Skill API, MCP, Telegram, or future channel
  dispatch can preview or apply work.
- Keep read-only skills executable in every mode while blocking mutating skills
  for read-only profiles, including crafted direct API calls.
- Permit local-write previews in the default mode, but require explicit
  confirmation before apply.
- Require separate external-write profile approval and per-action confirmation
  before any external-write skill can apply.
- Keep destructive skills behind action-specific strong confirmation in every
  mode.
- Persist policy and audit decisions without recording secrets or sensitive
  payloads.

## Capabilities

### Modified Capabilities

- `agent-skills-registry`: add profile-scoped authority policy review for skill
  preview/apply decisions and cross-entry-point enforcement.
- `agent-universal-channels`: require all Agent entry points to use the same
  profile authority decision before dispatching skills or lower-level actions.

## Impact

- Affected code: `internal/agentskills`, `internal/app`, `internal/chat`,
  `internal/mcpserver`, `internal/telegramcapture`, `ui.web`.
- Affected tests: Agent Skill guard unit tests, direct API tests, Chat and
  Assistant UI tests, Telegram/API bypass tests, backup/restore persistence
  coverage.
- Affected documentation: Settings > Skills copy, OpenSpec traceability, issue
  evidence.
- Related issues: `#1932`, `#1935`, `#1934`, `#1714`.
