package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProviderFamilyDetectReturnsFamilyConfidenceAndEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/family-detect",
		strings.NewReader(`{
  "provider_url":"https://example-woo.test",
  "html":"<script>var woo='woocommerce';</script><a href='/wp-json/wc/store/v1/products'>api</a>"
}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("family detect status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		ProposedAPIFamily string   `json:"proposed_api_family"`
		Confidence        float64  `json:"confidence"`
		Evidence          []string `json:"evidence"`
		ProviderDomain    string   `json:"provider_domain"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.ProposedAPIFamily != "woocommerce" {
		t.Fatalf("expected proposed_api_family=woocommerce got=%q", payload.ProposedAPIFamily)
	}
	if payload.Confidence <= 0 {
		t.Fatalf("expected positive confidence, got=%f", payload.Confidence)
	}
	if len(payload.Evidence) == 0 {
		t.Fatal("expected non-empty evidence markers")
	}
	if payload.ProviderDomain != "example-woo.test" {
		t.Fatalf("expected provider_domain=example-woo.test got=%q", payload.ProviderDomain)
	}
}

func TestProviderFamilyOverridePersistsIntoRegistry(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	override := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/family-override",
		strings.NewReader(`{"provider_domain":"frontlinehobbies.com.au","api_family":"doofinder"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if override.Code != http.StatusOK {
		t.Fatalf("override status=%d body=%s", override.Code, override.Body.String())
	}
	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	found := false
	for _, provider := range payload.Providers {
		if strings.EqualFold(strings.TrimSpace(provider["base_domain"].(string)), "frontlinehobbies.com.au") {
			found = true
			if strings.TrimSpace(provider["api_family"].(string)) != "doofinder" {
				t.Fatalf("expected override api_family=doofinder got=%v", provider["api_family"])
			}
		}
	}
	if !found {
		t.Fatal("expected frontlinehobbies.com.au provider in registry")
	}
}

