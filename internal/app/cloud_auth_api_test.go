package app

import (
	"net/http"
	"strings"
	"testing"
)

func TestCloudSessionBootstrapReturnsEntitlement(t *testing.T) {
	a := newTestApp(t)

	// header.payload.sig where payload is {"sub":"user_123","email":"owner@example.com","plan":"pro"}
	token := "e30.eyJzdWIiOiJ1c2VyXzEyMyIsImVtYWlsIjoib3duZXJAZXhhbXBsZS5jb20iLCJwbGFuIjoicHJvIn0.e30"
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"`+token+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)

	if resp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, needle := range []string{`"user_id":"user_123"`, `"email":"owner@example.com"`, `"plan":"pro"`, `"ai_assist"`} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected %q in body %s", needle, body)
		}
	}
}
