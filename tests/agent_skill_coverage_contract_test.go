package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Issue: #1702",
		"Parent: #1701",
		"Related registry epic: #1666",
		"Main Chat `/chats`",
		"Side-panel Chat",
		"Inbox review",
		"Telegram/external channel",
		"Skills page detail/actions",
		"| Dashboard |",
		"| Inventory |",
		"| Wishlist |",
		"| Collections |",
		"| Media |",
		"| Discoveries |",
		"| Market Watch / Scanner |",
		"| Purchases |",
		"| Integrations |",
		"| Inbox |",
		"| Users / workspace admin |",
		"| Settings / Profile |",
		"| Settings / Account |",
		"| Settings / Appearance |",
		"| Settings / Storage |",
		"| Import / Export / Backup / Restore / Maintenance |",
		"| Chats / Agent itself |",
		"Expected skill ids",
		"User-facing request examples",
		"Safety",
		"Required context/selection",
		"Required setup",
		"Bound capability ids",
		"Bound guided workflow ids",
		"Missing issue/PR links",
		"Validation evidence",
		"Blocked/deferred reason",
		"#1707",
		"#1708",
		"#1709",
		"#1710",
		"#1711",
		"#1715",
		"#1773",
		"cabinet.navigate.open_surface",
		"cabinet.guided.inventory.update_item",
		"cabinet.inbox.summarise_unhandled",
		"cabinet.users.remove_user",
		"Marketplace discovery/publishing/payments/reviews remain deferred",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected agent skill coverage matrix to include %q", fragment)
		}
	}
}

func TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(traceRaw)

	for _, fragment := range []string{
		"AGENT-SKILL-COVERAGE-001",
		"#1701/#1702/#1666",
		"openspec/traceability/agent-skill-coverage.md",
		"TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields",
		"TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec",
		"TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence",
		"#1985 reconciles the Skills page detail/actions entry point",
		"main Chat, side-panel Chat, Inbox review, and Telegram/external channels",
		"| partial |",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected agent skill coverage traceability to include %q", fragment)
		}
	}

	specPath := filepath.Join("..", "openspec", "specs", "agent-skills-registry", "spec.md")
	specRaw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read agent skill registry spec: %v", err)
	}
	spec := string(specRaw)

	for _, fragment := range []string{
		"### Requirement: Agent skill coverage SHALL stay mapped by Cabinet surface",
		"#### Scenario: Maintain per-surface coverage matrix",
		"skill ids, safety levels, dependencies, statuses, issue links, and validation evidence",
		"distinguish in-app main Chat, side-panel Chat, Inbox review, and Telegram/external channels",
	} {
		if !strings.Contains(spec, fragment) {
			t.Fatalf("expected agent skill registry spec to include %q", fragment)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var skillsPageRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Skills page detail/actions |") {
			skillsPageRow = line
			break
		}
	}
	if skillsPageRow == "" {
		t.Fatalf("expected Skills page detail/actions channel coverage row")
	}

	for _, stale := range []string{
		"| planned |",
		"Planned Cypress",
	} {
		if strings.Contains(skillsPageRow, stale) {
			t.Fatalf("expected Skills page coverage row to avoid stale planned #1670 wording %q: %s", stale, skillsPageRow)
		}
	}

	for _, fragment := range []string{
		"| implemented |",
		"#1670",
		"TestAgentSkillStateAPIEnablesAndDisablesImportedSkill",
		"TestAgentSkillStateAPIBlocksBuiltInAndHighRiskWithoutConfirmation",
		"AGENT-SKILLS-REGISTRY-008",
		"lists skills opens details and imports a local archive disabled by default",
	} {
		if !strings.Contains(skillsPageRow, fragment) {
			t.Fatalf("expected Skills page coverage row to include merged #1670 evidence %q: %s", fragment, skillsPageRow)
		}
	}
}
