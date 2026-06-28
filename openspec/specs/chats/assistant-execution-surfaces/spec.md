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

#### Scenario: Preview and apply actions use capability metadata
- **GIVEN** the capability registry exposes executable Inventory, Wishlist, Collections, and app-control capabilities
- **WHEN** Chat requests an action preview with a supported `capability_id`
- **THEN** Cabinet MUST resolve the capability to the canonical preview/apply handler declared by registry metadata
- **AND** legacy action aliases MUST normalize to the same canonical capability/action mapping instead of bypassing capability policy
- **AND** unsupported, read-only, preview-only, or unavailable capability ids MUST fail with deterministic setup or permission guidance
- **AND** confirmed apply MUST execute only the canonical handler for the previewed capability and preserve profile/thread confirmation boundaries

### Requirement ASSISTANT-EXECUTION-006: OpenAI-backed assistant capabilities SHALL expose truthful setup and readiness states
Cabinet MUST treat OpenAI/API-key/Browser Auth readiness as provider evidence that gates capabilities, not as copy or navigation state.

#### Scenario: Block false OpenAI readiness
- **GIVEN** an assistant capability requires OpenAI-backed processing
- **WHEN** the active profile lacks a verified API key, verified Browser Auth artifact, or passing provider test
- **THEN** the capability MUST report setup-needed or unavailable state
- **AND** the UI/API MUST NOT mark Browser Auth connected from outbound navigation alone
- **AND** the provider registry MUST NOT report OpenAI ready from an active auth-method value unless the selected method also has verified credential/proof evidence
- **AND** provider tests MUST return truthful readiness evidence instead of passing only because a credential-like value exists

#### Scenario: Record OpenAI provider-test evidence
- **GIVEN** OpenAI API-key mode is selected for the active profile
- **WHEN** Cabinet runs an OpenAI provider test
- **THEN** Cabinet MUST call an OpenAI-compatible connectivity endpoint with the stored secret without returning the secret to clients
- **AND** the response MUST include provider, profile, auth method, credential-present state, pass/fail status, timestamp, next action, and non-secret upstream failure evidence when the provider rejects the test
- **AND** Browser Auth MUST remain setup-needed for provider-test evidence until a verified runtime provider-test adapter proof exists
- **AND** Browser Auth provider-test readiness MUST require connected Browser Auth state, a verified auth artifact flag, a non-secret provider-test artifact id, and a passed provider-test state
- **AND** Browser Auth failed proof states MUST return failed provider-test evidence without marking the provider ready or exposing secret material
- **AND** the provider health and registry responses MUST only report Browser Auth ready when that passed proof is present

#### Scenario: Discover content and listing generation readiness
- **GIVEN** content_generate and listing_draft_generate require OpenAI-backed processing
- **WHEN** the assistant queries the governed capability registry before verified provider readiness exists
- **THEN** both capabilities MUST be discoverable with provider requirements, input schema, preview shape, audit behavior, and result destination
- **AND** content_generate MUST remain preview-only/no-mutation while setup is needed
- **AND** listing_draft_generate MUST require explicit confirmation before any listing draft mutation can be applied

#### Scenario: Discover image analysis and processing readiness
- **GIVEN** image_analyze and image_process require OpenAI-backed media processing
- **WHEN** the assistant queries the governed capability registry before verified provider and media-processing readiness exists
- **THEN** both capabilities MUST be discoverable with provider requirements, media access requirements, input schema, preview shape, audit behavior, and media result destination
- **AND** image_analyze MUST remain preview-only/no-mutation while setup is needed
- **AND** image_process MUST require explicit confirmation before any processed variant is linked or applied

### Requirement ASSISTANT-EXECUTION-007: Assistant workflow runs SHALL persist auditable execution records
Cabinet MUST persist OpenAI-backed and agent-backed work as durable workflow run records so previews, confirmations, provider traces, and outcomes are inspectable outside transient chat text.

#### Scenario: Persist assistant workflow run
- **GIVEN** an assistant capability creates, analyzes, enriches, reconciles, tests, or previews work
- **WHEN** the workflow is accepted for execution or queued for provider work
- **THEN** Cabinet MUST persist a run id, workflow id, capability id, profile id, source channel, status, progress timestamps, provider trace, result or error payload, and confirmation state
- **AND** bulk runs MUST expose per-item result state so one failed item does not hide other item outcomes

