package tests

import (
	"os"
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
