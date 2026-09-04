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

func TestFrontlineProviderRegistryAdvertisesBrowserCompanionFallbackTruthfully(t *testing.T) {
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
	provider := findRegistryProvider(payload.Providers, "au-webshop-frontlinehobbies-com-au")
	if provider == nil {
		t.Fatalf("Frontline provider missing from registry: %+v", payload.Providers)
	}
	for key, want := range map[string]string{
		"market_watch_scope":      "frontlinehobbies",
		"fallback_state":          "browser_companion_user_present",
		"browser_companion_state": "available_when_paired",
		"direct_search_state":     "best_effort_fail_closed",
		"live_evidence_state":     "external_user_present_evidence_required",
	} {
		if got := provider[key]; got != want {
			t.Fatalf("Frontline registry %s=%v want %q: %+v", key, got, want, provider)
		}
	}
	capabilities, ok := provider["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("Frontline capabilities got %T: %+v", provider["capabilities"], provider)
	}
	for key, want := range map[string]bool{
		"browser_companion": true, "user_present_search": true,
		"unattended_search": false, "challenge_bypass": false,
	} {
		if got := capabilities[key]; got != want {
			t.Fatalf("Frontline capability %s=%v want %t: %+v", key, got, want, capabilities)
		}
	}
	instructions := strings.ToLower(provider["setup_instructions"].(string))
	for _, token := range []string{"browser companion", "user", "challenge", "cookie"} {
		if !strings.Contains(instructions, token) {
			t.Fatalf("Frontline setup instructions omit %q: %s", token, instructions)
		}
	}
}

