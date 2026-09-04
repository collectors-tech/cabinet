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

func TestContinuousUIValidationPersistsRevisionStateAndPreventsOverlap(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", ".github", "workflows", "continuous-ui-validation.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"concurrency:",
		"group: continuous-ui-validation-${{ github.ref }}",
		"cancel-in-progress: false",
		"timeout-minutes: 240",
		"actions/cache/restore@v4",
		"id: hourly-state",
		".logs/hourly-ui-validation-state.json",
		"hourly-ui-validation-state-${{ runner.os }}-${{ github.run_id }}-${{ github.run_attempt }}",
		"restore-keys:",
		"actions/cache/save@v4",
		"if: always()",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected continuous UI validation workflow to contain fragment %q", fragment)
		}
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
		"cypress.ps1",
		"$cypressArgs",
		"gh issue create",
		"control_intent_results",
		"form_field_results",
		"intent_pass_count",
		"field_pass_count",
		"-RequireE2EHooks",
		"-ApiContractSmoke",
		"api_contract_smoke",
		"require_e2e_hooks",
		"allow_stale_runtime_version",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected hourly validation script to contain fragment %q", fragment)
		}
	}
}

func TestHourlyUIValidationOnlyMarksSuccessfulRunsValidated(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "hourly-ui-validation.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(raw)

	reportIndex := strings.Index(content, "Write-JsonFile $reportPath $reportPayload")
	if reportIndex < 0 {
		t.Fatalf("expected hourly validation to write a report before handling validation failure")
	}
	failureOffset := strings.Index(content[reportIndex:], "if ($hadFailures) {")
	failureIndex := reportIndex + failureOffset
	stateIndex := strings.Index(content, "$updatedState = [ordered]@")
	if failureOffset < 0 || stateIndex < 0 {
		t.Fatalf("expected hourly validation failure and state-persistence branches")
	}
	if stateIndex <= failureIndex {
		t.Fatalf("failed validation must not advance the successful validated revision before returning failure")
	}
}

func TestContinuousUIValidationUsesWritableRunScopedStateCache(t *testing.T) {
	t.Parallel()

	workflowPath := filepath.Join("..", ".github", "workflows", "continuous-ui-validation.yml")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	content := string(raw)

	if strings.Contains(content, "hourly-ui-validation-state-${{ runner.os }}-${{ github.sha }}") {
		t.Fatalf("hourly validation state cache must not use an immutable revision key")
	}
	for _, fragment := range []string{
		"hourly-ui-validation-state-${{ runner.os }}-${{ github.run_id }}-${{ github.run_attempt }}",
		"hourly-ui-validation-state-${{ runner.os }}-",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected writable run-scoped hourly validation state cache fragment %q", fragment)
		}
	}
}

func TestHourlyUIValidationReusesOneBoundedExactRuntime(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "hourly-ui-validation.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Start-HourlyValidationRuntime",
		"Stop-HourlyValidationRuntime",
		`"-ReuseServer"`,
		`"-SkipRuntimeBuild"`,
		`"-SkipDependencyPrep"`,
		`"-Retries", "0"`,
		`"-ExecutionTimeoutSec", "300"`,
		`"-LogDir", $reportDir`,
		`"-LogName", $logName`,
		"runtime_revision",
		"runner_phase",
		"execution_timed_out",
		"runner_failures",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected hourly validation exact-runtime contract fragment %q", fragment)
		}
	}
}

func TestHourlyUIValidationUsesCypressSafeLogNameLength(t *testing.T) {
	t.Parallel()

	hourlyPath := filepath.Join("..", "scripts", "hourly-ui-validation.ps1")
	hourlyRaw, err := os.ReadFile(hourlyPath)
	if err != nil {
		t.Fatalf("read hourly validation script: %v", err)
	}
	cypressPath := filepath.Join("..", "cypress.ps1")
	cypressRaw, err := os.ReadFile(cypressPath)
	if err != nil {
		t.Fatalf("read cypress wrapper script: %v", err)
	}
	hourly := string(hourlyRaw)
	cypress := string(cypressRaw)

	if !strings.Contains(cypress, "if ($safe.Length -gt 80)") {
		t.Fatalf("expected cypress wrapper to cap safe log segments at 80 characters")
	}
	if !strings.Contains(hourly, "if ($segment.Length -gt 80)") {
		t.Fatalf("hourly validation must use the cypress wrapper safe log segment length when later locating summary files")
	}
	for _, fragment := range []string{
		`$logNamePrefix = "hourly-{0:D3}-"`,
		"$safeSpecMaxLength = 80 - $logNamePrefix.Length",
		"$safeSpec = $safeSpec.Substring(0, $safeSpecMaxLength).Trim('-')",
		`$logName = ConvertTo-SafeLogSegment "$logNamePrefix$safeSpec"`,
	} {
		if !strings.Contains(hourly, fragment) {
			t.Fatalf("hourly validation must cap the full cypress LogName before summary lookup; missing %q", fragment)
		}
	}
	if strings.Contains(hourly, "if ($segment.Length -gt 120)") {
		t.Fatalf("hourly validation must not use a longer safe log segment than cypress.ps1")
	}
}

func TestHourlyUIValidationDeduplicatesSameRevisionIssues(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "hourly-ui-validation.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read script: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Find-ExistingValidationIssue",
		"- commit: $currentCommit",
		"Existing open validation issue",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected hourly validation issue-deduplication fragment %q", fragment)
		}
	}
}

func TestHourlyUIValidationRunnerLifecycleIsTraceable(t *testing.T) {
	t.Parallel()

	specPath := filepath.Join("..", "openspec", "specs", "general", "continuous-ui-validation", "spec.md")
	specRaw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read continuous UI validation spec: %v", err)
	}
	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	for _, fragment := range []string{
		"CONT-UI-CAB-013",
		"one workflow-built exact runtime",
		"zero retries",
		"runner failure",
		"same-revision",
	} {
		if !strings.Contains(string(specRaw), fragment) {
			t.Fatalf("expected continuous UI validation lifecycle requirement fragment %q", fragment)
		}
	}
	for _, fragment := range []string{
		"CONT-UI-CAB-013",
		"#2307",
		"#2475",
		"TestHourlyUIValidationReusesOneBoundedExactRuntime",
		"TestHourlyUIValidationDeduplicatesSameRevisionIssues",
		"TestHourlyUIValidationOnlyMarksSuccessfulRunsValidated",
	} {
		if !strings.Contains(string(traceRaw), fragment) {
			t.Fatalf("expected continuous UI validation traceability fragment %q", fragment)
		}
	}
}
