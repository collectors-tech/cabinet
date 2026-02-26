package app

import (
	"os"
	"strings"
	"testing"
)

func TestObservabilitySpecsAreSplitAndDefineTriggers(t *testing.T) {
	t.Parallel()

	removed := "../../openspec/specs/errors-logging-diagnostics/spec.md"
	if _, err := os.Stat(removed); err == nil {
		t.Fatalf("combined observability spec must be removed: %s", removed)
	}

	required := []string{
		"../../openspec/specs/errors/spec.md",
		"../../openspec/specs/logging/spec.md",
		"../../openspec/specs/diagnostics/spec.md",
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing required spec: %s", path)
		}
	}

	loggingBytes, err := os.ReadFile("../../openspec/specs/logging/spec.md")
	if err != nil {
		t.Fatalf("read logging spec: %v", err)
	}
	logging := string(loggingBytes)
	loggingMustContain := []string{
		"API request logging",
		"error logging",
		"WHEN",
		"THEN",
	}
	for _, token := range loggingMustContain {
		if !strings.Contains(logging, token) {
			t.Fatalf("logging spec missing token: %s", token)
		}
	}

	diagBytes, err := os.ReadFile("../../openspec/specs/diagnostics/spec.md")
	if err != nil {
		t.Fatalf("read diagnostics spec: %v", err)
	}
	diag := string(diagBytes)
	diagMustContain := []string{
		"Sentry",
		"user session",
		"opt-in",
	}
	for _, token := range diagMustContain {
		if !strings.Contains(strings.ToLower(diag), strings.ToLower(token)) {
			t.Fatalf("diagnostics spec missing token: %s", token)
		}
	}
}

