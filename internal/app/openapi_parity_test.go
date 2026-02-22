package app

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestOpenAPIDocumentsOnboardingSampleDataEndpoint(t *testing.T) {
	t.Parallel()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	specPath := filepath.Clean(filepath.Join(root, "..", "..", "docs", "api", "openapi.yaml"))
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}

	pathPattern := regexp.MustCompile(`(?m)^  /api/onboarding/sample-data:$`)
	if !pathPattern.Match(raw) {
		t.Fatalf("openapi missing /api/onboarding/sample-data path in %s", specPath)
	}
}

