package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiveUIAuditIssueBindingsExistInTraceability(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	content := string(raw)

	required := []string{
		"#239",
		"#258",
		"#259",
		"#260",
		"#261",
		"#262",
		"#263",
		"#264",
	}

	for _, issue := range required {
		if !strings.Contains(content, issue) {
			t.Fatalf("expected traceability to include live UI audit issue binding %s", issue)
		}
	}
}
