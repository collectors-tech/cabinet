## Purpose
Define Home screen behavior as the primary command center with testable, action-first use-cases.

## Requirements
### Requirement UI-SCREEN-HOME-001: Home SHALL render a command-center summary with actionable priorities
Home SHALL present actionable discovery, pricing, and health signals above the fold.

#### Scenario: Home priority view ready state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home screen data loads successfully
- **THEN** users SHALL see actionable priority signals with direct actions

### Requirement UI-SCREEN-HOME-002: Home SHALL support deterministic state handling
Home SHALL support loading, empty, error, and ready states for each major panel.

#### Scenario: Home loading state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home data is still loading
- **THEN** Home SHALL render loading placeholders without layout shift

#### Scenario: Home empty state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** no actionable items exist
- **THEN** Home SHALL render calm-state guidance with at least one next action

#### Scenario: Home error state
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** Home API fetch fails
- **THEN** Home SHALL render actionable retry/error messaging without route crash

### Requirement UI-SCREEN-HOME-003: Home quick actions SHALL route to correct workflows
Each quick action SHALL navigate to the correct destination with preserved context.

#### Scenario: Quick action navigation
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user clicks a Home quick action
- **THEN** app SHALL open the expected screen/workflow context

#### Scenario: Recently Added card routes to collections
- **GIVEN** Home dashboard data includes a `Recently Added` quick-action card for recently added inventory
- **WHEN** user opens that quick action from Home
- **THEN** the app SHALL route to `/collections`
- **AND** the action MUST NOT target invalid singular route `/collection`

### Requirement UI-SCREEN-HOME-004 (Deprecated): Home onboarding rail starter actions
This requirement is deprecated. Starter setup is now a pre-auth setup wizard and MUST NOT be rendered inside authenticated Home.
#### Scenario: Deprecated behavior reference
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** legacy Home onboarding panel behavior is evaluated
- **THEN** implementation SHALL NOT reintroduce setup wizard controls on Home

#### Scenario: Deprecated step-navigation controls
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user is in authenticated Home route
- **THEN** onboarding step controls SHALL be absent
- **AND** controls labeled `Back Step` and `Next Step` MUST NOT render on authenticated Home

### Requirement UI-SCREEN-HOME-005: Home toolbar SHALL support explicit refresh action
Home SHALL expose a `Refresh Dashboard` action that re-fetches Home data and preserves route context.

#### Scenario: Refresh dashboard action
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user clicks `Refresh Dashboard`
- **THEN** Home data SHALL refresh and render updated panel state without route transition

### Requirement UI-SCREEN-HOME-006: Home SHALL NOT render setup wizard content in authenticated shell
Authenticated Home route SHALL not include starter setup wizard cards or controls.

#### Scenario: Home excludes setup wizard
- **GIVEN** user is authenticated and route is Home
- **WHEN** Home renders dashboard content
- **THEN** labels and controls `Starter Onboarding`, `Start Setup`, `Import Existing Collection`, and `Use Sample Data` MUST NOT be present
### Requirement UI-SCREEN-HOME-007: Home SHALL use `/dashboard` as the canonical route
Authenticated Home SHALL load on `/dashboard`, while `/` SHALL redirect deterministically to `/dashboard` without route drift.

#### Scenario: Canonical dashboard route and redirect
- **GIVEN** an authenticated actor opens the app shell from a direct deep link, root entry, sidebar navigation, refresh, or browser history interaction
- **WHEN** Home route resolution occurs
- **THEN** `/dashboard` SHALL render the Home screen directly
- **AND** `/` SHALL redirect to `/dashboard`
- **AND** primary Dashboard navigation SHALL target `/dashboard`
- **AND** refresh/back-forward behavior on `/dashboard` SHALL remain stable without falling into 404

### Requirement UI-SCREEN-HOME-008: Home summary data SHALL be scoped to the active profile
Home API summary counts, action signals, and recent activity SHALL use the active profile as the data boundary.

#### Scenario: Active profile summary isolation
- **GIVEN** two local profiles have different inventory, wishlist, discovery, and pricing records
- **WHEN** the user opens Home with one profile active
- **THEN** Home summary totals, action cards, and recently-added items SHALL include only records for the active profile
- **AND** records from inactive profiles MUST NOT inflate Home counts or recent activity

## Acceptance Criteria
- Every Home critical flow has UC ID and deterministic expected outcome.
- E2E mapping exists for command-center render and quick actions.
- Starter setup is excluded from authenticated Home and covered by setup/auth specs.

## Success Criteria
- Users can identify and start next action from Home in one interaction.
- No placeholder-only content remains in Home ready state.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-HOME-01 | Open Home with active data | Actionable priority cards render | planned: `cypress/e2e/ui/home.cy.ts` `home-renders-priority-cards` |
| UC-HOME-02 | Open Home with no pending actions | Calm empty state with CTA renders | planned: `cypress/e2e/ui/home.cy.ts` `home-empty-state` |
| UC-HOME-03 | Home data fetch failure | Inline error + retry appears | planned: `cypress/e2e/ui/home.cy.ts` `home-error-retry` |
| UC-HOME-04 | Trigger quick action | Correct navigation/context opens | planned: `cypress/e2e/ui/home.cy.ts` `home-quick-action-routing` |
| UC-HOME-05 | Authenticated Home setup exclusion | Setup wizard actions are not shown in Home | planned: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target` |
| UC-HOME-06 | Click Refresh Dashboard | Home data re-fetches without route transition | planned: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `home-refresh-dashboard` |
| UC-HOME-07 | Home renders without legacy onboarding controls | Legacy setup step controls absent from authenticated shell | planned: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target` |
| UC-HOME-08 | Resolve Home via canonical route and shell nav | `/dashboard` loads directly, `/` redirects to `/dashboard`, and Dashboard nav targets canonical path | planned: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-007 resolves canonical /dashboard route, root redirect, and nav target stability` |
| UC-HOME-09 | Open Home after switching active profile | Summary counts and recent activity include only active profile records | Go: `internal/dashboard/service_test.go` `TestSummaryScopesSignalsToProfile`; `internal/app/dashboard_api_test.go` `TestDashboardEndpointScopesToActiveProfile` |