#### Scenario: Preserve original media through image workflow runs
- **GIVEN** an assistant image workflow analyzes or processes existing Cabinet media
- **WHEN** the workflow run completes with preview findings or a processed variant proposal
- **THEN** the result MUST link back to source media evidence
- **AND** image analysis MUST NOT create processed variants or mutate item/media records
- **AND** image processing results MUST preserve original media id, variant media id, provenance, provider trace, and pending confirmation state

### Requirement ASSISTANT-EXECUTION-008: External intake channels SHALL use the same governed capability model
Cabinet MUST route Telegram-originated photo, barcode, text, or mixed catalog intake through the same assistant capability registry, preview, confirmation, and audit model as in-app assistant actions.

#### Scenario: Govern Telegram catalog intake
- **GIVEN** a Telegram message contains item photos, barcode data, text notes, or a mixed capture session
- **WHEN** Cabinet maps the sender/chat to an authorized profile
- **THEN** the intake MUST create a draft preview through catalog_add_from_photo, catalog_add_from_barcode, or catalog_add_from_text capabilities before any catalog/inventory mutation
- **AND** the intake MUST persist a queryable workflow-run audit record that binds the Telegram source message/thread, selected catalog_add_from_* capability, preview id, confirmation state, and non-secret provider trace
- **AND** ambiguous recognition MUST ask for follow-up input rather than inventing item data
- **AND** the audit trail MUST preserve source channel, sender/chat identity, media ids, proposed fields, confirmation decision, and applied item links
- **AND** authorized Telegram/OpenAI external runtime proof MUST be recorded through an approved production-channel proof packet that verifies persisted sender/chat authorization, binds an existing source thread and preview, requires a catalog_add_from_* capability, rejects incomplete provider evidence, and stores only non-secret OpenAI request/result trace data

#### Scenario: Record authorized Telegram/OpenAI production proof packet
- **GIVEN** a Telegram catalog capture preview exists for the authorized sender/chat and active profile
- **WHEN** an approved external-intake proof packet is submitted for that source message/thread, preview id, catalog_add_from_* capability, and OpenAI provider result
- **THEN** Cabinet MUST reject the packet unless proof approval, provider=openai, live-provider evidence, request id, result id, and credential-returned=false are present
- **AND** Cabinet MUST persist a completed, queryable workflow-run record with source_channel=telegram, the source thread/message, selected capability, preview id, confirmation state, and non-secret provider trace

### Requirement ASSISTANT-EXECUTION-009: Cabinet Agent app-control tools SHALL be discoverable and governed
Cabinet Agent MUST expose safe app-control tools for main chat and side-panel Assistant UI through the same capability, preview, confirmation, and audit boundary as existing assistant actions.

#### Scenario: Discover route navigation tool
- **GIVEN** an active profile and assistant thread context
- **WHEN** the assistant queries the capability registry for app-control tools
- **THEN** `navigate.open_surface` MUST be returned as an available app-control capability
- **AND** the capability MUST restrict targets to known Cabinet surfaces, including Media
- **AND** the capability MUST be preview-only/read-only so route opening cannot mutate Cabinet records

#### Scenario: Confirm open item title mutation
- **GIVEN** a Cabinet Agent request targets the currently open inventory item
- **WHEN** the request proposes `update_open_item_title`
- **THEN** Cabinet MUST create a preview tied to the active profile and thread
- **AND** applying the preview MUST require explicit confirmation
- **AND** the mutation MUST only update the target item for the same profile
- **AND** the assistant audit message MUST include the action, item id, changed title, confirmation state, and mutation-applied evidence

### Requirement ASSISTANT-EXECUTION-010: Chat-guided walkthroughs SHALL expose typed guidance modes
Cabinet Agent MUST model Chat app-control guidance as explicit, typed modes: `explain`, `show_me`, `do_it_with_me`, and `do_it_for_me`. Each mode MUST declare whether it may navigate, highlight UI targets, collect user input, preview mutations, request confirmation, or apply a confirmed mutation.

#### Scenario: Select guided walkthrough mode
- **GIVEN** an active profile, assistant thread, current route context, and a user request for help completing a Cabinet workflow
- **WHEN** the assistant planner selects a guided walkthrough mode
- **THEN** the selected mode MUST be persisted on the workflow run
- **AND** the mode MUST determine which command bus operations are allowed for that step
- **AND** mutating modes MUST still require preview and explicit confirmation before save/apply
- **AND** `explain` and `show_me` modes MUST remain non-mutating

