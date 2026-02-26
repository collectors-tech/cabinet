## Purpose
Define Inventory AI Assist screen behavior and guarded suggestion use cases.

## Requirements
### Requirement: Inventory AI Assist screen SHALL expose suggestion workflows with confidence
The screen SHALL support title/photo suggestion requests and show confidence/error states.

#### Scenario: Use case - normalize title to metadata
- **WHEN** user submits listing title
- **THEN** screen SHALL show structured suggestion with confidence

### Requirement: Inventory AI Assist screen SHALL require explicit confirmation before apply
Mutating apply actions SHALL require explicit user confirmation.

#### Scenario: Use case - apply AI suggestion safely
- **WHEN** user attempts apply
- **THEN** app SHALL require confirm-before-apply and persist only after confirmation
