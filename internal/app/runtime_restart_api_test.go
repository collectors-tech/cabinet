package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeShutdownEndpointAllowsLoopbackPost(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/shutdown", strings.NewReader(`{"reason":"restart"}`))
	req.RemoteAddr = "127.0.0.1:40123"
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	a.srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"reason":"restart"`) {
		t.Fatalf("expected restart reason in body, got %s", rr.Body.String())
	}
}

func TestRuntimeShutdownEndpointRejectsNonLoopbackRequests(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/shutdown", strings.NewReader(`{"reason":"restart"}`))
	req.RemoteAddr = "192.168.1.55:40123"
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	a.srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rr.Code, rr.Body.String())
	}
}
