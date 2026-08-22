## Purpose
Define the dedicated `/chats` workspace as a real conversation surface distinct from Assistant and Inbox.

## Requirements
### Requirement CHATS-WORKSPACE-001: `/chats` SHALL present a real Cabinet conversation workspace rather than placeholder/template inbox content
Chats workspace MUST use Cabinet-specific purpose, copy, and thread semantics instead of generic stock inbox template content.

#### Scenario: Render chats workspace with domain-specific semantics
- **GIVEN** user opens `/chats`
- **WHEN** route renders
- **THEN** headings, empty-state copy, and thread list content MUST reflect Cabinet conversation workflows
- **AND** route MUST NOT render generic placeholder names/copy as if the feature were a stock template inbox

### Requirement CHATS-WORKSPACE-002: Chats send flow SHALL preserve active thread context after send
Sending a message in `/chats` MUST append to the active thread and keep the selected conversation context stable.

#### Scenario: Send message in selected thread
- **GIVEN** user selected a conversation on `/chats`
- **WHEN** user sends a message
- **THEN** outgoing message MUST appear in the active thread
- **AND** route MUST remain on the selected conversation state instead of collapsing to a generic landing/empty state

### Requirement CHATS-WORKSPACE-003: Cabinet SHALL define the boundary between Assistant and `/chats`
Cabinet MUST define the distinct responsibilities of shell Assistant workspace and `/chats` so they are related but non-duplicative.

#### Scenario: Compare Assistant and `/chats` responsibilities
- **GIVEN** shell Assistant workspace and `/chats` route both exist
- **WHEN** product contracts are inspected
- **THEN** Assistant MUST be documented as the AI helper workspace
- **AND** `/chats` MUST be documented as the intentional conversation/thread workspace
- **AND** overlap, if any, MUST be explicitly justified by product behavior

### Requirement CHATS-WORKSPACE-004: `/chats` SHALL preserve the original/example two-pane conversation layout
Chats workspace MUST render as a dark, sparse two-pane conversation surface with a compact conversation rail on the left and the active or empty conversation workspace on the right.

#### Scenario: Render original/example chats layout parity
- **GIVEN** user opens `/chats`
- **WHEN** no conversation is selected
- **THEN** the left rail MUST include a search input before the conversation list
- **AND** conversation rows MUST preserve compact avatar, participant/title, and message preview structure
- **AND** the right workspace MUST show a centered empty-state composition with icon, title, helper text, and primary action affordance
- **AND** the rail and workspace MUST remain separate, readable, and unclipped at normal desktop widths

### Requirement CHATS-WORKSPACE-005: `/chats` SHALL support thread discovery and new-thread entry from the two-pane rail
Chats workspace MUST let users start a new durable thread from both the empty workspace and selected-thread topbar while keeping thread search scoped to discoverable conversation rows.

#### Scenario: Create and filter conversation threads
- **GIVEN** user opens `/chats`
- **WHEN** user creates multiple chat threads and filters the conversation rail
- **THEN** matching thread rows MUST remain visible
- **AND** non-matching thread rows MUST be hidden
- **AND** New Chat / Start a conversation affordances MUST open a labeled new-thread dialog at desktop and compact widths without changing the active route
- **AND** a successfully created thread MUST close the dialog and become the active conversation without requiring access to the conversation rail
- **AND** created threads MUST persist through the Cabinet chat thread API for the active profile

### Requirement CHATS-WORKSPACE-006: `/chats` SHALL closely follow the assistant-ui example chat surface
Chats workspace MUST render close to the provided assistant-ui examples: a dark, compact thread rail, a broad uninterrupted conversation canvas, and a bottom-docked centered composer instead of stacked framed panels.

#### Scenario: Render assistant-ui example visual contract
- **GIVEN** user opens `/chats` and selects a thread
- **WHEN** the selected conversation workspace renders
- **THEN** the layout MUST identify the assistant-ui example visual contract
- **AND** the conversation rail MUST stay compact beside the main surface
- **AND** the message canvas MUST occupy the dominant vertical workspace
- **AND** the composer MUST be docked at the bottom center of the conversation surface
- **AND** the composer MUST NOT be followed by separate custom Attachments or Action Preview panels
- **AND** attachment affordances MUST live with the composer while action previews, when present, MUST appear in the conversation canvas instead of as stacked controls under the composer

