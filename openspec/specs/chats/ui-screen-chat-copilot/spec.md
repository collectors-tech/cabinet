## Purpose
Define Chat Copilot screen/rail behavior for conversational workflows and guarded actions.

## Requirements
### Requirement UI-SCREEN-CHAT-COPILOT-001: Chat Copilot SHALL support persistent threads and messages
Chat SHALL support local profile-scoped thread and message persistence.

#### Scenario: Reopen thread history
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user reopens existing thread
- **THEN** thread history SHALL render prior messages

### Requirement UI-SCREEN-CHAT-COPILOT-002: Chat Copilot SHALL support attachment and action preview flows
Chat SHALL support user-selected file attachments and preview-before-apply actions.

#### Scenario: Upload attachment from chat panel
- **GIVEN** user is on Chats screen with active thread context
- **WHEN** user clicks `Upload` and selects a file
- **THEN** attachment MUST be linked to chat context with deterministic success/error feedback

#### Scenario: Preview action before apply
- **GIVEN** chat action draft fields are available
- **WHEN** user clicks `Preview Action`
- **THEN** UI MUST render preview output without mutating persisted records

#### Scenario: Apply action from preview
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user confirms apply for previewed action
- **THEN** mutation SHALL execute and log outcome

### Requirement UI-SCREEN-CHAT-COPILOT-003: Chat Copilot SHALL support deterministic state handling
Chat SHALL support loading, empty, error, and ready states for threads/messages.

#### Scenario: Chat service error
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** thread/message API fails
- **THEN** chat SHALL show actionable error state with retry

### Requirement UI-SCREEN-CHAT-COPILOT-009: Chat Copilot SHALL disable contradictory thread-creation controls while bootstrap context is unavailable
When active profile or bootstrap chat context is unavailable, Chats SHALL not present enabled thread-creation controls that imply usable chat state.

#### Scenario: Active profile bootstrap failure blocks thread creation
- **GIVEN** user opens `/chats` and active profile bootstrap fails (for example `active_profile_404`)
- **WHEN** unavailable state renders
- **THEN** UI MUST show the unavailable/retry state
- **AND** `New thread title` input MUST be disabled
- **AND** `Create` MUST remain disabled until chat context recovers

### Requirement UI-SCREEN-CHAT-COPILOT-004: Chat Copilot SHALL support AI-enabled photo analysis conversations
When AI is enabled, chat SHALL support photo-driven analysis prompts and return structured suggestions linked to media assets.

#### Scenario: Analyze attached photo in chat
- **GIVEN** AI is enabled for active profile and user attaches one or more media assets in chat
- **WHEN** user asks assistant to analyze the photos
- **THEN** runtime MUST return analysis output with confidence and suggested actions/assignments, and preserve linkage to referenced `asset_id` values

#### Scenario: AI disabled guard for photo analysis
- **GIVEN** AI is disabled for active profile
- **WHEN** user requests photo analysis in chat
- **THEN** UI MUST return deterministic disabled-state guidance and SHALL NOT run provider inference

### Requirement UI-SCREEN-CHAT-COPILOT-005: Photo analysis in Cabinet SHALL persist rich metadata for cataloging and assignment
Cabinet SHALL store maximal practical analysis metadata for each analyzed asset to support search, dedupe, assignment, and audit workflows.

#### Scenario: Persist rich metadata after photo analysis
- **GIVEN** chat photo analysis completes for an asset
- **WHEN** metadata is persisted
- **THEN** record MUST include at minimum `asset_id`, `thread_id`, `provider`, `model`, `prompt_template_id`, `analysis_version`, `confidence`, `title`, `description`, `tags[]`, `brand`, `model_name`, `part_number`, `year`, `condition_hint`, `detected_text`, `ocr_text`, `possible_duplicates[]`, `suggested_inventory_links[]`, `suggested_wishlist_links[]`, `timestamp`, and raw/normalized payload references

#### Scenario: Reuse rich metadata in subsequent workflows
- **GIVEN** rich metadata exists for asset
- **WHEN** user opens media card, inventory assignment, wishlist assignment, or follow-up chat
- **THEN** runtime/UI MUST retrieve and use stored metadata without requiring re-analysis unless explicitly requested

### Requirement UI-SCREEN-CHAT-COPILOT-007: Chat Copilot SHALL support inventory and wishlist CRUD assistance with confirm-before-apply
Copilot SHALL assist with creating/updating inventory and wishlist records, but MUST require explicit user confirmation before mutating data.

#### Scenario: Update existing inventory item via chat
- **GIVEN** user asks copilot to update an inventory item
- **WHEN** copilot proposes field changes
- **THEN** UI MUST present confirmation summary and only apply changes after explicit confirm action

#### Scenario: Create new item/wishlist entry via chat
- **GIVEN** user requests creation of new record in chat
- **WHEN** copilot returns structured draft payload
- **THEN** user MUST be able to confirm creation and resulting record MUST be linked in chat outcome

#### Scenario: Cancel apply keeps preview pending
- **GIVEN** user has a previewed chat action and the confirm-before-apply dialog is open
- **WHEN** user cancels the apply confirmation
- **THEN** Cabinet MUST close the confirmation dialog without applying the action
- **AND** the pending preview MUST remain visible with actionable cancellation feedback

