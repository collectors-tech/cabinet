## MODIFIED Requirements

### Requirement: Cabinet Agent SHALL route work through skills and governed execution boundaries
Cabinet Agent SHALL use the Agent Skill Registry, capability registry, guided
workflow registry, shell command bus, preview/apply handlers, Action Timeline,
audit records, and active profile Agent authority policy as the governed
execution model for supported work.

#### Scenario: Execute read-only skill
- **GIVEN** a user invokes a read-only Agent skill with sufficient profile, route, and selection context
- **WHEN** Agent dispatches the skill
- **THEN** Cabinet SHALL allow the skill to read or summarize Cabinet state without mutating records
- **AND** the result SHALL identify the skill id, source surface, and non-secret evidence used

#### Scenario: Preview mutating skill before apply
- **GIVEN** a user invokes an Agent skill that can create, update, delete, import, export, or call an external write path
- **WHEN** Agent prepares the work
- **THEN** Cabinet SHALL return a preview, target summary, source context, confirmation requirement, and audit destination before any mutation is applied
- **AND** cancellation, stale context, missing permission, failed provider setup, and failed apply states SHALL leave retryable Action Timeline or audit evidence without mutating early

#### Scenario: Entry point cannot bypass profile Agent authority
- **GIVEN** a user or external channel invokes an Agent Skill from main Chat, Assistant side panel, direct Skill API, MCP, Telegram, or another approved Agent entry point
- **WHEN** the request reaches Cabinet's skill dispatch boundary
- **THEN** Cabinet SHALL apply the same active-profile Agent authority policy before dispatching the skill or lower-level capability
- **AND** a read-only profile SHALL block all mutation apply attempts even when the request is crafted directly against an API endpoint
- **AND** allowed and blocked decisions SHALL preserve the non-secret entry point and source context needed for audit review
