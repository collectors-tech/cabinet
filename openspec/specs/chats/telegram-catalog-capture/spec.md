## Purpose
Define Telegram as an authorized external intake channel for Cabinet catalog draft creation.

## Requirements
### Requirement TELEGRAM-CATALOG-CAPTURE-001: Telegram catalog capture SHALL require authorized sender/profile mapping
Cabinet MUST map each Telegram sender/chat to an explicitly authorized Cabinet profile before creating threads, attachments, previews, inbox items, or mutating catalog data.

#### Scenario: Reject unauthorized Telegram sender
- **GIVEN** a Telegram message is received from a sender/chat with no Cabinet profile authorization
- **WHEN** Cabinet evaluates the catalog capture request
- **THEN** Cabinet MUST reject the request before creating any chat thread, attachment, preview, inbox item, or catalog record

#### Scenario: Accept authorized Telegram API capture
- **GIVEN** a Cabinet profile has persisted Telegram catalog capture sender/chat authorization settings
- **WHEN** the Telegram catalog capture API receives matching sender/chat input for that profile
- **THEN** Cabinet MUST accept the request and create the capture through the governed preview-before-apply path
- **AND** a sender/chat mismatch MUST return an authorization failure without creating capture records

### Requirement TELEGRAM-CATALOG-CAPTURE-002: Telegram catalog capture SHALL preserve source and media metadata
Cabinet MUST persist Telegram source identifiers and media metadata with the assistant thread/message context so the capture remains auditable from Inbox and Assistant surfaces.

#### Scenario: Store Telegram photo and text context
- **GIVEN** an authorized Telegram sender sends item text, barcode data, and one or more media attachments
- **WHEN** Cabinet creates the assistant capture thread
- **THEN** the thread/message context MUST preserve source channel, sender/chat/message identifiers, barcode/grouping metadata, and attachment metadata including filename, MIME type, and Telegram file id where available
- **AND** the Inbox audit metadata MUST preserve the Telegram media source fields and source metadata needed to recover the capture without reading the assistant message payload

### Requirement TELEGRAM-CATALOG-CAPTURE-003: Telegram catalog capture SHALL create drafts before mutation
Telegram intake MUST produce a preview/inbox handoff first; it MUST NOT create or update catalog inventory records until an explicit confirmation is applied.

#### Scenario: Create preview without applying item
- **GIVEN** an authorized Telegram capture has enough text/barcode/media context to form a catalog draft
- **WHEN** Cabinet ingests the capture
- **THEN** Cabinet MUST create a previewed inventory action and Inbox review item
- **AND** no catalog item MUST be created until explicit confirmation is submitted
### Requirement TELEGRAM-CATALOG-CAPTURE-004: Barcode-only Telegram capture SHALL use a manual draft path when lookup is unavailable
When Telegram intake receives a barcode but no resolved product lookup, Cabinet MUST create a transparent manual draft path instead of inventing product attributes.

#### Scenario: Barcode-only manual draft
- **GIVEN** an authorized Telegram sender submits a barcode without a resolved lookup payload
- **WHEN** Cabinet creates the preview
- **THEN** the preview MUST use the barcode as the part number and a clear barcode-derived title
- **AND** any missing product facts MUST remain unset for user review

### Requirement TELEGRAM-CATALOG-CAPTURE-005: Telegram catalog capture SHALL return a confirmation handoff
Cabinet MUST return Telegram-facing confirmation copy and a Cabinet review link for each accepted capture so channel adapters can tell the user how to confirm or cancel the draft without applying it automatically.

#### Scenario: Return Telegram review instructions
- **GIVEN** an authorized Telegram capture creates a catalog draft preview
- **WHEN** Cabinet returns the capture result to the Telegram channel adapter
- **THEN** the response MUST include user-facing reply text, a Cabinet review URL, confirmation-required state, and available review actions
- **AND** the Inbox handoff metadata MUST preserve the same review URL and Telegram reply controls for audit/recovery

