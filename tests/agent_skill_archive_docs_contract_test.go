package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSkillArchiveTemplateDocsCoverLocalImportContract(t *testing.T) {
	t.Parallel()

	docPath := filepath.Join("..", "docs", "integrations", "agent-skill-archive-template.md")
	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read agent skill archive template doc: %v", err)
	}
	doc := string(raw)

	for _, fragment := range []string{
		"Issue: #1671",
		"AGENT-SKILLS-REGISTRY-005",
		"AGENT-SKILLS-REGISTRY-010",
		".cabinet-skill.zip",
		"cabinet.skill.json",
		"schemas/input.schema.json",
		"workflows/guided-workflow.json",
		"ui-targets/ui-targets.json",
		"skill-validation.json",
		"`schema`",
		"`id`",
		"`version`",
		"`displayName`",
		"`description`",
		"`category`",
		"`author`",
		"`source`",
		"`license`",
		"`safetyLevel`",
		"`status`",
		"`modes`",
		"`permissions`",
		"`inputSchemaRef`",
		"`outputSchemaRef`",
		"`audit`",
		"`checksums`",
		"`compatibility`",
		"cabinet.example.open_inventory_help",
		"cabinet.example.update_item_guided",
		"read-only",
		"confirm-required",
		"preview/confirm/apply",
		"Action Timeline",
		"valid-ready-to-install",
		"valid-with-warnings",
		"blocked-missing-dependency",
		"blocked-invalid-manifest",
		"blocked-unsafe-archive",
		"installed-disabled",
		"installed-enabled",
		"marketplace discovery",
		"publishing, payments, ratings, reviews",
	} {
		if !strings.Contains(doc, fragment) {
			t.Fatalf("expected archive template doc to include %q", fragment)
		}
	}
}

func TestAgentSkillArchiveDocsTraceabilityStaysBoundToIssue1671(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(raw)

	for _, fragment := range []string{
		"AGENT-SKILLS-REGISTRY-005",
		"#1667/#1669/#1671",
		"docs/integrations/agent-skill-archive-template.md",
		"TestAgentSkillArchiveTemplateDocsCoverLocalImportContract",
		"AGENT-SKILLS-REGISTRY-010",
		"#1667/#1670/#1671",
		"TestAgentSkillArchiveDocsTraceabilityStaysBoundToIssue1671",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected agent skill archive traceability to include %q", fragment)
		}
	}
}
