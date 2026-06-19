package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestImplementedIntegrationTraceabilityIDsAreUnique(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	idPattern := regexp.MustCompile("^\\| `?([^`| ]+)`? \\|")
	seen := map[string]int{}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "| ") || !strings.Contains(line, "| implemented |") {
			continue
		}
		matches := idPattern.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}
		if !strings.HasPrefix(matches[1], "INTEGRATION-") {
			continue
		}
		seen[matches[1]]++
	}

	var duplicates []string
	for id, count := range seen {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	if len(duplicates) > 0 {
		t.Fatalf("implemented integration traceability IDs must be unique, found duplicates: %s", strings.Join(duplicates, ", "))
	}
}
