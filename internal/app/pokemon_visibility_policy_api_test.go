package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPokemonVisibilityPrivateAnonymousDenied(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=private&actor=anonymous", nil, nil)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 status, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["error"] != "visibility_forbidden" {
		t.Fatalf("expected visibility_forbidden error, got %+v", payload)
	}
	if payload["required"] != "authenticated" {
		t.Fatalf("expected required=authenticated, got %+v", payload)
	}
}

func TestPokemonVisibilitySharedLinkRequiresToken(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=shared_link&actor=anonymous", nil, nil)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 status, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["error"] != "missing_share_token" {
		t.Fatalf("expected missing_share_token error, got %+v", payload)
	}

	allowed := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=shared_link&actor=anonymous&share_token=abc123", nil, nil)
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected 200 status with token, got=%d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestPokemonVisibilityTeamRequiresTeamMemberActor(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	denied := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=team&actor=authenticated", nil, nil)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("expected 403 status, got=%d body=%s", denied.Code, denied.Body.String())
	}
	allowed := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=team&actor=team_member", nil, nil)
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got=%d body=%s", allowed.Code, allowed.Body.String())
	}
}

func TestPokemonVisibilityRejectsInvalidPolicy(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/visibility-access?visibility=unknown&actor=anonymous", nil, nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["error"] != "invalid_visibility" {
		t.Fatalf("expected invalid_visibility error, got %+v", payload)
	}
}
