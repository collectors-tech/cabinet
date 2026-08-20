## Purpose
Define continuous, scheduled, executable UI validation for Cabinet including release-aware runtime checks and exhaustive user-action verification.

## Requirements
### Requirement CONT-UI-CAB-001: Cabinet hourly validation SHALL be release-aware
Hourly validation run SHALL compare active build/version against latest available build and skip full run when unchanged.

#### Scenario: No new build available
- **GIVEN** hourly validation starts
- **WHEN** latest build/version matches current validated build
- **THEN** runner SHALL record `no-change` status and wait for next schedule

### Requirement CONT-UI-CAB-002: Cabinet hourly validation SHALL execute real browser interactions
Validation run MUST open live app and execute real user actions (no synthetic-only claims).

#### Scenario: Exhaustive action execution
- **GIVEN** a runnable Cabinet instance is available
- **WHEN** validation run executes
- **THEN** every user-clickable action on audited screens MUST be interacted with and recorded pass/fail

### Requirement CONT-UI-CAB-003: Validation failures SHALL create/update issues with reproducible evidence
Failed actions MUST be linked to focused issues with expected vs actual behavior and reproduction details.

#### Scenario: Action failure recorded
- **GIVEN** any action fails during run
- **WHEN** run completes
- **THEN** issue(s) SHALL include screen, action, error text, and evidence path

### Requirement CONT-UI-CAB-004: Scheduled validation SHALL maintain OpenSpec + commit traceability
Validation outputs MUST update OpenSpec/traceability when contracts are missing and commit changes with issue-linked messages.

#### Scenario: Spec gap discovered
- **GIVEN** action exists without explicit requirement coverage
- **WHEN** validation run reconciles action-to-spec mapping
- **THEN** spec IDs SHALL be added append-only and committed with linked issue reference

### Requirement CONT-UI-CAB-005: Validation SHALL verify control intent outcomes, not click-only execution
For each interactive control, validation SHALL assert intended outcome (route/state/data/error feedback), not merely that click action executed.

#### Scenario: Intent verification per control
- **GIVEN** an interactive control is discovered in audit inventory
- **WHEN** control is executed
- **THEN** run output MUST record expected vs actual intent outcome and pass/fail status

### Requirement CONT-UI-CAB-006: Validation SHALL include full form-field behavior contract checks
Validation SHALL enumerate and test form fields (required/optional, valid/invalid input handling, error messages, submit/save behavior, keyboard accessibility).

#### Scenario: Form field validation contract
- **GIVEN** a form is present on audited screen
- **WHEN** field interactions and submissions are executed
- **THEN** run output MUST include field-level validation outcomes and any failures/backlog issue links

### Requirement CONT-UI-CAB-007: Validator exploration runtime SHALL preflight required Node dependencies
Before running exploratory UI scripts, validator runtime SHALL ensure required Node dependencies (including `playwright`) are installed and importable.

#### Scenario: Exploration preflight checks dependencies
- **GIVEN** validator is about to run `node scripts/cabinet_ui_cycle.mjs`
- **WHEN** dependency preflight executes
- **THEN** missing dependencies SHALL be installed or reported with actionable remediation prior to exploration execution

### Requirement CONT-UI-CAB-008: Cypress validation SHALL have a project-local container image path
Cabinet SHALL provide a repo-local container image definition for isolated Cypress runtime lanes so browser validation can start from a known build artifact instead of a stale shared desktop runtime.

#### Scenario: Build image from source
- **GIVEN** the Cabinet repository is checked out at a known commit
- **WHEN** the container image is built from the repository root
- **THEN** the image build MUST install UI dependencies, include UI raw-content inputs needed by the bundle, build the static UI bundle, compile the Go Cabinet runtime, and copy the UI bundle into the Go build context before producing the runtime image.

#### Scenario: Bound image build context
- **GIVEN** the Cabinet repository contains local runtime data, generated binaries, dependency folders, and work-agent logs
- **WHEN** the container image build context is prepared
- **THEN** `.dockerignore` MUST exclude those local artifacts while preserving source files and raw-content docs required for a reproducible image build.

#### Scenario: Run isolated validation runtime
- **GIVEN** a Cypress lane starts the Cabinet container image
- **WHEN** the container runtime launches
- **THEN** it MUST disable browser auto-open, listen on the container network interface at a deterministic HTTP port, use a writable mounted `/data` directory, set an E2E profile and instance name, and allow parallel runtime execution for lane isolation.

#### Scenario: Report runtime health to lane orchestration
- **GIVEN** a Cypress lane starts the Cabinet container image
- **WHEN** the container runtime is waiting for readiness
- **THEN** the image MUST expose a health check against `/healthz` so orchestration can distinguish app readiness from Cypress assertion failures.

### Requirement CONT-UI-CAB-009: Cypress matrix validation SHALL orchestrate isolated container lanes
Cabinet SHALL allow the bounded Cypress matrix runner to start one repo-local container runtime per active lane so matrix execution can validate against isolated data volumes and deterministic host ports instead of shared desktop runtimes.

