package app

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/companion"
	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestBonzaProviderRegistryAdvertisesBrowserCompanionFallbackTruthfully(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	response := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("provider registry status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode provider registry: %v", err)
	}
	provider := findRegistryProvider(payload.Providers, "au-webshop-bonzaslotcars-com-au")
	if provider == nil {
		t.Fatalf("Bonza provider missing from registry: %+v", payload.Providers)
	}
	for key, want := range map[string]string{
		"market_watch_scope":      "bonzaslotcars",
		"fallback_state":          "browser_companion_user_present",
		"browser_companion_state": "available_when_paired",
		"direct_search_state":     "best_effort_fail_closed",
		"live_evidence_state":     "external_user_present_evidence_required",
	} {
		if got := provider[key]; got != want {
			t.Fatalf("Bonza registry %s=%v want %q: %+v", key, got, want, provider)
		}
	}
	capabilities, ok := provider["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("Bonza capabilities got %T: %+v", provider["capabilities"], provider)
	}
	for key, want := range map[string]bool{
		"browser_companion": true, "user_present_search": true,
		"unattended_search": false, "challenge_bypass": false,
	} {
		if got := capabilities[key]; got != want {
			t.Fatalf("Bonza capability %s=%v want %t: %+v", key, got, want, capabilities)
		}
	}
	instructions := strings.ToLower(provider["setup_instructions"].(string))
	for _, token := range []string{"browser companion", "user", "challenge", "cookie"} {
		if !strings.Contains(instructions, token) {
			t.Fatalf("Bonza setup instructions omit %q: %s", token, instructions)
		}
	}
}

func TestBonzaCompleteCompanionCaptureProducesScopedLiveEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profiles := profile.NewRepository(a.db)
	created, err := profiles.Create(context.Background(), "Bonza Companion")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("set active profile: %v", err)
	}
	instance, err := profiles.UpsertIntegrationInstance(context.Background(), created.ID, profile.IntegrationInstancePatch{
		ProviderID: "au-webshop-bonzaslotcars-com-au",
	})
	if err != nil {
		t.Fatalf("upsert Bonza integration: %v", err)
	}
	if _, err := a.authService.CreateUnlockedSession(created.ID); err != nil {
		t.Fatalf("unlock profile: %v", err)
	}

	receipt := requestCompanionPairing(t, a, []string{companion.CapabilityModulesRead, companion.CapabilityCapturesSubmit})
	approved := doCompanionManagementRequest(t, a, http.MethodPost, "/api/companion/pairing/approvals", strings.NewReader(
		`{"request_id":"`+receipt.RequestID+`","profile_id":"`+created.ID+`"}`,
	), nil)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve Bonza pairing status=%d body=%s", approved.Code, approved.Body.String())
	}
	exchanged := exchangeCompanionPairing(t, a, receipt, true)
	var credential companion.CredentialResponse
	if err := json.NewDecoder(exchanged.Body).Decode(&credential); err != nil {
		t.Fatalf("decode companion exchange: %v", err)
	}
	authorization := "Bearer " + credential.Credential
	modules := doCompanionExtensionRequest(t, a, http.MethodGet, "/api/companion/modules", nil, map[string]string{"Authorization": authorization})
	if modules.Code != http.StatusOK || !strings.Contains(modules.Body.String(), `"id":"bonzaslotcars-search-capture"`) ||
		!strings.Contains(modules.Body.String(), `"integration_instance_id":"`+instance.ID+`"`) ||
		!strings.Contains(modules.Body.String(), `"sync_available":true`) {
		t.Fatalf("Bonza module projection status=%d body=%s", modules.Code, modules.Body.String())
	}

	rawFixture, err := os.ReadFile(filepath.Join("..", "companion", "testdata", "bonza-search-results-v1.json"))
	if err != nil {
		t.Fatalf("read Bonza fixture: %v", err)
	}
	data := map[string]any{}
	if err := json.Unmarshal(rawFixture, &data); err != nil {
		t.Fatalf("decode Bonza fixture: %v", err)
	}
	data["complete"] = true
	data["total_pages"] = float64(1)
	payload := companion.PayloadSubmission{
		ProtocolVersion: companion.ProtocolVersionV1, ProfileID: created.ID, ModuleID: "bonzaslotcars-search-capture",
		ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instance.ID,
		ProviderID: "au-webshop-bonzaslotcars-com-au", URL: "https://www.bonzaslotcars.com.au/?s=Scalextric",
		PayloadType: "search_results", CapturedAt: time.Now().UTC().Format(time.RFC3339), PageComplete: true,
		Passive: true, ConfidenceScore: 0.95, RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"},
		PayloadHash: companion.PayloadDigest(data), IdempotencyKey: "bonza-app-live-proof", Data: data,
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal Bonza payload: %v", err)
	}
	accepted := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(string(rawPayload)), map[string]string{
		"Authorization": authorization, "Content-Type": "application/json", "X-Cabinet-Idempotency-Key": payload.IdempotencyKey,
	})
	if accepted.Code != http.StatusAccepted || !strings.Contains(accepted.Body.String(), `"committed":true`) {
		t.Fatalf("Bonza capture status=%d body=%s", accepted.Code, accepted.Body.String())
	}
	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("Bonza post-capture registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var registryPayload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&registryPayload); err != nil {
		t.Fatalf("decode Bonza post-capture registry: %v", err)
	}
	bonza := findRegistryProvider(registryPayload.Providers, "au-webshop-bonzaslotcars-com-au")
	if bonza == nil || bonza["beta_release_status"] != "available_live_validated" || bonza["live_evidence_state"] != "validated" {
		t.Fatalf("complete Bonza companion capture did not produce scoped live health evidence: %+v", bonza)
	}
}

func TestBonzaCompanionHelpDocumentsPublicUserPresentLimits(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "help-center", "sections", "integrations.md"))
	if err != nil {
		t.Fatalf("read integrations help: %v", err)
	}
	section := strings.ToLower(string(raw))
	for _, token := range []string{
		"bonza slot cars browser companion",
		"user-present",
		"six sync attempts per minute",
		"never exports cookies",
		"sucuri challenge",
		"partial",
		"selector drift",
		"external live evidence",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("Bonza companion help omits %q", token)
		}
	}
}
