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

#### Scenario: Inventory update apply records changed fields
- **GIVEN** user has previewed an inventory update with exact part number and title changes
- **WHEN** the user confirms apply from the confirmation dialog
- **THEN** Cabinet MUST update the active-profile inventory item with those fields
- **AND** the apply result and chat thread history MUST record the changed part number and title evidence

#### Scenario: Create new item/wishlist entry via chat
- **GIVEN** user requests creation of new record in chat
- **WHEN** copilot returns structured draft payload
- **THEN** user MUST be able to confirm creation and resulting record MUST be linked in chat outcome

#### Scenario: Apply without explicit confirmation leaves state unchanged
- **GIVEN** user has previewed an inventory create action
- **WHEN** the apply request does not include explicit confirmation
- **THEN** Cabinet MUST reject the apply with confirmation-required feedback
- **AND** the preview MUST remain unapplied and available to apply for the same profile/thread
- **AND** inventory state MUST remain unchanged
- **AND** chat thread history MUST NOT record an applied assistant outcome

#### Scenario: Wishlist entry apply persists state and records target item
- **GIVEN** user has previewed a wishlist entry creation with part number and title fields
- **WHEN** the user confirms apply from the confirmation dialog
- **THEN** Cabinet MUST create a wishlist entry and backing inventory item in the active profile
- **AND** the resulting inventory item MUST appear through the wishlist item API/surface with wishlist status
- **AND** the chat thread history MUST record an assistant outcome message that links the wishlist entry to the created item

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

#### Scenario: Collection assignment apply persists state and records outcome
- **GIVEN** user has previewed a collection assignment with an exact item and collection target
- **WHEN** the user confirms apply from the confirmation dialog
- **THEN** Cabinet MUST persist the item membership in the active profile workspace collections state
- **AND** the Collections surface MUST show the item assigned to the requested collection after navigation
- **AND** the chat thread history MUST record an assistant outcome message with the applied action, target item, and collection

#### Scenario: Failed collection assignment leaves state unchanged
- **GIVEN** user has previewed a collection assignment for an item target that is not present in the active profile
- **WHEN** the user confirms apply from the confirmation dialog
- **THEN** Cabinet MUST reject the apply without creating workspace collection membership
- **AND** the preview MUST remain pending for correction or cancellation
- **AND** the chat thread history MUST NOT record an applied outcome message for the failed mutation

#### Scenario: Failed inventory update apply leaves state unchanged
- **GIVEN** user has previewed an inventory update for an item target that is not present in the active profile
- **WHEN** the user confirms apply from the confirmation dialog
- **THEN** Cabinet MUST reject the apply without creating or updating inventory records
- **AND** the preview MUST remain pending for correction or cancellation
- **AND** the chat thread history MUST NOT record an applied outcome message for the failed mutation

#### Scenario: Canceled inventory update apply leaves state unchanged and records outcome
- **GIVEN** user has previewed an inventory update with exact target item and changed field values
- **WHEN** the user cancels from the confirmation dialog
- **THEN** Cabinet MUST mark the preview canceled without applying the changed fields to inventory
- **AND** the apply action MUST no longer be enabled for that canceled preview
- **AND** the chat thread history MUST record a canceled outcome message with no-mutation evidence
- **AND** the chat thread history MUST NOT record an applied outcome message for the canceled mutation

#### Scenario: Canceled collection assignment records target without mutation
- **GIVEN** user has previewed a collection assignment with an exact item and collection target
- **WHEN** the user cancels from the confirmation dialog
- **THEN** Cabinet MUST mark the preview canceled without assigning the item to the collection
- **AND** the chat thread history MUST record the canceled action, target item, target collection, and no-mutation evidence
- **AND** the canceled preview MUST reject any later apply attempt

#### Scenario: Thread context change clears pending action state
- **GIVEN** user has a pending chat action preview, cancellation notice, or apply result in one thread
- **WHEN** user switches to another chat thread in the same active profile
- **THEN** Cabinet MUST clear the pending preview, apply notice, apply result, and confirmation dialog
- **AND** the next thread MUST NOT expose an enabled apply action for the previous thread preview

#### Scenario: Pending action preview resumes after route return and reload
- **GIVEN** user has a pending chat action preview in the selected chat thread
- **WHEN** user navigates to another route, returns to Chats, or reloads the Chats route in the same browser session
- **THEN** Cabinet MUST restore the same pending preview for the active profile and thread
- **AND** the restored preview MUST remain explicitly pending and applyable only for that same profile/thread context

#### Scenario: Stale thread apply attempts leave state unchanged
- **GIVEN** user has a pending chat action preview in one thread for the active profile
- **WHEN** an apply request is attempted from another thread in the same active profile using that preview id
- **THEN** Cabinet MUST reject the apply as unavailable to that thread
- **AND** the owning preview MUST remain pending in its original thread
- **AND** inventory state MUST remain unchanged
- **AND** neither thread history MUST record an applied assistant outcome for the stale apply attempt

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