### Requirement TELEGRAM-CATALOG-CAPTURE-006: Telegram webhook payloads SHALL normalize into governed capture inputs
Cabinet MUST normalize Telegram webhook updates into the same catalog capture input used by the governed intake service before authorization, preview creation, or mutation handling.

#### Scenario: Normalize mixed Telegram photo caption capture
- **GIVEN** Telegram sends a webhook update containing a sender, chat, message id, media group id, photo sizes, and a caption with barcode-like text
- **WHEN** Cabinet normalizes the webhook update for catalog intake
- **THEN** Cabinet MUST preserve sender/chat/message identifiers, media grouping, update metadata, caption text, inferred barcode, and the largest available photo as capture media

#### Scenario: Normalize text-only barcode capture
- **GIVEN** Telegram sends a webhook update containing text with a barcode-like number and no media
- **WHEN** Cabinet normalizes the webhook update for catalog intake
- **THEN** Cabinet MUST preserve the text and inferred barcode without creating synthetic media attachments

### Requirement TELEGRAM-CATALOG-CAPTURE-007: Telegram webhook catalog capture API SHALL resolve profile authorization before preview creation
Cabinet MUST accept raw Telegram webhook updates through a dedicated catalog capture endpoint, resolve the sender/chat against persisted profile authorization settings, and then create the capture through the governed preview-before-apply service.

#### Scenario: Accept authorized Telegram webhook capture
- **GIVEN** a Cabinet profile has persisted Telegram catalog capture sender/chat authorization settings
- **WHEN** the webhook endpoint receives a matching Telegram update with text, barcode-like content, and media metadata
- **THEN** Cabinet MUST resolve the profile from the sender/chat, normalize the update, and create the preview/inbox handoff without requiring a profile id in the webhook payload
- **AND** no catalog item MUST be created until explicit confirmation is submitted

#### Scenario: Reject unauthorized Telegram webhook capture
- **GIVEN** no Cabinet profile has matching Telegram sender/chat authorization settings
- **WHEN** the webhook endpoint receives a Telegram update from that sender/chat
- **THEN** Cabinet MUST reject the request before creating capture records

### Requirement TELEGRAM-CATALOG-CAPTURE-008: Ambiguous Telegram captures SHALL request follow-up instead of inventing draft fields
When authorized Telegram text or media does not contain enough item identity to create a safe catalog preview, Cabinet MUST return a Telegram-visible follow-up prompt and MUST NOT create preview, Inbox, attachment, or catalog records from invented details.

#### Scenario: Request follow-up for ambiguous text-only capture
- **GIVEN** a Cabinet profile has persisted Telegram catalog capture sender/chat authorization settings
- **WHEN** the webhook endpoint receives an authorized text-only update without a barcode, part number, or resolved draft title
- **THEN** Cabinet MUST return a follow-up-required response with Telegram-facing reply copy and missing identity fields
- **AND** no catalog item MUST be created from the ambiguous text

### Requirement TELEGRAM-CATALOG-CAPTURE-009: Telegram replies SHALL expose structured channel action buttons
Telegram capture responses MUST include structured channel action button descriptors so Telegram adapters can render URL, callback, and reply controls without scraping human-readable copy.

#### Scenario: Return structured review action buttons
- **GIVEN** an authorized Telegram capture creates a catalog draft preview
- **WHEN** Cabinet returns the capture result to the Telegram channel adapter
- **THEN** the Telegram reply MUST include a URL button for Cabinet review plus callback buttons for confirm and cancel actions that reference the preview id
- **AND** the Inbox handoff metadata MUST preserve those action button descriptors for audit/recovery

#### Scenario: Return structured follow-up reply buttons
- **GIVEN** an authorized Telegram capture is too ambiguous for a safe preview
- **WHEN** Cabinet returns a follow-up-required response
- **THEN** the Telegram reply MUST include structured reply button descriptors for barcode, part number, and item title follow-up

### Requirement TELEGRAM-CATALOG-CAPTURE-010: Telegram callback actions SHALL apply or cancel only authorized previews
Cabinet MUST accept Telegram catalog capture callback actions only from the authorized sender/chat for the profile that owns the preview, and MUST translate confirm/cancel callback data into the same governed preview apply/cancel lifecycle used by Cabinet review.

