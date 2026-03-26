## Purpose
Define how Cabinet Inbox and Assistant interact so notifications, async results, and execution outcomes can move cleanly between the two surfaces.

## Requirements
### Requirement ASSISTANT-INBOX-001: Inbox SHALL store surfaced assistant execution results and failures
Cabinet Inbox MUST persist important assistant execution outcomes and failures instead of relying only on transient toasts.

#### Scenario: Assistant run emits inbox item
- **GIVEN** assistant run completes with failure, warning, or important result
- **WHEN** runtime surfaces the outcome
- **THEN** Inbox MUST create a readable item with status and timestamp metadata

### Requirement ASSISTANT-INBOX-002: Inbox items SHALL link back to related assistant thread or execution detail
Inbox items originating from assistant activity MUST provide deterministic linkage back to the related assistant thread or execution record.

#### Scenario: Open assistant result from inbox
- **GIVEN** Inbox item originated from assistant execution
- **WHEN** user opens that Inbox item
- **THEN** Cabinet MUST open the related assistant thread or execution detail surface with sufficient context to continue work

### Requirement ASSISTANT-INBOX-003: Assistant SHALL surface async execution states consistently across thread and Inbox
When assistant actions are asynchronous, Cabinet MUST expose queued/running/completed/failed states in both Assistant and Inbox without conflicting status narratives.

#### Scenario: Async execution transitions
- **GIVEN** assistant action starts an async job
- **WHEN** job moves through queued, running, and completed states
- **THEN** Assistant thread state and Inbox item state MUST remain consistent and auditable
