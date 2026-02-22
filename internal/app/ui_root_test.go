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
