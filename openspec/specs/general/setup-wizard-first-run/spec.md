## Purpose
Define the first-run Setup Wizard contract for Cabinet, based on required behavior: full-screen pre-auth flow only when config is missing, and deterministic completion that launches the app.

## Requirements

### Requirement SETUP-WIZ-001: Wizard trigger MUST be config-missing only and pre-auth
Setup Wizard MUST render as full-screen entry flow before login/app shell, only when `cabinet.json` is missing or invalid.

#### Scenario: Config missing
- **GIVEN** app starts and `cabinet.json` is missing
- **WHEN** entry route is resolved
- **THEN** full-screen Setup Wizard MUST render
- **AND** sign-in and authenticated shell MUST NOT render

#### Scenario: Config present
- **GIVEN** app starts and valid `cabinet.json` exists
- **WHEN** entry route is resolved
- **THEN** Setup Wizard MUST be skipped
- **AND** normal auth/session route MUST load

### Requirement SETUP-WIZ-002: Wizard MUST use step-based form UX with progress + nav controls
Wizard MUST provide step progress and deterministic controls: `Previous`, `Next`, and `Save`/`Save Draft` where applicable.

#### Scenario: Step navigation controls
- **GIVEN** user is in Setup Wizard
- **WHEN** user navigates steps
- **THEN** progress indicator MUST update deterministically
- **AND** `Previous`/`Next` controls MUST preserve entered state

### Requirement SETUP-WIZ-006: Wizard step layout MUST follow explicit progress-first template
Wizard steps MUST render with clear `STEP X OF N` labeling, visible progress bar/percentage, and deterministic footer actions.

#### Scenario: Progress-first step rendering
- **GIVEN** wizard has 3+ steps
- **WHEN** user is on any step
- **THEN** header MUST show step index (`STEP 1 OF N`, etc.) and progress percentage
- **AND** footer MUST show `Next` on intermediate steps, `Previous` from step 2 onward, and `Complete` on final step

#### Scenario: Final-step to completion transition
- **GIVEN** user is on final step and clicks `Complete`
- **WHEN** setup processing succeeds
- **THEN** UI MUST transition to setup completion state (`Config complete`) and show `Start App`/`Open Cabinet`
- **AND** registration-success/email-activation template copy MUST NOT render

### Requirement SETUP-WIZ-003: Wizard completion MUST show config-complete state and start action
Final step MUST show explicit config completion and provide primary action to start/open app.

#### Scenario: Setup complete screen
- **GIVEN** setup data validates and config write succeeds
- **WHEN** wizard reaches final step
- **THEN** UI MUST show `Config complete` (or equivalent setup-complete title)
- **AND** UI MUST provide primary `Start App`/`Open Cabinet` action
- **AND** UI MUST NOT show registration/email-activation template copy

### Requirement SETUP-WIZ-004: Completion MUST persist resolved runtime details
Wizard completion MUST display and persist resolved config/runtime metadata.

#### Scenario: Completion details
- **GIVEN** setup completed
- **WHEN** completion screen renders
- **THEN** UI MUST show resolved config path, data directory, and runtime URL/port

### Requirement SETUP-WIZ-005: Home dashboard MUST NOT embed setup-starter card
Setup onboarding controls MUST NOT appear as dashboard card in authenticated app shell.

#### Scenario: Authenticated home
- **GIVEN** user is authenticated and app shell is loaded
- **WHEN** home/dashboard renders
- **THEN** setup starter controls (`Start Setup`, `Import Existing Collection`, `Use Sample Data`, etc.) MUST NOT render in home card region

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SW-01 | Missing config startup | Full-screen setup wizard appears pre-auth | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-missing-config` |
| UC-SW-02 | Existing config startup | Wizard skipped; normal auth/shell loads | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-existing-config-skip` |
| UC-SW-03 | Step navigation | Progress + prev/next/save controls behave deterministically | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-step-controls` |
| UC-SW-04 | Completion state | Config-complete screen shows start action and no registration template copy | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-completion-state` |
| UC-SW-05 | Dashboard guard | Home contains no embedded setup starter card | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-not-in-home-shell` |
| UC-SW-06 | Progress template parity | Step header shows `STEP X OF N` + progress %, footer actions match step state | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-progress-template` |
| UC-SW-07 | Final complete transition | `Complete` transitions to `Config complete` + `Start App` action | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-complete-to-launch` |
