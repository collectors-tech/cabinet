package companion

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestCompanionProviderFixturesPersistObservationsCandidatesAndIdempotentReplay(t *testing.T) {
	t.Parallel()

	for _, fixture := range []struct {
		provider      string
		host          string
		filename      string
		complete      bool
		expectedState string
	}{
		{provider: "frontline", host: "www.frontlinehobbies.com.au", filename: "frontline-search-results-v1.json", complete: true, expectedState: captureStateCompleted},
		{provider: "bonza", host: "www.bonzaslotcars.com.au", filename: "bonza-search-results-v1.json", complete: false, expectedState: captureStatePartial},
	} {
		fixture := fixture
		t.Run(fixture.provider, func(t *testing.T) {
			t.Parallel()
			svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
			instance, err := profiles.UpsertIntegrationInstance(context.Background(), profileID, profile.IntegrationInstancePatch{ProviderID: fixture.provider, Enabled: boolPointer(true)})
			if err != nil {
				t.Fatalf("create fixture integration: %v", err)
			}
			modules := append(DefaultModules(), fixtureSearchModule(fixture.provider, fixture.host))
			svc, err = NewPersistentService(context.Background(), svc.db, profiles, modules, Options{})
			if err != nil {
				t.Fatalf("restart fixture companion: %v", err)
			}
			authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
			payload := fixtureSearchPayload(t, svc, profileID, instance.ID, fixture.provider, fixture.host, fixture.filename, "fixture-"+fixture.provider, fixture.complete)

			accepted, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata)
			if err != nil || !accepted.Accepted || !accepted.Committed || accepted.State != fixture.expectedState {
				t.Fatalf("fixture AcceptPayload() = %+v, %v", accepted, err)
			}
			assertCaptureDomainCounts(t, svc, profileID, fixture.provider, 1, 2, 2, 1)
			assertCompanionCandidateProvenance(t, svc, profileID, fixture.provider, fixture.host)

			replayed, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata)
			if err != nil || !replayed.Replayed || replayed.CaptureID != accepted.CaptureID || replayed.State != fixture.expectedState {
				t.Fatalf("idempotent replay = %+v, %v", replayed, err)
			}
			assertCaptureDomainCounts(t, svc, profileID, fixture.provider, 1, 2, 2, 1)

			payload.Data["query"] = "conflicting replay"
			payload.PayloadHash = PayloadDigest(payload.Data)
			if _, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata); ErrorCode(err) != "companion_idempotency_conflict" {
				t.Fatalf("idempotency conflict error = %v", err)
			}
		})
	}
}

