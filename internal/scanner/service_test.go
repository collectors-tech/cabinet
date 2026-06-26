package scanner

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

type testProvider struct {
	providerID string
	failures   int
	calls      int
	items      []CandidateInput
}

func (p *testProvider) ProviderID() string {
	return p.providerID
}

func (p *testProvider) Search(_ context.Context, _ QuerySet) ([]CandidateInput, error) {
	p.calls++
	if p.calls <= p.failures {
		return nil, errors.New("temporary failure")
	}
	return p.items, nil
}

func TestQuerySetAndRunNowLifecycle(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySet(context.Background(), QuerySet{
		Name:          "AFX US",
		Keywords:      []string{"afx", "slot car"},
		Exclusions:    []string{"damaged"},
		MaxPrice:      100,
		Region:        "US",
		Condition:     "used",
		ScheduleCron:  "*/15 * * * *",
		Enabled:       true,
		RateLimitRPS:  2,
		MaxRetryCount: 2,
	})
	if err != nil {
		t.Fatalf("CreateQuerySet() error = %v", err)
	}
	if len(qs.Keywords) != 2 || qs.Region != "US" || qs.Condition != "used" {
		t.Fatalf("unexpected query set: %+v", qs)
	}

	provider := &testProvider{
		failures: 1,
		items: []CandidateInput{
			{ListingID: "L1", Title: "AFX P-1 car", Price: 55, URL: "http://x/1", Seller: "s1", StockState: "low_stock", StockCount: 2},
		},
	}
	run, err := svc.RunNow(context.Background(), qs.ID, provider)
	if err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}
	if run.Attempts < 2 {
		t.Fatalf("expected retry attempts, got %+v", run)
	}
	cands, err := svc.ListCandidates(context.Background(), qs.ID)
	if err != nil {
		t.Fatalf("ListCandidates() error = %v", err)
	}
	if len(cands) != 1 || cands[0].Status == "" || cands[0].FirstSeen == "" || cands[0].LastSeen == "" {
		t.Fatalf("unexpected candidates: %+v", cands)
	}
	if cands[0].StockState != "low_stock" || cands[0].StockCount != 2 {
		t.Fatalf("expected stock persistence in candidates, got %+v", cands[0])
	}
	health, err := svc.ProviderHealth(context.Background(), "ebay")
	if err != nil {
		t.Fatalf("ProviderHealth() error = %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("expected provider health ok, got %+v", health)
	}
	failures, err := svc.ListFailures(context.Background())
	if err != nil {
		t.Fatalf("ListFailures() error = %v", err)
	}
	if len(failures) == 0 {
		t.Fatal("expected failure log entries from retry sequence")
	}
}