### Requirement CHATS-WORKSPACE-008: `/chats` SHALL dispatch normal user text through governed app-control planning
Chats workspace MUST send normal conversation text with profile, thread, route, and assistant context so deterministic app-control requests can return governed route or action results instead of defaulting to Inbox handoff.

#### Scenario: Dispatch route-opening chat text without creating Inbox noise
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `open media`
- **THEN** the chat message API response MUST include a `navigate.open_surface` app-control result for `/media`
- **AND** the route-opening action MUST create durable workflow-run audit evidence for the selected thread
- **AND** the response MUST NOT create a default assistant Inbox handoff for the handled app-control request

### Requirement CHATS-WORKSPACE-009: `/chats` SHALL preview normal-text mutations before apply
Chats workspace MUST let deterministic app-control planning convert normal user text into preview-required mutation actions without mutating records before explicit confirmation.

#### Scenario: Dispatch item-create chat text as a pending preview
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `create an inventory item ...`
- **THEN** the chat message API response MUST include an `inventory.item.create` app-control result with a `create_inventory_item` preview
- **AND** the preview MUST include the parsed item payload and pending confirmation workflow-run audit evidence
- **AND** inventory MUST NOT include the item before the user explicitly applies the preview
- **AND** the response MUST NOT create a default assistant Inbox handoff for the handled app-control request

### Requirement CHATS-WORKSPACE-010: `/chats` SHALL render governed app-control route and setup-needed results
Chats workspace MUST expose deterministic app-control outcomes from normal chat messages as visible, user-activated UI instead of relying only on API metadata.

#### Scenario: Activate a route-opening result and expose provider setup-needed guidance
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `open media`
- **THEN** the workspace MUST render a route action card for `/media`
- **AND** Cabinet MUST stay on `/chats` until the user activates the card
- **AND** activating the card MUST navigate to `/media`
- **WHEN** user sends a provider-backed request that cannot run without provider readiness
- **THEN** the workspace MUST show visible setup-needed guidance instead of pretending success

### Requirement CHATS-WORKSPACE-011: Chat messages SHALL answer normal text through the selected assistant provider without default Inbox handoff noise
Chats workspace and shell Assistant messages MUST dispatch ordinary non-action text through the selected usable assistant provider, persist provider/model provenance, and MUST reserve durable Inbox handoffs for background, review-required, queued, or failure work. A selected provider failure MUST remain an explicit provider failure or setup state rather than a deterministic success response.

#### Scenario: Respond to greeting without durable Inbox item
- **GIVEN** user sends `hello` with profile, thread, route, and assistant context
- **WHEN** `/api/chat/messages` accepts the message
- **THEN** the response MUST include a direct `assistant_response` thread message
- **AND** the assistant thread MUST show the direct response instead of `Assistant handoff queued in Inbox.`
- **AND** `/api/chat/inbox` MUST NOT receive an `assistant_handoff` item for that normal message

#### Scenario: Use the selected provider for an ordinary conversation turn
- **GIVEN** the active profile selects a usable assistant provider and model
- **WHEN** the user sends ordinary non-action text through `/api/chat/messages`
- **THEN** Cabinet MUST send bounded profile/thread conversation history to that provider
- **AND** the provider MUST receive no Cabinet tool, database, filesystem, browser, plugin, or mutation authority
- **AND** Cabinet MUST persist the provider response with exact provider/model provenance
- **AND** Cabinet MUST NOT replace the provider response with `deterministic_chat_fallback`

#### Scenario: Fail closed when the selected provider cannot reply
- **GIVEN** the active profile selects a provider whose adapter, authentication, model, or transport is unavailable
- **WHEN** the user sends ordinary non-action text
- **THEN** Cabinet MUST persist a normalized setup-required or retryable failure response
- **AND** Cabinet MUST NOT present the deterministic fallback copy as a successful provider reply