func TestCompanionCaptureSchemaRedactionProviderAndPartialRangeFailClosed(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	instance, err := profiles.UpsertIntegrationInstance(context.Background(), profileID, profile.IntegrationInstancePatch{ProviderID: "frontline", Enabled: boolPointer(true)})
	if err != nil {
		t.Fatalf("create Frontline integration: %v", err)
	}
	svc, err = NewPersistentService(context.Background(), svc.db, profiles, append(DefaultModules(), fixtureSearchModule("frontline", "www.frontlinehobbies.com.au")), Options{})
	if err != nil {
		t.Fatalf("restart companion: %v", err)
	}
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	base := fixtureSearchPayload(t, svc, profileID, instance.ID, "frontline", "www.frontlinehobbies.com.au", "frontline-search-results-v1.json", "schema-base", true)
	base.URL += "?session=must-not-persist"

	unknown := clonePayload(t, base)
	unknown.IdempotencyKey = "unknown-field"
	unknown.Data["unexpected"] = "unbounded"
	unknown.PayloadHash = PayloadDigest(unknown.Data)
	if _, err := svc.AcceptPayload(context.Background(), unknown, authorization, metadata); ErrorCode(err) != "companion_payload_field_rejected" {
		t.Fatalf("unknown field error = %v", err)
	}

	secret := clonePayload(t, base)
	secret.IdempotencyKey = "secret-field"
	secret.Data["access_token"] = "must-not-persist"
	secret.PayloadHash = PayloadDigest(secret.Data)
	if _, err := svc.AcceptPayload(context.Background(), secret, authorization, metadata); ErrorCode(err) != "companion_payload_redaction_failed" {
		t.Fatalf("secret field error = %v", err)
	}

	crossProvider := clonePayload(t, base)
	crossProvider.IdempotencyKey = "cross-provider"
	items := crossProvider.Data["items"].([]any)
	items[0].(map[string]any)["url"] = "https://www.bonzaslotcars.com.au/product/not-frontline/"
	crossProvider.PayloadHash = PayloadDigest(crossProvider.Data)
	if _, err := svc.AcceptPayload(context.Background(), crossProvider, authorization, metadata); ErrorCode(err) != "companion_payload_schema_invalid" && ErrorCode(err) != "companion_payload_cross_provider" {
		t.Fatalf("cross-provider item error = %v", err)
	}

	complete, err := svc.AcceptPayload(context.Background(), base, authorization, metadata)
	if err != nil || complete.State != captureStateCompleted {
		t.Fatalf("complete fixture = %+v, %v", complete, err)
	}
	var persistedRaw string
	if err := svc.db.QueryRow(`SELECT raw_payload_json FROM companion_captures WHERE id = ?`, complete.CaptureID).Scan(&persistedRaw); err != nil || strings.Contains(persistedRaw, "must-not-persist") {
		t.Fatalf("capture raw envelope retained source URL query secret: raw=%s err=%v", persistedRaw, err)
	}
	lowConfidence := clonePayload(t, base)
	lowConfidence.IdempotencyKey = "low-confidence-review"
	lowItems := lowConfidence.Data["items"].([]any)
	lowItems[0].(map[string]any)["field_confidence"].(map[string]any)["price"] = 0.5
	lowConfidence.PayloadHash = PayloadDigest(lowConfidence.Data)
	reviewed, err := svc.AcceptPayload(context.Background(), lowConfidence, authorization, metadata)
	if err != nil || reviewed.State != captureStateReview {
		t.Fatalf("low-confidence fixture = %+v, %v", reviewed, err)
	}
	partial := clonePayload(t, base)
	partial.IdempotencyKey = "partial-range"
	partial.PageComplete = false
	partial.Data["complete"] = false
	partial.Data["range_end"] = float64(0)
	partial.Data["items"] = partial.Data["items"].([]any)[:1]
	partial.PayloadHash = PayloadDigest(partial.Data)
	acceptedPartial, err := svc.AcceptPayload(context.Background(), partial, authorization, metadata)
	if err != nil || acceptedPartial.State != captureStatePartial {
		t.Fatalf("partial fixture = %+v, %v", acceptedPartial, err)
	}
	var candidates int
	if err := svc.db.QueryRow(`SELECT COUNT(*) FROM scanner_candidates WHERE profile_id = ? AND source = 'frontline'`, profileID).Scan(&candidates); err != nil || candidates != 2 {
		t.Fatalf("partial range removed prior candidates: count=%d err=%v", candidates, err)
	}
}

func TestCompanionQueueResumesAfterRestartAndCanCancelWithoutDuplicates(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	payload := validPurchasePayload(t, svc, profileID, "restart-purchase", "RESTART-1")
	accepted, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata)
	if err != nil || accepted.State != captureStateReview {
		t.Fatalf("initial purchase capture = %+v, %v", accepted, err)
	}
	if _, err := svc.db.Exec(`UPDATE companion_captures SET state = 'processing' WHERE id = ?`, accepted.CaptureID); err != nil {
		t.Fatalf("simulate interrupted processing: %v", err)
	}
	restarted, err := NewPersistentService(context.Background(), svc.db, profiles, DefaultModules(), Options{})
	if err != nil {
		t.Fatalf("restart companion: %v", err)
	}
	record, err := restarted.captureByID(context.Background(), profileID, accepted.CaptureID)
	if err != nil || record.State != captureStateReview || record.AttemptCount < 2 {
		t.Fatalf("resumed capture = %+v, %v", record, err)
	}
	var purchases int
	if err := svc.db.QueryRow(`SELECT COUNT(*) FROM companion_purchase_inbox WHERE profile_id = ?`, profileID).Scan(&purchases); err != nil || purchases != 1 {
		t.Fatalf("restart duplicated purchase rows: count=%d err=%v", purchases, err)
	}

	if _, err := svc.db.Exec(`UPDATE companion_captures SET state = 'retryable-failed' WHERE id = ?`, accepted.CaptureID); err != nil {
		t.Fatalf("prepare cancellable capture: %v", err)
	}
	cancelled, err := restarted.CancelCapture(context.Background(), authorization, metadata, accepted.CaptureID)
	if err != nil || cancelled.State != captureStateCancelled {
		t.Fatalf("CancelCapture() = %+v, %v", cancelled, err)
	}
}

