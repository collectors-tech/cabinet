package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderWorkflowTraceabilityPartialRowsAreActionable(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range []string{
			"PROVIDER-WORKFLOW-001",
			"PROVIDER-WORKFLOW-002",
			"PROVIDER-WORKFLOW-003",
			"PROVIDER-WORKFLOW-004",
			"PROVIDER-WORKFLOW-FULL-ASSESSMENT",
		} {
			if strings.HasPrefix(line, "| "+id+" ") || strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"PROVIDER-WORKFLOW-001": {
			"#827",
			"#841",
			"#842",
			"provider-specific Cypress/API checklists",
			"connect/auth, validate token, search/query, candidate mapping, apply/import",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"| partial |",
		},
		"PROVIDER-WORKFLOW-002": {
			"#827",
			"#841",
			"#842",
			"auth failure, empty results, rate limiting, and upstream failures",
			"deterministic error-code and next-action suites",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"| partial |",
		},
		"PROVIDER-WORKFLOW-003": {
			"#827",
			"#841",
			"#842",
			"mock/live parity comparisons",
			"normalized candidate/output schema",
			"live provider lane unavailable without verified credentials",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"| partial |",
		},
		"PROVIDER-WORKFLOW-004": {
			"#827",
			"#841",
			"#842",
			"setup docs QA checklist",
			"validate-token suites",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"| partial |",
		},
		"PROVIDER-WORKFLOW-FULL-ASSESSMENT": {
			"#827",
			"#841",
			"#842",
			"provider-workflow-full-assessment.md",
			"execution issues per provider",
			"stage evidence",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"| partial |",
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
