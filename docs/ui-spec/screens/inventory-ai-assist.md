# Inventory AI Assist Screen Spec

## Use Cases
1. User enables/disables AI features by preference.
2. User normalizes listing title into structured fields.
3. User identifies item candidate from photo URL.
4. User confirms before applying AI suggestion.

## UI Sections
1. AI enable/disable controls
2. Title normalize form
3. Photo identify form
4. Suggestion preview + confidence
5. Confirm apply action

## State Behavior
- Loading: action-specific pending state.
- Empty: no suggestion yet.
- Error: scoped AI error with retry.
- Success: suggestion preview and optional apply.

## Acceptance Criteria
- [ ] AI does not auto-write values without confirmation.
- [ ] Confidence is displayed when provided by API.
- [ ] Failed AI actions do not affect other inventory tabs.
- [ ] Retry action reuses latest request context safely.

## Test Cases
- `INV-AI-001` enable/disable AI.
- `INV-AI-002` title normalize request + preview.
- `INV-AI-003` apply suggestion requires explicit click.
- `INV-AI-004` error then retry flow.

