package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProfileStorageSecretAndLicenseEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	storage := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/storage", nil, nil)
	if storage.Code != http.StatusOK {
		t.Fatalf("storage status=%d body=%s", storage.Code, storage.Body.String())
	}
	var st map[string]string
	if err := json.NewDecoder(storage.Body).Decode(&st); err != nil {
		t.Fatalf("decode storage: %v", err)
	}
	if strings.TrimSpace(st["db_path"]) == "" || strings.TrimSpace(st["media_dir"]) == "" {
		t.Fatalf("expected non-empty storage paths: %+v", st)
	}

	putSecret := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-test"}`), map[string]string{"Content-Type": "application/json"})
	if putSecret.Code != http.StatusOK {
		t.Fatalf("put secret status=%d body=%s", putSecret.Code, putSecret.Body.String())
	}
	getSecret := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/secrets?key=openai_api_key", nil, nil)
	if getSecret.Code != http.StatusOK {
		t.Fatalf("get secret status=%d body=%s", getSecret.Code, getSecret.Body.String())
	}

	putLicense := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/license", strings.NewReader(`{"license_json":"{\"tier\":\"pro\"}"}`), map[string]string{"Content-Type": "application/json"})
	if putLicense.Code != http.StatusOK {
		t.Fatalf("put license status=%d body=%s", putLicense.Code, putLicense.Body.String())
	}
	getLicense := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/license", nil, nil)
	if getLicense.Code != http.StatusOK {
		t.Fatalf("get license status=%d body=%s", getLicense.Code, getLicense.Body.String())
	}
}