### Requirement UI-SCREEN-CHAT-COPILOT-010: Assistant sidebar SHALL operate as a compact chat with explicit screen-opening actions
The authenticated shell Assistant sidebar SHALL let users select an existing assistant chat, create a new assistant chat, send messages from the compact panel, and expose assistant-proposed navigation actions without changing the current page until the user invokes the action.

#### Scenario: Select an existing assistant chat from the sidebar
- **GIVEN** an authenticated actor has an active local profile and two or more assistant chat threads exist
- **WHEN** user opens the Assistant sidebar and selects an existing chat
- **THEN** the sidebar MUST render that chat's message history and clear pending action state from the previously selected chat
- **AND** the current Cabinet page route MUST remain unchanged

#### Scenario: Create a new assistant chat from the sidebar
- **GIVEN** an authenticated actor has an active local profile and the Assistant sidebar is open
- **WHEN** user invokes the new-chat control
- **THEN** Cabinet MUST create a fresh assistant chat for the active profile
- **AND** the sidebar MUST select the new chat with an empty compact conversation state

#### Scenario: Assistant exposes a screen-opening action from chat output
- **GIVEN** an authenticated actor is using the Assistant sidebar from another Cabinet page
- **WHEN** user sends a request for layout configuration help
- **THEN** the sidebar MUST expose an explicit action to open the relevant settings screen
- **AND** Cabinet MUST keep the current page route unchanged until user invokes that action
- **AND** invoking the action MUST navigate to the relevant settings screen

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
| UC-CHAT-13 | Collection assignment preview/apply | Preview and confirm summary preserve target item, collection name, and assistant defaults before apply; confirmed apply persists collection membership and records thread outcome | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-013 previews structured collection assignment targets before apply`; `internal/chat/service_test.go` `TestServiceThreadMessagePreviewApplyLifecycle` |
| UC-CHAT-14 | Wrong-profile preview apply | Preview apply is scoped to the owning profile/thread and rejected stale profile attempts leave inventory unchanged | implemented: `internal/chat/service_test.go` `TestServiceActionPreviewRejectsCrossProfileApply` |
| UC-CHAT-15 | Failed update apply | Missing inventory target rejects apply, keeps the preview pending, leaves inventory unchanged, and avoids false assistant applied history | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-014 keeps failed update apply pending without false history`; `internal/chat/service_test.go` `TestServiceUpdatePreviewApplyRejectsMissingTarget` |
| UC-CHAT-16 | Thread context reset | Pending preview/apply UI state is cleared when switching chat threads | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-015 clears pending action state when thread context changes` |
| UC-CHAT-17 | Pending preview route return/reload | Pending action preview restores for the same active profile/thread after route return and reload while remaining scoped to that context | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-016 restores pending action preview after route return and reload` |
| UC-CHAT-18 | Wishlist apply outcome | Confirmed wishlist creation persists the wishlist item and records entry/item linkage in chat history | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-008 supports confirm-before-apply for inventory and wishlist mutations`; `internal/chat/service_test.go` `TestServiceThreadMessagePreviewApplyLifecycle` |
| UC-CHAT-19 | Inventory update apply outcome | Confirmed inventory updates persist changed fields and record part/title evidence in UI and thread history | implemented: `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` `UI-SCREEN-CHAT-COPILOT-008 supports confirm-before-apply for inventory and wishlist mutations`; `internal/chat/service_test.go` `TestServiceThreadMessagePreviewApplyLifecycle` |
| UC-CHAT-20 | Failed collection assignment apply | Missing collection assignment target rejects apply, keeps the preview pending, leaves workspace collections unchanged, and avoids false assistant applied history | implemented: `internal/chat/service_test.go` `TestServiceCollectionAssignmentRejectsMissingTarget` |
| UC-CHAT-21 | Missing confirmation apply rejection | Apply attempts without explicit confirmation reject before mutation, keep the preview unapplied, and avoid applied assistant history | implemented: `internal/chat/service_test.go` `TestServiceThreadMessagePreviewApplyLifecycle` |
| UC-CHAT-22 | Stale thread apply rejection | Same-profile apply attempts from the wrong thread reject as unavailable, leave inventory unchanged, keep the owner preview pending, and avoid false assistant history in either thread | implemented: `internal/chat/service_test.go` `TestServiceActionPreviewRejectsCrossThreadApply` |
| UC-CHAT-23 | Assistant sidebar compact chat selection/new-chat/navigation action | Sidebar selects existing assistant chats, creates a new chat, sends a layout configuration prompt, exposes an explicit screen-opening action, and navigates only after invocation | implemented: `ui.web/cypress/e2e/chats/assistant-workspace/spec.cy.ts` `ASSISTANT-WORKSPACE-005 selects chats, creates a new chat, and exposes a layout navigation action` |
