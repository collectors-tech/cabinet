package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContinuousUIValidationMatrixTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| CONT-UI-CAB-009 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatal("expected traceability row for CONT-UI-CAB-009")
	}

	requiredFragments := []string{
		"#1095/#1058",
		"TestCypressMatrixRunnerProvidesIsolatedLanes",
		"TestCypressMatrixContainerImagePreflightFailsBeforeLaneFanout",
		"TestCypressMatrixSuccessFixtureWritesLiveMultiLaneSummary",
		"TestCypressMatrixSuccessFixtureWritesTimingMetadata",
		"TestApiContractSmokeScriptAcceptsAllowedRuntimePortForMappedContainers",
		"| implemented |",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected CONT-UI-CAB-009 traceability row to include %q; row: %s", fragment, row)
		}
	}
}
