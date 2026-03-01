## Purpose
Define Inventory Photos screen behavior for upload, media management, and fullscreen inspection.

## Requirements
### Requirement UI-SCREEN-INVENTORY-PHOTOS-001: Inventory Photos SHALL support full media lifecycle
Photos screen SHALL support upload, list, primary selection, and delete workflows.

#### Scenario: Upload and primary update
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user uploads image and sets primary
- **THEN** media list SHALL reflect updated primary state

### Requirement UI-SCREEN-INVENTORY-PHOTOS-002: Inventory Photos SHALL support deterministic state handling
Photos screen SHALL support loading, empty, error, and ready states.

#### Scenario: Photos empty state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** selected item has no photos
- **THEN** screen SHALL show empty guidance for upload actions

#### Scenario: Photos error state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** media API fails
- **THEN** screen SHALL show retry-capable error state

### Requirement UI-SCREEN-INVENTORY-PHOTOS-003: Inventory Photos SHALL support fullscreen inspection
Photos SHALL open in fullscreen view with stable controls.

#### Scenario: Fullscreen open
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user requests fullscreen view
- **THEN** fullscreen viewer SHALL render selected media

## Acceptance Criteria
- UC IDs cover upload, primary, fullscreen, and error-state paths.
- E2E mappings defined for media lifecycle.

## Success Criteria
- Users complete photo workflows without leaving screen context.
- Media actions never require manual DB recovery due to UI failures.

## Data Profiles
- Sample: 300 photos across 100 items
- Bulk: 150,000 media metadata rows

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-PHO-01 | Upload photo | Photo appears in list | planned: `cypress/e2e/ui/photos.cy.ts` `photo-upload` |
| UC-PHO-02 | Set primary | Primary indicator updates | planned: `cypress/e2e/ui/photos.cy.ts` `photo-set-primary` |
| UC-PHO-03 | No photos for item | Empty state guidance appears | planned: `cypress/e2e/ui/photos.cy.ts` `photo-empty-state` |
| UC-PHO-04 | Media API failure | Error + retry appears | planned: `cypress/e2e/ui/photos.cy.ts` `photo-error-state` |
| UC-PHO-05 | Open fullscreen | Fullscreen viewer renders media | planned: `cypress/e2e/ui/photos.cy.ts` `photo-fullscreen` |