func TestRunNowPersistsObservedCurrency(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySet(context.Background(), QuerySet{
		Name:     "eBay currency",
		Keywords: []string{"afx"},
	})
	if err != nil {
		t.Fatalf("CreateQuerySet() error = %v", err)
	}
	provider := &testProvider{items: []CandidateInput{{
		ListingID: "EBAY-CURRENCY-1",
		Title:     "AFX currency candidate",
		Price:     42.5,
		Currency:  " aud ",
		URL:       "https://example.test/ebay/currency",
		Source:    "ebay",
	}}}
	if _, err := svc.RunNow(context.Background(), qs.ID, provider); err != nil {
		t.Fatalf("RunNow() error = %v", err)
	}

	var observedCurrency string
	if err := conn.QueryRow(`SELECT observed_currency FROM scanner_candidates WHERE listing_id = ?`, "EBAY-CURRENCY-1").Scan(&observedCurrency); err != nil {
		t.Fatalf("query observed currency: %v", err)
	}
	if observedCurrency != "AUD" {
		t.Fatalf("expected observed currency AUD, got %q", observedCurrency)
	}
	candidates, err := svc.ListCandidates(context.Background(), qs.ID)
	if err != nil {
		t.Fatalf("ListCandidates() error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Currency != "AUD" {
		t.Fatalf("expected candidate read model to expose observed currency AUD, got %+v", candidates)
	}

	provider.items[0].Currency = " usd "
	if _, err := svc.RunNow(context.Background(), qs.ID, provider); err != nil {
		t.Fatalf("RunNow() second pass error = %v", err)
	}
	if err := conn.QueryRow(`SELECT observed_currency FROM scanner_candidates WHERE listing_id = ?`, "EBAY-CURRENCY-1").Scan(&observedCurrency); err != nil {
		t.Fatalf("query updated observed currency: %v", err)
	}
	if observedCurrency != "USD" {
		t.Fatalf("expected updated observed currency USD, got %q", observedCurrency)
	}
	candidates, err = svc.ListCandidates(context.Background(), qs.ID)
	if err != nil {
		t.Fatalf("ListCandidates() after update error = %v", err)
	}
	if len(candidates) != 1 || candidates[0].Currency != "USD" {
		t.Fatalf("expected candidate read model to expose refreshed observed currency USD, got %+v", candidates)
	}
}

func TestUpdateQuerySetPreservesProviderScopeWhenOmitted(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	created, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "eBay Slot Cars",
		Keywords:      []string{"afx"},
		ProviderScope: []string{"ebay"},
		ScheduleCron:  "0 */6 * * *",
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile() error = %v", err)
	}

	updated, err := svc.UpdateQuerySetForProfile(context.Background(), "profile-a", created.ID, QuerySet{
		Name:         "eBay Slot Cars Edited",
		Keywords:     []string{"afx", "mega g+"},
		ScheduleCron: "30 */8 * * *",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("UpdateQuerySetForProfile() error = %v", err)
	}
	if len(updated.ProviderScope) != 1 || updated.ProviderScope[0] != "ebay" {
		t.Fatalf("expected omitted provider scope update to preserve [ebay], got %+v", updated.ProviderScope)
	}
	if updated.ScheduleCron != "30 */8 * * *" {
		t.Fatalf("expected updated schedule to persist, got %+v", updated)
	}
}

func TestRunScheduled(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	svc := NewService(conn)
	_, err = svc.CreateQuerySet(context.Background(), QuerySet{
		Name:         "scheduled",
		Keywords:     []string{"afx"},
		ScheduleCron: "*/5 * * * *",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("CreateQuerySet() error = %v", err)
	}
	p := &testProvider{items: []CandidateInput{{ListingID: "S1", Title: "AFX S1", URL: "http://x/s1"}}}
	ran, err := svc.RunScheduled(context.Background(), p)
	if err != nil {
		t.Fatalf("RunScheduled() error = %v", err)
	}
	if ran != 1 {
		t.Fatalf("expected 1 scheduled run, got %d", ran)
	}
}

func TestFailureSnapshotsAreProfileScoped(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	failingQuery, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Profile A failures",
		Keywords:      []string{"afx"},
		Enabled:       true,
		MaxRetryCount: 0,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(profile-a) error = %v", err)
	}
	otherQuery, err := svc.CreateQuerySetForProfile(context.Background(), "profile-b", QuerySet{
		Name:     "Profile B clean",
		Keywords: []string{"afx"},
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(profile-b) error = %v", err)
	}

	_, err = svc.RunNowForProfile(context.Background(), "profile-a", failingQuery.ID, &testProvider{failures: 1})
	if err == nil {
		t.Fatal("expected profile-a run failure")
	}

	profileAFailures, err := svc.ListFailuresByProfile(context.Background(), "profile-a")
	if err != nil {
		t.Fatalf("ListFailuresByProfile(profile-a) error = %v", err)
	}
	if len(profileAFailures) != 1 || profileAFailures[0]["query_set_id"] != failingQuery.ID {
		t.Fatalf("expected profile-a failure for its query set, got %+v", profileAFailures)
	}
	for key, expected := range map[string]string{
		"provider":       "ebay",
		"reason":         "temporary failure",
		"retry_guidance": "Check provider health, credentials, and retry the operation.",
		"next_action":    "check_provider_health_and_credentials",
	} {
		if got := profileAFailures[0][key]; got != expected {
			t.Fatalf("expected profile-a failure %s=%q, got %q in %+v", key, expected, got, profileAFailures[0])
		}
	}
	profileBFailures, err := svc.ListFailuresByProfile(context.Background(), "profile-b")
	if err != nil {
		t.Fatalf("ListFailuresByProfile(profile-b) error = %v", err)
	}
	if len(profileBFailures) != 0 {
		t.Fatalf("expected no profile-b failures, got %+v", profileBFailures)
	}

	reloadedFailing, err := svc.GetQuerySetForProfile(context.Background(), "profile-a", failingQuery.ID)
	if err != nil {
		t.Fatalf("GetQuerySetForProfile(profile-a) error = %v", err)
	}
	if reloadedFailing.LastRunStatus != "failed" || reloadedFailing.LastRunMessage != "temporary failure" {
		t.Fatalf("expected profile-a failed snapshot, got %+v", reloadedFailing)
	}
	reloadedOther, err := svc.GetQuerySetForProfile(context.Background(), "profile-b", otherQuery.ID)
	if err != nil {
		t.Fatalf("GetQuerySetForProfile(profile-b) error = %v", err)
	}
	if reloadedOther.LastRunStatus != "never" || reloadedOther.LastRunMessage != "" {
		t.Fatalf("expected profile-b clean snapshot, got %+v", reloadedOther)
	}
}

func TestRunNowRecordsProviderHealthForExecutingProvider(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Bonza scoped health",
		Keywords:      []string{"afx"},
		ProviderScope: []string{"bonzaslotcars"},
		Enabled:       true,
		MaxRetryCount: 0,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile() error = %v", err)
	}

	_, err = svc.RunNowForProfile(context.Background(), "profile-a", qs.ID, &testProvider{
		providerID: "bonzaslotcars",
		failures:   1,
	})
	if err == nil {
		t.Fatal("expected bonzaslotcars run failure")
	}

	bonzaHealth, err := svc.ProviderHealth(context.Background(), "bonzaslotcars")
	if err != nil {
		t.Fatalf("ProviderHealth(bonzaslotcars) error = %v", err)
	}
	if bonzaHealth["status"] != "error" || bonzaHealth["message"] != "temporary failure" {
		t.Fatalf("expected bonzaslotcars health failure, got %+v", bonzaHealth)
	}
	ebayHealth, err := svc.ProviderHealth(context.Background(), "ebay")
	if err != nil {
		t.Fatalf("ProviderHealth(ebay) error = %v", err)
	}
	if ebayHealth["status"] != "unknown" || ebayHealth["message"] != "" {
		t.Fatalf("non-ebay provider failure must not poison ebay health, got %+v", ebayHealth)
	}
	failures, err := svc.ListFailuresByProfile(context.Background(), "profile-a")
	if err != nil {
		t.Fatalf("ListFailuresByProfile(profile-a) error = %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected one provider-scoped failure, got %+v", failures)
	}
	for key, expected := range map[string]string{
		"provider":       "bonzaslotcars",
		"retry_guidance": "Review provider status and retry the operation.",
		"next_action":    "review_provider_status",
	} {
		if got := failures[0][key]; got != expected {
			t.Fatalf("expected scoped failure %s=%q, got %q in %+v", key, expected, got, failures[0])
		}
	}
}

func TestBuildRecognitionReviewNormalizesCandidatesAndRequiresConfirmation(t *testing.T) {
	t.Parallel()

	review, err := BuildRecognitionReview([]RecognitionCandidateInput{
		{
			ID:         "low",
			Title:      "Fuzzy slot car",
			Confidence: 0.41,
			Source:     "camera",
			Provenance: "ocr-v1",
			MediaID:    "media-1",
			Target:     "wishlist",
		},
		{
			ID:         "top",
			Title:      "AFX Camaro",
			Confidence: 0.93,
			Source:     "catalog",
			Provenance: "matcher-v2",
			MediaURL:   "https://example.test/scan.jpg",
			Target:     "inventory",
		},
	})
	if err != nil {
		t.Fatalf("BuildRecognitionReview() error = %v", err)
	}
	if review.TopCandidate.ID != "top" || review.SelectedCandidate.ID != "top" {
		t.Fatalf("expected top candidate selected, got %+v", review)
	}
	if review.ConfidenceLabel != "high" || review.RequiresManualReview {
		t.Fatalf("expected high confidence without manual review, got %+v", review)
	}
	if !review.ConfirmBeforeCreate || review.Target != "inventory" {
		t.Fatalf("expected confirm-before-create inventory preview, got %+v", review)
	}
	if got := review.MediaEvidence["media_url"]; got != "https://example.test/scan.jpg" {
		t.Fatalf("expected selected media evidence, got %+v", review.MediaEvidence)
	}
	if len(review.Alternates) != 1 || review.Alternates[0].ID != "low" {
		t.Fatalf("expected lower-confidence alternate retained, got %+v", review.Alternates)
	}
	if len(review.Provenance) != 4 {
		t.Fatalf("expected unique source/provenance evidence, got %+v", review.Provenance)
	}
}

func TestBuildRecognitionReviewPreservesManualOverridePreview(t *testing.T) {
	t.Parallel()

	review, err := BuildRecognitionReview([]RecognitionCandidateInput{
		{
			ID:         "auto",
			Title:      "Auto match",
			Confidence: 0.96,
			Source:     "catalog",
			Provenance: "matcher",
			MediaID:    "media-2",
		},
		{
			ID:           "manual",
			Title:        "Manual override",
			Confidence:   0.72,
			Source:       "user",
			Provenance:   "manual-search",
			OverrideID:   "manual",
			OverrideBy:   "reviewer",
			OverrideNote: "selected exact variant",
			Target:       "wishlist",
		},
	})
	if err != nil {
		t.Fatalf("BuildRecognitionReview() error = %v", err)
	}
	if review.TopCandidate.ID != "auto" {
		t.Fatalf("expected auto top candidate retained, got %+v", review.TopCandidate)
	}
	if review.SelectedCandidate.ID != "manual" || !review.ManualOverrideApplied {
		t.Fatalf("expected manual override selected, got %+v", review)
	}
	if review.ConfidenceLabel != "medium" || !review.RequiresManualReview {
		t.Fatalf("expected medium-confidence manual review, got %+v", review)
	}
	if review.Target != "wishlist" || !review.ConfirmBeforeCreate {
		t.Fatalf("expected wishlist confirm-before-create preview, got %+v", review)
	}
	if len(review.Alternates) != 1 || review.Alternates[0].ID != "auto" {
		t.Fatalf("expected original top as alternate, got %+v", review.Alternates)
	}
}