#### Scenario: Preserve explicit handoff durability
- **GIVEN** user sends a message that asks Cabinet to follow up, queue, review, notify, or run background work
- **WHEN** `/api/chat/messages` accepts the message
- **THEN** Cabinet MAY create an `assistant_handoff` Inbox item with thread metadata
- **AND** Inbox triage status lifecycle MUST remain durable for that queued handoff

### Requirement CHATS-WORKSPACE-012: Full Chat Agent controls SHALL avoid compact-layout overflow traps
Full Chat MUST keep its message result, composer, attachment, Retry, Apply, and Cancel controls keyboard reachable through intended scrolling at compact desktop and 200-percent-zoom-equivalent layouts.

#### Scenario: Operate a governed result at 200-percent-zoom equivalent
- **GIVEN** full Chat is rendered at a 640 by 360 viewport with a selected thread
- **WHEN** a governed Agent result exposes retry or preview actions
- **THEN** the result and composer regions MUST use bounded independent scrolling without fixed minimum-height or fixed-width overflow
- **AND** attachment, Retry, Apply, and Cancel controls MUST remain visible and keyboard focusable when applicable
- **AND** focus MUST remain on the initiating action after a terminal dialog is canceled

### Requirement CHATS-WORKSPACE-013: Public Chat ingress SHALL reject client-authored trusted Agent evidence
The public `/api/chat/messages` endpoint MUST accept only user-authored messages and MUST fail closed when request context carries trusted Agent result, preview, execution, authority, capability, or success evidence. Server-side planner, provider, dispatcher, preview, apply, cancel, and audit paths MAY still persist assistant/system evidence through internal service calls.

#### Scenario: Reject forged public Chat evidence
- **GIVEN** a client calls `POST /api/chat/messages`
- **WHEN** the request uses `role=assistant` or `role=system`
- **THEN** Cabinet MUST reject the request before storing a chat message
- **WHEN** a `role=user` request carries `agent_planner`, `agent_capabilities`, preview, execution, authority, assistant response, assistant handoff, admin-session, or success evidence in public context fields
- **THEN** Cabinet MUST reject the request before storing a chat message
- **AND** a normal user message MAY still trigger server-owned planner/provider/dispatcher persistence for trusted assistant evidence

### Requirement CHATS-WORKSPACE-014: Literal response instructions SHALL remain ordinary provider Chat
Cabinet MUST distinguish literal response instructions from governed Cabinet operation requests before dispatching a user message. A requested literal value MAY contain words that otherwise identify an Agent operation, but those words MUST NOT grant or invoke Cabinet tool authority.

#### Scenario: Return literal text containing an operation word
- **GIVEN** ordinary Chat has a selected assistant provider
- **WHEN** the user asks the provider to reply, respond, say, return, or echo an exact non-empty value containing an Agent action word such as `restore`, `delete`, `import`, or `upload`
- **THEN** Cabinet MUST dispatch the message through the ordinary selected-provider conversation path
- **AND** Cabinet MUST NOT create an Agent planner result or governed action preview

#### Scenario: Preserve genuine governed operations
- **GIVEN** the user intends Cabinet to inspect or change managed product data
- **WHEN** the user directly asks Cabinet to restore, delete, import, export, upload, or otherwise manage Cabinet data
- **THEN** Cabinet MUST continue to classify the request for server-owned Agent planning and governed execution

### Requirement CHATS-WORKSPACE-015: Read-only Agent results SHALL surface bounded server-owned facts

When Cabinet executes a read-only Agent skill from Chat, the resulting assistant message SHALL include a presentation-neutral, server-owned summary of the matching records instead of merely repeating the provider's plan text.

#### Scenario: Inventory lookup returns exact safe record facts

- **GIVEN** the active profile contains an inventory item with a known part number and title
- **WHEN** the user asks Chat to find that item and report its exact title
- **THEN** Cabinet MUST execute the profile-scoped read skill and render the matching part number and title in both full and contextual Chat
- **AND** the summary MUST survive reload and thread handoff
- **AND** the summary MUST contain no write action or confirmation control
- **AND** Cabinet MUST cap the number and length of rendered fields and exclude notes, URLs, credentials, secrets, and raw provider or execution blobs

