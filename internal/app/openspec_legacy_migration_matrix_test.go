package app

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLegacyDocsMigrationMatrixCoversAllHistoricMarkdownFiles(t *testing.T) {
	t.Parallel()

	const baselineCommit = "82294546bf0b715fe49394e1c5a885d3045294d2"
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", baselineCommit)
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("list docs tree from %s: %v", baselineCommit, err)
	}

	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasSuffix(line, ".md") || !strings.HasPrefix(line, "docs/") {
			continue
		}
		files = append(files, line)
	}
	if len(files) == 0 {
		t.Fatalf("no markdown files found in baseline commit %s", baselineCommit)
	}

	b, err := os.ReadFile("../../openspec/migrations/legacy-docs-file-audit.yaml")
	if err != nil {
		t.Fatalf("read migration audit file: %v", err)
	}
	audit := string(b)

	for _, f := range files {
		token := "source: " + f
		if !strings.Contains(audit, token) {
			t.Fatalf("migration audit missing file mapping: %s", f)
		}
	}

	if !strings.Contains(audit, "status: migrated") {
		t.Fatal("migration audit missing migrated status entries")
	}
	if !strings.Contains(audit, "status: reference_only") {
		t.Fatal("migration audit missing reference_only status entries")
	}
	if bytes.Count(b, []byte("targets:")) < len(files) {
		t.Fatalf("migration audit must include targets for every source file")
	}
}