func TestCompanionCaptureInboxIsProfileScopedAndRevocationStopsAccess(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	if _, err := svc.AcceptPayload(context.Background(), validPurchasePayload(t, svc, profileID, "profile-one", "PROFILE-1"), authorization, metadata); err != nil {
		t.Fatalf("capture profile one: %v", err)
	}
	other, err := profiles.Create(context.Background(), "Other capture profile")
	if err != nil {
		t.Fatalf("create other profile: %v", err)
	}
	if _, err := profiles.UpsertIntegrationInstance(context.Background(), other.ID, profile.IntegrationInstancePatch{ProviderID: "ebay", Enabled: boolPointer(true)}); err != nil {
		t.Fatalf("create other profile module: %v", err)
	}
	otherAuthorization := pairCompanionTestSession(t, svc, other.ID, metadata, []string{CapabilityCapturesSubmit})
	inbox, err := svc.ListCaptures(context.Background(), otherAuthorization, metadata, "", 50)
	if err != nil || len(inbox.Captures) != 0 {
		t.Fatalf("capture inbox leaked profiles: %+v, %v", inbox, err)
	}
	if err := svc.RevokeCredential(context.Background(), authorization, metadata); err != nil {
		t.Fatalf("revoke capture session: %v", err)
	}
	if _, err := svc.ListCaptures(context.Background(), authorization, metadata, "", 50); ErrorCode(err) != "companion_session_revoked" {
		t.Fatalf("revoked capture inbox access error = %v", err)
	}
}

func TestCompanionCaptureQuotaRejectsBeforePersistence(t *testing.T) {
	t.Parallel()
	svc, _, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	if _, err := svc.db.Exec(`WITH RECURSIVE seq(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM seq WHERE x < ?)
		INSERT INTO companion_captures(id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, payload_hash, idempotency_key,
			redaction_summary_json, raw_payload_json, state, created_at, updated_at)
		SELECT printf('quota-%05d', x), ?, 'session', 'module', '1', '1', 'provider', 'instance', 'item_detail',
			'https://example.test/item', '2026-08-06T00:00:00Z', printf('sha256:%064d', x), printf('quota-key-%05d', x),
			'[]', '{}', 'completed', '2026-08-06T00:00:00Z', '2026-08-06T00:00:00Z' FROM seq`, maxCaptureRecordsPerProfile, profileID); err != nil {
		t.Fatalf("seed capture quota: %v", err)
	}
	if _, err := svc.AcceptPayload(context.Background(), validPurchasePayload(t, svc, profileID, "over-quota", "OVER-QUOTA"), authorization, metadata); ErrorCode(err) != "companion_capture_quota_exceeded" {
		t.Fatalf("capture quota error = %v", err)
	}
	var count int
	if err := svc.db.QueryRow(`SELECT COUNT(*) FROM companion_captures WHERE profile_id = ?`, profileID).Scan(&count); err != nil || count != maxCaptureRecordsPerProfile {
		t.Fatalf("capture quota persisted overflow count=%d err=%v", count, err)
	}
}

func fixtureSearchModule(providerID, host string) Module {
	return Module{
		ID: providerID + "-search-capture", ModuleVersion: "1.0.0", Site: providerID, ProviderID: providerID,
		Actions: []string{"capture_search_results"}, PassiveOnly: true,
		CaptureSchemas: []CaptureSchema{{PayloadType: "search_results", Fields: []string{"query", "page", "page_size", "range_start", "range_end", "total_pages", "complete", "items"}, MediaFields: []string{"items.image_url"}}},
		Workflows:      []string{"manual_search_capture"}, RedactionRules: []string{"no_cookies", "no_raw_page", "no_tokens"}, FixtureVersion: "1",
		Display: ModuleDisplay{Name: providerID + " search"},
		Browser: BrowserContract{StartURL: "https://" + host + "/", Origins: []string{"https://" + host + "/*"}, URLPatterns: []string{"https://" + host + "/*"}, CaptureScript: "modules/" + providerID + ".js",
			Readiness: BrowserReadiness{Ready: []string{"#results"}, LoggedOut: []string{"#login"}, Challenge: []string{"#challenge"}}},
		Configuration: ModuleConfiguration{CaptureMode: "manual_user_present", ItemFields: []string{"listing_id", "title", "price", "currency", "url"}, MediaPolicy: "references_by_default", ReviewDestination: "discoveries", RateLimitPerMinute: 6, HelpURL: "/help-center/integrations", SyncAvailable: true},
	}
}

