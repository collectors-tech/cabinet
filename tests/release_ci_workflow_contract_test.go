package tests

import (
	"encoding/json"
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
		"cypress_specs_override:",
		"Validate fixed beta core Cypress acceptance pack",
		"node scripts/validate-beta-core-cypress-pack.mjs --manifest release/beta-core-cypress-pack.json",
		".logs/release-candidate/cypress-pack.json",
		"${{ steps.cypress_pack.outputs.specs }}",
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

func TestBetaCoreCypressPackManifestContract(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	manifestPath := filepath.Join(repoRoot, "release", "beta-core-cypress-pack.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read beta core Cypress pack manifest: %v", err)
	}

	var manifest struct {
		Version            int      `json:"version"`
		Issue              int      `json:"issue"`
		RequiredCategories []string `json:"required_categories"`
		Specs              []struct {
			Category string `json:"category"`
			Path     string `json:"path"`
		} `json:"specs"`
		ManualPackagedSteps []string `json:"manual_packaged_steps"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parse beta core Cypress pack manifest: %v", err)
	}
	if manifest.Version < 1 {
		t.Fatalf("manifest version must be positive, got %d", manifest.Version)
	}
	if manifest.Issue != 2055 {
		t.Fatalf("manifest should stay bound to #2055, got #%d", manifest.Issue)
	}

	required := map[string]bool{
		"login_profile":    false,
		"inventory":        false,
		"wishlist":         false,
		"collections":      false,
		"media":            false,
		"recovery":         false,
		"provider_handoff": false,
	}
	declared := map[string]bool{}
	for _, category := range manifest.RequiredCategories {
		declared[category] = true
	}
	for category := range required {
		if !declared[category] {
			t.Fatalf("manifest does not declare required category %s", category)
		}
	}

	seen := map[string]bool{}
	for _, spec := range manifest.Specs {
		if _, ok := required[spec.Category]; ok {
			required[spec.Category] = true
		}
		if seen[spec.Path] {
			t.Fatalf("duplicate Cypress spec in beta pack: %s", spec.Path)
		}
		seen[spec.Path] = true
		if !strings.HasPrefix(spec.Path, "cypress/e2e/") || !strings.HasSuffix(spec.Path, ".cy.ts") {
			t.Fatalf("spec path is not a Cypress e2e TypeScript spec: %s", spec.Path)
		}
		if _, err := os.Stat(filepath.Join(repoRoot, "ui.web", filepath.FromSlash(spec.Path))); err != nil {
			t.Fatalf("manifest spec %s is not present under ui.web: %v", spec.Path, err)
		}
	}
	for category, covered := range required {
		if !covered {
			t.Fatalf("beta core Cypress pack does not cover required category %s", category)
		}
	}
	for _, requiredStep := range []string{"#1869", "#1944", "#1945"} {
		found := false
		for _, step := range manifest.ManualPackagedSteps {
			if strings.Contains(step, requiredStep) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("manifest does not keep manual packaged/live step %s explicit", requiredStep)
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