#### Scenario: Chat action timeline returns typed safe workflow facts

- **GIVEN** the active profile has governed Agent workflow runs in a Chat thread
- **WHEN** the user asks Chat to show that thread's governed action timeline
- **THEN** Cabinet MUST execute the thread-scoped read-only Chat timeline skill and persist a `chat_action_timeline` result summary with bounded workflow run identifiers, capability identifiers, execution status, and operation labels
- **AND** the summary MUST exclude raw input prompts, provider traces, bulk item payloads, source message identifiers, authority details, mutation evidence, timestamps, preview tokens, provider secrets, stack traces, execution internals, and cross-thread timeline entries
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Dashboard summary returns typed safe activity facts

- **GIVEN** the active profile has Dashboard totals, attention signals, and recent item identifiers
- **WHEN** the user asks Chat to summarise current Dashboard activity
- **THEN** Cabinet MUST execute the profile-scoped read-only Dashboard skill and persist a `dashboard_activity` result summary with bounded attention metrics and recent record labels
- **AND** the summary MUST distinguish unavailable Dashboard dependencies from empty or no-attention states
- **AND** the summary MUST contain no write action or confirmation control
- **AND** Cabinet MUST exclude cross-profile records, provider secrets, raw provider payloads, private source URLs, seller details, stack traces, and execution internals

#### Scenario: Storage status returns typed safe operational facts

- **GIVEN** the active profile can ask for storage and backup readiness from Chat
- **WHEN** the user asks Chat to show current storage or backup status
- **THEN** Cabinet MUST execute the profile-scoped read-only Storage skill and persist a `storage_status` result summary with bounded storage and backup state facts
- **AND** the summary MUST contain no write action, confirmation control, backup target path, provider secret, raw provider payload, stack trace, or execution internals

#### Scenario: Wishlist search returns typed safe entry facts

- **GIVEN** the active profile contains wishlist entries with planning and purchase metadata
- **WHEN** the user asks Chat to find matching wishlist entries
- **THEN** Cabinet MUST execute the profile-scoped read-only Wishlist skill and persist a `wishlist_entries` result summary with bounded entry identifiers, priority, and purchase-state facts
- **AND** the summary MUST exclude notes, purchase URLs, provider secrets, raw provider payloads, stack traces, execution internals, and cross-profile wishlist records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Collections search returns typed safe collection facts

- **GIVEN** the active profile contains workspace collections and assigned collection item membership
- **WHEN** the user asks Chat to find matching collections
- **THEN** Cabinet MUST execute the profile-scoped read-only Collections skill and persist a `collections` result summary with bounded collection names, assigned item identifiers, assigned item titles, and collection membership names
- **AND** the summary MUST exclude workspace item detail text, provider secrets, raw provider payloads, stack traces, execution internals, and cross-profile collection workspace records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Integration provider search returns typed safe provider facts

- **GIVEN** the active profile can ask Chat to find configured or available integration providers
- **WHEN** the user asks Chat to search integration providers
- **THEN** Cabinet MUST execute the read-only Integrations provider search skill and persist an `integration_providers` result summary with bounded provider identifiers, availability status, and setup-required state
- **AND** the summary MUST exclude provider secrets, raw provider payloads, execution internals, write-claim evidence, preview tokens, stack traces, and configuration payloads
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Data export preparation returns typed safe readiness facts

- **GIVEN** the active profile can ask Chat to prepare a non-mutating Cabinet data export bundle
- **WHEN** the user asks Chat to export Cabinet data
- **THEN** Cabinet MUST execute the profile-scoped export preparation path and persist a `data_export_bundle` result summary with bounded export scope and readiness status facts
- **AND** the summary MUST exclude provider secrets, raw provider payloads, backup paths, export artifact paths, preview tokens, stack traces, and execution internals
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Maintenance safe check returns typed safe operational facts

