## ADDED Requirements

### Requirement: Inventory routes MUST not return fatal 500 user states for expected datasets
The system SHALL render inventory and wishlist pages successfully for empty, seeded, and high-volume datasets without surfacing generic fatal 500 error pages.

#### Scenario: Empty dataset route load
- **WHEN** a user opens inventory with no items
- **THEN** the page SHALL render a non-error empty state and actionable controls

#### Scenario: Seeded dataset route load
- **WHEN** a user opens inventory with valid seeded data
- **THEN** the page SHALL render list content without fatal runtime errors

#### Scenario: Large dataset route load
- **WHEN** a user opens inventory with large fixture data
- **THEN** the page SHALL remain usable and SHALL not transition to a 500 page

### Requirement: User-facing error states MUST be recoverable and explicit
The system SHALL present structured error feedback for failed data operations and SHALL not collapse entire screen routes when a single operation fails.

#### Scenario: API operation fails in-page
- **WHEN** an inventory mutation request fails
- **THEN** the system SHALL show operation-level error feedback and keep the route usable
