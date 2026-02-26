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
