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
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

func TestCompanionProviderCaptureIsReviewableAndPersistsWishlistHandoff(t *testing.T) {
	providers := []struct {
		name          string
		providerID    string
		moduleID      string
		providerScope string
		fixture       string
		captureURL    string
		pageComplete  bool
	}{
		{
			name: "Frontline", providerID: "au-webshop-frontlinehobbies-com-au",
			moduleID: "frontlinehobbies-search-capture", providerScope: "frontlinehobbies",
			fixture: "frontline-search-results-v1.json", captureURL: "https://www.frontlinehobbies.com.au/?s=AFX", pageComplete: true,
		},
		{
			name: "Bonza", providerID: "au-webshop-bonzaslotcars-com-au",
			moduleID: "bonzaslotcars-search-capture", providerScope: "bonzaslotcars",
			fixture: "bonza-search-results-v1.json", captureURL: "https://www.bonzaslotcars.com.au/?s=Scalextric", pageComplete: false,
		},
	}

	for _, provider := range providers {
		provider := provider
		t.Run(provider.name, func(t *testing.T) {
			a := newTestApp(t)
			profiles := profile.NewRepository(a.db)
			created, err := profiles.Create(context.Background(), provider.name+" discovery handoff")
			if err != nil {
				t.Fatalf("create profile: %v", err)
			}
			if err := profiles.SetActiveProfile(context.Background(), created.ID); err != nil {
				t.Fatalf("set active profile: %v", err)
			}
			instance, err := profiles.UpsertIntegrationInstance(context.Background(), created.ID, profile.IntegrationInstancePatch{
				ProviderID: provider.providerID,
			})
			if err != nil {
				t.Fatalf("upsert %s integration: %v", provider.name, err)
			}
			if _, err := a.authService.CreateUnlockedSession(created.ID); err != nil {
				t.Fatalf("unlock profile: %v", err)
			}

			receipt := requestCompanionPairing(t, a, []string{companion.CapabilityCapturesSubmit})
			approved := doCompanionManagementRequest(t, a, http.MethodPost, "/api/companion/pairing/approvals", strings.NewReader(
				`{"request_id":"`+receipt.RequestID+`","profile_id":"`+created.ID+`"}`,
			), nil)
			if approved.Code != http.StatusOK {
				t.Fatalf("approve %s pairing status=%d body=%s", provider.name, approved.Code, approved.Body.String())
			}
			exchanged := exchangeCompanionPairing(t, a, receipt, true)
			var credential companion.CredentialResponse
			if err := json.NewDecoder(exchanged.Body).Decode(&credential); err != nil {
				t.Fatalf("decode companion exchange: %v", err)
			}

			rawFixture, err := os.ReadFile(filepath.Join("..", "companion", "testdata", provider.fixture))
			if err != nil {
				t.Fatalf("read %s fixture: %v", provider.name, err)
			}
			data := map[string]any{}
			if err := json.Unmarshal(rawFixture, &data); err != nil {
				t.Fatalf("decode %s fixture: %v", provider.name, err)
			}
			payload := companion.PayloadSubmission{
				ProtocolVersion: companion.ProtocolVersionV1, ProfileID: created.ID, ModuleID: provider.moduleID,
				ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instance.ID,
				ProviderID: provider.providerID, URL: provider.captureURL, PayloadType: "search_results",
				CapturedAt: time.Now().UTC().Format(time.RFC3339), PageComplete: provider.pageComplete,
				Passive: true, ConfidenceScore: 0.95, RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"},
				PayloadHash: companion.PayloadDigest(data), IdempotencyKey: "discovery-handoff-" + strings.ToLower(provider.name), Data: data,
			}
			rawPayload, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal %s payload: %v", provider.name, err)
			}
			accepted := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(string(rawPayload)), map[string]string{
				"Authorization": "Bearer " + credential.Credential, "Content-Type": "application/json",
				"X-Cabinet-Idempotency-Key": payload.IdempotencyKey,
			})
			if accepted.Code != http.StatusAccepted || !strings.Contains(accepted.Body.String(), `"committed":true`) {
				t.Fatalf("%s capture status=%d body=%s", provider.name, accepted.Code, accepted.Body.String())
			}
			replayed := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(string(rawPayload)), map[string]string{
				"Authorization": "Bearer " + credential.Credential, "Content-Type": "application/json",
				"X-Cabinet-Idempotency-Key": payload.IdempotencyKey,
			})
			if replayed.Code != http.StatusAccepted || !strings.Contains(replayed.Body.String(), `"replayed":true`) {
				t.Fatalf("%s capture replay status=%d body=%s", provider.name, replayed.Code, replayed.Body.String())
			}

			discoveries := doRequest(t, a, http.MethodGet, "/api/discovery/not-in-collection", nil, nil)
			if discoveries.Code != http.StatusOK {
				t.Fatalf("%s discoveries status=%d body=%s", provider.name, discoveries.Code, discoveries.Body.String())
			}
			var discoveryPayload struct {
				Items []discovery.Item `json:"items"`
			}
			if err := json.NewDecoder(discoveries.Body).Decode(&discoveryPayload); err != nil {
				t.Fatalf("decode %s discoveries: %v", provider.name, err)
			}
			if len(discoveryPayload.Items) != 2 {
				t.Fatalf("%s fresh capture discoveries=%d want=2: %+v", provider.name, len(discoveryPayload.Items), discoveryPayload.Items)
			}
			var matchCount int
			if err := a.db.QueryRow(`SELECT COUNT(*) FROM scanner_matches m JOIN scanner_candidates c ON c.id = m.candidate_id WHERE c.profile_id = ?`, created.ID).Scan(&matchCount); err != nil || matchCount != 2 {
				t.Fatalf("%s capture replay matches=%d want=2 err=%v", provider.name, matchCount, err)
			}
			candidate := discoveryPayload.Items[0]
			if candidate.SourceProvider != provider.providerScope || candidate.QuerySetID == "" || candidate.ListingID == "" ||
				candidate.ExtractedPart != candidate.ListingID || candidate.ObservedCurrency != "AUD" || candidate.SourceResultURL == "" ||
				candidate.URL == "" || !candidate.NeedsReview || candidate.Confidence != 0 {
				t.Fatalf("%s discovery lost review/provenance fields: %+v", provider.name, candidate)
			}

			actionBody, _ := json.Marshal(map[string]any{
				"candidate_id": candidate.CandidateID,
				"type":         "add_to_wishlist",
				"payload": map[string]any{
					"source": "browser_companion", "reviewer_notes": "Reviewed provider capture",
				},
			})
			for attempt := 1; attempt <= 2; attempt++ {
				action := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(string(actionBody)), map[string]string{"Content-Type": "application/json"})
				if action.Code != http.StatusOK || !strings.Contains(action.Body.String(), `"ok":true`) {
					t.Fatalf("%s Wishlist handoff attempt=%d status=%d body=%s", provider.name, attempt, action.Code, action.Body.String())
				}
			}

			wishlistResponse := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
			if wishlistResponse.Code != http.StatusOK {
				t.Fatalf("%s Wishlist status=%d body=%s", provider.name, wishlistResponse.Code, wishlistResponse.Body.String())
			}
			var wishlistPayload struct {
				Items []wishlist.Entry `json:"items"`
			}
			if err := json.NewDecoder(wishlistResponse.Body).Decode(&wishlistPayload); err != nil {
				t.Fatalf("decode %s Wishlist: %v", provider.name, err)
			}
			if len(wishlistPayload.Items) != 1 || wishlistPayload.Items[0].ItemID == "" || wishlistPayload.Items[0].Owned {
				t.Fatalf("%s Wishlist handoff entries=%+v", provider.name, wishlistPayload.Items)
			}
			var profileID, status, partNumber string
			if err := a.db.QueryRow(`SELECT profile_id, status, part_number FROM canonical_items WHERE id = ?`, wishlistPayload.Items[0].ItemID).
				Scan(&profileID, &status, &partNumber); err != nil {
				t.Fatalf("load %s linked canonical item: %v", provider.name, err)
			}
			if profileID != created.ID || status != "wishlist" || partNumber != candidate.ListingID {
				t.Fatalf("%s linked item profile=%q status=%q part=%q candidate=%+v", provider.name, profileID, status, partNumber, candidate)
			}
			var itemCount, wishlistCount int
			if err := a.db.QueryRow(`SELECT COUNT(*) FROM canonical_items WHERE profile_id = ?`, created.ID).Scan(&itemCount); err != nil {
				t.Fatalf("count %s canonical items: %v", provider.name, err)
			}
			if err := a.db.QueryRow(`SELECT COUNT(*) FROM wishlist_entries WHERE profile_id = ?`, created.ID).Scan(&wishlistCount); err != nil {
				t.Fatalf("count %s Wishlist rows: %v", provider.name, err)
			}
			if itemCount != 1 || wishlistCount != 1 {
				t.Fatalf("%s replay duplicated handoff items=%d wishlist=%d", provider.name, itemCount, wishlistCount)
			}

			other, err := profiles.Create(context.Background(), provider.name+" isolated profile")
			if err != nil {
				t.Fatalf("create %s isolated profile: %v", provider.name, err)
			}
			if err := profiles.SetActiveProfile(context.Background(), other.ID); err != nil {
				t.Fatalf("activate %s isolated profile: %v", provider.name, err)
			}
			isolatedWishlist := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
			if isolatedWishlist.Code != http.StatusOK || isolatedWishlist.Body.String() != "{\"items\":[]}\n" {
				t.Fatalf("%s Wishlist leaked into another profile status=%d body=%s", provider.name, isolatedWishlist.Code, isolatedWishlist.Body.String())
			}
		})
	}
}
