package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevelopQualityGateWorkflowContract(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	workflowPath := filepath.Join(repoRoot, ".github", "workflows", "develop-quality-gate.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read develop gate workflow: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"name: Develop Quality Gate",
		"pull_request:",
		"branches: [develop]",
		"Build runtime UI static bundle for contract tests",
		"Verify deployment infrastructure contracts",
		"npm run test:infra",
		"openspec validate --all --strict --no-interactive",
		"npm run build",
		"fetch-depth: 0",
		"go test ./internal/... ./cmd/...",
		"Build runtime UI static bundle for OpenAPI tests",
		"go test ./internal/app -run TestOpenAPIDocumentsOnboardingSampleDataEndpoint -count=1",
		"@redocly/cli@latest lint docs/api/openapi.yaml",
		"@redocly/cli@latest build-docs docs/api/openapi.yaml -o docs/api/index.html",
		"npm run e2e:ci-smoke",
		"actions/upload-artifact@v4",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("develop quality gate missing %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"branches: [main]",
		"git push",
		"merge develop",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("develop quality gate contains forbidden fragment %q", forbidden)
		}
	}
}

func TestBetaReleaseCandidateWorkflowContract(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	workflowPath := filepath.Join(repoRoot, ".github", "workflows", "beta-release-candidate.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read release-candidate workflow: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"name: Beta Release Candidate Gate",
		"workflow_dispatch:",
		"commit_sha:",
		"^[0-9a-fA-F]{40}$",
		"ref: ${{ inputs.commit_sha }}",
		"git rev-parse HEAD",
		"git status --porcelain",
		"openspec validate --all --strict --no-interactive",
		"npm --prefix ui.web run build",
		"go test ./...",
		"@redocly/cli@latest lint docs/api/openapi.yaml",
		"@redocly/cli@latest build-docs docs/api/openapi.yaml -o docs/api/index.html",
		"./cypress.ps1 -Spec",
		"Upload release-candidate logs",
		"does not merge develop into main",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("release-candidate gate missing %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"pull_request:",
		"git push",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("release-candidate gate contains forbidden fragment %q", forbidden)
		}
	}
}
