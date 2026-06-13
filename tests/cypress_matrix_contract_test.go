package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCypressMatrixRunnerProvidesIsolatedLanes(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "run-cypress-matrix.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress matrix runner: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		`[int]$BasePort = 17880`,
		`[int]$LaneCount = 2`,
		`[int]$MaxWorkers = 2`,
		`$workerLimit = [Math]::Min($LaneCount, $MaxWorkers)`,
		`$laneResults += Wait-MatrixJobSlot $jobs $workerLimit`,
		`[switch]$PlanOnly`,
		`[switch]$UseContainerImage`,
		`[string]$ContainerImage = "cabinet:e2e"`,
		`[int]$ContainerStartupTimeoutSec = 60`,
		`[switch]$KeepContainers`,
		`[ValidateSet("", "container_image", "container_start", "runtime_health", "cypress")]`,
		`[string]$FailureFixtureStage = ""`,
		`[int]$FailureFixtureLane = 1`,
		`[ValidateSet("", "pass")]`,
		`[string]$CypressFixtureMode = ""`,
		`"run", "-d"`,
		`"--name", $containerName`,
		`"-e", "CABINET_E2E_MODE=1"`,
		`"-p", "$($lanePort):17880"`,
		`"-v", "$($containerVolume):/data"`,
		`"--listen", "0.0.0.0:17880"`,
		`$args += "-ReuseServer"`,
		`$args += "-ApiContractSmoke"`,
		`api_contract_smoke = $ApiContractSmoke.IsPresent`,
		`api_contract_smoke = $apiContractSmoke`,
		`container_started = $containerStarted`,
		`failure_stage = $failureStage`,
		`error_message = $errorMessage`,
		`$failureStage = "container_start"`,
		`$failureStage = "runtime_health"`,
		`$failureStage = "cypress"`,
		`docker image inspect $ContainerImage`,
		`failure_stage = "container_image"`,
		`Container image preflight failed for $ContainerImage`,
		`$errorMessage = $_.Exception.Message`,
		`Cypress spec $spec failed with exit code $exitCode.`,
		`"-BaseUrl", "http://127.0.0.1:$lanePort"`,
		`data_dir = Join-Path $repoRoot ".tmp\cypress-runtime-$lanePort"`,
		`profile = "e2e-cypress-$lanePort"`,
		`instance_name = "cypress-$lanePort"`,
		`source_commit = $sourceCommit`,
		`use_container_image = $UseContainerImage.IsPresent`,
		`container_image = if ($UseContainerImage) { $ContainerImage } else { $null }`,
		`container_startup_timeout_sec = if ($UseContainerImage) { $ContainerStartupTimeoutSec } else { $null }`,
		`keep_containers = $KeepContainers.IsPresent`,
		`failure_fixture_stage = if (-not [string]::IsNullOrWhiteSpace($FailureFixtureStage)) { $FailureFixtureStage } else { $null }`,
		`failure_fixture_lane = if (-not [string]::IsNullOrWhiteSpace($FailureFixtureStage)) { $FailureFixtureLane } else { $null }`,
		`cypress_fixture_mode = if (-not [string]::IsNullOrWhiteSpace($CypressFixtureMode)) { $CypressFixtureMode } else { $null }`,
		`active_lane_count = $activeLaneCount`,
		`empty_lane_count = $emptyLaneCount`,
		`completed_spec_count = $null`,
		`passed_spec_count = $null`,
		`failed_spec_count = $null`,
		`passed_lane_count = $null`,
		`failed_lane_count = $null`,
		`completed_spec_count = $completedSpecCount`,
		`passed_spec_count = $passedSpecCount`,
		`failed_spec_count = $failedSpecCount`,
		`passed_lane_count = $passedLaneCount`,
		`failed_lane_count = $failedLaneCount`,
		`spec_counts_by_lane = $specCountsByLane`,
		`".work-agent\logs\cypress-matrix"`,
		`matrix.summary.json`,
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected matrix runner to contain fragment %q", fragment)
		}
	}
}

