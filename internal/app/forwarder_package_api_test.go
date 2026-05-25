package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestForwarderPackageAPIUpsertsAndListsPackages(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := "{\"profile_id\":\"profile-api\",\"provider\":\"Stackry\",\"source\":\"manual\",\"external_package_id\":\"PKG-API-1\",\"status\":\"received\",\"weight_grams\":250,\"raw_payload\":{\"note\":\"manual entry\"}}"
	created := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if created.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	updated := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader("{\"profile_id\":\"profile-api\",\"provider\":\"stackry\",\"source\":\"manual\",\"external_package_id\":\"PKG-API-1\",\"status\":\"ready_to_ship\",\"weight_grams\":300}"), map[string]string{"Content-Type": "application/json"})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
	listed := doRequest(t, a, http.MethodGet, "/api/forwarding/packages?profile_id=profile-api&status=ready_to_ship", nil, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var payload struct {
		Packages []map[string]any
		Summary  map[string]int
	}
	if err := json.NewDecoder(listed.Body).Decode(&payload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if payload.Summary["count"] != 1 || len(payload.Packages) != 1 {
		t.Fatalf("expected one package summary, got %+v", payload)
	}
	pkg := payload.Packages[0]
	if pkg["provider"] != "stackry" || pkg["source"] != "manual" || pkg["external_package_id"] != "PKG-API-1" || pkg["status"] != "ready_to_ship" || int(pkg["weight_grams"].(float64)) != 300 {
		t.Fatalf("unexpected listed package: %+v", pkg)
	}
}

func TestForwarderPackageAPIRejectsInvalidImports(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader("{\"profile_id\":\"profile-api\",\"provider\":\"stackry\",\"source\":\"manual\",\"status\":\"received\"}"), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid package import to fail, status=%d body=%s", resp.Code, resp.Body.String())
	}
}
