package app

import (
	"net/http"
	"strings"
	"testing"
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
