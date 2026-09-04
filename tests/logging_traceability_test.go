package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggingTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range []string{"LOGGING-001", "LOGGING-002"} {
			if strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"LOGGING-001": {
			"Structured runtime access log",
			"method/path/status/duration/profile/request ID",
			"TestRuntimeLifecycleMetadataAndStructuredLogs",
			"TestLoggingTraceabilityImplemented",
			"| implemented |",
		},
		"LOGGING-002": {
			"Structured runtime error log",
			"non-2xx failure context",
			"request ID",
			"TestRuntimeLifecycleMetadataAndStructuredLogs",
			"TestLogListExportAndRedaction",
			"TestLoggingTraceabilityImplemented",
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
