package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNonFunctionalTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range []string{"NON-FUNCTIONAL-001", "NON-FUNCTIONAL-002"} {
			if strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"NON-FUNCTIONAL-001": {
			"TestNFRGates",
			"startup fast-exit",
			"search <=200ms target",
			"scanner local provider run",
			"TestNonFunctionalTraceabilityImplemented",
			"| implemented |",
		},
		"NON-FUNCTIONAL-002": {
			"TestNFRGates",
			"crash-free beta objective",
			"startup/search/scanner diagnostics",
			"strict startup failure gate",
			"TestNonFunctionalTraceabilityImplemented",
			"| implemented |",
		},
	}

	for id, requiredFragments := range requiredByID {
		row := rows[id]
		if row == "" {
			t.Fatalf("expected traceability row for %s", id)
		}
		for _, fragment := range requiredFragments {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected %s traceability row to include %q; row: %s", id, fragment, row)
			}
		}
	}
}
