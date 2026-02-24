# Discover Screen Spec

## Use Cases
1. User reviews new items not in collection.
2. User filters candidates by query/price/date.
3. User triages each candidate via ignore/wishlist/track/create.

## UI Sections
1. Filter bar
2. Candidate list
3. Candidate row actions

## State Behavior
- Loading: candidate list loading.
- Empty: no current discoveries with guidance.
- Error: load/action error with retry.
- Success: filtered candidate rows.

## Acceptance Criteria
- [ ] Filters map correctly to query parameters.
- [ ] Row action gives immediate inline feedback.
- [ ] Candidate list reflects action outcomes.
- [ ] Empty state includes CTA to run scanner.

## Test Cases
- `DISC-001` filter application.
- `DISC-002` ignore candidate action.
- `DISC-003` add-to-wishlist action.
- `DISC-004` create-item action.

