package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSpecScenariosRequireGivenWhenThen(t *testing.T) {
	t.Parallel()

	specFiles, err := filepath.Glob("../../openspec/specs/*/spec.md")
	if err != nil {
		t.Fatalf("glob spec files: %v", err)
	}
	if len(specFiles) == 0 {
		t.Fatal("no openspec spec files found")
	}

	var failures []string

	for _, file := range specFiles {
		b, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatalf("read %s: %v", file, readErr)
		}
		lines := strings.Split(string(b), "\n")

		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if !strings.HasPrefix(line, "#### Scenario:") {
				continue
			}

			hasGiven := false
			hasWhen := false
			hasThen := false

			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "#### Scenario:") || strings.HasPrefix(next, "### Requirement:") || strings.HasPrefix(next, "## ") {
					break
				}
				if strings.Contains(next, "**GIVEN**") {
					hasGiven = true
				}
				if strings.Contains(next, "**WHEN**") {
					hasWhen = true
				}
				if strings.Contains(next, "**THEN**") {
					hasThen = true
				}
			}

			if !hasGiven || !hasWhen || !hasThen {
				failures = append(failures, file+": "+line)
			}
		}
	}

	if len(failures) > 0 {
		t.Fatalf("scenarios missing GIVEN/WHEN/THEN: %v", failures)
	}
}

