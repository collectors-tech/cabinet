package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSpecCapabilitiesDoNotUseCombinedAndNaming(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("../../openspec/specs")
	if err != nil {
		t.Fatalf("read openspec specs dir: %v", err)
	}

	var bad []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		specPath := filepath.Join("../../openspec/specs", name, "spec.md")
		if _, statErr := os.Stat(specPath); statErr != nil {
			continue
		}
		if strings.Contains(name, "-and-") {
			bad = append(bad, name)
		}
	}

	if len(bad) > 0 {
		t.Fatalf("combined capability names are not allowed: %v", bad)
	}
}

