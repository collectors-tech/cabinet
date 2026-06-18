package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCabinetBrandIconAssetsAndMetadata(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))

	indexHTML := readFile(t, filepath.Join(repoRoot, "ui.web", "index.html"))
	requiredIndexRefs := []string{
		`href="/images/favicon.svg"`,
		`href="/images/favicon_light.svg"`,
		`href="/images/favicon.ico"`,
		`href="/images/apple-touch-icon.png"`,
		`href="/images/site.webmanifest"`,
	}
	for _, ref := range requiredIndexRefs {
		if !strings.Contains(indexHTML, ref) {
			t.Fatalf("ui.web/index.html missing Cabinet icon metadata reference %s", ref)
		}
	}

	logoTSX := readFile(t, filepath.Join(repoRoot, "ui.web", "src", "assets", "logo.tsx"))
	if !strings.Contains(logoTSX, `viewBox='0 0 511.26 529.94'`) {
		t.Fatal("ui.web/src/assets/logo.tsx should render the supplied Cabinet icon geometry")
	}
	if strings.Contains(logoTSX, `M15 6v12`) {
		t.Fatal("ui.web/src/assets/logo.tsx still contains the prior template icon path")
	}

	publicImages := filepath.Join(repoRoot, "ui.web", "public", "images")
	staticImages := filepath.Join(repoRoot, "internal", "ui", "static", "images")

	for _, imageRoot := range []string{publicImages, staticImages} {
		assertCabinetIconSVG(t, filepath.Join(imageRoot, "favicon.svg"), false)
		assertCabinetIconSVG(t, filepath.Join(imageRoot, "favicon_light.svg"), true)
		assertNonEmptyFile(t, filepath.Join(imageRoot, "favicon.ico"))
		assertNonEmptyFile(t, filepath.Join(imageRoot, "favicon-96x96.png"))
		assertNonEmptyFile(t, filepath.Join(imageRoot, "apple-touch-icon.png"))
		assertNonEmptyFile(t, filepath.Join(imageRoot, "web-app-manifest-192x192.png"))
		assertNonEmptyFile(t, filepath.Join(imageRoot, "web-app-manifest-512x512.png"))

		manifest := readFile(t, filepath.Join(imageRoot, "site.webmanifest"))
		for _, token := range []string{"Cabinet", "web-app-manifest-192x192.png", "web-app-manifest-512x512.png"} {
			if !strings.Contains(manifest, token) {
				t.Fatalf("%s missing expected manifest token %q", filepath.Join(imageRoot, "site.webmanifest"), token)
			}
		}
	}
}

func assertCabinetIconSVG(t *testing.T, path string, darkVariant bool) {
	t.Helper()

	content := readFile(t, path)
	if !strings.Contains(content, `viewBox="0 0 511.26 529.94"`) {
		t.Fatalf("%s does not contain the supplied Cabinet icon viewBox", path)
	}
	if !strings.Contains(content, "#0E4A4A") {
		t.Fatalf("%s does not contain the supplied Cabinet brand color", path)
	}
	hasBackground := strings.Contains(content, `<rect width="511.26" height="529.94" fill="#0E4A4A"/>`)
	if darkVariant && !hasBackground {
		t.Fatalf("%s should contain the dark variant background rectangle", path)
	}
	if !darkVariant && hasBackground {
		t.Fatalf("%s should be the light variant without the dark background rectangle", path)
	}
}

func assertNonEmptyFile(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		t.Fatalf("%s is empty", path)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
