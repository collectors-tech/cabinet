package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestImplementedIntegrationTraceabilityIDsAreUnique(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	idPattern := regexp.MustCompile("^\\| `?([^`| ]+)`? \\|")
	seen := map[string]int{}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "| ") || !strings.Contains(line, "| implemented |") {
			continue
		}
		matches := idPattern.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		if !strings.HasPrefix(matches[1], "INTEGRATION-") {
			continue
		}
		seen[matches[1]]++
	}

	var duplicates []string
	for id, count := range seen {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	if len(duplicates) > 0 {
		t.Fatalf("implemented integration traceability IDs must be unique, found duplicates: %s", strings.Join(duplicates, ", "))
	}
}

func TestRemainingTraceabilityBacklogRowsAreExplicit(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	idPattern := regexp.MustCompile("^\\| `?([^`| ]+)`? \\|")
	allowed := map[string][]string{
		"AGENT-SKILLS-REGISTRY-009": {
			"| partial |",
			"#1667/#1668/#1672/#1715",
			"TestAgentSkillAPIPropagatesInvocationSourceContext",
			"broader skill execution timeline coverage remains planned",
		},
		"AGENT-SKILLS-REGISTRY-012": {
			"| partial |",
			"#1942/#1701",
			"TestDashboardActivitySummarySkillExposesReadOnlyProfileBoundary",
			"side-panel Agent Skill dispatch",
			"main Chat natural-language routing remains sequenced behind #1933",
		},
		"AGENT-SKILL-COVERAGE-001": {
			"| partial |",
			"#1701/#1702/#1666",
			"openspec/traceability/agent-skill-coverage.md",
			"TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields",
			"TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec",
		},
		"AGENT-UNIVERSAL-CHANNELS-001": {
			"| planned |",
			"#1701/#1712/#1714",
			"AGENT-UNIVERSAL-CHANNELS-001 opens Agent from supported surfaces with preserved context",
			"preserve profile/route/thread/selection/source context",
		},
		"AGENT-UNIVERSAL-CHANNELS-002": {
			"| partial |",
			"#1701/#1977/#1712/#1708/#1709/#1710/#1711/#1715",
			"TestAgentCapabilityExplanationDerivesFromRegistryAndProfileAuthority",
			"TestChatMessagesExplainAgentCapabilitiesForMainAndSidePanel",
			"read-only, preview-only, confirm-required, external-write/setup-required",
		},
		"AGENT-UNIVERSAL-CHANNELS-004": {
			"| planned |",
			"#1701/#1712/#1704/#1705/#1706",
			"TestTelegramAgentIntakeRequiresAuthorizationAndSetupProof",
			"non-secret proof",
		},
		"AGENT-UNIVERSAL-CHANNELS-005": {
			"| planned |",
			"#1701/#1712/#1705/#1706",
			"TestTelegramAgentIntakeRoutesTextMediaThroughPreviewConfirmApply",
			"TELEGRAM-AGENT-REVIEW-001 opens external review thread",
		},
		"AGENT-ACCEPTANCE-SUITE-001": {
			"| partial |",
			"#1716/#1773",
			"openspec/traceability/agent-acceptance-suite.md",
			"TestAgentAcceptanceSuiteEvidenceMapCoversIssue1716Scope",
			"fixture/proof-packet validation from live production-channel validation",
		},
		"INTEGRATION-063": {
			"| partial |",
			"#1463",
			"docs/integrations/provider-authoring.md",
			"TestIntegrationProviderAuthoringGuideCoversIssue1463Workflow",
			"provider-authoring workflow",
		},
		"ASSISTANT-EXECUTION-010": {
			"| planned |",
			"#1509/#1514",
			"TestGuidedWalkthroughModesGovernCommandPermissions",
			"ASSISTANT-EXECUTION-010 preserves confirm-before-apply across walkthrough modes",
		},
		"UI-SCREEN-CHAT-COPILOT-018": {
			"| planned |",
			"#1205",
			"Future affected main chat or side-panel Assistant issues/PRs",
			"TestAssistantUIDirectionTraceabilityRowIsActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"PROVIDER-WORKFLOW-001": {
			"| partial |",
			"#827",
			"#841",
			"#842",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"PROVIDER-WORKFLOW-002": {
			"| partial |",
			"#827",
			"#841",
			"#842",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"PROVIDER-WORKFLOW-003": {
			"| partial |",
			"#827",
			"#841",
			"#842",
			"verified credentials",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"PROVIDER-WORKFLOW-004": {
			"| partial |",
			"#827",
			"#841",
			"#842",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"PROVIDER-WORKFLOW-FULL-ASSESSMENT": {
			"| partial |",
			"#827",
			"#841",
			"#842",
			"live credential/capability evidence",
			"TestProviderWorkflowTraceabilityPartialRowsAreActionable",
			"TestRemainingTraceabilityBacklogRowsAreExplicit",
		},
		"RUNTIME-CORE-020": {
			"| partial |",
			"#1869",
			"openspec/migration/beta-packaged-core-workflow-acceptance.md",
			"TestPackagedCoreWorkflowAcceptanceChecklistCoversIssue1869",
			"#1864 release approval guardrails",
		},
	}

	seen := map[string]bool{}
	var unexpected []string
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "| ") || (!strings.Contains(line, "| partial |") && !strings.Contains(line, "| planned |")) {
			continue
		}
		matches := idPattern.FindStringSubmatch(line)
		if len(matches) != 2 {
			unexpected = append(unexpected, line)
			continue
		}
		id := matches[1]
		requiredFragments, ok := allowed[id]
		if !ok {
			unexpected = append(unexpected, id)
			continue
		}
		seen[id] = true
		for _, fragment := range requiredFragments {
			if !strings.Contains(line, fragment) {
				t.Fatalf("expected %s non-implemented traceability row to include %q; row: %s", id, fragment, line)
			}
		}
	}
	if len(unexpected) > 0 {
		sort.Strings(unexpected)
		t.Fatalf("unexpected non-implemented traceability rows: %s", strings.Join(unexpected, ", "))
	}

	var missing []string
	for id := range allowed {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("expected allowed non-implemented traceability rows to remain explicit: %s", strings.Join(missing, ", "))
	}
}
