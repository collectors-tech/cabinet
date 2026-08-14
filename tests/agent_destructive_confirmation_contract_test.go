package tests

import (
	"os"
	"strings"
	"testing"
)

func TestAgentDestructiveStrongConfirmationTraceability(t *testing.T) {
	t.Parallel()

	spec, err := os.ReadFile("../openspec/specs/agent-skills-registry/spec.md")
	if err != nil {
		t.Fatalf("read Agent Skills registry spec: %v", err)
	}
	trace, err := os.ReadFile("../openspec/traceability.md")
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	openAPI, err := os.ReadFile("../docs/api/openapi.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}

	for _, required := range []string{
		"AGENT-SKILLS-REGISTRY-014",
		"action-specific strong confirmation",
		"single-use token",
		"protected-owner rules",
		"backup compatibility",
		"pre-restore recovery backup",
	} {
		if !strings.Contains(string(spec), required) {
			t.Fatalf("destructive confirmation spec missing %q", required)
		}
	}
	row := traceabilityRow(t, string(trace), "AGENT-SKILLS-REGISTRY-014")
	for _, required := range []string{"#2089", "agent_destructive_confirmation_api_test.go", "agent-admin-session-authority/spec.cy.ts", "| implemented |"} {
		if !strings.Contains(row, required) {
			t.Fatalf("destructive confirmation traceability missing %q: %s", required, row)
		}
	}
	for _, required := range []string{"/api/agent/skills/confirm-destructive:", "five-minute single-use confirmation token", "strong confirmation for a destructive Agent preview"} {
		if !strings.Contains(string(openAPI), required) {
			t.Fatalf("destructive confirmation OpenAPI missing %q", required)
		}
	}
}