func fixtureSearchPayload(t *testing.T, svc *Service, profileID, instanceID, providerID, host, filename, key string, complete bool) PayloadSubmission {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	data := map[string]any{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return PayloadSubmission{
		ProtocolVersion: ProtocolVersionV1, ProfileID: profileID, ModuleID: providerID + "-search-capture",
		ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instanceID, ProviderID: providerID,
		URL: "https://" + host + "/search", PayloadType: "search_results", CapturedAt: svc.options.Now().UTC().Format(time.RFC3339),
		PageComplete: complete, Passive: true, ConfidenceScore: 0.95,
		RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"}, PayloadHash: PayloadDigest(data), IdempotencyKey: key, Data: data,
	}
}

func clonePayload(t *testing.T, input PayloadSubmission) PayloadSubmission {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal payload clone: %v", err)
	}
	var cloned PayloadSubmission
	if err := json.Unmarshal(raw, &cloned); err != nil {
		t.Fatalf("unmarshal payload clone: %v", err)
	}
	return cloned
}

func assertCaptureDomainCounts(t *testing.T, svc *Service, profileID, providerID string, captures, observations, candidates, runs int) {
	t.Helper()
	queries := []struct {
		name  string
		want  int
		query string
	}{
		{name: "captures", want: captures, query: `SELECT COUNT(*) FROM companion_captures WHERE profile_id = ? AND provider_id = ?`},
		{name: "observations", want: observations, query: `SELECT COUNT(*) FROM companion_observations WHERE profile_id = ? AND provider_id = ?`},
		{name: "candidates", want: candidates, query: `SELECT COUNT(*) FROM scanner_candidates WHERE profile_id = ? AND source = ?`},
		{name: "runs", want: runs, query: `SELECT COUNT(*) FROM scanner_runs WHERE profile_id = ? AND provider = ? AND trigger_type = 'browser_companion'`},
	}
	for _, check := range queries {
		var got int
		if err := svc.db.QueryRow(check.query, profileID, providerID).Scan(&got); err != nil || got != check.want {
			t.Fatalf("%s count=%d want=%d err=%v", check.name, got, check.want, err)
		}
	}
}

func assertCompanionCandidateProvenance(t *testing.T, svc *Service, profileID, providerID, host string) {
	t.Helper()
	var source, currency, sourceResultURL, listingURL, imageURL, firstSeen, lastSeen, stockState string
	var price float64
	var stockCount int
	if err := svc.db.QueryRow(`SELECT source, observed_currency, source_result_url, url, image, first_seen, last_seen, price, stock_state, stock_count
		FROM scanner_candidates WHERE profile_id = ? AND source = ? ORDER BY listing_id LIMIT 1`, profileID, providerID).
		Scan(&source, &currency, &sourceResultURL, &listingURL, &imageURL, &firstSeen, &lastSeen, &price, &stockState, &stockCount); err != nil {
		t.Fatalf("load companion candidate provenance: %v", err)
	}
	providerPrefix := "https://" + host + "/"
	if source != providerID || currency != "AUD" || sourceResultURL != "https://"+host+"/search" ||
		!strings.HasPrefix(listingURL, providerPrefix) || !strings.HasPrefix(imageURL, providerPrefix) ||
		firstSeen == "" || lastSeen == "" || price <= 0 || stockState == "" || stockCount < 0 {
		t.Fatalf("incomplete companion candidate provenance source=%q currency=%q result=%q url=%q image=%q first=%q last=%q price=%v stock=%s/%d",
			source, currency, sourceResultURL, listingURL, imageURL, firstSeen, lastSeen, price, stockState, stockCount)
	}
}

func boolPointer(value bool) *bool { return &value }