#### Scenario: Confirm authorized Telegram capture callback
- **GIVEN** an authorized Telegram capture has created a pending catalog item preview
- **WHEN** the Telegram channel adapter submits the confirm callback data from the same authorized sender/chat
- **THEN** Cabinet MUST apply the preview, create the catalog item, and return Telegram-visible confirmed reply state

#### Scenario: Reject unauthorized Telegram capture callback
- **GIVEN** a pending Telegram catalog capture preview exists for one authorized sender/chat
- **WHEN** a different sender or chat submits callback data for that preview
- **THEN** Cabinet MUST reject the callback before applying or cancelling the preview

#### Scenario: Cancel authorized Telegram capture callback
- **GIVEN** an authorized Telegram capture has created a pending catalog item preview
- **WHEN** the Telegram channel adapter submits the cancel callback data from the same authorized sender/chat
- **THEN** Cabinet MUST mark the preview cancelled and MUST NOT create a catalog item

### Requirement TELEGRAM-CATALOG-CAPTURE-011: Telegram bot adapter SHALL route updates and render Bot API payloads deterministically
Cabinet MUST provide a Telegram bot adapter contract that maps Telegram message and callback updates to the correct Cabinet catalog capture APIs and renders structured Cabinet replies into Telegram Bot API payloads without scraping human-readable copy.

#### Scenario: Route Telegram bot updates to Cabinet catalog capture APIs
- **GIVEN** Telegram delivers either a message update or a callback query update to the bot adapter
- **WHEN** the adapter prepares the Cabinet handoff
- **THEN** message updates MUST route to the raw webhook catalog capture endpoint with original update context
- **AND** callback query updates MUST route to the catalog capture callback endpoint with sender id, chat id, and callback data preserved

#### Scenario: Render Telegram review and follow-up controls
- **GIVEN** Cabinet returns a structured Telegram reply for a preview, follow-up, confirmation, or cancellation state
- **WHEN** the adapter prepares a Telegram Bot API response
- **THEN** URL and callback action buttons MUST render as inline keyboard controls
- **AND** follow-up reply actions MUST render as a one-time reply keyboard
- **AND** callback result replies MUST be renderable as an edit to the original Telegram message

#### Scenario: Acknowledge Telegram callback queries
- **GIVEN** Telegram delivers a callback query for a Cabinet catalog capture action
- **WHEN** Cabinet returns a confirmation or cancellation Telegram reply
- **THEN** the adapter MUST render an answerCallbackQuery payload using the callback query id and user-visible result text
- **AND** the acknowledgement MUST be non-alert by default so the callback interaction clears without interrupting the chat

### Requirement TELEGRAM-CATALOG-CAPTURE-012: Telegram bot runtime wiring SHALL dispatch updates through Cabinet and bind Bot API requests at the edge
Cabinet MUST provide runtime-safe wiring that dispatches Telegram updates through the governed Cabinet catalog capture APIs, then converts the returned structured Telegram reply into Bot API requests using a caller-provided bot token.

#### Scenario: Dispatch message update and send Cabinet reply
- **GIVEN** Telegram delivers a message update to the bot runtime wiring
- **WHEN** Cabinet returns a structured capture or follow-up Telegram reply
- **THEN** the runtime wiring MUST route the update through the webhook catalog capture API
- **AND** it MUST render a sendMessage Bot API call to the originating Telegram chat

#### Scenario: Dispatch callback update and update Telegram message
- **GIVEN** Telegram delivers a callback query update to the bot runtime wiring
- **WHEN** Cabinet returns a structured confirmation or cancellation Telegram reply
- **THEN** the runtime wiring MUST route the update through the catalog capture callback API
- **AND** it MUST render both answerCallbackQuery and editMessageText Bot API calls for the originating callback/message

