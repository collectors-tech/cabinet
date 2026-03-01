package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUICopyExcludesTemplatePlaceholders(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))

	banned := []string{
		"Cabinet admin workspace UI template.",
		"Shadcn Admin",
		"Shadcn-Admin",
		"Vite + ShadcnUI",
		"Lorem ipsum",
	}

	paths := []string{
		filepath.Join(repoRoot, "ui.web", "index.html"),
	}

	srcRoot := filepath.Join(repoRoot, "ui.web", "src")
	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".ts", ".tsx", ".html":
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk ui sources: %v", err)
	}

	for _, path := range paths {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		content := string(data)
		for _, phrase := range banned {
			if strings.Contains(content, phrase) {
				t.Fatalf("banned placeholder copy %q found in %s", phrase, path)
			}
		}
	}
}
