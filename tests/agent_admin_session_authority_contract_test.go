package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentAdminSessionAuthorityContractIsTraceable(t *testing.T) {
	t.Parallel()

	specRaw, err := os.ReadFile(filepath.Join("..", "openspec", "specs", "agent-skills-registry", "spec.md"))
	if err != nil {
		t.Fatalf("read Agent Skills spec: %v", err)
	}
	for _, required := range []string{
		"AGENT-SKILLS-REGISTRY-013",
		"server-validated session",
		"active profile membership",
		"client-supplied `admin_session`, role, permission, or authority",
		"planner, read, preview, apply, cancel, replay, and audit",
		"opaque session token",
		"protected owner",
	} {
		if !strings.Contains(string(specRaw), required) {
			t.Fatalf("Agent Skills authority spec missing %q", required)
		}
	}

	openAPIRaw, err := os.ReadFile(filepath.Join("..", "docs", "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	for _, required := range []string{
		"Users/admin Agent requests derive authority from the validated Cabinet session",
		"agent_admin_authentication_required",
		"agent_admin_profile_forbidden",
	} {
		if !strings.Contains(string(openAPIRaw), required) {
			t.Fatalf("OpenAPI authority contract missing %q", required)
		}
	}

	traceRaw, err := os.ReadFile(filepath.Join("..", "openspec", "traceability.md"))
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	row := traceabilityRow(t, string(traceRaw), "AGENT-SKILLS-REGISTRY-013")
	for _, required := range []string{
		"#2088",
		"TestAgentUsersAPIDoesNotTrustClientAssertedAdminAuthority",
		"TestAgentUsersAPIFailsClosedForMissingAndWrongProfileSessions",
		"TestAgentUsersAPIAcceptsServerDerivedActiveOwnerSession",
		"| implemented |",
	} {
		if !strings.Contains(row, required) {
			t.Fatalf("AGENT-SKILLS-REGISTRY-013 traceability row missing %q: %s", required, row)
		}
	}
}
