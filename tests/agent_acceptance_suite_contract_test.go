package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentAcceptanceSuiteEvidenceMapCoversIssue1716Scope(t *testing.T) {
	t.Parallel()

	mapPath := filepath.Join("..", "openspec", "traceability", "agent-acceptance-suite.md")
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatalf("read agent acceptance evidence map: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Issue: #1716",
		"fixture/proof-packet evidence from live production Telegram-channel evidence",
		"Main Chat read-only app work",
		"Side-panel Chat read-only app work",
		"Preview/cancel/apply mutation",
		"Attachment success/failure",
		"Telegram authorized text/media intake",
		"Telegram unauthorized sender rejection",
		"Telegram proof-packet versus live-channel distinction",
		"implemented-fixture",
		"partial",
		"CHATS-WORKSPACE-008/#1503 dispatches normal main Chat text to app-control route planning without Inbox noise",
		"ASSISTANT-WORKSPACE-009/#1503 dispatches normal side-panel text to app-control route planning without Inbox noise",
		"ASSISTANT-EXECUTION-001/002/003/004 renders preview-before-apply with confirm and explicit permission guidance",
		"TestServiceThreadMessagePreviewApplyLifecycle",
		"TestChatAPIsThreadMessageAttachmentAndPreviewApply",
		"cross-thread attachment reuse is rejected",
		"AGENT-ATTACHMENTS-001 handles side-panel attachments with the same scoped message binding as main Chat",
		"side-panel Chat uploads explicit user-selected attachments",
		"TestTelegramCatalogCaptureAPIRequiresPersistedSenderAuthorization",
		"TestTelegramCatalogCaptureWebhookAPIResolvesProfileAuthorization",
		"TestTelegramExternalIntakeProofRequiresAuthorizedProviderEvidence",
		"manual live Telegram-channel checklist",
		"openspec/traceability/agent-live-telegram-channel-checklist.md",
		"#1773",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected agent acceptance evidence map to include %q", fragment)
		}
	}
}

func TestAgentLiveTelegramChecklistNamesNonSecretProofRequirements(t *testing.T) {
	t.Parallel()

	checklistPaths := []string{
		filepath.Join("..", "openspec", "traceability", "agent-live-telegram-channel-checklist.md"),
		filepath.Join("..", "docs", "validation", "agent-live-telegram-channel-checklist.md"),
	}

	requiredFragments := []string{
		"Issue: #1773",
		"Parent: #1716",
		"non-secret evidence",
		"Do not record bot tokens",
		"Authorized Text Intake",
		"Authorized Media Intake",
		"Unauthorized Sender Rejection",
		"Source message id",
		"Workflow run / preview id",
		"Response or deep-link state",
		"Mutation state before confirmation",
		"Record absence check",
		"Mutation absence check",
		"linked issue/PR comment",
	}

	for _, checklistPath := range checklistPaths {
		raw, err := os.ReadFile(checklistPath)
		if err != nil {
			t.Fatalf("read live Telegram channel checklist %s: %v", checklistPath, err)
		}
		content := string(raw)

		for _, fragment := range requiredFragments {
			if !strings.Contains(content, fragment) {
				t.Fatalf("expected live Telegram checklist %s to include %q", checklistPath, fragment)
			}
		}
	}
}

func TestAgentAcceptanceSuiteTraceabilityStaysBoundToOpenSpec(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(traceRaw)

	for _, fragment := range []string{
		"AGENT-ACCEPTANCE-SUITE-001",
		"#1716",
		"openspec/traceability/agent-acceptance-suite.md",
		"TestAgentAcceptanceSuiteEvidenceMapCoversIssue1716Scope",
		"AGENT-UNIVERSAL-CHANNELS-003",
		"AGENT-UNIVERSAL-CHANNELS-004",
		"AGENT-UNIVERSAL-CHANNELS-005",
		"fixture/proof-packet validation from live production-channel validation",
		"| partial |",
		"docs/validation/agent-live-telegram-channel-checklist.md",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected #1716 acceptance traceability to include %q", fragment)
		}
	}

	specPath := filepath.Join("..", "openspec", "changes", "define-universal-agent-channel-contracts", "specs", "agent-universal-channels", "spec.md")
	specRaw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read universal Agent channel spec: %v", err)
	}
	spec := string(specRaw)
	for _, fragment := range []string{
		"#### Scenario: Maintain an acceptance evidence map",
		"Cabinet SHALL maintain a #1716 Agent acceptance evidence map",
		"fixture/proof-packet validation from live production-channel validation",
		"live Telegram-channel checklist",
	} {
		if !strings.Contains(spec, fragment) {
			t.Fatalf("expected universal Agent channel spec to include %q", fragment)
		}
	}
}
