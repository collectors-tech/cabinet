package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrimaryNavigationLinksExposeVisibleLabelsAndAccessibleCollapsedState(t *testing.T) {
	root := repoRoot(t)
	navGroupPath := filepath.Join(root, "ui.web", "src", "components", "layout", "nav-group.tsx")
	sourceBytes, err := os.ReadFile(navGroupPath)
	if err != nil {
		t.Fatalf("read nav group source: %v", err)
	}
	source := string(sourceBytes)

	if !strings.Contains(source, "aria-label={item.title}") {
		t.Fatalf("primary nav links must expose item.title as a stable aria-label")
	}
	if !strings.Contains(source, "data-testid={`sidebar-nav-label-${itemKey}`}") {
		t.Fatalf("expanded primary nav links must render the visible sidebar-nav-label span")
	}
	if !strings.Contains(source, "!isIconOnly ?") {
		t.Fatalf("primary nav links must still hide labels in collapsed desktop icon-only state")
	}
	if !strings.Contains(source, "justify-center") {
		t.Fatalf("primary nav links must center icon-only controls in their row")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if filepath.Base(wd) == "tests" {
		return filepath.Dir(wd)
	}
	return wd
}
