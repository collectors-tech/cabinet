package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonDynamicListTemplatesListAndApply(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	listResp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/list-templates", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list templates status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Templates []map[string]any `json:"templates"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(listPayload.Templates) == 0 {
		t.Fatalf("expected non-empty template list")
	}

	applyResp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/list-templates/apply", strings.NewReader(`{"template_id":"trade_binder","list_name":"Trade Binder"}`), map[string]string{"Content-Type": "application/json"})
	if applyResp.Code != http.StatusCreated {
		t.Fatalf("apply template status=%d body=%s", applyResp.Code, applyResp.Body.String())
	}
	var applyPayload map[string]any
	if err := json.NewDecoder(applyResp.Body).Decode(&applyPayload); err != nil {
		t.Fatalf("decode apply payload: %v", err)
	}
	if applyPayload["template_id"] != "trade_binder" {
		t.Fatalf("expected trade_binder template id, got %+v", applyPayload)
	}
}

func TestPokemonDynamicListTemplatesRejectInvalidTemplateID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/list-templates/apply", strings.NewReader(`{"template_id":"unknown-template"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid template, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invalid template payload: %v", err)
	}
	if payload["error"] != "invalid_template_id" {
		t.Fatalf("expected invalid_template_id, got %+v", payload)
	}
}
