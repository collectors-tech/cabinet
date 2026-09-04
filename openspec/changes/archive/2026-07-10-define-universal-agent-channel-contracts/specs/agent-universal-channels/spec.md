## ADDED Requirements

### Requirement: Cabinet Agent SHALL expose universal governed entry points
Cabinet SHALL make Agent available from the main `/chats` workspace, side-panel Chat, relevant table/detail screens, Inbox/action review flows, and approved external channels without changing the safety boundary for the requested work.

#### Scenario: Open Agent from supported surfaces
- **GIVEN** an authenticated user is on `/chats`, a side-panel Chat surface, a supported table/detail screen, or an Inbox/action review flow
- **WHEN** the user opens Cabinet Agent
- **THEN** Agent SHALL preserve the active profile, route, thread, selected entity, and source surface context needed to answer or guide the request
- **AND** missing route, selection, profile, provider, permission, or setup context SHALL be reported as setup-needed or clarification-required guidance instead of guessed
- **AND** side-panel Chat SHALL remain available while non-mutating navigation or highlighting changes the main app route

#### Scenario: Explain available work before execution
- **GIVEN** Agent is opened from a supported surface
- **WHEN** the user asks what Agent can do
- **THEN** Agent SHALL explain available skills and capabilities for the current context
- **AND** it SHALL distinguish read-only, preview-only, confirm-required, external-write, blocked, and unavailable work
- **AND** unavailable or unsupported work SHALL include the next safe user action rather than silent omission or hallucinated capability

### Requirement: Cabinet Agent SHALL route work through skills and governed execution boundaries
Cabinet Agent SHALL use the Agent Skill Registry, capability registry, guided workflow registry, shell command bus, preview/apply handlers, and Action Timeline/audit records as the governed execution model for supported work.

#### Scenario: Execute read-only skill
- **GIVEN** a user invokes a read-only Agent skill with sufficient profile, route, and selection context
- **WHEN** Agent dispatches the skill
- **THEN** Cabinet SHALL allow the skill to read or summarize Cabinet state without mutating records
- **AND** the result SHALL identify the skill id, source surface, and non-secret evidence used

#### Scenario: Preview mutating skill before apply
- **GIVEN** a user invokes an Agent skill that can create, update, delete, import, export, or call an external write path
- **WHEN** Agent prepares the work
- **THEN** Cabinet SHALL create a preview or confirmation-required workflow before applying any mutation
- **AND** apply SHALL require explicit confirmation from the authorized user/channel
- **AND** cancellation, stale context, missing permission, failed provider setup, and failed apply states SHALL leave retryable Action Timeline or audit evidence without mutating early

### Requirement: Agent Chat attachments SHALL be explicit, scoped, and auditable
Cabinet SHALL handle attachments consistently for main Chat, side-panel Chat, Telegram-originated media, and future approved external channels while requiring explicit user-selected or authorized channel-provided files.

#### Scenario: Attach local file to in-app Agent chat
- **GIVEN** an authenticated user has an Agent thread open in main Chat or side-panel Chat
- **WHEN** the user selects a local file through an attachment control
- **THEN** Cabinet SHALL persist the attachment only for the active profile, thread, and message context
- **AND** attachment metadata SHALL include filename, byte size, MIME type, provenance, source surface, and created timestamp
- **AND** a queued attachment removed before send SHALL NOT be included in the next message request or persisted message attachment list
- **AND** local files SHALL NOT be attached from arbitrary paths without explicit user selection

#### Scenario: Reject unsupported or unsafe attachment
- **GIVEN** a user or authorized channel submits an unsupported type, oversized file, unsafe filename/path, malformed metadata, or attachment that cannot be associated with an active profile/thread/message
- **WHEN** Cabinet validates the attachment
- **THEN** Cabinet SHALL reject the attachment before Agent uses it as context
- **AND** the user/channel SHALL receive deterministic validation guidance
- **AND** no unrelated profile, thread, message, workflow run, or external channel record SHALL receive the failed attachment

#### Scenario: Preserve Telegram-originated media provenance
- **GIVEN** an authorized Telegram sender/chat provides media for an Agent workflow
- **WHEN** Cabinet ingests the media
- **THEN** Cabinet SHALL preserve Telegram source channel, sender/chat/message identifiers, media group identifiers where present, file id or resolved file evidence, MIME type, filename when available, byte size when available, and caption/text context
- **AND** Agent SHALL treat the media as authorized external-channel context for the mapped profile rather than as a local user-selected file
- **AND** the resulting thread, message, attachment, preview, Inbox item, and workflow-run audit metadata SHALL be sufficient to recover the source context without reading secrets

### Requirement: External Agent channels SHALL require explicit setup and authorization
Cabinet SHALL require setup and sender/chat authorization before Telegram or another external channel can create Agent threads, messages, attachments, workflow runs, previews, or mutations.

