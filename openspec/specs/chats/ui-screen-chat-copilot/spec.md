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

### Requirement UI-SCREEN-CHAT-COPILOT-006: Header chat trigger SHALL render as icon-only control
Header chat/copilot trigger in Cabinet shell SHALL render as icon-only action (no inline text label in header row).

#### Scenario: Header chat action visual contract
- **GIVEN** authenticated shell header is visible on desktop
- **WHEN** chat trigger renders
- **THEN** trigger MUST render as icon-only control
- **AND** visible text label next to the icon MUST NOT render in header row
- **AND** control MUST retain accessible name via `aria-label` and/or tooltip

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
