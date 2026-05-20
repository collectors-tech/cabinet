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