#### Scenario: Bind Bot API request with runtime token
- **GIVEN** the runtime wiring has a Telegram Bot API call and a bot token from runtime configuration
- **WHEN** it prepares the outbound Telegram request
- **THEN** it MUST POST JSON to the Telegram Bot API method URL for that token
- **AND** it MUST require the token at request construction time rather than embedding it in persisted specs or fixtures

#### Scenario: Execute rendered Bot API calls and report outbound failures
- **GIVEN** the runtime wiring has dispatched a Telegram update through Cabinet and rendered one or more Bot API calls
- **WHEN** it executes those calls with the caller-provided Telegram endpoint, token, and HTTP client
- **THEN** it MUST POST each rendered Bot API call in order
- **AND** any outbound Telegram Bot API failure MUST be returned with method/status/body evidence while preserving the Cabinet dispatch path and rendered calls for handoff logging

### Requirement TELEGRAM-CATALOG-CAPTURE-020: Telegram bot adapter SHALL return user-visible failure feedback from Cabinet dispatch errors
Cabinet Telegram bot dispatch MUST convert structured Cabinet capture/callback failures into Telegram-visible Bot API replies, and MUST provide a safe fallback reply when Cabinet cannot return structured copy.

#### Scenario: Render structured Cabinet failure reply for message capture
- **GIVEN** Telegram delivers a message update to the bot adapter
- **WHEN** Cabinet rejects the capture with a structured Telegram reply such as follow-up-required or authorization failure
- **THEN** the adapter MUST preserve the Cabinet dispatch path and error for logging
- **AND** it MUST render the structured Telegram reply as a sendMessage payload to the originating chat

#### Scenario: Render fallback callback failure feedback
- **GIVEN** Telegram delivers a callback query for a Cabinet catalog capture action
- **WHEN** Cabinet callback dispatch fails without a structured Telegram reply
- **THEN** the adapter MUST preserve the Cabinet dispatch path and error for logging
- **AND** it MUST render answerCallbackQuery and editMessageText payloads with safe user-visible failure copy for the originating callback/message

### Requirement TELEGRAM-CATALOG-CAPTURE-021: Telegram bot adapter SHALL cap rendered reply text to Bot API limits
Cabinet Telegram bot dispatch MUST cap rendered message, edit, and callback acknowledgement text before constructing Telegram Bot API payloads so long Cabinet reply copy does not create avoidable outbound Bot API failures.

#### Scenario: Cap rendered Telegram reply payload text
- **GIVEN** Cabinet returns a structured Telegram reply whose text exceeds a Telegram Bot API payload limit
- **WHEN** the adapter renders sendMessage, editMessageText, or answerCallbackQuery payloads
- **THEN** message and edit payload text MUST fit within the Telegram message text limit
- **AND** callback acknowledgement text MUST fit within the Telegram callback answer text limit
- **AND** truncated text MUST remain non-empty and visibly indicate truncation

### Requirement TELEGRAM-CATALOG-CAPTURE-022: Telegram bot adapter SHALL keep rendered button payloads within Bot API limits
Cabinet Telegram bot dispatch MUST cap rendered button labels and avoid invalid callback payloads before constructing Telegram Bot API reply markup so long Cabinet action descriptors do not create avoidable outbound Bot API failures.

#### Scenario: Cap rendered Telegram reply markup buttons
- **GIVEN** Cabinet returns structured Telegram reply buttons with labels longer than the Bot API button text limit
- **WHEN** the adapter renders inline or reply keyboard markup
- **THEN** each rendered button label MUST fit within the Telegram button text limit
- **AND** truncated labels MUST remain non-empty and visibly indicate truncation
- **AND** callback buttons whose callback data exceeds the Telegram callback data limit MUST be omitted while valid URL, reply, and callback alternatives remain renderable

### Requirement TELEGRAM-CATALOG-CAPTURE-013: Telegram webhook media SHALL resolve file ids into persisted attachment bytes
Cabinet MUST support a channel-edge media resolver that turns Telegram webhook photo/document file identifiers into attachment readers before capture ingestion, while preserving the Telegram file id, filename, MIME type, and source metadata used for audit.