func TestCypressMatrixPlanSummaryExposesAPISmokeMetadata(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-api-smoke-plan-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-ApiContractSmoke",
		"-RequireE2EHooks",
		"-PlanOnly",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run API-smoke matrix plan: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read API-smoke matrix summary: %v", err)
	}

	var summary struct {
		ApiContractSmoke bool `json:"api_contract_smoke"`
		Lanes            []struct {
			ApiContractSmoke bool `json:"api_contract_smoke"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse API-smoke matrix summary: %v\n%s", err, raw)
	}

	if !summary.ApiContractSmoke {
		t.Fatalf("expected run-level API smoke metadata, got %+v", summary)
	}
	if len(summary.Lanes) != 2 {
		t.Fatalf("expected two lane plans, got %d", len(summary.Lanes))
	}
	for index, lane := range summary.Lanes {
		if !lane.ApiContractSmoke {
			t.Fatalf("lane %d missing API smoke metadata: %+v", index+1, lane)
		}
	}
}

func TestCypressMatrixSuccessFixtureWritesLiveMultiLaneSummary(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-success-fixture-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-*/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-CypressFixtureMode",
		"pass",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run matrix success fixture: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read matrix summary: %v", err)
	}

	var summary struct {
		ExitCode           int    `json:"exit_code"`
		PlanOnly           bool   `json:"plan_only"`
		CypressFixtureMode string `json:"cypress_fixture_mode"`
		LaneCount          int    `json:"lane_count"`
		ActiveLaneCount    int    `json:"active_lane_count"`
		EmptyLaneCount     int    `json:"empty_lane_count"`
		CompletedSpecCount int    `json:"completed_spec_count"`
		PassedSpecCount    int    `json:"passed_spec_count"`
		FailedSpecCount    int    `json:"failed_spec_count"`
		PassedLaneCount    int    `json:"passed_lane_count"`
		FailedLaneCount    int    `json:"failed_lane_count"`
		Lanes              []struct {
			Lane               int     `json:"lane"`
			Port               int     `json:"port"`
			DataDir            string  `json:"data_dir"`
			Profile            string  `json:"profile"`
			InstanceName       string  `json:"instance_name"`
			CypressFixtureMode string  `json:"cypress_fixture_mode"`
			FailureStage       *string `json:"failure_stage"`
			ErrorMessage       *string `json:"error_message"`
			Results            []struct {
				Spec               string `json:"spec"`
				BaseURL            string `json:"base_url"`
				CypressFixtureMode string `json:"cypress_fixture_mode"`
				CypressSummaryPath string `json:"cypress_summary_path"`
				CypressLogPath     string `json:"cypress_log_path"`
				ExitCode           int    `json:"exit_code"`
			} `json:"results"`
			ExitCode int `json:"exit_code"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse matrix summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 0 || summary.PlanOnly || summary.CypressFixtureMode != "pass" {
		t.Fatalf("unexpected success fixture run metadata: %+v", summary)
	}
	if summary.LaneCount != 2 || summary.ActiveLaneCount != 2 || summary.EmptyLaneCount != 0 {
		t.Fatalf("expected two active live lanes, got %+v", summary)
	}
	if summary.PassedLaneCount != 2 || summary.FailedLaneCount != 0 {
		t.Fatalf("expected two passing live lanes, got %+v", summary)
	}
	if summary.CompletedSpecCount == 0 || summary.PassedSpecCount != summary.CompletedSpecCount || summary.FailedSpecCount != 0 {
		t.Fatalf("expected all completed fixture specs to pass, got %+v", summary)
	}
	if len(summary.Lanes) != 2 {
		t.Fatalf("expected two completed lane summaries, got %d in %s", len(summary.Lanes), raw)
	}
	for _, lane := range summary.Lanes {
		if lane.ExitCode != 0 || lane.FailureStage != nil || lane.ErrorMessage != nil {
			t.Fatalf("expected passing lane summary, got %+v", lane)
		}
		if lane.Port == 0 || lane.DataDir == "" || lane.Profile == "" || lane.InstanceName == "" {
			t.Fatalf("expected lane isolation metadata, got %+v", lane)
		}
		if lane.CypressFixtureMode != "pass" {
			t.Fatalf("expected lane cypress fixture metadata, got %+v", lane)
		}
		if len(lane.Results) == 0 {
			t.Fatalf("expected fixture result entries for lane %+v", lane)
		}
		seenSummaryPaths := map[string]bool{}
		for _, result := range lane.Results {
			if result.ExitCode != 0 || result.CypressFixtureMode != "pass" || result.BaseURL == "" || result.Spec == "" {
				t.Fatalf("unexpected fixture result: %+v", result)
			}
			if result.CypressSummaryPath == "" || result.CypressLogPath == "" {
				t.Fatalf("fixture result should link artifacts: %+v", result)
			}
			if seenSummaryPaths[result.CypressSummaryPath] {
				t.Fatalf("fixture result summary path reused within lane: %s", result.CypressSummaryPath)
			}
			seenSummaryPaths[result.CypressSummaryPath] = true
			artifactRaw, err := os.ReadFile(result.CypressSummaryPath)
			if err != nil {
				t.Fatalf("read Cypress fixture summary: %v", err)
			}
			var artifact struct {
				Spec    string `json:"spec"`
				LogPath string `json:"log_path"`
			}
			if err := json.Unmarshal(artifactRaw, &artifact); err != nil {
				t.Fatalf("parse Cypress fixture summary: %v\n%s", err, artifactRaw)
			}
			if artifact.Spec != result.Spec || artifact.LogPath != result.CypressLogPath {
				t.Fatalf("fixture artifact does not match result: artifact=%+v result=%+v", artifact, result)
			}
		}
	}
}

func TestCypressMatrixApiSmokeSuccessFixtureWritesLiveResultMetadata(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-api-smoke-live-fixture-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-*/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-ApiContractSmoke",
		"-CypressFixtureMode",
		"pass",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run API-smoke success fixture: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read API-smoke success summary: %v", err)
	}

	var summary struct {
		ExitCode         int  `json:"exit_code"`
		PlanOnly         bool `json:"plan_only"`
		ApiContractSmoke bool `json:"api_contract_smoke"`
		Lanes            []struct {
			ApiContractSmoke bool `json:"api_contract_smoke"`
			Results          []struct {
				Spec             string `json:"spec"`
				ApiContractSmoke bool   `json:"api_contract_smoke"`
				ExitCode         int    `json:"exit_code"`
			} `json:"results"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse API-smoke success summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 0 || summary.PlanOnly || !summary.ApiContractSmoke {
		t.Fatalf("unexpected API-smoke success run metadata: %+v", summary)
	}
	if len(summary.Lanes) != 2 {
		t.Fatalf("expected two live lane summaries, got %d in %s", len(summary.Lanes), raw)
	}
	for laneIndex, lane := range summary.Lanes {
		if !lane.ApiContractSmoke {
			t.Fatalf("lane %d missing API smoke metadata: %+v", laneIndex+1, lane)
		}
		if len(lane.Results) == 0 {
			t.Fatalf("lane %d missing per-spec result entries: %+v", laneIndex+1, lane)
		}
		for _, result := range lane.Results {
			if result.Spec == "" || result.ExitCode != 0 || !result.ApiContractSmoke {
				t.Fatalf("lane %d missing per-spec API smoke result metadata: %+v", laneIndex+1, result)
			}
		}
	}
}

func TestCypressMatrixMixedLaneFixtureSummarizesOutcomeCounts(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-mixed-lane-outcome-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-*/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-CypressFixtureMode",
		"pass",
		"-FailureFixtureStage",
		"cypress",
		"-FailureFixtureLane",
		"2",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected mixed lane fixture to exit nonzero\n%s", output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read mixed lane summary: %v\n%s", err, output)
	}

	var summary struct {
		ExitCode           int `json:"exit_code"`
		ActiveLaneCount    int `json:"active_lane_count"`
		CompletedSpecCount int `json:"completed_spec_count"`
		PassedSpecCount    int `json:"passed_spec_count"`
		FailedSpecCount    int `json:"failed_spec_count"`
		PassedLaneCount    int `json:"passed_lane_count"`
		FailedLaneCount    int `json:"failed_lane_count"`
		Lanes              []struct {
			Lane         int     `json:"lane"`
			ExitCode     int     `json:"exit_code"`
			FailureStage *string `json:"failure_stage"`
			ErrorMessage *string `json:"error_message"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse mixed lane summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 1 || summary.ActiveLaneCount != 2 || summary.PassedLaneCount != 1 || summary.FailedLaneCount != 1 {
		t.Fatalf("unexpected mixed lane outcome counts: %+v", summary)
	}
	if summary.CompletedSpecCount == 0 || summary.PassedSpecCount == 0 || summary.FailedSpecCount != 1 {
		t.Fatalf("unexpected mixed spec outcome counts: %+v", summary)
	}
	if len(summary.Lanes) != 2 {
		t.Fatalf("expected two completed lane summaries, got %d in %s", len(summary.Lanes), raw)
	}
	var sawPassingLane, sawFailingLane bool
	for _, lane := range summary.Lanes {
		switch lane.ExitCode {
		case 0:
			sawPassingLane = true
			if lane.FailureStage != nil || lane.ErrorMessage != nil {
				t.Fatalf("passing lane should not record failure diagnostics: %+v", lane)
			}
		case 1:
			sawFailingLane = true
			if lane.FailureStage == nil || *lane.FailureStage != "cypress" {
				t.Fatalf("failing lane should record cypress failure stage: %+v", lane)
			}
			if lane.ErrorMessage == nil || !strings.Contains(*lane.ErrorMessage, "forced Cypress failure") {
				t.Fatalf("failing lane should record fixture diagnostic: %+v", lane)
			}
		default:
			t.Fatalf("unexpected lane exit code: %+v", lane)
		}
	}
	if !sawPassingLane || !sawFailingLane {
		t.Fatalf("expected one passing and one failing lane, got %+v", summary.Lanes)
	}
}

func TestCypressMatrixSuccessFixtureLinksPerSpecArtifacts(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-result-artifact-links-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-ApiContractSmoke",
		"-CypressFixtureMode",
		"pass",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run artifact-link fixture: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read artifact-link matrix summary: %v", err)
	}

	var summary struct {
		ExitCode int `json:"exit_code"`
		Lanes    []struct {
			Results []struct {
				Spec               string `json:"spec"`
				CypressSummaryPath string `json:"cypress_summary_path"`
				CypressLogPath     string `json:"cypress_log_path"`
				ExitCode           int    `json:"exit_code"`
			} `json:"results"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse artifact-link matrix summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 0 {
		t.Fatalf("expected passing artifact-link summary, got %+v", summary)
	}
	var linkedResults int
	for laneIndex, lane := range summary.Lanes {
		for _, result := range lane.Results {
			linkedResults++
			if result.Spec == "" || result.ExitCode != 0 {
				t.Fatalf("lane %d has unexpected result: %+v", laneIndex+1, result)
			}
			if result.CypressSummaryPath == "" || result.CypressLogPath == "" {
				t.Fatalf("lane %d result missing Cypress artifact links: %+v", laneIndex+1, result)
			}
			if _, err := os.Stat(result.CypressSummaryPath); err != nil {
				t.Fatalf("lane %d summary path is not readable: %v", laneIndex+1, err)
			}
			if _, err := os.Stat(result.CypressLogPath); err != nil {
				t.Fatalf("lane %d log path is not readable: %v", laneIndex+1, err)
			}
			artifactRaw, err := os.ReadFile(result.CypressSummaryPath)
			if err != nil {
				t.Fatalf("read Cypress fixture summary: %v", err)
			}
			var artifact struct {
				ExitCode int    `json:"exit_code"`
				Spec     string `json:"spec"`
				LogPath  string `json:"log_path"`
			}
			if err := json.Unmarshal(artifactRaw, &artifact); err != nil {
				t.Fatalf("parse Cypress fixture summary: %v\n%s", err, artifactRaw)
			}
			if artifact.ExitCode != 0 || artifact.Spec != result.Spec || artifact.LogPath != result.CypressLogPath {
				t.Fatalf("Cypress fixture summary does not match matrix result: artifact=%+v result=%+v", artifact, result)
			}
		}
	}
	if linkedResults != 1 {
		t.Fatalf("expected one linked result for the one-spec fixture, got %d", linkedResults)
	}
}

func TestCypressMatrixFailureFixturesWriteLaneDiagnostics(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		name             string
		stage            string
		lane             string
		useContainer     bool
		wantStarted      bool
		wantResult       bool
		wantMessagePiece string
	}{
		{
			name:             "container-start",
			stage:            "container_start",
			lane:             "1",
			useContainer:     true,
			wantStarted:      false,
			wantResult:       false,
			wantMessagePiece: "forced container_start failure",
		},
		{
			name:             "runtime-health",
			stage:            "runtime_health",
			lane:             "1",
			useContainer:     true,
			wantStarted:      false,
			wantResult:       false,
			wantMessagePiece: "forced runtime_health failure",
		},
		{
			name:             "cypress",
			stage:            "cypress",
			lane:             "1",
			useContainer:     false,
			wantStarted:      false,
			wantResult:       true,
			wantMessagePiece: "forced Cypress failure",
		},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			logRoot := t.TempDir()
			runID := "matrix-failure-fixture-" + fixture.name
			args := []string{
				"-NoLogo",
				"-NoProfile",
				"-File",
				filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
				"-SpecGlob",
				"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
				"-LaneCount",
				"2",
				"-MaxWorkers",
				"2",
				"-FailureFixtureStage",
				fixture.stage,
				"-FailureFixtureLane",
				fixture.lane,
				"-RunId",
				runID,
				"-LogRoot",
				logRoot,
			}
			if fixture.useContainer {
				args = append(args, "-UseContainerImage")
			}
			cmd := exec.Command("pwsh", args...)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected failure fixture %s to exit nonzero\n%s", fixture.name, output)
			}

			summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
			raw, readErr := os.ReadFile(summaryPath)
			if readErr != nil {
				t.Fatalf("read failure fixture summary: %v\n%s", readErr, output)
			}

			var summary struct {
				ExitCode            int    `json:"exit_code"`
				FailureFixtureStage string `json:"failure_fixture_stage"`
				FailureFixtureLane  int    `json:"failure_fixture_lane"`
				Lanes               []struct {
					Lane             int     `json:"lane"`
					ExitCode         int     `json:"exit_code"`
					ContainerStarted bool    `json:"container_started"`
					FailureStage     *string `json:"failure_stage"`
					ErrorMessage     *string `json:"error_message"`
					Results          []struct {
						Spec     string `json:"spec"`
						ExitCode int    `json:"exit_code"`
					} `json:"results"`
				} `json:"lanes"`
			}
			if err := json.Unmarshal(raw, &summary); err != nil {
				t.Fatalf("parse failure fixture summary: %v\n%s", err, raw)
			}

			if summary.ExitCode != 1 {
				t.Fatalf("expected summary exit_code=1, got %+v", summary)
			}
			if summary.FailureFixtureStage != fixture.stage || summary.FailureFixtureLane != 1 {
				t.Fatalf("unexpected fixture metadata: %+v", summary)
			}
			if len(summary.Lanes) != 1 {
				t.Fatalf("expected only active lane result, got %d lanes in %s", len(summary.Lanes), raw)
			}
			lane := summary.Lanes[0]
			if lane.Lane != 1 || lane.ExitCode != 1 {
				t.Fatalf("unexpected failed lane summary: %+v", lane)
			}
			if lane.ContainerStarted != fixture.wantStarted {
				t.Fatalf("unexpected container_started for %s: %+v", fixture.name, lane)
			}
			if lane.FailureStage == nil || *lane.FailureStage != fixture.stage {
				t.Fatalf("expected failure_stage %q, got %+v", fixture.stage, lane)
			}
			if lane.ErrorMessage == nil || !strings.Contains(*lane.ErrorMessage, fixture.wantMessagePiece) {
				t.Fatalf("expected diagnostic message containing %q, got %+v", fixture.wantMessagePiece, lane)
			}
			if fixture.wantResult {
				if len(lane.Results) != 1 || lane.Results[0].ExitCode != 1 {
					t.Fatalf("expected Cypress fixture to record failed spec result, got %+v", lane.Results)
				}
			} else if len(lane.Results) != 0 {
				t.Fatalf("expected setup/runtime fixture to avoid spec results, got %+v", lane.Results)
			}
		})
	}
}

func TestCypressMatrixContainerImagePreflightFailsBeforeLaneFanout(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-container-image-preflight-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-UseContainerImage",
		"-ContainerImage",
		"cabinet:e2e",
		"-FailureFixtureStage",
		"container_image",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected container image preflight fixture to exit nonzero\n%s", output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read container image preflight summary: %v\n%s", err, output)
	}

	var summary struct {
		ExitCode            int    `json:"exit_code"`
		PlanOnly            bool   `json:"plan_only"`
		UseContainerImage   bool   `json:"use_container_image"`
		FailureFixtureStage string `json:"failure_fixture_stage"`
		ActiveLaneCount     int    `json:"active_lane_count"`
		PassedLaneCount     int    `json:"passed_lane_count"`
		FailedLaneCount     int    `json:"failed_lane_count"`
		Lanes               []struct {
			Lane             int      `json:"lane"`
			ContainerStarted bool     `json:"container_started"`
			FailureStage     string   `json:"failure_stage"`
			ErrorMessage     string   `json:"error_message"`
			Results          []string `json:"results"`
			ExitCode         int      `json:"exit_code"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse container image preflight summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 1 || summary.PlanOnly || !summary.UseContainerImage || summary.FailureFixtureStage != "container_image" {
		t.Fatalf("unexpected run-level preflight metadata: %+v", summary)
	}
	if summary.ActiveLaneCount != 1 || summary.PassedLaneCount != 0 || summary.FailedLaneCount != 1 {
		t.Fatalf("unexpected lane outcome counts for image preflight: %+v", summary)
	}
	if len(summary.Lanes) != 1 {
		t.Fatalf("expected only active lane failure summary, got %d in %s", len(summary.Lanes), raw)
	}
	lane := summary.Lanes[0]
	if lane.Lane != 1 || lane.ExitCode != 1 || lane.ContainerStarted {
		t.Fatalf("expected failed lane without container start: %+v", lane)
	}
	if lane.FailureStage != "container_image" || !strings.Contains(lane.ErrorMessage, "forced container_image failure") {
		t.Fatalf("expected container_image diagnostic, got %+v", lane)
	}
	if len(lane.Results) != 0 {
		t.Fatalf("image preflight should fail before Cypress result fanout, got %+v", lane.Results)
	}
}

func TestCypressMatrixContainerImagePreflightReportsMissingDockerCLI(t *testing.T) {
	t.Parallel()

	pwshPath, err := exec.LookPath("pwsh")
	if err != nil {
		t.Fatalf("resolve pwsh: %v", err)
	}

	logRoot := t.TempDir()
	runID := "matrix-docker-cli-missing-contract"
	emptyPath := t.TempDir()
	cmd := exec.Command(
		pwshPath,
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-UseContainerImage",
		"-ContainerImage",
		"cabinet:e2e",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	cmd.Env = append(os.Environ(), "PATH="+emptyPath, "Path="+emptyPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected missing Docker CLI preflight to exit nonzero\n%s", output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read missing Docker CLI preflight summary: %v\n%s", err, output)
	}

	var summary struct {
		ExitCode          int  `json:"exit_code"`
		PlanOnly          bool `json:"plan_only"`
		UseContainerImage bool `json:"use_container_image"`
		ActiveLaneCount   int  `json:"active_lane_count"`
		PassedLaneCount   int  `json:"passed_lane_count"`
		FailedLaneCount   int  `json:"failed_lane_count"`
		Lanes             []struct {
			Lane             int      `json:"lane"`
			ContainerStarted bool     `json:"container_started"`
			FailureStage     string   `json:"failure_stage"`
			ErrorMessage     string   `json:"error_message"`
			Results          []string `json:"results"`
			ExitCode         int      `json:"exit_code"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse missing Docker CLI preflight summary: %v\n%s", err, raw)
	}

	if summary.ExitCode != 1 || summary.PlanOnly || !summary.UseContainerImage {
		t.Fatalf("unexpected run-level missing Docker CLI metadata: %+v", summary)
	}
	if summary.ActiveLaneCount != 1 || summary.PassedLaneCount != 0 || summary.FailedLaneCount != 1 {
		t.Fatalf("unexpected lane outcome counts for missing Docker CLI: %+v", summary)
	}
	if len(summary.Lanes) != 1 {
		t.Fatalf("expected only active lane failure summary, got %d in %s", len(summary.Lanes), raw)
	}
	lane := summary.Lanes[0]
	if lane.Lane != 1 || lane.ExitCode != 1 || lane.ContainerStarted {
		t.Fatalf("expected failed lane without container start: %+v", lane)
	}
	if lane.FailureStage != "container_image" || !strings.Contains(lane.ErrorMessage, "Docker CLI is unavailable") {
		t.Fatalf("expected missing Docker CLI diagnostic, got %+v", lane)
	}
	if len(lane.Results) != 0 {
		t.Fatalf("missing Docker CLI preflight should fail before Cypress result fanout, got %+v", lane.Results)
	}
}

func TestPackageJsonExposesCypressMatrixScript(t *testing.T) {
	t.Parallel()

	packagePath := filepath.Join("..", "ui.web", "package.json")
	raw, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("read ui package: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, `"e2e:matrix": "pwsh -NoLogo -NoProfile -File ../scripts/run-cypress-matrix.ps1"`) {
		t.Fatalf("expected ui package to expose e2e:matrix script")
	}
}

func TestCypressMatrixPlanSummaryExposesLaneCounts(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-plan-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"3",
		"-MaxWorkers",
		"5",
		"-PlanOnly",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run matrix plan: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read matrix summary: %v", err)
	}

	var summary struct {
		SpecCount         int   `json:"spec_count"`
		LaneCount         int   `json:"lane_count"`
		MaxWorkers        int   `json:"max_workers"`
		WorkerLimit       int   `json:"worker_limit"`
		ActiveLaneCount   int   `json:"active_lane_count"`
		EmptyLaneCount    int   `json:"empty_lane_count"`
		SpecCountsByLane  []int `json:"spec_counts_by_lane"`
		UseContainerImage bool  `json:"use_container_image"`
		ContainerImage    any   `json:"container_image"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse matrix summary: %v\n%s", err, raw)
	}

	if summary.SpecCount != 1 {
		t.Fatalf("expected one matched spec, got %d", summary.SpecCount)
	}
	if summary.LaneCount != 3 || summary.MaxWorkers != 5 || summary.WorkerLimit != 3 {
		t.Fatalf("unexpected lane worker bounds: %+v", summary)
	}
	if summary.ActiveLaneCount != 1 || summary.EmptyLaneCount != 2 {
		t.Fatalf("unexpected active/empty lane counts: %+v", summary)
	}
	if len(summary.SpecCountsByLane) != 3 {
		t.Fatalf("expected spec count for each lane, got %+v", summary.SpecCountsByLane)
	}
	if summary.UseContainerImage {
		t.Fatalf("plan without -UseContainerImage should record use_container_image=false")
	}
	if summary.ContainerImage != nil {
		t.Fatalf("plan without -UseContainerImage should not record a container image, got %#v", summary.ContainerImage)
	}
	expectedCounts := []int{1, 0, 0}
	for index, expected := range expectedCounts {
		if summary.SpecCountsByLane[index] != expected {
			t.Fatalf("lane %d expected %d specs, got %d", index+1, expected, summary.SpecCountsByLane[index])
		}
	}
}

func TestCypressMatrixPlanSummaryExposesContainerLaneMetadata(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	runID := "matrix-container-plan-contract"
	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File",
		filepath.Join("..", "scripts", "run-cypress-matrix.ps1"),
		"-SpecGlob",
		"ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts",
		"-LaneCount",
		"2",
		"-MaxWorkers",
		"2",
		"-UseContainerImage",
		"-ContainerImage",
		"cabinet:e2e",
		"-PlanOnly",
		"-RunId",
		runID,
		"-LogRoot",
		logRoot,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run container matrix plan: %v\n%s", err, output)
	}

	summaryPath := filepath.Join(logRoot, runID, "matrix.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read container matrix summary: %v", err)
	}

	var summary struct {
		UseContainerImage       bool   `json:"use_container_image"`
		ContainerImage          string `json:"container_image"`
		ContainerStartupTimeout int    `json:"container_startup_timeout_sec"`
		KeepContainers          bool   `json:"keep_containers"`
		Lanes                   []struct {
			UseContainerImage bool    `json:"use_container_image"`
			ContainerImage    string  `json:"container_image"`
			ContainerName     string  `json:"container_name"`
			ContainerVolume   string  `json:"container_volume"`
			FailureStage      *string `json:"failure_stage"`
			ErrorMessage      *string `json:"error_message"`
		} `json:"lanes"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse container matrix summary: %v\n%s", err, raw)
	}

	if !summary.UseContainerImage || summary.ContainerImage != "cabinet:e2e" {
		t.Fatalf("expected container matrix metadata, got %+v", summary)
	}
	if summary.ContainerStartupTimeout != 60 || summary.KeepContainers {
		t.Fatalf("unexpected container lifecycle defaults: %+v", summary)
	}
	if len(summary.Lanes) != 2 {
		t.Fatalf("expected two lane plans, got %d", len(summary.Lanes))
	}
	for index, lane := range summary.Lanes {
		if !lane.UseContainerImage || lane.ContainerImage != "cabinet:e2e" {
			t.Fatalf("lane %d missing container image metadata: %+v", index+1, lane)
		}
		if !strings.Contains(lane.ContainerName, "matrix-container-plan-contract-lane-") {
			t.Fatalf("lane %d has unexpected container name %q", index+1, lane.ContainerName)
		}
		if lane.ContainerVolume != lane.ContainerName+"-data" {
			t.Fatalf("lane %d has unexpected container volume %q for name %q", index+1, lane.ContainerVolume, lane.ContainerName)
		}
		if lane.FailureStage != nil || lane.ErrorMessage != nil {
			t.Fatalf("lane %d plan should not report a failure: %+v", index+1, lane)
		}
	}
}
