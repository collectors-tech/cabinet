## Purpose
Define Telegram as an authorized external intake channel for Cabinet catalog draft creation.

## Requirements
### Requirement TELEGRAM-CATALOG-CAPTURE-001: Telegram catalog capture SHALL require authorized sender/profile mapping
Cabinet MUST map each Telegram sender/chat to an explicitly authorized Cabinet profile before creating threads, attachments, previews, inbox items, or mutating catalog data.

#### Scenario: Reject unauthorized Telegram sender
- **GIVEN** a Telegram message is received from a sender/chat with no Cabinet profile authorization
- **WHEN** Cabinet evaluates the catalog capture request
- **THEN** Cabinet MUST reject the request before creating any chat thread, attachment, preview, inbox item, or catalog record

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


