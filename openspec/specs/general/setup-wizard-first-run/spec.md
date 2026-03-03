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

### Requirement SETUP-WIZ-007: Initial config payload MUST follow deterministic `cabinet.json` schema
Setup Wizard MUST produce a deterministic initial config object containing required runtime/bootstrap fields before app launch.

#### Scenario: Build initial config payload
- **GIVEN** wizard required fields are completed
- **WHEN** user clicks `Complete`
- **THEN** wizard MUST write `cabinet.json` with required sections: `instance`, `storage`, `runtime`, `auth`, `bootstrap`, and `meta`
- **AND** missing required fields MUST block completion with inline validation

### Requirement SETUP-WIZ-008: Startup MUST synchronize current runtime URL into config metadata
When `cabinet.json` exists, runtime startup MUST reconcile and persist the resolved runtime URL into metadata for deterministic post-launch introspection.

#### Scenario: Startup metadata sync
- **GIVEN** `cabinet.json` exists and runtime starts with resolved URL `http://<host>:<port>`
- **WHEN** startup initialization runs before serving requests
- **THEN** config metadata `meta.currentUrl` MUST match the resolved runtime URL
- **AND** metadata update MUST preserve existing config sections and schema validity

### Requirement SETUP-WIZ-009: PID lifecycle MUST use a PID-only runtime file
Runtime process lifecycle MUST maintain a PID-only file separate from `cabinet.json` configuration metadata.

#### Scenario: Runtime PID file contract
- **GIVEN** runtime startup succeeds
- **WHEN** PID file is written
- **THEN** PID file content MUST contain only the numeric PID value (no JSON/config payload)
- **AND** PID file lifecycle MUST remain independent from `cabinet.json`

#### Scenario: Clerk mode config requirements
- **GIVEN** user selects auth mode `clerk`
- **WHEN** completion is attempted
- **THEN** required Clerk keys/settings refs MUST be present in config payload or completion MUST fail with actionable validation message

#### Initial `cabinet.json` schema (v1)
```json
{
  "version": 1,
  "instance": {
    "name": "Primary",
    "profile": "primary"
  },
  "storage": {
    "dataDir": "./data",
    "mediaDir": "./data/media"
  },
  "runtime": {
    "portMode": "auto",
    "port": null,
    "resolvedUrl": "http://127.0.0.1:17880"
  },
  "auth": {
    "mode": "local",
    "clerk": {
      "publishableKey": "",
      "enabled": false
    }
  },
  "bootstrap": {
    "workspace": "Local Workspace",
    "databaseProfile": "Primary DB"
  },
  "features": {
    "chat": true,
    "providers": true,
    "scanner": true
  },
  "meta": {
    "createdAt": "<ISO-8601>",
    "updatedAt": "<ISO-8601>",
    "wizardVersion": "1"
  }
}
```

#### Scenario: Authenticated home
- **GIVEN** user is authenticated and app shell is loaded
- **WHEN** home/dashboard renders
- **THEN** setup starter controls (`Start Setup`, `Import Existing Collection`, `Use Sample Data`, etc.) MUST NOT render in home card region

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SW-01 | Missing config startup | Full-screen setup wizard appears pre-auth | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-missing-config` |
| UC-SW-02 | Existing config startup | Wizard skipped; normal auth/shell loads | planned: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `setup-wizard-existing-config-skip` |
| UC-SW-03 | Step navigation | Progress + prev/next/save controls behave deterministically | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-03 setup-wizard-step-controls preserves step form state while navigating previous/next` |
| UC-SW-04 | Completion state | Config-complete screen shows start action and no registration template copy | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-04 setup-wizard-completion-state shows runtime and storage details with start action` |
| UC-SW-05 | Dashboard guard | Home contains no embedded setup starter card | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-05 setup-wizard-not-in-home-shell keeps starter setup controls out of authenticated home` |
| UC-SW-06 | Progress template parity | Step header shows `STEP X OF N` + progress %, footer actions match step state | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-06 setup-wizard-progress-template shows step header, percentage, and footer actions` |
| UC-SW-07 | Final complete transition | `Complete` transitions to `Config complete` + `Start App` action | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-07 setup-wizard-complete-to-launch transitions to config complete with start action` |
| UC-SW-08 | Initial config schema write | `cabinet.json` contains deterministic required sections/fields after completion | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-08 setup-wizard-config-schema-write persists deterministic cabinet.json payload`; `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupStatusAndCompleteContract` |
| UC-SW-09 | Clerk config validation | Clerk mode blocks completion when required keys/settings refs are missing | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-09 setup-wizard-clerk-required-fields blocks completion when clerk key is missing`; `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupCompleteRequiresClerkPublishableKey` |
| UC-SW-10 | Startup runtime URL sync | Existing config receives `meta.currentUrl` matching resolved runtime URL on startup | implemented: `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupSyncCurrentURLUpdatesConfigMetadata` |
| UC-SW-11 | PID-only runtime lifecycle file | Runtime writes PID-only file and cleanup keeps PID lifecycle separate from config | implemented: `internal/app/runtime_setup_api_test.go` `TestRuntimePIDFileContainsPIDOnly` |