- **GIVEN** the active profile can ask Chat to run a read-only Cabinet maintenance safe check
- **WHEN** the user asks Chat to check Cabinet maintenance or storage health
- **THEN** Cabinet MUST execute the profile-scoped read-only Maintenance skill and persist a `maintenance_safe_check` result summary with bounded check name, check level, and health status facts
- **AND** the summary MUST exclude provider secrets, raw provider payloads, backup paths, export artifact paths, preview tokens, stack traces, and execution internals
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Inbox notification search returns typed safe notification facts

- **GIVEN** the active profile contains Inbox notification records from one or more Cabinet sources
- **WHEN** the user asks Chat to find or summarise matching Inbox notifications
- **THEN** Cabinet MUST execute the profile-scoped read-only Inbox skill and persist an `inbox_notifications` result summary with bounded notification identifiers, titles, statuses, and source labels
- **AND** the summary MUST exclude notification body/summary text, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile Inbox records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Inbox unhandled summary returns typed safe triage facts

- **GIVEN** the active profile contains unread, queued, handled, and cross-profile Inbox notification records
- **WHEN** the user asks Chat to summarise unhandled Inbox notifications
- **THEN** Cabinet MUST execute the profile-scoped read-only Inbox unhandled summary skill and persist an `inbox_unhandled` result summary with bounded unhandled notification identifiers, titles, statuses, and source labels
- **AND** the summary MUST include only unread or queued active-profile Inbox records
- **AND** the summary MUST exclude handled Inbox records, notification body/summary text, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile Inbox records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Workspace user search returns typed safe authority facts

- **GIVEN** an authorized active profile can ask Chat to inspect workspace users
- **WHEN** the user asks Chat to find matching workspace users
- **THEN** Cabinet MUST execute the profile-scoped read-only Users skill and persist a `workspace_users` result summary with bounded user identifiers, display labels, statuses, and role labels
- **AND** the summary MUST exclude email addresses, phone numbers, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile workspace users
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Media search returns typed safe asset facts

- **GIVEN** the active profile contains workspace media assets from Chat attachments or inventory photos
- **WHEN** the user asks Chat to find matching media assets
- **THEN** Cabinet MUST execute the profile-scoped read-only Media skill and persist a `media_assets` result summary with bounded media identifiers, display titles, linkage states, and source labels
- **AND** the summary MUST exclude stored paths, private source URLs, notes, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile media records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Discovery search returns typed safe result facts

- **GIVEN** the active profile contains provider discovery results requiring review
- **WHEN** the user asks Chat to find matching discovery results for a provider
- **THEN** Cabinet MUST execute the profile-scoped read-only Discoveries skill and persist a `discovery_results` result summary with bounded candidate identifiers, display titles, triage status, provider label, and review-needed state
- **AND** the summary MUST exclude listing URLs, source result URLs, seller labels, reviewer notes, extracted part numbers, prices, shipping, confidence scores, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile discovery records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Purchase order search returns typed safe order facts

- **GIVEN** the active profile contains purchase order lifecycle records with line-item state
- **WHEN** the user asks Chat to find matching purchase orders
- **THEN** Cabinet MUST execute the profile-scoped read-only Purchases skill and persist a `purchase_orders` result summary with bounded order identifiers, order status, source label, and line-item count facts
- **AND** the summary MUST exclude sellers, tracking codes, private notes, amounts, arrival identifiers, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile purchase records
- **AND** the summary MUST contain no write action or confirmation control

#### Scenario: Market Watch saved-watch search returns typed safe watch facts

- **GIVEN** the active profile contains Market Watch saved-watch definitions
- **WHEN** the user asks Chat to find matching saved watches for a provider
- **THEN** Cabinet MUST execute the profile-scoped read-only Market Watch search skill and persist a `market_watch_watches` result summary with bounded saved-watch identifiers, display names, enabled state, provider-scope count, and last-result count facts
- **AND** the summary MUST exclude search keywords, exclusions, provider health internals, listing URLs, provider secrets, raw provider payloads, preview tokens, stack traces, execution internals, mutation evidence, and cross-profile saved watches
- **AND** the summary MUST contain no write action or confirmation control