#### Scenario: Show Telegram setup and connection state
- **GIVEN** a user opens Integrations, Settings, Agent skill details, or an Agent setup-needed response for Telegram
- **WHEN** Cabinet reports Telegram channel state
- **THEN** it SHALL derive connection status from persisted sender/chat authorization and non-secret runtime/provider proof
- **AND** it SHALL show missing setup requirements without exposing bot tokens, API keys, or secret material
- **AND** navigation, copy, or a selected provider option alone SHALL NOT mark Telegram production intake ready

#### Scenario: Reject unauthorized external message
- **GIVEN** Telegram or another external channel delivers a text or media message whose sender/chat is not mapped to an authorized Cabinet profile
- **WHEN** Cabinet evaluates the intake
- **THEN** Cabinet SHALL reject the message before creating Agent thread, message, attachment, Inbox, preview, workflow-run, or mutation records
- **AND** the rejection SHALL preserve non-secret operational evidence for logs and support

### Requirement: External Agent intake SHALL create Cabinet reviewable work before mutation
Telegram and future approved external channels SHALL enter Cabinet Agent through the same skill routing, preview, confirmation, and audit boundaries as in-app Agent work.

#### Scenario: Route external text or media into Agent skill workflow
- **GIVEN** an authorized Telegram sender/chat submits text, photo/media, or mixed context for a supported Agent skill
- **WHEN** Cabinet normalizes and routes the intake
- **THEN** Cabinet SHALL create or select a profile-scoped Agent thread/message for the source channel
- **AND** it SHALL select a supported skill/capability or return a follow-up prompt when the intent, setup, or item identity is ambiguous
- **AND** approved Telegram bot-adapter commands SHALL preserve non-secret update, sender, chat, message, skill, and parameter evidence when routing to Agent text endpoints
- **AND** mutating work SHALL create a preview or Inbox review handoff before apply
- **AND** Telegram-visible responses SHALL include safe reply copy, review links, or structured action descriptors without scraping human-readable text

#### Scenario: Confirm or cancel external preview
- **GIVEN** an authorized external channel created a pending Agent preview
- **WHEN** the same authorized sender/chat submits a confirm or cancel action
- **THEN** Cabinet SHALL apply or cancel only that preview for the owning profile/thread
- **AND** approved Telegram bot-adapter callbacks SHALL route Agent preview confirmations to the Agent text callback endpoint instead of catalog-capture callbacks
- **AND** confirmation from a different sender/chat, stale preview, missing permission, or failed apply SHALL be rejected with retryable audit evidence
- **AND** successful apply SHALL record the source channel, skill id, preview id, confirmation state, mutation result, and non-secret provider/runtime evidence

#### Scenario: Preserve Market Watch and Purchases external skill source context
- **GIVEN** an authorized Telegram or approved external channel invokes a Market Watch or Purchases Agent Skill
- **WHEN** Cabinet creates the skill preview or confirmed apply response
- **THEN** the response MUST preserve the external source surface, source channel, source thread id, and source message id
- **AND** Market Watch result handoff preview MUST remain non-mutating until confirmed
- **AND** Purchases order creation apply MUST preserve purchase provenance and confirmation evidence without claiming an external provider write

### Requirement: Universal Agent specification SHALL classify implementation state
Cabinet SHALL distinguish implemented, planned, blocked, and deferred Agent behavior so follow-up issues cannot treat this parent specification as a broad implementation claim.

#### Scenario: Report current state honestly
- **GIVEN** a developer, validator, or Agent skill details view inspects this universal Agent contract
- **WHEN** it maps requirements to implementation issues
- **THEN** existing Chat, assistant execution, Telegram catalog capture, and Agent Skill Registry behaviors MAY be marked implemented only when linked tests already prove them
- **AND** #1703, #1704, #1705, #1706, #1708, #1709, #1710, #1711, #1715, and #1716 SHALL remain planned validation or implementation work until their own issues provide code, tests, runtime evidence, and closure proof
- **AND** public marketplace, unauthorised external write access, arbitrary local file access, secret exposure, and external-channel bypass of preview/confirmation SHALL remain deferred or disallowed

#### Scenario: Maintain an acceptance evidence map
- **GIVEN** #1716 is the Agent acceptance-suite issue for in-app work, attachments, and Telegram intake
- **WHEN** Cabinet reports acceptance coverage for Agent work
- **THEN** Cabinet SHALL maintain a #1716 Agent acceptance evidence map that names the exact requirement IDs, scenario groups, test targets, and evidence status for each required acceptance area
- **AND** the map SHALL distinguish implemented fixture/proof-packet validation from live production-channel validation
- **AND** #1716 SHALL NOT be claimed complete until attachment success/failure coverage and a live Telegram-channel checklist or explicit live-channel blocker are linked from durable issue/PR evidence
