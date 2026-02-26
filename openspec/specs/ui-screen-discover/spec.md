## Purpose
Define Discover screen behavior and not-in-collection triage use cases.

## Requirements
### Requirement: Discover screen SHALL list triage-ready not-in-collection candidates
The screen SHALL load discovery candidates with price/query/date filters.

#### Scenario: Use case - triage filtered discoveries
- **WHEN** user applies filters and loads discoveries
- **THEN** screen SHALL show filtered candidate list

### Requirement: Discover screen SHALL support per-candidate actions
The screen SHALL support ignore, wishlist, track, and create-item actions.

#### Scenario: Use case - add discovery to wishlist
- **WHEN** user chooses wishlist action for candidate
- **THEN** app SHALL persist wishlist linkage and update discovery state
