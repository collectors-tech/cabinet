package app

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/ui"
)

func TestRootServesAppShell(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	body := resp.Body.String()
	if strings.Contains(body, "Application runtime is active.") {
		t.Fatalf("expected scaffold copy to be removed from root")
	}
	if !strings.Contains(body, "id=\"root\"") {
		t.Fatalf("expected SPA mount node in root page")
	}
	if !strings.Contains(body, "<script type=\"module\"") {
		t.Fatalf("expected SPA entry script in root page")
	}
	if !strings.Contains(body, "<title>Cabinet</title>") {
		t.Fatalf("expected Cabinet title in root page")
	}
	if strings.Contains(body, "UI template") {
		t.Fatalf("expected template placeholder copy to be removed from root page metadata")
	}
}

func TestSPADeepLinksServeAppShell(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/_authenticated/inventory/", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	body := resp.Body.String()
	if !strings.Contains(body, "id=\"root\"") {
		t.Fatalf("expected SPA mount node in deep-link response")
	}
}

func TestUnderscorePrefixedUIChunksServeFromEmbeddedStatic(t *testing.T) {
	t.Parallel()

	matches, err := fs.Glob(ui.Static, "static/assets/_*.js")
	if err != nil {
		t.Fatalf("glob embedded underscore chunks: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected embedded underscore-prefixed UI chunks")
	}

	a := newTestApp(t)
	assetPath := "/" + strings.TrimPrefix(matches[0], "static/")
	resp := doRequest(t, a, http.MethodGet, assetPath, nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want %d", assetPath, resp.Code, http.StatusOK)
	}
	if !strings.Contains(resp.Body.String(), "export") {
		t.Fatalf("expected %s to return a JavaScript module", assetPath)
	}
}

func TestAPIDocsRoutes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	docs := doRequest(t, a, http.MethodGet, "/apidocs", nil, nil)
	if docs.Code != http.StatusOK {
		t.Fatalf("apidocs status = %d, want %d", docs.Code, http.StatusOK)
	}
	if !strings.Contains(docs.Body.String(), "Cabinet API Docs") {
		t.Fatalf("expected docs page content in /apidocs")
	}

	legacy := doRequest(t, a, http.MethodGet, "/redoc.html", nil, nil)
	if legacy.Code != http.StatusMovedPermanently {
		t.Fatalf("redoc redirect status = %d, want %d", legacy.Code, http.StatusMovedPermanently)
	}
	if location := legacy.Header().Get("Location"); location != "/apidocs" {
		t.Fatalf("redoc redirect location = %q, want %q", location, "/apidocs")
	}

	spec := doRequest(t, a, http.MethodGet, "/api/openapi.yaml", nil, nil)
	if spec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", spec.Code, http.StatusOK)
	}
	if !strings.Contains(spec.Body.String(), "openapi:") {
		t.Fatalf("expected openapi yaml payload from /api/openapi.yaml")
	}
}

func TestAPIDocsSpecLoadsWhenStartedOutsideRepoRoot(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	temp := t.TempDir()
	if err := os.Chdir(temp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	a := newTestApp(t)
	spec := doRequest(t, a, http.MethodGet, "/api/openapi.yaml", nil, nil)
	if spec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d body=%s", spec.Code, http.StatusOK, spec.Body.String())
	}
	if !strings.Contains(spec.Body.String(), "openapi:") {
		t.Fatalf("expected openapi yaml payload from /api/openapi.yaml")
	}
}
