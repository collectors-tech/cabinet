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
		`"run", "-d"`,
		`"--name", $containerName`,
		`"-p", "$($lanePort):17880"`,
		`"-v", "$($containerVolume):/data"`,
		`"--listen", "0.0.0.0:17880"`,
		`$args += "-ReuseServer"`,
		`container_started = $containerStarted`,
		`"-BaseUrl", "http://127.0.0.1:$lanePort"`,
		`data_dir = Join-Path $repoRoot ".tmp\cypress-runtime-$lanePort"`,
		`profile = "e2e-cypress-$lanePort"`,
		`instance_name = "cypress-$lanePort"`,
		`source_commit = $sourceCommit`,
		`use_container_image = $UseContainerImage.IsPresent`,
		`container_image = if ($UseContainerImage) { $ContainerImage } else { $null }`,
		`container_startup_timeout_sec = if ($UseContainerImage) { $ContainerStartupTimeoutSec } else { $null }`,
		`keep_containers = $KeepContainers.IsPresent`,
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
		UseContainerImage        bool   `json:"use_container_image"`
		ContainerImage           string `json:"container_image"`
		ContainerStartupTimeout  int    `json:"container_startup_timeout_sec"`
		KeepContainers           bool   `json:"keep_containers"`
		Lanes                    []struct {
			UseContainerImage bool   `json:"use_container_image"`
			ContainerImage    string `json:"container_image"`
			ContainerName     string `json:"container_name"`
			ContainerVolume   string `json:"container_volume"`
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
	}
}
