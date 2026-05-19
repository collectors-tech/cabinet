## Purpose
Define how Cabinet presents assistant proposals, confirmation boundaries, execution states, and result summaries for governed action-taking workflows.

## Requirements
### Requirement ASSISTANT-EXECUTION-001: Assistant mutations SHALL render preview-before-apply execution surfaces
Assistant-proposed create/update/delete/import actions MUST render a preview surface before execution.

#### Scenario: Preview assistant mutation
- **GIVEN** assistant proposes a structured mutation
- **WHEN** proposal is rendered to user
- **THEN** UI MUST show a preview/summary of intended action and affected entities before apply becomes possible

### Requirement ASSISTANT-EXECUTION-002: Assistant SHALL require explicit confirmation for state-changing actions
Assistant MUST require explicit confirmation for actions that mutate records, import files, or change important application state.

#### Scenario: Confirm before apply
- **GIVEN** assistant action would change Cabinet state
- **WHEN** user has not explicitly confirmed
- **THEN** runtime MUST NOT execute the action

### Requirement ASSISTANT-EXECUTION-003: Assistant SHALL expose deterministic execution lifecycle states
Assistant MUST expose deterministic queued/running/success/failure states for governed action execution.

#### Scenario: Show execution lifecycle
- **GIVEN** assistant action is accepted for execution
- **WHEN** execution lifecycle progresses
- **THEN** UI MUST show deterministic lifecycle states and outcome summary instead of vague loading behavior

### Requirement ASSISTANT-EXECUTION-004: Assistant SHALL make tool permission boundaries visible
Cabinet MUST expose which classes of assistant actions are read-only, preview-only, confirm-required, or unavailable under the active policy.

#### Scenario: Show permission boundary on blocked action
- **GIVEN** assistant proposes an action outside allowed permission scope
- **WHEN** UI renders the proposal or rejection
- **THEN** user MUST receive explicit permission-state guidance instead of silent failure or hidden omission

### Requirement ASSISTANT-EXECUTION-005: Assistant SHALL expose an app-wide capability registry
Cabinet MUST expose a deterministic assistant capability registry for the active profile/workspace so agents can discover supported app functions and their governance boundaries before proposing work.

#### Scenario: Discover governed assistant capabilities
- **GIVEN** an active profile and assistant thread context
- **WHEN** the assistant queries the capability registry for the current route
- **THEN** the response MUST include representative Inventory, Collections, Wishlist, Settings/Data, and Integrations capabilities
- **AND** each capability MUST declare user-facing purpose, required context, permission state, execution mode, preview/apply behavior, audit behavior, and result destination
- **AND** mutating capabilities MUST be marked confirm-required or preview-only rather than directly executable from chat
- **AND** unavailable provider-backed capabilities MUST be returned with setup-needed state instead of being omitted or hallucinated

### Requirement ASSISTANT-EXECUTION-006: OpenAI-backed assistant capabilities SHALL expose truthful setup and readiness states
Cabinet MUST treat OpenAI/API-key/Browser Auth readiness as provider evidence that gates capabilities, not as copy or navigation state.

#### Scenario: Block false OpenAI readiness
- **GIVEN** an assistant capability requires OpenAI-backed processing
- **WHEN** the active profile lacks a verified API key, verified Browser Auth artifact, or passing provider test
- **THEN** the capability MUST report setup-needed or unavailable state
- **AND** the UI/API MUST NOT mark Browser Auth connected from outbound navigation alone
- **AND** provider tests MUST return truthful readiness evidence instead of passing only because a credential-like value exists

### Requirement ASSISTANT-EXECUTION-007: Assistant workflow runs SHALL persist auditable execution records
Cabinet MUST persist OpenAI-backed and agent-backed work as durable workflow run records so previews, confirmations, provider traces, and outcomes are inspectable outside transient chat text.

#### Scenario: Persist assistant workflow run
- **GIVEN** an assistant capability creates, analyzes, enriches, reconciles, tests, or previews work
- **WHEN** the workflow is accepted for execution or queued for provider work
- **THEN** Cabinet MUST persist a run id, workflow id, capability id, profile id, source channel, status, progress timestamps, provider trace, result or error payload, and confirmation state
- **AND** bulk runs MUST expose per-item result state so one failed item does not hide other item outcomes

### Requirement ASSISTANT-EXECUTION-008: External intake channels SHALL use the same governed capability model
Cabinet MUST route Telegram-originated photo, barcode, text, or mixed catalog intake through the same assistant capability registry, preview, confirmation, and audit model as in-app assistant actions.

#### Scenario: Govern Telegram catalog intake
- **GIVEN** a Telegram message contains item photos, barcode data, text notes, or a mixed capture session
- **WHEN** Cabinet maps the sender/chat to an authorized profile
- **THEN** the intake MUST create a draft preview through catalog_add_from_photo, catalog_add_from_barcode, or catalog_add_from_text capabilities before any catalog/inventory mutation
- **AND** ambiguous recognition MUST ask for follow-up input rather than inventing item data
- **AND** the audit trail MUST preserve source channel, sender/chat identity, media ids, proposed fields, confirmation decision, and applied item links
