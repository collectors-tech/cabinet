## MODIFIED Requirements

### Requirement: Chat-driven mutations SHALL require preview and confirmation
Cabinet SHALL require preview and explicit confirm for mutation actions.

#### Scenario: Plan natural-language Chat actions without free-form mutation
- **GIVEN** a user sends a natural-language request in main Chat or side-panel Chat
- **WHEN** the request could map to a Cabinet Agent Skill
- **THEN** Chat SHALL route the request through the governed planner contract instead of inferring mutations from free-form phrase matching alone
- **AND** Chat SHALL preserve profile, thread, route, source surface, selected-record, attachment, permission, setup, and audit context from the canonical Agent context envelope
- **AND** Chat SHALL ask a focused clarification when required identifiers, selected records, provider setup, or user intent are missing or ambiguous
- **AND** Chat SHALL show an actionable fallback when no healthy provider is configured or when the provider cannot return a structured supported selection

#### Scenario: Confirm planned Chat mutation exactly once
- **GIVEN** the governed planner selected a mutating Agent Skill from a natural-language Chat request
- **WHEN** Cabinet creates a preview and the user confirms apply
- **THEN** Chat SHALL apply through the existing Agent Skill confirmation endpoint and revalidate the token, profile, authority policy, selected target, and target state
- **AND** replaying the same apply token SHALL NOT duplicate the mutation
- **AND** read-only mode, disabled skills, external-write policy, destructive confirmation, stale-preview, and validation failures SHALL block apply with truthful redacted guidance
- **AND** provider, entry point, selected skill, preview token state, confirmation outcome, and final result SHALL be recorded in workflow or action evidence without secrets or raw provider payloads
