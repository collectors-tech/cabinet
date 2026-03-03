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

### Requirement SETUP-WIZ-010: Step 1 MUST render explicit welcome actions for setup start and config import
When setup is required, the first wizard screen MUST render `Start Setup` and `Import Existing Config` actions before the user enters multi-step configuration fields.

#### Scenario: Welcome actions on missing config
- **GIVEN** app startup detects `cabinet.json` missing and setup route is active
- **WHEN** wizard Step 1 renders
- **THEN** UI MUST show primary `Start Setup` and secondary `Import Existing Config` actions
- **AND** instance/profile field editors MUST remain hidden until user chooses `Start Setup`

#### Scenario: Start Setup deterministic transition
- **GIVEN** wizard Step 1 welcome actions are visible
- **WHEN** user clicks `Start Setup`
- **THEN** wizard MUST transition deterministically to form step mode
- **AND** instance/profile editors MUST become visible without route reload

### Requirement SETUP-WIZ-011: Import Existing Config MUST support deterministic import flow
Wizard Step 1 MUST provide an import path entry flow that copies a valid existing config file into the runtime setup config path and exits setup-required state.

#### Scenario: Successful setup config import
- **GIVEN** setup-required state with Step 1 visible and a valid external config file path
- **WHEN** user submits import action
- **THEN** runtime MUST validate and write imported config to active `cabinet.json` path
- **AND** response MUST return status `200` with `ok=true`, `setup_required=false`, and `config_path`
- **AND** subsequent `GET /api/runtime/setup-status` MUST return `setup_required=false`

#### Scenario: Invalid import path validation
- **GIVEN** setup-required state and import action is submitted with missing or unreadable source path
- **WHEN** runtime processes import request
- **THEN** runtime MUST return status `400` with deterministic `error_code` and `message`
- **AND** setup-required state MUST remain `true`

### Requirement SETUP-WIZ-012: Identity step MUST support optional profile key with inline validation and config path preview
Identity form step MUST capture instance name, accept optional profile key, and show deterministic config-path preview while preserving entered state.

#### Scenario: Identity step path preview
- **GIVEN** setup wizard is in identity form mode
- **WHEN** identity step renders
- **THEN** UI MUST show current config destination path preview
- **AND** path preview MUST match `/api/runtime/setup-status` `config_path`

#### Scenario: Optional profile key auto-derivation
- **GIVEN** instance name is provided and profile key is blank
- **WHEN** setup completion is submitted
- **THEN** runtime MUST derive deterministic profile key from instance name
- **AND** persisted config `instance.profile` MUST be non-empty and normalized

#### Scenario: Inline identity validation
- **GIVEN** identity step is active
- **WHEN** user clicks `Next` with blank instance name
- **THEN** UI MUST render inline validation error without leaving identity step
- **AND** previously entered identity fields MUST remain intact

### Requirement SETUP-WIZ-013: Storage step MUST support exe-local default, custom path validation, and portable mode
Storage form step MUST expose default exe-local storage, optional custom data path, writable/free-space checks, and portable-mode toggle before proceeding.

#### Scenario: Exe-local default storage
- **GIVEN** setup wizard storage step is active
- **WHEN** step first renders
- **THEN** storage mode MUST default to `exe_local`
- **AND** data directory preview MUST resolve to `<exe_dir>/data`
- **AND** portable mode toggle MUST be visible and off by default

#### Scenario: Custom storage inline validation
- **GIVEN** storage mode is set to `custom`
- **WHEN** user attempts to continue with blank custom data path
- **THEN** UI MUST show inline validation error
- **AND** wizard MUST remain on storage step

#### Scenario: Storage validation contract
- **GIVEN** storage step submits a candidate data path
- **WHEN** runtime validates the path
- **THEN** runtime MUST return deterministic validation payload including `writable` and free-space check fields
- **AND** wizard MUST block next transition when `writable=false`

#### Scenario: Persisted storage selection
- **GIVEN** setup completion succeeds after storage step selection
- **WHEN** setup config is written
- **THEN** `storage.dataDir` and `storage.mediaDir` MUST match selected storage mode/path
- **AND** storage selection MUST be reflected in completion details

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
| UC-SW-12 | Welcome actions before form | Step 1 shows `Start Setup` + `Import Existing Config` and hides editors until start | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-12 setup-wizard-welcome-actions renders start/import actions before form fields` |
| UC-SW-13 | Import existing config action | Importing valid existing config clears setup-required state and returns sign-in form | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-13 setup-wizard-import-existing-config loads external config and exits setup mode`; `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupImportExistingConfigContract` |
| UC-SW-14 | Identity path preview | Identity step shows deterministic config path preview from setup status payload | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-14 setup-wizard-identity-path-preview shows config destination path` |
| UC-SW-15 | Optional profile key derivation | Blank profile key auto-derives normalized profile during completion | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-15 setup-wizard-optional-profile-key auto-derives profile key from instance name`; `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupCompleteDerivesProfileKeyWhenBlank` |
| UC-SW-16 | Inline identity validation | Blank instance name shows inline validation and stays on identity step | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-16 setup-wizard-identity-inline-validation blocks next on missing instance name` |
| UC-SW-17 | Exe-local storage defaults | Storage step defaults to exe-local mode with `<exe_dir>/data` preview and portable toggle | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-17 setup-wizard-storage-defaults shows exe-local mode and default data path` |
| UC-SW-18 | Storage custom path validation | Custom storage mode blocks next with blank path and shows inline error | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-18 setup-wizard-storage-custom-path-validation blocks blank custom path` |
| UC-SW-19 | Storage persistence | Selected storage mode/path persists into setup config payload | implemented: `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts` `UC-SW-19 setup-wizard-storage-selection persists data and media dirs`; `internal/app/runtime_setup_api_test.go` `TestRuntimeSetupCompletePersistsSelectedStoragePath` |
