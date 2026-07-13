package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContinuousUIValidationWorkflowContract(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", ".github", "workflows", "continuous-ui-validation.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "cron: '0 * * * *'") {
		t.Fatalf("expected hourly cron schedule in continuous-ui-validation workflow")
	}
	if !strings.Contains(content, "workflow_dispatch:") {
		t.Fatalf("expected manual dispatch trigger in continuous-ui-validation workflow")
	}
	if !strings.Contains(content, "scripts/hourly-ui-validation.ps1") {
		t.Fatalf("expected workflow to execute scripts/hourly-ui-validation.ps1")
	}
}

func TestContinuousUIValidationUsesCanonicalCabinetBuild(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", ".github", "workflows", "continuous-ui-validation.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	content := string(raw)

	buildIndex := strings.Index(content, "scripts/build-cabinet.ps1")
	validationIndex := strings.Index(content, "scripts/hourly-ui-validation.ps1")
	if buildIndex < 0 {
		t.Fatalf("continuous UI validation must use scripts/build-cabinet.ps1 so UI assets are built before the Go runtime")
	}
	if validationIndex < 0 {
		t.Fatalf("expected workflow to execute scripts/hourly-ui-validation.ps1")
	}
	if buildIndex > validationIndex {
		t.Fatalf("continuous UI validation must build Cabinet before running hourly validation")
	}
	if strings.Contains(content, "go build -o bin/cabinet.exe ./cmd/cabinet") {
		t.Fatalf("continuous UI validation must not bypass the canonical Cabinet build entrypoint")
	}
}

func TestHourlyUIValidationScriptContract(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "hourly-ui-validation.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"no-change",
		"last_validated_version",
		"ui.web/cypress/e2e",
		"pwsh -File ./cypress.ps1",
		"gh issue create",
		"control_intent_results",
		"form_field_results",
		"intent_pass_count",
		"field_pass_count",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected hourly validation script to contain fragment %q", fragment)
		}
	}
}
