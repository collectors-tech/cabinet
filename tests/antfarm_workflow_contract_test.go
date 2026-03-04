package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func resolveRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
}

func TestAntFarmCabinetWorkflowIsRepoLocalAndSelfContained(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)

	workflowPath := filepath.Join(repoRoot, ".antfarm", "workflows", "cabinet", "workflow.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow file: %v", err)
	}
	content := string(data)

	requiredRefs := []string{
		"baseDir: agents/planner",
		"baseDir: agents/setup",
		"baseDir: agents/developer",
		"baseDir: agents/verifier",
		"baseDir: agents/tester",
		"baseDir: agents/reviewer",
	}
	for _, required := range requiredRefs {
		if !strings.Contains(content, required) {
			t.Fatalf("workflow missing required local agent reference %q", required)
		}
	}

	for _, forbidden := range []string{
		"shared-agents",
		"/Users/",
		"C:\\",
		"\\\\",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("workflow includes forbidden external/shared path reference %q", forbidden)
		}
	}
}

func TestAntFarmCabinetMetadataMatchesLocalWorkflow(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	metadataPath := filepath.Join(repoRoot, ".antfarm", "workflows", "cabinet", "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("read metadata file: %v", err)
	}
	var payload struct {
		WorkflowID string `json:"workflowId"`
		Source     string `json:"source"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if payload.WorkflowID != "cabinet" {
		t.Fatalf("expected metadata workflowId=cabinet, got %q", payload.WorkflowID)
	}
	if payload.Source != "bundled:cabinet" {
		t.Fatalf("expected metadata source=bundled:cabinet, got %q", payload.Source)
	}
}

func TestAntFarmCabinetWorkflowRoleProfilesAreComplete(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	base := filepath.Join(repoRoot, ".antfarm", "workflows", "cabinet", "agents")
	roles := []string{"planner", "setup", "developer", "verifier", "tester", "reviewer"}
	requiredFiles := []string{"AGENTS.md", "SOUL.md", "IDENTITY.md"}

	for _, role := range roles {
		role := role
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			for _, name := range requiredFiles {
				path := filepath.Join(base, role, name)
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("required profile file missing (%s): %v", path, err)
				}
				if info.IsDir() {
					t.Fatalf("expected file, got directory: %s", path)
				}
			}
		})
	}
}
