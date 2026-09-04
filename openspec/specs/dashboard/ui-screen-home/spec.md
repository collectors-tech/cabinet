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
- **AND** Home SHALL preserve the failure status in Notification Inbox history with source, level, category, title, and summary metadata

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

### Requirement UI-SCREEN-HOME-009: Home SHALL render a collector and inventory manager signal hub
Home SHALL present the first viewport as an operational signal hub that balances collector health, purchase pipeline, inventory readiness, and actions needed from real dashboard data.

#### Scenario: Collector and inventory manager signal hub
- **GIVEN** an authenticated actor opens Home with dashboard data containing collection totals, wishlist hits, price drops, restocks, recent additions, and action cards
- **WHEN** the Home ready state renders
- **THEN** the first viewport SHALL show a prioritized signal band with collection size, wishlist hits, operational alerts, and collection value
- **AND** Home SHALL expose distinct sections for collector health, purchase pipeline, inventory readiness, and actions needed
- **AND** each drill-through action SHALL target a live Cabinet route such as `/inventory`, `/wishlist`, `/purchases`, `/media`, `/discoveries`, or a normalized replacement for unavailable legacy destinations

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
| UC-HOME-01 | Open Home with active data | Actionable priority cards render | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-001 renders actionable priority cards with direct actions` |
| UC-HOME-02 | Open Home with no pending actions | Calm empty state with CTA renders | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-002 renders deterministic loading and empty states` |
| UC-HOME-03 | Home data fetch failure | Inline error + retry appears | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-003 + UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves dashboard fetch errors in Inbox history` |
| UC-HOME-04 | Trigger quick action | Correct navigation/context opens | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-003 routes dashboard action links to live Cabinet destinations` |
| UC-HOME-05 | Authenticated Home setup exclusion | Setup wizard actions are not shown in Home | planned: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target` |
| UC-HOME-06 | Click Refresh Dashboard | Home data re-fetches without route transition | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-005 refreshes dashboard data without route transition` |
| UC-HOME-07 | Home renders without legacy onboarding controls | Legacy setup step controls absent from authenticated shell | planned: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` `UI-SCREEN-ONBOARDING-AUTH-001 locks workspace until sign-in then unlocks redirect target` |
| UC-HOME-08 | Resolve Home via canonical route and shell nav | `/dashboard` loads directly, `/` redirects to `/dashboard`, and Dashboard nav targets canonical path | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-007 resolves canonical /dashboard route, root redirect, and nav target stability` |
| UC-HOME-09 | Open Home after switching active profile | Summary counts and recent activity include only active profile records | Go: `internal/dashboard/service_test.go` `TestSummaryScopesSignalsToProfile`; `internal/app/dashboard_api_test.go` `TestDashboardEndpointScopesToActiveProfile` |
| UC-HOME-10 | Open Home with collector and inventory manager data | Signal hub renders KPI band, collector health, purchase pipeline, inventory readiness, actions needed, and live drill-through links | implemented: `ui.web/cypress/e2e/dashboard/ui-screen-home/spec.cy.ts` `UI-SCREEN-HOME-009 renders the collector and inventory manager signal hub` |
