package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentationGovernanceTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	ids := []string{
		"DOCUMENTATION-GOVERNANCE-001",
		"DOCUMENTATION-GOVERNANCE-002",
		"DOCUMENTATION-GOVERNANCE-003",
		"DOCUMENTATION-GOVERNANCE-004",
		"DOCUMENTATION-GOVERNANCE-005",
	}

	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range ids {
			if strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"DOCUMENTATION-GOVERNANCE-001": {
			"OpenSpec normative source policy",
			"openspec/specs/general/documentation-governance/spec.md",
			"TestOpenSpecScenariosRequireGivenWhenThen",
			"TestDocumentationGovernanceTraceabilityImplemented",
			"| implemented |",
		},
		"DOCUMENTATION-GOVERNANCE-002": {
			"docs markdown exception policy",
			"docs/help-center/**/*.md",
			"docs/auth/exploration-auth-setup.md",
			"TestDocumentationGovernanceTraceabilityImplemented",
			"| implemented |",
		},
		"DOCUMENTATION-GOVERNANCE-003": {
			"docs/api/openapi.yaml",
			"OpenAPI contract remains stable",
			"TestOpenAPIParitySuite",
			"TestDocumentationGovernanceTraceabilityImplemented",
			"| implemented |",
		},
		"DOCUMENTATION-GOVERNANCE-004": {
			"Migration Inventory",
			"docs/FULL_FEATURE_LIST.md",
			"docs/SPEC.md",
			"docs/USE_CASES_AND_SCENARIOS.md",
			"TestDocumentationGovernanceTraceabilityImplemented",
			"| implemented |",
		},
		"DOCUMENTATION-GOVERNANCE-005": {
			"openspec/migrations/legacy-docs-file-audit.yaml",
			"baseline commit 82294546bf0b715fe49394e1c5a885d3045294d2",
			"migration audit covers entire baseline",
			"TestDocumentationGovernanceTraceabilityImplemented",
			"| implemented |",
		},
	}

	for _, id := range ids {
		row := rows[id]
		if row == "" {
			t.Fatalf("expected traceability row for %s", id)
		}
		for _, fragment := range requiredByID[id] {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected %s traceability row to include %q; row: %s", id, fragment, row)
			}
		}
	}
}

func TestAPIDocsReleaseGateTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "openspec", "traceability.md"))
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	content := string(raw)
	for id, required := range map[string][]string{
		"API-DOCS-004": {"Redocly lint", "implemented"},
		"API-DOCS-005": {"#2037", "TestOpenAPIParitySuite", "TestVerifyParityTestOutputRejectsSuccessfulPackageWithNoNamedTest", "implemented"},
		"API-DOCS-006": {"CompanionProfileBearer", "critical_security_and_payload_contracts_are_explicit", "implemented"},
	} {
		marker := "| `" + id + "` "
		start := strings.Index(content, marker)
		if start < 0 {
			t.Fatalf("expected traceability row for %s", id)
		}
		end := strings.Index(content[start:], "\n")
		if end < 0 {
			end = len(content) - start
		}
		row := content[start : start+end]
		for _, fragment := range required {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected %s traceability row to include %q; row: %s", id, fragment, row)
			}
		}
	}
}