#### Scenario: Plan container-backed lanes
- **GIVEN** the Cypress matrix runner is invoked in plan-only mode with container image lanes enabled
- **WHEN** it writes `matrix.summary.json`
- **THEN** the summary MUST record the container image, startup timeout, keep-container setting, and per-lane container name and data volume alongside the lane port, profile, instance, source commit, and assigned specs.
- **AND** the summary MUST record a run-level `runner_command` array so validators can confirm the exact matrix invocation, selected spec glob, lane bounds, container mode, API smoke, E2E hook, fixture, and run-id inputs without scraping console logs.

#### Scenario: Execute container-backed lanes
- **GIVEN** the Cypress matrix runner executes with container image lanes enabled
- **WHEN** an active lane starts
- **THEN** the runner MUST start the configured Cabinet container image with a unique host port, named writable data volume, E2E hooks enabled, E2E profile, instance name, and parallel-runtime flag before invoking Cypress against that lane.
- **AND** the runner MUST reuse the already-started lane runtime for Cypress, wait for `/healthz` before assertions, and stop/remove the lane container and volume unless keep-container diagnostics are explicitly requested.
- **AND** `matrix.summary.json` MUST record a machine-readable container command for each container-backed lane so validators can confirm the host port mapping, writable data volume, E2E hook flag, listen address, profile, instance name, and parallel-runtime flag without scraping console logs.

#### Scenario: Preflight container image availability before lane fanout
- **GIVEN** the Cypress matrix runner executes with container image lanes enabled
- **WHEN** the configured container image is unavailable before any lane starts
- **THEN** the runner MUST fail before lane fanout and write `matrix.summary.json` with `failure_stage=container_image`, a diagnostic `error_message`, `container_started=false`, no per-spec Cypress results, and aggregate failed lane counts for the active lane set.
- **AND** if the Docker CLI is unavailable, the same preflight MUST write the machine-readable summary instead of terminating before evidence is recorded.

#### Scenario: Preflight host port availability before lane fanout
- **GIVEN** the Cypress matrix runner executes with container image lanes enabled
- **WHEN** an active lane's host port is already accepting TCP connections before any lane starts
- **THEN** the runner MUST fail before lane fanout and write `matrix.summary.json` with `failure_stage=port_preflight`, a diagnostic `error_message`, `container_started=false`, no per-spec Cypress results, and aggregate failed lane counts for the active lane set.

#### Scenario: Preflight lane runtime contracts before browser assertions
- **GIVEN** the Cypress matrix runner is invoked with API contract smoke enabled
- **WHEN** the runner plans or executes isolated lanes
- **THEN** `matrix.summary.json` MUST record `api_contract_smoke` at the run and lane levels.
- **AND** each active lane MUST pass `-ApiContractSmoke` through to `cypress.ps1` so `/healthz`, `/api/runtime`, `/api/openapi.yaml`, `/sign-in`, and required E2E hook checks can fail before browser assertions.
- **AND** container-backed lanes MUST allow the Cabinet runtime's internal container port in `/api/runtime` while still validating the externally mapped host `BaseUrl` and preserving the allowed-port metadata in the API smoke summary.
- **AND** the repo-local runtime image MUST include the OpenAPI YAML artifact needed by `/api/openapi.yaml` so API smoke failures distinguish stale or incomplete images from browser assertion failures.

#### Scenario: Preserve preflight metadata on live per-spec results
- **GIVEN** the Cypress matrix runner executes non-plan lane work with API contract smoke enabled
- **WHEN** it writes per-spec result entries to `matrix.summary.json`
- **THEN** each result entry MUST record `api_contract_smoke=true` so downstream triage can distinguish runs that performed API preflight checks from browser-only assertions.

#### Scenario: Link matrix results to Cypress artifacts
- **GIVEN** the Cypress matrix runner executes non-plan lane work
- **WHEN** it writes per-spec result entries to `matrix.summary.json`
- **THEN** each Cypress-executed or deterministic fixture result entry MUST record readable `cypress_summary_path` and `cypress_log_path` values so downstream validation can inspect the exact per-spec evidence without scraping lane stdout.

#### Scenario: Report lane failures in machine-readable summary
- **GIVEN** a Cypress matrix lane fails during container cleanup, container start, runtime health, or Cypress execution
- **WHEN** the runner writes `matrix.summary.json`
- **THEN** the failed lane MUST record a nonzero exit code plus `failure_stage` and `error_message` fields so operators can distinguish setup/runtime failures from browser assertion failures.

#### Scenario: Exercise lane failure diagnostics deterministically
- **GIVEN** the Cypress matrix runner is invoked with an explicit failure fixture for `port_preflight`, `container_start`, `runtime_health`, or `cypress`
- **WHEN** the selected lane reaches the requested fixture stage
- **THEN** the runner MUST fail that lane, write `failure_fixture_stage` and `failure_fixture_lane` at the run level, and preserve the selected lane's `failure_stage`, `error_message`, and nonzero `exit_code` in `matrix.summary.json`.

