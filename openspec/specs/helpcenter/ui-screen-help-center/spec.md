## Purpose
Define Help Center screen behavior for in-app documentation discovery and reading.

## Requirements
### Requirement UI-SCREEN-HELP-CENTER-001: Help Center route SHALL surface an in-app article library
Help Center route SHALL render the available Help Center article set as a browsable in-app library rather than only a placeholder state.

#### Scenario: Open help center route
- **GIVEN** authenticated user navigates to `/help-center`
- **WHEN** route loads
- **THEN** UI MUST render a visible Help Center article library with selectable article entries and MUST avoid route-level error fallback state

### Requirement UI-SCREEN-HELP-CENTER-002: Help Center route SHALL render selected article content in-app
Help Center route SHALL let the user open readable article content inside the Help Center route without leaving the shell.

#### Scenario: Open article from help center library
- **GIVEN** help center route is active with article entries visible
- **WHEN** user selects a Help Center article
- **THEN** the route MUST show that article title and readable guide content in the in-app article viewer

### Requirement UI-SCREEN-HELP-CENTER-003: Help Center route SHALL preserve shell-level controls
Help Center route SHALL preserve global shell controls and layout contracts while the article library and reader are active.

#### Scenario: Use shell controls on help center
- **GIVEN** help center route is active
- **WHEN** user uses global search, theme, language, sidebar, or profile controls
- **THEN** controls MUST behave the same as other authenticated routes

### Requirement UI-SCREEN-HELP-CENTER-004: Help Center route SHALL filter articles by search text
Help Center route SHALL provide an in-page article search control that filters the article library by title, summary, category, or related article metadata without navigating away from `/help-center`.

#### Scenario: Search help center articles
- **GIVEN** help center route is active with article entries visible
- **WHEN** user enters a search query
- **THEN** the article library MUST show matching articles, hide non-matching articles, keep the article reader populated from the filtered result set, and show an explicit empty-results state when no articles match

### Requirement UI-SCREEN-HELP-CENTER-005: Help Center route SHALL support route-addressable article selection
Help Center route SHALL preserve the selected article in route search parameters so copied links, refreshes, and direct navigation reopen the same article.

#### Scenario: Open route-addressed article
- **GIVEN** authenticated user navigates to `/help-center?article=<article-id>`
- **WHEN** the route loads
- **THEN** the matching article MUST be selected in the article library and rendered in the in-app article viewer

### Requirement UI-SCREEN-HELP-CENTER-006: Help Center route SHALL provide section/category navigation
Help Center route SHALL expose article category controls that let users narrow the library to a section such as Getting Started, Sections, or Reference.

#### Scenario: Filter by help center category
- **GIVEN** help center route is active with article category controls visible
- **WHEN** user chooses a category control
- **THEN** the article library MUST narrow to that category and the article reader MUST select a readable article from the narrowed result set

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-HLP-01 | Open help center | Article library loads without 500/404 fallback | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-001 renders article library on help-center route` |
| UC-HLP-02 | Open article from help center | Selected article content renders in the in-app reader | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-002 renders selected article content in-app` |
| UC-HLP-03 | Use shell controls from help center | Search/theme/profile/sidebar controls remain usable | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-003 preserves shell controls on help-center route` |
| UC-HLP-04 | Search help center articles | Article library filters by query and renders empty-results feedback | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-004 filters article library by search query and empty state` |
| UC-HLP-05 | Open route-addressed article | Article query parameter selects the matching in-app article after sign-in/refresh | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-005 opens route-addressed articles from query parameters` |
| UC-HLP-06 | Filter by help center category | Category controls narrow the article library and keep a readable article selected | `ui.web/cypress/e2e/helpcenter/ui-screen-help-center/spec.cy.ts` `UI-SCREEN-HELP-CENTER-006 filters article library by category controls` |
