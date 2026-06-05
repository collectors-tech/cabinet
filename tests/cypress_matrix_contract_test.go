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
		`"-BaseUrl", "http://127.0.0.1:$lanePort"`,
		`data_dir = Join-Path $repoRoot ".tmp\cypress-runtime-$lanePort"`,
		`profile = "e2e-cypress-$lanePort"`,
		`instance_name = "cypress-$lanePort"`,
		`source_commit = $sourceCommit`,
		`active_lane_count = $activeLaneCount`,
		`empty_lane_count = $emptyLaneCount`,
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
		SpecCount        int   `json:"spec_count"`
		LaneCount        int   `json:"lane_count"`
		MaxWorkers       int   `json:"max_workers"`
		WorkerLimit      int   `json:"worker_limit"`
		ActiveLaneCount  int   `json:"active_lane_count"`
		EmptyLaneCount   int   `json:"empty_lane_count"`
		SpecCountsByLane []int `json:"spec_counts_by_lane"`
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
	expectedCounts := []int{1, 0, 0}
	for index, expected := range expectedCounts {
		if summary.SpecCountsByLane[index] != expected {
			t.Fatalf("lane %d expected %d specs, got %d", index+1, expected, summary.SpecCountsByLane[index])
		}
	}
}
