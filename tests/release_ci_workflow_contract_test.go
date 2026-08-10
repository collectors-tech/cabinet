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
		"Validate Docker Compose deployments",
		"bash scripts/validate-compose-deployments.sh",
		"Verify ZITADEL application auth contracts",
		"npm run test:auth-infra",
		"openspec validate --all --strict --no-interactive",
		"npm run build",
		"node scripts/verify-browser-extension-load.mjs",
		"fetch-depth: 0",
		"go test ./...",
		"Build runtime UI static bundle for OpenAPI tests",
		"go run ./cmd/openapi-parity-gate",
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
		"go test ./internal/... ./cmd/...",
		"git push",
		"merge develop",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("develop quality gate contains forbidden fragment %q", forbidden)
		}
	}
}

func TestReleaseWorkflowsRunCompleteGoRepositorySuite(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	for _, relativePath := range []string{
		filepath.Join(".github", "workflows", "develop-quality-gate.yml"),
		filepath.Join(".github", "workflows", "beta-release-candidate.yml"),
		filepath.Join(".github", "workflows", "release-installers.yml"),
		filepath.Join(".github", "workflows", "main-gate.yml"),
	} {
		relativePath := relativePath
		t.Run(relativePath, func(t *testing.T) {
			t.Parallel()

			raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
			if err != nil {
				t.Fatalf("read %s: %v", relativePath, err)
			}
			content := string(raw)
			if !strings.Contains(content, "go test ./...") {
				t.Fatalf("%s does not run the complete Go repository suite", relativePath)
			}
			if strings.Contains(content, "go test ./internal/... ./cmd/...") {
				t.Fatalf("%s still omits the root contract package", relativePath)
			}
		})
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

func TestReleaseWorkflowsUseVerifiedOpenAPIParityGate(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	for _, relativePath := range []string{
		filepath.Join(".github", "workflows", "develop-quality-gate.yml"),
		filepath.Join(".github", "workflows", "beta-release-candidate.yml"),
		filepath.Join(".github", "workflows", "main-gate.yml"),
	} {
		relativePath := relativePath
		t.Run(relativePath, func(t *testing.T) {
			t.Parallel()

			raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
			if err != nil {
				t.Fatalf("read %s: %v", relativePath, err)
			}
			content := string(raw)
			if !strings.Contains(content, "go run ./cmd/openapi-parity-gate") {
				t.Fatalf("%s does not invoke the verified OpenAPI parity gate", relativePath)
			}
			if strings.Contains(content, "TestOpenAPIDocumentsOnboardingSampleDataEndpoint") {
				t.Fatalf("%s still invokes the nonexistent OpenAPI test target", relativePath)
			}
		})
	}
}
