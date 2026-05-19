package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssistantOpenAIWorkflowPlanBindsIssue847Contracts(t *testing.T) {
	t.Parallel()

	planPath := filepath.Join("..", "docs", "assistant-openai-workflow-plan.md")
	planRaw, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read assistant OpenAI workflow plan: %v", err)
	}
	plan := string(planRaw)

	requiredPlanFragments := []string{
		"issue #847",
		"softlystudio/scheduling-assistant#740",
		"softlystudio/scheduling-assistant#770",
		"catalog_add_from_photo",
		"catalog_add_from_barcode",
		"catalog_add_from_text",
		"image_analyze",
		"image_process",
		"content_generate",
		"listing_draft_generate",
		"purchase_reconcile",
		"package_reconcile",
		"provider_test",
		"app_action_preview",
		"app_action_apply_after_confirm",
		"Browser Auth must never be marked connected from navigation alone",
		"Telegram is a channel into the same governed capability system",
	}
	for _, fragment := range requiredPlanFragments {
		if !strings.Contains(plan, fragment) {
			t.Fatalf("expected plan to include %q", fragment)
		}
	}

	specPath := filepath.Join("..", "openspec", "specs", "chats", "assistant-execution-surfaces", "spec.md")
	specRaw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read assistant execution spec: %v", err)
	}
	spec := string(specRaw)
	for _, requirement := range []string{"ASSISTANT-EXECUTION-006", "ASSISTANT-EXECUTION-007", "ASSISTANT-EXECUTION-008"} {
		if !strings.Contains(spec, requirement) {
			t.Fatalf("expected assistant execution spec to include %s", requirement)
		}
	}

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(traceRaw)
	for _, requirement := range []string{"ASSISTANT-EXECUTION-006", "ASSISTANT-EXECUTION-007", "ASSISTANT-EXECUTION-008"} {
		if !strings.Contains(trace, requirement) || !strings.Contains(trace, "#847") {
			t.Fatalf("expected traceability to bind %s to issue #847 planning evidence", requirement)
		}
	}
}
