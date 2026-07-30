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
		"TestChatAgentPlannerDeniesExternalWriteSelectionWithoutApproval",
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