#### Scenario: Exercise successful multi-lane orchestration deterministically
- **GIVEN** the Cypress matrix runner is invoked with a success fixture mode across multiple active lanes
- **WHEN** it writes a non-plan `matrix.summary.json`
- **THEN** the summary MUST record `cypress_fixture_mode`, all active lane summaries, each lane's unique port/data/profile/instance metadata, and passing per-spec result entries without requiring a shared desktop runtime.

#### Scenario: Summarize multi-lane pass and failure outcomes
- **GIVEN** the Cypress matrix runner executes non-plan work across multiple active lanes
- **WHEN** it writes `matrix.summary.json`
- **THEN** the summary MUST record aggregate `passed_lane_count` and `failed_lane_count` values matching the completed lane exit codes so validators can classify mixed lane outcomes without parsing every lane entry.
- **AND** the summary MUST record aggregate `completed_spec_count`, `passed_spec_count`, and `failed_spec_count` values matching the completed per-spec result entries so validators can classify multi-spec outcomes without parsing every lane entry.

#### Scenario: Record run, lane, and per-spec timing metadata
- **GIVEN** the Cypress matrix runner executes non-plan work
- **WHEN** it writes `matrix.summary.json`
- **THEN** the run summary, each completed lane summary, and each per-spec result entry MUST record `started_at`, `finished_at`, and `duration_ms` values so operators can identify slow setup, runtime health, and browser assertion phases without scraping logs.

### Requirement CONT-UI-CAB-010: Isolated Cypress harness documentation SHALL define local prerequisites and fallback behavior
Cabinet SHALL document how operators run isolated Cypress validation locally, including Docker/container prerequisites, single-spec and bounded-matrix commands, machine-readable evidence paths, failure-stage interpretation, and fallback behavior when Docker is unavailable.

#### Scenario: Operator runs or triages an isolated Cypress harness command
- **GIVEN** an operator needs to run a Cabinet Cypress spec or bounded matrix against isolated runtime lanes
- **WHEN** they read the harness documentation
- **THEN** the documentation MUST name the local prerequisites, the `cabinet:e2e` image build command, one-spec and bounded-matrix runner commands, API smoke and E2E hook switches, `matrix.summary.json` evidence path, failure stages, and the rule that stale shared desktop runtimes are not valid isolated-lane proof.

### Requirement CONT-UI-CAB-011: Single-spec Cypress validation SHALL emit machine-readable invocation evidence
Cabinet SHALL make each `cypress.ps1` run summary sufficient to prove the exact one-spec invocation and runtime timing without scraping transcript logs.

#### Scenario: Record single-spec runner command and timing evidence
- **GIVEN** `cypress.ps1` runs a focused Cypress spec against an isolated or fallback runtime
- **WHEN** it writes the per-run Cypress summary JSON
- **THEN** the summary MUST record `runner_command`, `started_at`, `finished_at`, and `duration_ms` alongside the existing spec, base URL, runtime port, data directory, profile, instance name, executable path, source commit, and log path evidence.

### Requirement CONT-UI-CAB-012: Hourly validation SHALL reject stale runtime baselines before browser assertions
Cabinet hourly UI validation SHALL run each selected Cypress spec through the project Cypress wrapper with API contract smoke and E2E hook preflight enabled so stale or non-test runtimes are reported as setup/freshness failures instead of broad product regressions.

#### Scenario: Scheduled spec invocation records runtime freshness preflight
- **GIVEN** hourly validation selects a Cypress spec for scheduled execution
- **WHEN** it invokes `cypress.ps1`
- **THEN** the invocation MUST include `-RequireE2EHooks` and `-ApiContractSmoke`
- **AND** the hourly report MUST record `api_contract_smoke`, `require_e2e_hooks`, and `allow_stale_runtime_version` for the spec result.

### Requirement CONT-UI-CAB-013: Hourly validation SHALL own a bounded reusable runtime lifecycle
Cabinet hourly UI validation SHALL start one workflow-built exact runtime for the selected spec set, reuse it without rebuilding dependencies or runtime assets, run Cypress with zero retries and a bounded per-spec watchdog, and stop only that owned runtime after report generation.

#### Scenario: Separate product failures from runner exhaustion
- **GIVEN** an hourly run executes multiple selected specs against one exact runtime revision
- **WHEN** a completed Cypress spec reports assertion failures
- **THEN** the report SHALL preserve that result as a product failure and continue with the next selected spec.
- **AND** when a timeout, missing summary, stale revision, or other runner failure occurs, the run SHALL stop before launching another spec and record the runner phase, timeout state, revision, exit code, and evidence path.

#### Scenario: Skip duplicate same-revision validation
- **GIVEN** scheduled Windows runners are ephemeral
- **WHEN** a validated revision is scheduled again
- **THEN** the workflow SHALL restore the persisted revision state and record `no-change` without starting Cabinet or Cypress.
- **AND** if state restoration is unavailable, failure handling SHALL avoid creating another open hourly issue whose body already records the same commit.
