package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave3ActionMatchHasNoUnresolvedUnmatchedActions(t *testing.T) {
	t.Parallel()

	reportPath := filepath.Join("..", "workdocs", "ui-action-spec-coverage-wave3-action-match.md")
	raw, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read wave3 action report: %v", err)
	}

	content := string(raw)
	sectionMarker := "## Unmatched actions (need spec text/ID refinement)"
	idx := strings.Index(content, sectionMarker)
	if idx == -1 {
		t.Fatalf("expected report section %q", sectionMarker)
	}

	section := content[idx+len(sectionMarker):]
	lines := strings.Split(section, "\n")
	unresolved := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			if strings.HasPrefix(strings.ToLower(trimmed), "- none") {
				continue
			}
			unresolved = append(unresolved, trimmed)
		}
	}

	if len(unresolved) > 0 {
		t.Fatalf("expected 0 unresolved unmatched controls; found %d first=%q", len(unresolved), unresolved[0])
	}
}
