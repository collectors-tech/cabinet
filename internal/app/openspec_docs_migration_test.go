package app

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocsFolderContainsNoMarkdownAfterMigration(t *testing.T) {
	t.Parallel()

	allowedMarkdownFiles := map[string]struct{}{
		"../../docs/CONSOLE-OUTPUT-STANDARD.md":                                    {},
		"../../docs/PRODUCT-OVERVIEW.md":                                           {},
		"../../docs/PRODUCT_OVERVIEW_PLAN.md":                                      {},
		"../../docs/auth/exploration-auth-setup.md":                                {},
		"../../docs/backlog/reviews/chat-app-control-issue-plan-2026-06-25.md":     {},
		"../../docs/backlog/reviews/chat-application-control-review-2026-06-25.md": {},
		"../../docs/integrations/provider-authoring.md":                            {},
		"../../docs/validation/agent-live-telegram-channel-checklist.md":           {},
	}

	var markdownFiles []string
	err := filepath.WalkDir("../../docs", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			normalized := filepath.ToSlash(path)
			if strings.HasPrefix(normalized, "../../docs/help-center/") {
				return nil
			}
			if _, ok := allowedMarkdownFiles[normalized]; ok {
				return nil
			}
			markdownFiles = append(markdownFiles, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs directory: %v", err)
	}
	if len(markdownFiles) > 0 {
		t.Fatalf("docs folder still contains markdown files: %v", markdownFiles)
	}
}

func TestOpenSpecDocumentationGovernanceContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../openspec/specs/general/documentation-governance/spec.md")
	if err != nil {
		t.Fatalf("read documentation governance spec: %v", err)
	}
	src := string(b)

	required := []string{
		"## Requirements",
		"### Requirement DOCUMENTATION-GOVERNANCE-001: OpenSpec Is The Normative Documentation Source",
		"### Requirement DOCUMENTATION-GOVERNANCE-002: Legacy Docs Directory Contains No Markdown Sources",
		"### Requirement DOCUMENTATION-GOVERNANCE-003: OpenAPI Contract Remains in docs/api/openapi.yaml",
		"docs/PRODUCT-OVERVIEW.md",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("documentation governance spec missing token: %s", token)
		}
	}
}

func TestOpenSpecWorkflowReferencesDocumentationGovernanceSpec(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../openspec/specs/README.md")
	if err != nil {
		t.Fatalf("read openspec specs readme: %v", err)
	}
	src := string(b)
	if !strings.Contains(src, "documentation-governance") {
		t.Fatalf("openspec/specs/README.md must reference documentation-governance spec")
	}
}

func TestNoLegacyDocsReferencesInAgentPolicy(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../AGENTS.md")
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	src := string(b)
	forbidden := []string{
		"docs/OPENSPEC_MIGRATION_CATALOG.md",
		"docs/OPENSPEC_MIGRATION_TODO.md",
		"docs/OPENSPEC_WORKFLOW.md",
	}
	for _, token := range forbidden {
		if strings.Contains(src, token) {
			t.Fatalf("AGENTS.md contains legacy docs reference: %s", token)
		}
	}
}

func TestOpenSpecWorkflowContentMigrated(t *testing.T) {
	t.Parallel()

	read := func(path string) string {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(b)
	}

	governance := read("../../openspec/specs/general/documentation-governance/spec.md")
	requiredTokens := []string{
		"docs/FULL_FEATURE_LIST.md",
		"docs/SPEC.md",
		"docs/USE_CASES_AND_SCENARIOS.md",
		"docs/ui-spec/02-SCREEN-SPECS.md",
		"docs/ui-spec/05-TEST-MATRIX-UI.md",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(governance, token) {
			t.Fatalf("documentation governance spec missing migrated legacy token: %s", token)
		}
	}
}