### Requirement ASSISTANT-EXECUTION-011: Guided workflow registry SHALL provide typed recipes
Cabinet MUST expose a deterministic guided workflow registry that maps user intents to typed recipe definitions before the assistant dispatches app-control steps.

#### Scenario: Match workflow recipe
- **GIVEN** the assistant receives a request to update an existing inventory item
- **WHEN** the guided workflow registry evaluates the request against known recipes
- **THEN** it MUST return a typed `inventory.item.update` recipe when required route, target item, editable field, and confirmation capabilities are available
- **AND** the recipe MUST declare required context, ordered steps, allowed guidance modes, UI targets, command bus operations, validation expectations, and result/audit destinations
- **AND** unknown or under-specified requests MUST return a follow-up prompt rather than inventing workflow steps

### Requirement ASSISTANT-EXECUTION-012: UI target registry SHALL provide stable selectors for guided steps
Cabinet MUST expose a UI target registry for guided workflows so route navigation, highlights, callouts, and Cypress validation bind to stable target ids and selectors instead of fragile copy or layout assumptions.

#### Scenario: Resolve walkthrough UI target
- **GIVEN** a guided recipe references an inventory title field target
- **WHEN** the shell command bus resolves that target for the current route
- **THEN** the registry MUST return a stable target id, route/surface ownership, selector or `data-testid`, accessible label expectation, highlight/callout placement metadata, and unavailable-state guidance
- **AND** command execution MUST fail safely when a target cannot be resolved
- **AND** missing targets MUST be testable as registry coverage gaps before a walkthrough is marked complete

### Requirement ASSISTANT-EXECUTION-013: Shell command bus SHALL govern Chat-driven navigation and highlighting
Cabinet MUST route Chat-driven navigation, target highlighting, callouts, user-prompt waits, preview creation, and confirmation requests through a shell command bus with deterministic status and audit events.

#### Scenario: Dispatch non-mutating walkthrough commands
- **GIVEN** a guided recipe step requests navigation to Inventory and highlighting of an editable field
- **WHEN** the shell command bus dispatches the step
- **THEN** it MUST emit command status events for queued, running, success, failure, and skipped states
- **AND** navigation and highlight commands MUST remain non-mutating
- **AND** failures MUST preserve current route context and provide retry or fallback guidance
- **AND** side-panel Chat MUST remain open while the main app route changes

### Requirement ASSISTANT-EXECUTION-014: Guided walkthrough Action Timeline SHALL persist step records
Cabinet MUST persist guided walkthrough Action Timeline records from assistant workflow runs so each route, target, prompt, preview, confirmation, apply, and result step is inspectable after the chat message scrollback changes.

#### Scenario: Record guided walkthrough steps
- **GIVEN** a guided inventory update walkthrough is running
- **WHEN** route navigation, target highlighting, preview creation, confirmation, apply, or failure events occur
- **THEN** Cabinet MUST append ordered Action Timeline records with workflow run id, recipe id, mode, command id, target id, status, timestamp, and non-secret result/error evidence
- **AND** records MUST be queryable from the assistant thread and compact side-panel Action Timeline
- **AND** failed or paused steps MUST remain visible with the next required user or system action

### Requirement ASSISTANT-EXECUTION-015: Inventory item update walkthrough SHALL be the first validated recipe
The first guided walkthrough implementation MUST validate `inventory.item.update` end to end before broader guided recipes are treated as complete.

#### Scenario: Guide inventory item update with confirmation boundary
- **GIVEN** an existing inventory item is open or can be selected from the Inventory surface
- **WHEN** the user asks Chat to help update an editable item field
- **THEN** Cabinet MUST match the `inventory.item.update` recipe, navigate or focus the Inventory item editor as needed, highlight the target field, collect the intended value, and create a preview before mutation
- **AND** Cabinet MUST pause before save/apply until explicit user confirmation is received
- **AND** confirmed apply MUST persist the field update for the active profile and record the result in the Action Timeline and assistant thread audit
- **AND** cancellation, missing target, stale profile/thread, and failed apply states MUST avoid mutation and leave retryable evidence
