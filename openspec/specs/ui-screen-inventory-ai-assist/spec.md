## Purpose
Define Inventory AI Assist screen behavior for suggestion flows with guarded apply.

## Requirements
### Requirement: Inventory AI Assist SHALL support title and photo suggestion workflows
The screen SHALL support title normalization and photo-based identification requests.

#### Scenario: Title normalize request
- **WHEN** user submits listing title
- **THEN** structured suggestion response SHALL render

### Requirement: Inventory AI Assist SHALL enforce confirm-before-apply
AI-driven mutations SHALL not execute without explicit user confirmation.

#### Scenario: Apply suggested mutation
- **WHEN** user chooses apply
- **THEN** confirmation SHALL be required before mutation executes

### Requirement: Inventory AI Assist SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states.

#### Scenario: AI service error
- **WHEN** AI call fails
- **THEN** screen SHALL show actionable error state and retry path

## Acceptance Criteria
- UC IDs cover suggest, apply, and failure paths.
- E2E mappings exist for guarded mutation behavior.

## Success Criteria
- AI enhances workflows without unsafe auto-mutation.
- Error states are clear and recoverable.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-AI-01 | Suggest from title | Structured suggestion shown | planned: `cypress/e2e/ui/ai-assist.cy.ts` `ai-title-suggest` |
| UC-AI-02 | Suggest from photo | Suggestion with confidence shown | planned: `cypress/e2e/ui/ai-assist.cy.ts` `ai-photo-suggest` |
| UC-AI-03 | Apply suggestion | Confirm-before-apply enforced | planned: `cypress/e2e/ui/ai-assist.cy.ts` `ai-guarded-apply` |
| UC-AI-04 | AI failure | Error + retry shown | planned: `cypress/e2e/ui/ai-assist.cy.ts` `ai-error-state` |
