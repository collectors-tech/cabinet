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
		"actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7.0.1",
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
		"Build controlled Telegram Bot API fixture",
		"$fixtureExecutable = Join-Path $env:RUNNER_TEMP \"cabinet-telegram-test-fixture.exe\"",
		"go build -o $fixtureExecutable ./cmd/telegram-test-fixture",
		"$telegramFixtureSpecs = @(",
		"if ($telegramFixtureSpecs -contains $spec)",
		"if (-not $telegramFixtureStarted)",
		`$telegramFixtureStatusURL = "$($env:CYPRESS_telegramFixtureControlURL.TrimEnd('/'))/control/status"`,
		"Start-Process -FilePath $fixtureExecutable",
		"Controlled Telegram fixture is unavailable before ${spec}",
		"CABINET_TELEGRAM_TEST_API_BASE_URL",
		"CYPRESS_telegramRuntimeFixture: \"true\"",
		"if ($env:CYPRESS_telegramRuntimeFixture -ne \"true\"",
		"Controlled Telegram fixture flag is required; skipping fixture-controlled specs is forbidden.",
		"$pack.version -lt 6 -or $pack.spec_count -ne 26 -or $pack.specs.Count -ne 26",
		"Fixed beta Cypress pack must resolve version 6 with exactly 26 specs.",
		"timeout-minutes: 30",
		"go test ./... -count=1 -p 1 -parallel 4 -timeout 900s",
		"$name = Split-Path (Split-Path $spec -Parent) -Leaf",
		"$summaryPaths.Count -ne 1",
		"Candidate Cypress must produce exactly one summary",
		"-Retries 0",
		"-ExecutionTimeoutSec 300",
		"-RequireE2EHooks",
		"-LogDir .logs/release-candidate/cypress",
		"$summary.runtime_revision -ne \"${{ steps.commit.outputs.sha }}\"",
		"$summary.execution_timeout_sec -ne 300",
		"$summary.runner_phase -ne \"completed\"",
		".logs/release-candidate/telegram-fixture.pid",
		"[int]::TryParse($rawFixturePID, [ref]$fixturePID)",
		"Get-Process -Id $fixturePID",
		"$ownedFixture.Path -eq $fixtureExecutable",
		"Stop-Process -Id $fixturePID -Force",
		"foreach ($spec in $pack.specs)",
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
		"Get-NetTCPConnection -LocalAddress 127.0.0.1 -LocalPort 17994",
		"Start-Process -FilePath go",
		`"run", "./cmd/telegram-test-fixture"`,
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("release-candidate gate contains forbidden fragment %q", forbidden)
		}
	}

	loopIndex := strings.Index(content, "foreach ($spec in $pack.specs)")
	telegramGuardIndex := strings.Index(content, "if ($telegramFixtureSpecs -contains $spec)")
	fixtureStartIndex := strings.Index(content, "Start-Process -FilePath $fixtureExecutable")
	cypressRunIndex := strings.Index(content, "pwsh -NoLogo -NoProfile -File ./cypress.ps1 -Spec $spec")
	if loopIndex < 0 || telegramGuardIndex < loopIndex || fixtureStartIndex < telegramGuardIndex || cypressRunIndex < fixtureStartIndex {
		t.Fatal("controlled Telegram fixture must start inside the pack loop immediately before the Telegram specs")
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
	if manifest.Version < 6 {
		t.Fatalf("renderer-bounded Market Watch acceptance pack must be version 6 or newer, got %d", manifest.Version)
	}
	if len(manifest.Specs) != 26 {
		t.Fatalf("renderer-bounded Market Watch acceptance pack must contain exactly 26 specs, got %d", len(manifest.Specs))
	}

	required := map[string]bool{
		"login_profile":         false,
		"inventory":             false,
		"wishlist":              false,
		"collections":           false,
		"media":                 false,
		"recovery":              false,
		"provider_handoff":      false,
		"provider_auth":         false,
		"agent_primary":         false,
		"agent_authority":       false,
		"telegram_connector":    false,
		"telegram_conversation": false,
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
	for requiredPath, requiredCategory := range map[string]string{
		"cypress/e2e/integrations/ui-screen-market-watch/01-core.cy.ts":       "provider_handoff",
		"cypress/e2e/integrations/ui-screen-market-watch/02-create-run.cy.ts": "provider_handoff",
		"cypress/e2e/integrations/ui-screen-market-watch/03-results.cy.ts":    "provider_handoff",
		"cypress/e2e/integrations/ui-screen-market-watch/04-handoffs.cy.ts":   "provider_handoff",
		"cypress/e2e/integrations/provider-openai-chatgpt-ux/spec.cy.ts":      "provider_auth",
		"cypress/e2e/chats/agent-attachment-continuity/spec.cy.ts":            "agent_primary",
		"cypress/e2e/chats/agent-response-state-matrix/spec.cy.ts":            "agent_primary",
		"cypress/e2e/chats/agent-compact-accessibility/spec.cy.ts":            "agent_primary",
		"cypress/e2e/chats/cabinet-agent-collection-workflows/spec.cy.ts":     "agent_primary",
		"cypress/e2e/chats/assistant-acquisition-workflows/spec.cy.ts":        "agent_primary",
		"cypress/e2e/chats/assistant-workspace-dashboard-summary/spec.cy.ts":  "agent_primary",
		"cypress/e2e/chats/chat-integration-management/spec.cy.ts":            "agent_primary",
	} {
		foundCategory := ""
		for _, spec := range manifest.Specs {
			if spec.Path == requiredPath {
				foundCategory = spec.Category
				break
			}
		}
		if foundCategory != requiredCategory {
			t.Fatalf("expanded #2091 Agent acceptance spec %s must use category %s, got %q", requiredPath, requiredCategory, foundCategory)
		}
	}
	for category, covered := range required {
		if !covered {
			t.Fatalf("beta core Cypress pack does not cover required category %s", category)
		}
	}
	for _, requiredStep := range []string{"#1869", "#1944", "#1945", "#1716", "#1773"} {
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