func TestFrontlineCompanionCandidateCanBeConfirmedIntoWishlist(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profiles := profile.NewRepository(a.db)
	created, err := profiles.Create(context.Background(), "Frontline Companion")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("set active profile: %v", err)
	}
	instance, err := profiles.UpsertIntegrationInstance(context.Background(), created.ID, profile.IntegrationInstancePatch{
		ProviderID: "au-webshop-frontlinehobbies-com-au",
	})
	if err != nil {
		t.Fatalf("upsert Frontline integration: %v", err)
	}
	if _, err := a.authService.CreateUnlockedSession(created.ID); err != nil {
		t.Fatalf("unlock profile: %v", err)
	}

	receipt := requestCompanionPairing(t, a, []string{companion.CapabilityModulesRead, companion.CapabilityCapturesSubmit})
	approved := doCompanionManagementRequest(t, a, http.MethodPost, "/api/companion/pairing/approvals", strings.NewReader(
		`{"request_id":"`+receipt.RequestID+`","profile_id":"`+created.ID+`"}`,
	), nil)
	if approved.Code != http.StatusOK {
		t.Fatalf("approve Frontline pairing status=%d body=%s", approved.Code, approved.Body.String())
	}
	exchanged := exchangeCompanionPairing(t, a, receipt, true)
	var credential companion.CredentialResponse
	if err := json.NewDecoder(exchanged.Body).Decode(&credential); err != nil {
		t.Fatalf("decode companion exchange: %v", err)
	}
	authorization := "Bearer " + credential.Credential
	modules := doCompanionExtensionRequest(t, a, http.MethodGet, "/api/companion/modules", nil, map[string]string{"Authorization": authorization})
	if modules.Code != http.StatusOK || !strings.Contains(modules.Body.String(), `"id":"frontlinehobbies-search-capture"`) ||
		!strings.Contains(modules.Body.String(), `"integration_instance_id":"`+instance.ID+`"`) ||
		!strings.Contains(modules.Body.String(), `"sync_available":true`) {
		t.Fatalf("Frontline module projection status=%d body=%s", modules.Code, modules.Body.String())
	}

	rawFixture, err := os.ReadFile(filepath.Join("..", "companion", "testdata", "frontline-search-results-v1.json"))
	if err != nil {
		t.Fatalf("read Frontline fixture: %v", err)
	}
	data := map[string]any{}
	if err := json.Unmarshal(rawFixture, &data); err != nil {
		t.Fatalf("decode Frontline fixture: %v", err)
	}
	payload := companion.PayloadSubmission{
		ProtocolVersion: companion.ProtocolVersionV1, ProfileID: created.ID, ModuleID: "frontlinehobbies-search-capture",
		ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instance.ID,
		ProviderID: "au-webshop-frontlinehobbies-com-au", URL: "https://www.frontlinehobbies.com.au/?s=AFX",
		PayloadType: "search_results", CapturedAt: time.Now().UTC().Format(time.RFC3339), PageComplete: true,
		Passive: true, ConfidenceScore: 0.95, RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"},
		PayloadHash: companion.PayloadDigest(data), IdempotencyKey: "frontline-app-handoff", Data: data,
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal Frontline payload: %v", err)
	}
	accepted := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(string(rawPayload)), map[string]string{
		"Authorization": authorization, "Content-Type": "application/json", "X-Cabinet-Idempotency-Key": payload.IdempotencyKey,
	})
	if accepted.Code != http.StatusAccepted || !strings.Contains(accepted.Body.String(), `"committed":true`) {
		t.Fatalf("Frontline capture status=%d body=%s", accepted.Code, accepted.Body.String())
	}
	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("Frontline post-capture registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var registryPayload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&registryPayload); err != nil {
		t.Fatalf("decode Frontline post-capture registry: %v", err)
	}
	frontline := findRegistryProvider(registryPayload.Providers, "au-webshop-frontlinehobbies-com-au")
	if frontline == nil || frontline["beta_release_status"] != "available_live_validated" || frontline["live_evidence_state"] != "validated" {
		t.Fatalf("complete companion capture did not produce scoped live health evidence: %+v", frontline)
	}

	var candidateID, querySetID, listingID, title, listingURL, sourceResultURL string
	var price float64
	if err := a.db.QueryRow(`SELECT id, query_set_id, listing_id, title, price, url, source_result_url FROM scanner_candidates
		WHERE profile_id = ? AND source = 'frontlinehobbies' ORDER BY listing_id LIMIT 1`, created.ID).
		Scan(&candidateID, &querySetID, &listingID, &title, &price, &listingURL, &sourceResultURL); err != nil {
		t.Fatalf("load Frontline companion candidate: %v", err)
	}
	runs := doRequest(t, a, http.MethodGet, "/api/scanner/runs?query_set_id="+querySetID, nil, nil)
	if runs.Code != http.StatusOK {
		t.Fatalf("Frontline companion run history status=%d body=%s", runs.Code, runs.Body.String())
	}
	var runHistory struct {
		Runs []struct {
			Provider    string `json:"provider"`
			TriggerType string `json:"trigger_type"`
			Status      string `json:"status"`
			ResultCount int    `json:"result_count"`
		} `json:"runs"`
	}
	if err := json.NewDecoder(runs.Body).Decode(&runHistory); err != nil {
		t.Fatalf("decode Frontline companion run history: %v", err)
	}
	if len(runHistory.Runs) != 1 {
		t.Fatalf("Frontline companion run history got %+v", runHistory.Runs)
	}
	run := runHistory.Runs[0]
	if run.Provider != "frontlinehobbies" || run.TriggerType != "browser_companion" || run.Status != "succeeded" || run.ResultCount <= 0 {
		t.Fatalf("Frontline companion run history lost canonical success evidence: %+v", run)
	}
	action := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(
		`{"candidate_id":"`+candidateID+`","type":"add_to_wishlist","payload":{"source":"browser_companion","query_set_id":"`+querySetID+`","reviewer_notes":"Reviewed Frontline capture"}}`,
	), map[string]string{"Content-Type": "application/json"})
	if action.Code != http.StatusOK {
		t.Fatalf("Frontline Wishlist hand-off status=%d body=%s", action.Code, action.Body.String())
	}
	var result struct {
		OK    bool           `json:"ok"`
		Audit map[string]any `json:"audit"`
	}
	if err := json.NewDecoder(action.Body).Decode(&result); err != nil {
		t.Fatalf("decode Frontline hand-off: %v", err)
	}
	if !result.OK || result.Audit["source_provider"] != "frontlinehobbies" || result.Audit["source_result_url"] != sourceResultURL ||
		result.Audit["listing_url"] != listingURL || result.Audit["observed_currency"] != "AUD" || result.Audit["listing_id"] != listingID {
		t.Fatalf("Frontline hand-off lost provenance: %+v", result)
	}
	var wishlistCount int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM wishlist_entries w JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_matches m ON m.item_id = i.id
		WHERE w.profile_id = ? AND i.profile_id = ? AND m.candidate_id = ? AND i.part_number = ? AND w.owned = 0`,
		created.ID, created.ID, candidateID, listingID).Scan(&wishlistCount); err != nil || wishlistCount != 1 {
		t.Fatalf("Frontline Wishlist hand-off count=%d err=%v price=%v", wishlistCount, err, price)
	}
}

func TestFrontlineCompanionHelpDocumentsPublicUserPresentLimits(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "help-center", "sections", "integrations.md"))
	if err != nil {
		t.Fatalf("read integrations help: %v", err)
	}
	section := strings.ToLower(string(raw))
	for _, token := range []string{
		"frontline hobbies browser companion",
		"user-present",
		"six sync attempts per minute",
		"never exports cookies",
		"challenge",
		"partial",
		"selector drift",
		"external live evidence",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("Frontline companion help omits %q", token)
		}
	}
}