#### Scenario: Empty thread cannot preview actions without source context
- **GIVEN** a chat thread has no messages and no uploaded attachment context
- **WHEN** the user opens Action Preview controls
- **THEN** `Preview Action` MUST remain disabled until source conversation context exists
- **AND** the UI MUST NOT generate preview artifacts from seeded defaults alone

#### Scenario: Provider defaults are visible on previewed chat actions
- **GIVEN** the active profile has assistant provider/model defaults configured
- **WHEN** the user previews a structured chat action
- **THEN** the chat action surface MUST show the active assistant provider/model defaults before apply
- **AND** the preview and confirm-before-apply summary MUST preserve the same provider/model context

#### Scenario: Collection assignment previews show exact target boundaries
- **GIVEN** user asks copilot to assign an inventory item to a workspace collection
- **WHEN** the user previews the collection assignment action
- **THEN** the preview MUST show the target inventory item and collection name before apply
- **AND** the confirm-before-apply summary MUST preserve the same target item, collection, and assistant provider/model context

### Requirement UI-SCREEN-CHAT-COPILOT-008: Mobile chat SHALL support image attachment for analysis and record creation workflows
Mobile chat flow SHALL allow attaching or capturing images, sending them to copilot, and using results to create/update inventory or wishlist entries.

#### Scenario: Mobile image-to-inventory flow
- **GIVEN** user is on mobile chat and attaches an image
- **WHEN** copilot analyzes image and suggests structured fields
- **THEN** user MUST be able to confirm and create/update target record with linked media asset

### Requirement UI-SCREEN-CHAT-COPILOT-006: Header chat trigger SHALL render as icon-only control
Header chat/copilot trigger in Cabinet shell SHALL render as icon-only action (no inline text label in header row).

#### Scenario: Header chat action visual contract
- **GIVEN** authenticated shell header is visible on desktop
- **WHEN** chat trigger renders
- **THEN** trigger MUST render as icon-only control
- **AND** visible text label next to the icon MUST NOT render in header row
- **AND** control MUST retain accessible name via `aria-label` and/or tooltip

### Requirement UI-SCREEN-CHAT-COPILOT-009: Top-level `/inbox` SHALL resolve to a real communications surface
Cabinet SHALL provide a reachable authenticated `/inbox` route for communications access instead of falling through to the not-found page.

#### Scenario: Open top-level inbox route
- **GIVEN** an authenticated actor has an active local profile
- **WHEN** user navigates directly to `/inbox`
- **THEN** Cabinet MUST render a communications surface
- **AND** route MUST NOT render the app 404 page

## Acceptance Criteria
- UC IDs cover thread persistence, attachments, and guarded action apply.
- E2E mapping includes chat open/close and action safety flows.

## Success Criteria
- Chat remains usable across screen context changes.
- Mutating chat actions are always explicit and auditable.

## Data Profiles
- Sample: 10 threads, 100 messages, 10 attachments
- Bulk: 1,000 threads, 20,000 messages, 2,000 attachments

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-CHAT-01 | Open chat and list threads | Thread list renders from API | planned: `cypress/e2e/ui/chat.cy.ts` `chat-thread-list` |
| UC-CHAT-02 | Reopen thread | Prior messages render | planned: `cypress/e2e/ui/chat.cy.ts` `chat-thread-history` |
| UC-CHAT-03 | Add attachment | Attachment linked to thread | planned: `cypress/e2e/ui/chat.cy.ts` `chat-attachment` |
| UC-CHAT-04 | Preview and apply action | Confirm-before-apply enforced | planned: `cypress/e2e/ui/chat.cy.ts` `chat-guarded-apply` |
| UC-CHAT-05 | Chat API failure | Error + retry appears | planned: `cypress/e2e/ui/chat.cy.ts` `chat-error-state` |
| UC-CHAT-06 | Upload attachment | `Upload` links file to chat context | planned: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `chat-upload-attachment` |
| UC-CHAT-07 | Empty-thread preview gating | `Preview Action` stays disabled until source chat context exists | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-007 keeps Preview Action gated until thread context exists` |
| UC-CHAT-08 | Preview action | `Preview Action` renders dry-run output before apply | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-008 supports confirm-before-apply for inventory and wishlist mutations` |
| UC-CHAT-09 | Mobile image attachment flow | image attachment supports confirm-before-apply workflow | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-009 supports mobile image attachment and confirm-before-apply flow` |
| UC-CHAT-10 | Unavailable bootstrap state | Thread creation controls stay disabled until chat context recovers | planned: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `chat-unavailable-disables-thread-create` |
| UC-CHAT-11 | Cancel preview apply | Preview remains pending and no applied result is shown | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-011 cancels preview apply without mutating the pending action` |
| UC-CHAT-12 | Provider defaults in preview | Preview and confirm summary preserve active assistant provider/model defaults | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-012 reflects assistant provider defaults in chat action previews` |
| UC-CHAT-13 | Collection assignment preview | Preview and confirm summary preserve target item, collection name, and assistant defaults before apply | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-013 previews structured collection assignment targets before apply` |
