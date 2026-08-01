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
			"#1667/#1668/#1672/#1715/#1981/#1999/#2003",
			"TestAgentSkillAPIPropagatesInvocationSourceContext",
			"TestAgentSkillDirectAPIRecordsGovernedTimelineEvidence",
			"preview-required non-mutation",
			"UI target ids",
			"remaining shell command and provider-readiness execution dispatch proof",
		},
		"AGENT-SKILL-COVERAGE-001": {
			"| partial |",
			"#1701/#1702/#1666/#1985/#1987/#1989/#2001",
			"openspec/traceability/agent-skill-coverage.md",
			"TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields",
			"TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec",
			"TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence",
			"TestAgentSkillCoverageMatrixBindsInboxReviewToIssue1987Evidence",
			"TestAgentSkillCoverageMatrixBindsInboxMissingStaleContextToIssue1987Evidence",
			"remaining partial status is limited to live Telegram/external-channel validation, broader unfinished matrix surfaces, and #1701 parent closure",
		},
		"AGENT-UNIVERSAL-CHANNELS-001": {
			"| partial |",
			"#1701/#1979/#1987/#1712/#1714",
			"TestChatMessagesNormalizeAgentContextEnvelopeForMainAndSidePanel",
			"assistant-inbox-agent-context/spec.cy.ts",
			"AGENT-CONTEXT-003/#1714 sends selected inventory row context",
			"AGENT-CONTEXT-004/#1714 preserves side-panel Agent context",
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

func TestAssistantExecution010TraceabilityNamesMergedGuidedWalkthroughEvidence(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "ASSISTANT-EXECUTION-010")
	for _, stale := range []string{
		"| planned |",
		"planned:",
		"guided-inventory-update/spec.cy.ts",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("ASSISTANT-EXECUTION-010 must not keep stale planned guided-walkthrough wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| implemented |",
		"#1509/#1514/#1991",
		"TestGuidedWorkflowRegistryMatchesInventoryItemUpdateRecipe",
		"TestChatMessageAppControlPlannerStartsGuidedInventoryWalkthrough",
		"ASSISTANT-WORKSPACE-008/#1509 starts a show-mode guided item-update walkthrough without mutation",
		"TestGuidedInventoryUpdatePersistsTimelineAndConfirmedMutation",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("ASSISTANT-EXECUTION-010 traceability row must include %q; row: %s", required, row)
		}
	}
}

func TestAssistantExecution012TraceabilityNamesMergedTargetHighlightEvidence(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "ASSISTANT-EXECUTION-012")
	for _, stale := range []string{
		"planned follow-on Cypress",
		"guided-inventory-update/spec.cy.ts",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("ASSISTANT-EXECUTION-012 must not keep stale planned target-highlight wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| implemented |",
		"#1511/#1514/#1993",
		"TestAssistantUITargetRegistryBindsInventoryWalkthroughTargets",
		"TestAssistantUITargetTraceabilityIsImplemented",
		"ASSISTANT-WORKSPACE-008/#1503 renders app-control route and preview cards from assistant thread context",
		"ui-guidance-highlight",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("ASSISTANT-EXECUTION-012 traceability row must include %q; row: %s", required, row)
		}
	}
}

func TestDiscovery003TraceabilityNamesMergedCandidateProvenanceEvidence(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "DISCOVERY-003")
	for _, stale := range []string{
		"planned Cypress",
		"planned:",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("DISCOVERY-003 must not keep stale planned UI evidence wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| implemented |",
		"#1111/#1124/#1125/#1995",
		"TestDiscoveriesPurposeAndHandoffSpecContracts",
		"TestDiscoveryCandidateContractIncludesStatusAndSourceResultAuditLink",
		"UI-SCREEN-DISCOVER-005 renders candidate provenance and destination actions",
		"source/provider",
		"triage status",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("DISCOVERY-003 traceability row must include %q; row: %s", required, row)
		}
	}
}

func TestAgentUniversalChannels002TraceabilityNamesCapabilityExplanationEvidence(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "AGENT-UNIVERSAL-CHANNELS-002")
	for _, stale := range []string{
		"| partial |",
		"| planned |",
		"planned:",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("AGENT-UNIVERSAL-CHANNELS-002 must not keep stale capability-explanation wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| implemented |",
		"#1701/#1977/#1997",
		"registry-derived capability/setup-state explanation",
		"TestAgentCapabilityExplanationDerivesFromRegistryAndProfileAuthority",
		"TestChatMessagesExplainAgentCapabilitiesForMainAndSidePanel",
		"direct API, main Chat, and side-panel Chat",
		"without claiming live Telegram/external-channel completion",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("AGENT-UNIVERSAL-CHANNELS-002 traceability row must include %q; row: %s", required, row)
		}
	}
}

func TestAgentSkillsRegistry009TraceabilityNamesRemainingExecutionBoundary(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "AGENT-SKILLS-REGISTRY-009")
	for _, stale := range []string{
		"#1667/#1668/#1672/#1715/#1981;",
		"routes through governed capability/workflow/UI target/command/provider readiness/preview-apply/Action Timeline boundaries",
		"remaining in-app UI target, shell command, and provider-readiness execution dispatch proof",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("AGENT-SKILLS-REGISTRY-009 must not keep stale broad execution wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| partial |",
		"#1667/#1668/#1672/#1715/#1981/#1999/#2003",
		"TestAgentSkillAPIPropagatesInvocationSourceContext",
		"TestAgentSkillDirectAPIRecordsGovernedTimelineEvidence",
		"preview-required non-mutation",
		"confirmed mutation",
		"read-only non-mutating execution",
		"UI target ids",
		"remaining shell command and provider-readiness execution dispatch proof",
		"without duplicating #1981 direct timeline evidence",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("AGENT-SKILLS-REGISTRY-009 traceability row must include %q; row: %s", required, row)
		}
	}
}

func TestAgentSkillCoverage001TraceabilityNamesMergedCoverageMatrixEvidence(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	row := traceabilityRow(t, string(raw), "AGENT-SKILL-COVERAGE-001")
	for _, stale := range []string{
		"#1701/#1702/#1666/#1987/#1989;",
		"per-surface Agent skill coverage matrix lists required Cabinet surfaces",
	} {
		if strings.Contains(row, stale) {
			t.Fatalf("AGENT-SKILL-COVERAGE-001 must not keep stale broad coverage wording %q; row: %s", stale, row)
		}
	}
	for _, required := range []string{
		"| partial |",
		"#1701/#1702/#1666/#1985/#1987/#1989/#2001",
		"#1985 reconciles the Skills page detail/actions entry point",
		"#1987 binds Inbox review Agent launch context",
		"#1989 reconciles Inbox missing/stale context evidence",
		"TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence",
		"TestAgentSkillCoverageMatrixBindsInboxReviewToIssue1987Evidence",
		"TestAgentSkillCoverageMatrixBindsInboxMissingStaleContextToIssue1987Evidence",
		"remaining partial status is limited to live Telegram/external-channel validation, broader unfinished matrix surfaces, and #1701 parent closure",
		"without duplicating closed #1985/#1989 scope",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("AGENT-SKILL-COVERAGE-001 traceability row must include %q; row: %s", required, row)
		}
	}
}

func traceabilityRow(t *testing.T, raw, id string) string {
	t.Helper()
	prefixWithTicks := "| `" + id + "` |"
	prefixWithoutTicks := "| " + id + " |"
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, prefixWithTicks) || strings.HasPrefix(line, prefixWithoutTicks) {
			return line
		}
	}
	t.Fatalf("missing traceability row for %s", id)
	return ""
}
