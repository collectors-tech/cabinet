package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHelpCenterDocumentationSetIsComplete(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))

	requiredDocs := []string{
		"docs/help-center/README.md",
		"docs/help-center/getting-started/login-and-database-setup.md",
		"docs/help-center/sections/inventory.md",
		"docs/help-center/sections/wishlist.md",
		"docs/help-center/sections/collections.md",
		"docs/help-center/sections/integrations.md",
		"docs/help-center/sections/settings.md",
		"docs/help-center/sections/chats.md",
		"docs/help-center/ui-elements.md",
	}

	for _, relative := range requiredDocs {
		path := filepath.Join(repoRoot, filepath.FromSlash(relative))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("required help center document missing: %s (%v)", relative, err)
		}
	}
}

func TestGettingStartedGuideCoversLoginDatabaseAndProfile(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	path := filepath.Join(repoRoot, "docs", "help-center", "getting-started", "login-and-database-setup.md")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read getting-started guide: %v", err)
	}
	content := strings.ToLower(string(data))

	for _, requiredTerm := range []string{"login", "database", "profile"} {
		if !strings.Contains(content, requiredTerm) {
			t.Fatalf("getting-started guide must mention %q", requiredTerm)
		}
	}
}

func TestUIElementsGuideDocumentsSharedControls(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	path := filepath.Join(repoRoot, "docs", "help-center", "ui-elements.md")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ui-elements guide: %v", err)
	}
	content := strings.ToLower(string(data))

	for _, requiredTerm := range []string{"new", "create", "filter", "row", "toast"} {
		if !strings.Contains(content, requiredTerm) {
			t.Fatalf("ui-elements guide must mention %q behavior", requiredTerm)
		}
	}
}

