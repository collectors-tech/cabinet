package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentChatPlannerTraceabilityRowsBindIssue1933Evidence(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(raw)

	for _, fragment := range []string{
		"#1933",
		"route-natural-language-chat-through-governed-agent-skills",
		"TestChatAgentPlannerUsesDeterministicFakeProviderSkillSelection",
		"TestChatAgentPlannerExecutesReadOnlySelectionWithProfileIsolation",
		"TestChatAgentPlannerConvertsLocalWriteSelectionToPreviewOnly",
		"TestChatAgentPlannerPreviewsExternalWriteSelectionWithoutApplyAuthority",
		"TestChatMessagesUseSharedAgentPlannerContractForMainAndSidePanel",
		"run-chat-agent-planner-packaged-smoke.ps1",
		"provider-backed read",
		"preview-only local write",
		"| implemented |",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected #1933 traceability to include %q", fragment)
		}
	}
}

func TestAgentCoverageTraceabilityNamesIssue1933PlannerSlice(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	matrix := string(raw)

	for _, fragment := range []string{
		"#1933",
		"natural-language Chat planner",
		"provider-backed read",
		"preview-only write",
		"confirmed apply",
		"replay/idempotency",
		"does not close #1701",
	} {
		if !strings.Contains(matrix, fragment) {
			t.Fatalf("expected #1701 coverage matrix to include #1933 planner slice fragment %q", fragment)
		}
	}
}

func TestDashboardSummaryPlannerTraceabilityNamesIssue1983Evidence(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(raw)

	for _, fragment := range []string{
		"AGENT-SKILLS-REGISTRY-012",
		"#1942/#1983/#1701",
		"TestChatAgentPlannerRoutesDashboardActivitySummaryFromMainChat",
		"main Chat natural-language routing can select and execute the skill",
		"without mutation previews, confirmation tokens, external writes, or fabricated historical deltas",
		"| implemented |",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected #1983 Dashboard summary traceability to include %q", fragment)
		}
	}

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	matrixRaw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read coverage matrix: %v", err)
	}
	matrix := string(matrixRaw)
	for _, fragment := range []string{
		"| Dashboard |",
		"#1714, #1942, #1983",
		"TestChatAgentPlannerRoutesDashboardActivitySummaryFromMainChat",
		"main Chat natural-language Dashboard summary planner selection/execution",
		"None |",
	} {
		if !strings.Contains(matrix, fragment) {
			t.Fatalf("expected #1983 Dashboard summary coverage matrix to include %q", fragment)
		}
	}
	if strings.Contains(matrix, "Dashboard summary skill has no focused child issue yet") ||
		strings.Contains(matrix, "main Chat natural-language routing remains #1933") {
		t.Fatalf("Dashboard coverage matrix still contains stale #1983 gap language")
	}
}