#### Scenario: Resolve Telegram photo bytes before capture ingestion
- **GIVEN** Telegram delivers a webhook message with one or more photo sizes
- **WHEN** the channel edge resolves the selected Telegram file id into media bytes
- **THEN** Cabinet MUST persist those bytes as the capture attachment
- **AND** the persisted attachment metadata MUST preserve the Telegram file id-derived filename, MIME type, and photo kind
- **AND** the capture MUST still create only a preview/inbox handoff until explicit confirmation

#### Scenario: Resolve Telegram image document bytes before capture ingestion
- **GIVEN** Telegram delivers a webhook message with an image document, caption, and file id
- **WHEN** the channel edge resolves the document file id into media bytes
- **THEN** Cabinet MUST persist those bytes as the capture attachment
- **AND** the persisted attachment and Inbox audit metadata MUST preserve the Telegram file id, filename, MIME type, document-image kind, caption payload type, and inferred barcode context
- **AND** the capture MUST still create only a preview/inbox handoff until explicit confirmation

### Requirement TELEGRAM-CATALOG-CAPTURE-014: Telegram bot media resolver SHALL retrieve Bot API files before Cabinet ingestion
Cabinet MUST provide a Telegram Bot API media resolver that uses a runtime-supplied bot token to call \`getFile\`, download the returned file path, and return attachment bytes plus media metadata to the governed capture intake path.

#### Scenario: Retrieve Telegram file bytes through Bot API
- **GIVEN** Telegram delivers a photo or image document with a \`file_id\`
- **WHEN** the bot edge resolves the file for catalog capture
- **THEN** Cabinet MUST call the Telegram Bot API \`getFile\` method with that \`file_id\`
- **AND** it MUST download the returned file path through the token-bound file endpoint
- **AND** it MUST return a media reader with preserved file id, resolved filename, MIME type, and original media kind for capture ingestion
- **AND** it MUST require the runtime token at request construction time rather than persisting it in specs, fixtures, or capture records

### Requirement TELEGRAM-CATALOG-CAPTURE-015: Cabinet Inbox review links SHALL open Telegram capture threads
Telegram catalog capture Inbox items MUST expose the capture review URL as an actionable Cabinet link, and the Chats review surface MUST select the requested capture thread from that URL.

#### Scenario: Inbox opens the Telegram capture review thread
- **GIVEN** a Telegram catalog capture creates an Inbox item with metadata.review_url, thread_id, and preview_id
- **WHEN** the user opens the Inbox catch-up card and follows the review link
- **THEN** the link MUST point to the Cabinet Chats review URL
- **AND** the Chats surface MUST select the requested Telegram capture thread instead of defaulting to another thread

### Requirement TELEGRAM-CATALOG-CAPTURE-016: Cabinet SHALL establish Telegram capture authorization through private pairing
Cabinet MUST establish profile-scoped Telegram sender/chat authorization from a short-lived pairing code instead of requiring users to copy Telegram identifiers manually.

#### Scenario: Pair an exact private Telegram chat to a profile
- **GIVEN** the user has validated a Telegram bot for the active Cabinet profile and created an unexpired single-use pairing code
- **WHEN** the same Telegram user sends `/start <code>` from a private chat whose chat id equals the sender id
- **THEN** Cabinet MUST persist that exact sender/chat mapping for the profile
- **AND** the pairing code MUST become unusable after the first successful consume
- **AND** Cabinet MUST reject group chats, mismatched sender/chat identities, expired codes, and reused codes without changing the existing authorization mapping
- **AND** persisted pairing state MUST contain only a one-way code digest, expiry, and consumption state rather than the plaintext code

### Requirement TELEGRAM-CATALOG-CAPTURE-025: Integrations SHALL expose Telegram polling and private-pairing state
Cabinet MUST list Telegram in the Integrations registry and UI as an assistant capture channel whose readiness is derived from a validated write-only bot token, no webhook conflict, an exact private-chat pairing, and an active outbound long-polling connector.

#### Scenario: Show Telegram channel state in Integrations
- **GIVEN** the active profile has a validated Telegram bot and a paired private sender/chat mapping
- **WHEN** the user opens the Integrations page and reviews the Telegram provider
- **THEN** the provider registry MUST expose Telegram with `assistant_capture_channel` mode, bot-token/private-pairing auth metadata, outbound-long-polling transport, and assistant/media/text capture capabilities
- **AND** the Integrations UI MUST show non-secret bot identity, pairing, pause, offset/last-success, and degraded/error state without returning the token or implying that Cabinet exposes a public Telegram listener

### Requirement TELEGRAM-CATALOG-CAPTURE-026: Cabinet SHALL validate Telegram Bot API identity and webhook compatibility before polling
Cabinet MUST validate a candidate BotFather token against `getMe`, inspect `getWebhookInfo`, and require an explicit user action before removing a conflicting webhook.

#### Scenario: Validate bot identity without leaking the token
- **GIVEN** a user enters a candidate Telegram bot token for a Cabinet profile
- **WHEN** Cabinet tests the connection
- **THEN** Cabinet MUST call `getMe` and report non-secret bot identity
- **AND** Cabinet MUST persist the token only in the profile secret store after successful bot validation
- **AND** status, errors, logs, and API responses MUST NOT return the token

#### Scenario: Resolve a webhook conflict explicitly
- **GIVEN** `getWebhookInfo` reports a non-empty webhook URL
- **WHEN** Cabinet tests the connection
- **THEN** Cabinet MUST report a blocking webhook conflict without mutating Telegram configuration
- **AND** Cabinet MUST call `deleteWebhook` only after the user explicitly chooses the conflict-resolution action
- **AND** webhook removal MUST preserve pending updates

### Requirement TELEGRAM-CATALOG-CAPTURE-027: Cabinet SHALL poll Telegram outbound-only with durable exactly-once handoff boundaries
Cabinet MUST use outbound Bot API `getUpdates` long polling without opening an inbound Telegram listener, and MUST persist update offsets only after governed Cabinet processing succeeds.

#### Scenario: Poll restricted updates and advance after success
- **GIVEN** the connector has a validated bot and a persisted update offset
- **WHEN** it polls Telegram
- **THEN** it MUST send that offset, a bounded long-poll timeout, and only the supported message/callback update types
- **AND** it MUST process returned updates in update-id order through the existing governed Cabinet adapter
- **AND** it MUST persist `update_id + 1` only after that update finishes successfully

#### Scenario: Restart without duplicate processing
- **GIVEN** Cabinet has persisted an offset after successfully processing Telegram updates
- **WHEN** the runtime restarts and Telegram returns an older update plus the next update
- **THEN** Cabinet MUST resume from the persisted offset, skip the older update, and process the next update once
- **AND** a failed governed handoff MUST leave its update eligible for a later retry

#### Scenario: Poll the paired profile independently of the UI-active profile
- **GIVEN** one Cabinet profile has Telegram polling enabled and a different profile is currently active in the Cabinet UI
- **WHEN** the runtime selects Telegram connectors to poll
- **THEN** it MUST poll the Telegram-enabled profile instead of assuming the UI-active profile owns the bot
- **AND** it MUST keep every update, thread, preview, sender/chat mapping, and offset bound to that configured profile

### Requirement TELEGRAM-CATALOG-CAPTURE-028: Telegram polling SHALL use bounded retry and visible runtime health
Cabinet MUST run the Telegram connector with app lifecycle ownership, cancellation, bounded retry, Telegram `retry_after` handling, and non-secret health evidence.

#### Scenario: Recover from transient Telegram failure
- **GIVEN** Telegram returns a rate limit or transient network failure
- **WHEN** the connector schedules another poll
- **THEN** it MUST honor a positive Telegram `retry_after` value when present and otherwise use bounded exponential backoff
- **AND** status MUST report degraded state, safe error code, retry delay, last successful poll, and last processed update without exposing credentials
- **AND** pause, resume, token rotation, and disconnect controls MUST take effect without starting a public listener

### Requirement TELEGRAM-CATALOG-CAPTURE-029: Telegram connector acceptance SHALL include controlled real-runtime pairing proof
The release gate MUST exercise the running Cabinet API and UI against a controlled Telegram Bot API fixture, with live Telegram proof remaining an explicit operator-owned gate.

#### Scenario: Complete setup and pairing through a running Cabinet runtime
- **GIVEN** Cabinet is running with a controlled Bot API fixture and no real Telegram credential
- **WHEN** the acceptance journey enters a token, validates bot identity, resolves a fixture webhook conflict, creates a pairing code, and delivers a private `/start` update
- **THEN** the running connector MUST persist the exact profile/sender/chat mapping and offset
- **AND** the UI MUST show connected long-polling state, bot identity, and no public listener while keeping the token write-only
- **AND** this fixture proof MUST NOT be described as live Telegram production proof

### Requirement TELEGRAM-AGENT-CONVERSATION-001: Paired private Telegram text SHALL use the shared Cabinet Agent planner
After private pairing, Cabinet MUST treat ordinary Telegram text as natural-language Cabinet Agent input, preserve one stable profile-scoped thread for the paired sender/chat, and dispatch through the same governed planner used by in-app Chat.

#### Scenario: Continue one Cabinet conversation from Telegram
- **GIVEN** an exact private sender/chat is paired to a Cabinet profile
- **WHEN** that sender sends consecutive natural-language text messages
- **THEN** Cabinet MUST persist both messages in the same profile-scoped Chat thread
- **AND** planner context MUST identify the Telegram channel and conversation surface without exposing provider tools, raw skill ids, or a `key=value` command grammar
- **AND** a read-only result MUST return as natural Telegram text without confirmation buttons

#### Scenario: Reject text outside the paired private boundary
- **GIVEN** Telegram supplies an unpaired sender, a group chat, a mismatched private sender/chat, a caller-selected profile, or legacy Agent command grammar
- **WHEN** Cabinet receives the request
- **THEN** Cabinet MUST reject it before planner dispatch or mutation
- **AND** it MUST NOT create a cross-profile thread or grant authority from caller-authored fields

### Requirement TELEGRAM-AGENT-CONVERSATION-002: Telegram mutations SHALL use opaque durable confirmation callbacks
When the shared planner proposes a permitted non-admin mutation, Cabinet MUST persist the complete plan behind an opaque `asp_*` preview id and return only Apply and Cancel callbacks bound to that preview and exact Telegram conversation.

#### Scenario: Apply a Telegram preview exactly once
- **GIVEN** a paired private conversation has a pending unexpired `asp_*` preview
- **WHEN** the same sender/chat activates its Apply callback
- **THEN** Cabinet MUST revalidate profile authority, preview provenance, and target state before applying
- **AND** duplicate update or callback delivery MUST NOT apply the mutation more than once
- **AND** cross-sender, cross-profile, stale, expired, cancelled, malformed, or non-Telegram preview callbacks MUST fail closed

#### Scenario: Cancel without mutation
- **GIVEN** a paired private conversation has a pending `asp_*` preview
- **WHEN** the same sender/chat activates its Cancel callback
- **THEN** Cabinet MUST terminally cancel the preview without applying it
- **AND** later Apply delivery MUST not mutate Cabinet

### Requirement TELEGRAM-AGENT-CONVERSATION-003: Administrative work SHALL require authenticated in-app approval
Telegram identity and pairing MUST NOT confer Cabinet administrator authority or carry an administrator session.

#### Scenario: Route administrative intent to Cabinet
- **GIVEN** a paired Telegram sender asks for user, role, or other administrator-elevated work
- **WHEN** the shared planner classifies or selects that work
- **THEN** Cabinet MUST return a non-secret link to the stable in-app conversation
- **AND** it MUST NOT create an actionable Telegram preview or callback
- **AND** execution MUST require current authenticated in-app administrator approval

### Requirement TELEGRAM-CATALOG-CAPTURE-017: Lookup-backed Telegram drafts SHALL preserve lookup evidence
When Telegram catalog intake receives a barcode/product lookup result, Cabinet MUST preserve lookup source evidence with the preview, assistant message context, and Inbox audit trail so reviewers can distinguish resolved lookup-backed drafts from the manual barcode fallback.

#### Scenario: Preserve resolved lookup evidence in preview and audit metadata
- **GIVEN** an authorized Telegram capture includes a barcode and a resolved draft from a lookup source
- **WHEN** Cabinet creates the confirmation-required catalog preview
- **THEN** the preview payload MUST include the lookup source, URL, and confidence when available
- **AND** the assistant message context and Inbox item metadata MUST preserve the same lookup evidence for audit/review
- **AND** the preview MUST still require explicit confirmation before any catalog or inventory mutation

### Requirement TELEGRAM-CATALOG-CAPTURE-018: Telegram media-group captures SHALL group album photos before draft creation
Telegram channel adapters MUST be able to group webhook updates that share the same sender, chat, and Telegram media group id into one governed catalog capture input before preview creation.

#### Scenario: Group multi-photo Telegram album into one draft input
- **GIVEN** Telegram delivers multiple photo webhook updates with the same sender, chat, and media group id
- **WHEN** Cabinet prepares catalog capture input for the channel adapter
- **THEN** Cabinet MUST combine those updates into one capture input with all media attachments preserved in order
- **AND** the combined input MUST preserve captions/text, inferred barcode, draft fields, grouped message ids, distinct grouped update ids in arrival order, media group id, and album payload metadata
- **AND** updates from a different sender or chat MUST remain separate even when the media group id matches

### Requirement TELEGRAM-CATALOG-CAPTURE-023: Ungrouped Telegram photos SHALL remain separate capture inputs
Telegram channel adapters MUST NOT merge independent photo messages just because they come from the same authorized sender/chat. Only updates with a shared Telegram media group id may be combined into one draft input.

#### Scenario: Keep independent same-chat photo messages separate
- **GIVEN** Telegram delivers multiple photo webhook updates from the same sender and chat without a media group id
- **WHEN** Cabinet prepares catalog capture inputs for the channel adapter
- **THEN** Cabinet MUST keep each update as a separate capture input
- **AND** each separate input MUST preserve its own message id, photo media, barcode, text/caption, and draft fields for separate preview confirmation

### Requirement TELEGRAM-CATALOG-CAPTURE-024: Telegram media audit metadata SHALL preserve source file identifiers and sizes
Telegram media capture metadata MUST preserve Telegram source file identifiers and file size evidence in assistant message context and Inbox audit metadata so captured media remains traceable after Bot API file resolution and attachment persistence.

#### Scenario: Preserve Telegram media source identifiers and sizes
- **GIVEN** Telegram delivers a photo or image document with file id, file unique id, and file size metadata
- **WHEN** Cabinet normalizes, resolves, and persists the media as part of a catalog capture
- **THEN** the capture media input MUST preserve the Telegram file id, file unique id, and file size
- **AND** the assistant message media context and Inbox media metadata MUST include the same source identifiers and size evidence alongside attachment id, filename, MIME type, and media kind

### Requirement TELEGRAM-CATALOG-CAPTURE-019: Telegram webhook barcode captures SHALL use local barcode lookup evidence
When an authorized Telegram webhook capture contains a barcode that already matches a Cabinet item in the authorized profile, Cabinet MUST use that local match as lookup-backed draft evidence instead of falling back to a generic manual barcode draft.

#### Scenario: Draft webhook barcode capture from local match
- **GIVEN** a Cabinet profile has Telegram capture authorization and an existing catalog item with a matching barcode
- **WHEN** the Telegram webhook catalog capture endpoint receives a message containing that barcode from the authorized sender/chat
- **THEN** Cabinet MUST create the preview using the matched item part number, title, brand, and category
- **AND** the preview and Inbox metadata MUST preserve local barcode lookup source evidence
- **AND** no catalog item MUST be created until explicit confirmation is submitted
