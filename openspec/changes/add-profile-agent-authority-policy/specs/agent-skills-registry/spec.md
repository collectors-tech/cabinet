## MODIFIED Requirements

### Requirement: Skill permissions and context SHALL be explicit before execution
Cabinet SHALL require skills to declare the profile, route, selection, provider,
permission, and data-access context needed to run safely, and SHALL evaluate the
active profile's Agent authority policy before any skill preview or apply
dispatch.

#### Scenario: Missing context asks for clarification
- **GIVEN** a skill requires an active profile, selected item, selected collection, provider setup, integration credential, attachment, or route context
- **WHEN** the user invokes the skill without the required context
- **THEN** Cabinet SHALL return a clarification or setup-needed response instead of guessing a target
- **AND** the response SHALL identify the missing context and the next safe user action

#### Scenario: Permissions are visible and auditable
- **GIVEN** a skill can read, write, import, export, delete, configure, or call an external provider
- **WHEN** the registry or Skills page displays the skill
- **THEN** Cabinet SHALL display a permission declaration that separates Cabinet-local reads, Cabinet-local writes, external reads, external writes, secret access, and destructive operations
- **AND** any applied write or external operation SHALL create Action Timeline/audit evidence with non-secret payload references

#### Scenario: Profile authority policy gates skill execution
- **GIVEN** a profile has an Agent authority mode of `read_only`, `ask_before_local_changes`, or `approved_external_actions`
- **WHEN** Cabinet reviews a skill invocation from Chat, Assistant side panel, direct API, MCP, Telegram, or another approved Agent entry point
- **THEN** Cabinet SHALL evaluate the skill permission declaration, requested profile, entry point, confirmation state, and external-write approval through one shared guard before preview or apply
- **AND** read-only skills SHALL remain executable when their required context is present
- **AND** mutating local, external-write, and destructive skills SHALL be blocked for read-only profiles, including crafted direct API calls
- **AND** default `ask_before_local_changes` mode SHALL allow local-write previews but SHALL block apply until explicit confirmation
- **AND** external-write skills SHALL remain blocked until the profile separately approves external actions and the individual invocation is confirmed
- **AND** destructive skills SHALL require action-specific strong confirmation in every mode
- **AND** profile mismatches SHALL be blocked before dispatch
- **AND** the decision SHALL identify entry point, skill id, mode, decision, blocker, and next action without recording secrets or sensitive payload values
