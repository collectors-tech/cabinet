## Purpose
Define Chat Copilot screen/rail behavior and assistant use cases.

## Requirements
### Requirement: Chat Copilot SHALL support thread/message and context-aware assistant flow
The screen/rail SHALL support thread list, message history, and context-aware queries.

#### Scenario: Use case - ask collection question
- **WHEN** user sends collection question in chat
- **THEN** assistant response SHALL be stored in thread history

### Requirement: Chat Copilot SHALL support guarded action preview/apply workflow
The screen/rail SHALL support preview and explicit confirmation before applying mutations.

#### Scenario: Use case - apply suggested mutation
- **WHEN** user confirms a proposed action from preview
- **THEN** app SHALL apply mutation and log action outcome
