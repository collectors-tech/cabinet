package scanner

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
)

type testProvider struct {
	providerID string
	failures   int
	calls      int
	items      []CandidateInput
}

type scheduledMixedProvider struct{}

func (p *scheduledMixedProvider) ProviderID() string {
	return "ebay"
}

func (p *scheduledMixedProvider) Search(_ context.Context, q QuerySet) ([]CandidateInput, error) {
	switch q.Name {
	case "Scheduled failing watch":
		return nil, errors.New("provider outage")
	case "Scheduled healthy watch":
		return []CandidateInput{{
			ListingID: "SCHEDULED-OK-1",
			Title:     "Scheduled healthy result",
			Price:     33,
			Currency:  "AUD",
			URL:       "https://market.test/scheduled-ok-1",
			Source:    "ebay",
		}}, nil
	default:
		return nil, errors.New("unexpected scheduled watch")
	}
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

func TestRunNowPersistsDurableRunRecordAndDedupesResults(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Durable AFX watch",
		Keywords:      []string{"afx"},
		ProviderScope: []string{"bonzaslotcars"},
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile() error = %v", err)
	}
	provider := &testProvider{
		providerID: "bonzaslotcars",
		items: []CandidateInput{
			{ListingID: "AFX-100", Title: "AFX listing", Price: 42.5, Currency: "AUD", Shipping: 9.5, URL: "https://shop.test/afx-100", Source: "bonzaslotcars"},
			{Title: "URL-only listing", Price: 18, Currency: "AUD", URL: "https://shop.test/url-only", Source: "bonzaslotcars"},
		},
	}
	if _, err := svc.RunNowForProfile(context.Background(), "profile-a", qs.ID, provider); err != nil {
		t.Fatalf("RunNowForProfile() first pass error = %v", err)
	}
	provider.items[0].Price = 39.95
	provider.items[1].Price = 17.5
	if _, err := conn.Exec(`UPDATE scanner_candidates SET status = 'wishlisted' WHERE query_set_id = ? AND listing_id = ?`, qs.ID, "AFX-100"); err != nil {
		t.Fatalf("seed decision state: %v", err)
	}
	if _, err := conn.Exec(`UPDATE scanner_candidates SET status = 'archived' WHERE query_set_id = ? AND url = ?`, qs.ID, "https://shop.test/url-only"); err != nil {
		t.Fatalf("seed URL-only decision state: %v", err)
	}
	if _, err := svc.RunNowForProfile(context.Background(), "profile-a", qs.ID, provider); err != nil {
		t.Fatalf("RunNowForProfile() second pass error = %v", err)
	}

	var runCount, resultCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND provider = ? AND status = 'succeeded'`, "profile-a", qs.ID, "bonzaslotcars").Scan(&runCount); err != nil {
		t.Fatalf("query run records: %v", err)
	}
	if runCount != 2 {
		t.Fatalf("expected two durable run records, got %d", runCount)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ?`, "profile-a", qs.ID).Scan(&resultCount); err != nil {
		t.Fatalf("query result records: %v", err)
	}
	if resultCount != 2 {
		t.Fatalf("expected listing-id and source-url dedupe to keep two results after rerun, got %d", resultCount)
	}
	var status string
	var price float64
	if err := conn.QueryRow(`SELECT status, price FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND listing_id = ?`, "profile-a", qs.ID, "AFX-100").Scan(&status, &price); err != nil {
		t.Fatalf("load deduped listing result: %v", err)
	}
	if status != "wishlisted" {
		t.Fatalf("expected rerun to preserve user decision status, got %q", status)
	}
	if price != 39.95 {
		t.Fatalf("expected rerun to refresh listing metadata, got price=%v", price)
	}
	var urlOnlyListingID, urlOnlyStatus string
	var urlOnlyPrice float64
	if err := conn.QueryRow(`SELECT listing_id, status, price FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND url = ?`, "profile-a", qs.ID, "https://shop.test/url-only").Scan(&urlOnlyListingID, &urlOnlyStatus, &urlOnlyPrice); err != nil {
		t.Fatalf("load URL-only result: %v", err)
	}
	if urlOnlyListingID == "" {
		t.Fatal("expected URL-only result to receive a durable synthetic listing key")
	}
	if urlOnlyStatus != "archived" {
		t.Fatalf("expected URL-only rerun to preserve archived decision status, got %q", urlOnlyStatus)
	}
	if urlOnlyPrice != 17.5 {
		t.Fatalf("expected URL-only rerun to refresh result metadata, got price=%v", urlOnlyPrice)
	}
}

func TestRunNowPreservesDownstreamDecisionStatuses(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Decision handoff watch",
		Keywords:      []string{"afx"},
		ProviderScope: []string{"ebay"},
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile() error = %v", err)
	}
	provider := &testProvider{
		providerID: "ebay",
		items: []CandidateInput{
			{ListingID: "DECISION-IGNORED", Title: "Ignored result", Price: 10, Currency: "AUD", URL: "https://market.test/ignored", Source: "ebay"},
			{ListingID: "DECISION-ARCHIVED", Title: "Archived result", Price: 20, Currency: "AUD", URL: "https://market.test/archived", Source: "ebay"},
			{ListingID: "DECISION-WISHLISTED", Title: "Wishlisted result", Price: 30, Currency: "AUD", URL: "https://market.test/wishlisted", Source: "ebay"},
			{ListingID: "DECISION-PURCHASE", Title: "Purchase result", Price: 40, Currency: "AUD", URL: "https://market.test/purchase", Source: "ebay"},
			{ListingID: "DECISION-INVENTORY", Title: "Inventory result", Price: 50, Currency: "AUD", URL: "https://market.test/inventory", Source: "ebay"},
		},
	}
	if _, err := svc.RunNowForProfile(context.Background(), "profile-a", qs.ID, provider); err != nil {
		t.Fatalf("RunNowForProfile() first pass error = %v", err)
	}

	decisions := map[string]string{
		"DECISION-IGNORED":    "ignored",
		"DECISION-ARCHIVED":   "archived",
		"DECISION-WISHLISTED": "wishlisted",
		"DECISION-PURCHASE":   "purchase_candidate",
		"DECISION-INVENTORY":  "inventory_candidate",
	}
	refreshedPrices := map[string]float64{}
	for listingID, status := range decisions {
		if _, err := conn.Exec(`UPDATE scanner_candidates SET status = ? WHERE profile_id = ? AND query_set_id = ? AND listing_id = ?`, status, "profile-a", qs.ID, listingID); err != nil {
			t.Fatalf("seed %s decision status: %v", listingID, err)
		}
	}
	for i := range provider.items {
		provider.items[i].Price += 0.75
		refreshedPrices[provider.items[i].ListingID] = provider.items[i].Price
	}
	if _, err := svc.RunNowForProfile(context.Background(), "profile-a", qs.ID, provider); err != nil {
		t.Fatalf("RunNowForProfile() second pass error = %v", err)
	}

	candidates, err := svc.ListCandidates(context.Background(), qs.ID)
	if err != nil {
		t.Fatalf("ListCandidates() error = %v", err)
	}
	if len(candidates) != len(decisions) {
		t.Fatalf("expected %d decision results after rerun, got %d: %+v", len(decisions), len(candidates), candidates)
	}
	for _, candidate := range candidates {
		expected, ok := decisions[candidate.ListingID]
		if !ok {
			t.Fatalf("unexpected candidate after rerun: %+v", candidate)
		}
		if candidate.Status != expected {
			t.Fatalf("expected %s to preserve status %q, got %q", candidate.ListingID, expected, candidate.Status)
		}
		if candidate.Price != refreshedPrices[candidate.ListingID] || candidate.Currency != "AUD" {
			t.Fatalf("expected rerun to refresh metadata while preserving %s, got %+v", candidate.ListingID, candidate)
		}
	}

	var runCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND provider = ?`, "profile-a", qs.ID, "ebay").Scan(&runCount); err != nil {
		t.Fatalf("count run records: %v", err)
	}
	if runCount != 2 {
		t.Fatalf("expected both decision-preserving runs to be durable, got %d", runCount)
	}
}

func TestRunNowDedupeIsScopedByProfileAndWatch(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	profileAWatch, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{Name: "A watch", Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(profile-a) error = %v", err)
	}
	profileASecondWatch, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{Name: "A second watch", Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(profile-a second) error = %v", err)
	}
	profileBWatch, err := svc.CreateQuerySetForProfile(context.Background(), "profile-b", QuerySet{Name: "B watch", Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(profile-b) error = %v", err)
	}
	provider := &testProvider{providerID: "ebay", items: []CandidateInput{{
		ListingID: "SHARED-LISTING",
		Title:     "Shared listing",
		URL:       "https://market.test/shared-listing",
		Source:    "ebay",
	}}}

	for _, run := range []struct {
		profileID  string
		querySetID string
	}{
		{"profile-a", profileAWatch.ID},
		{"profile-a", profileASecondWatch.ID},
		{"profile-b", profileBWatch.ID},
	} {
		if _, err := svc.RunNowForProfile(context.Background(), run.profileID, run.querySetID, provider); err != nil {
			t.Fatalf("RunNowForProfile(%s, %s) error = %v", run.profileID, run.querySetID, err)
		}
	}

	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_candidates WHERE listing_id = 'SHARED-LISTING'`).Scan(&count); err != nil {
		t.Fatalf("count shared listing candidates: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected shared provider listing to persist once per profile/watch, got %d", count)
	}
}

func TestRunNowDoesNotReintroduceArchivedDiscoveryCandidate(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySet(context.Background(), QuerySet{
		Name:     "Archived duplicate",
		Keywords: []string{"afx"},
	})
	if err != nil {
		t.Fatalf("CreateQuerySet() error = %v", err)
	}
	provider := &testProvider{items: []CandidateInput{{
		ListingID: "ARCHIVED-DISCOVERY-1",
		Title:     "Archived AFX candidate",
		Price:     42.5,
		Currency:  "AUD",
		URL:       "https://example.test/provider/archived",
		Source:    "frontlinehobbies",
	}}}
	if _, err := svc.RunNow(context.Background(), qs.ID, provider); err != nil {
		t.Fatalf("RunNow() first pass error = %v", err)
	}
	if _, err := conn.Exec(`UPDATE scanner_candidates SET status = 'archived' WHERE listing_id = 'ARCHIVED-DISCOVERY-1'`); err != nil {
		t.Fatalf("archive candidate: %v", err)
	}
	provider.items[0].Price = 39.95
	if _, err := svc.RunNow(context.Background(), qs.ID, provider); err != nil {
		t.Fatalf("RunNow() duplicate pass error = %v", err)
	}

	var status string
	var price float64
	if err := conn.QueryRow(`SELECT status, price FROM scanner_candidates WHERE listing_id = 'ARCHIVED-DISCOVERY-1'`).Scan(&status, &price); err != nil {
		t.Fatalf("load candidate: %v", err)
	}
	if status != "archived" {
		t.Fatalf("duplicate provider refresh reintroduced archived candidate with status=%q", status)
	}
	if price != 39.95 {
		t.Fatalf("expected duplicate refresh to update metadata without changing status, price=%v", price)
	}

	var matchCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_matches WHERE candidate_id = (SELECT id FROM scanner_candidates WHERE listing_id = 'ARCHIVED-DISCOVERY-1')`).Scan(&matchCount); err != nil {
		t.Fatalf("load match count: %v", err)
	}
	if matchCount != 1 {
		t.Fatalf("expected one stable discovery match row, got %d", matchCount)
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

func TestQuerySetRunSnapshotUsesDurableRunRecordsAndComputesNextRun(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	qs, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Scheduled durable watch",
		Keywords:      []string{"afx"},
		ProviderScope: []string{"bonzaslotcars"},
		ScheduleCron:  "*/15 * * * *",
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile() error = %v", err)
	}
	provider := &testProvider{
		providerID: "bonzaslotcars",
		items: []CandidateInput{{
			ListingID: "SCHEDULE-1",
			Title:     "Scheduled result",
			URL:       "https://shop.test/schedule-1",
			Source:    "bonzaslotcars",
		}},
	}
	if _, err := svc.RunScheduledForProfile(context.Background(), "profile-a", provider); err != nil {
		t.Fatalf("RunScheduledForProfile() error = %v", err)
	}

	reloaded, err := svc.GetQuerySetForProfile(context.Background(), "profile-a", qs.ID)
	if err != nil {
		t.Fatalf("GetQuerySetForProfile() error = %v", err)
	}
	if reloaded.LastRunStatus != "succeeded" {
		t.Fatalf("expected latest durable run status succeeded, got %+v", reloaded)
	}
	if reloaded.LastRunAt == "" {
		t.Fatalf("expected latest durable run timestamp, got %+v", reloaded)
	}
	if reloaded.LastCandidateCount != 1 {
		t.Fatalf("expected candidate count from durable watch scope, got %+v", reloaded)
	}
	nextRun, ok := parseScannerTime(reloaded.NextRunAt)
	if !ok {
		t.Fatalf("expected computed next_run_at timestamp, got %+v", reloaded)
	}
	lastRun, ok := parseScannerTime(reloaded.LastRunAt)
	if !ok {
		t.Fatalf("expected parseable last_run_at timestamp, got %+v", reloaded)
	}
	if !nextRun.After(lastRun) {
		t.Fatalf("expected next_run_at after last_run_at, got last=%s next=%s", reloaded.LastRunAt, reloaded.NextRunAt)
	}
}

func TestComputeNextRunAtSupportsCommonMarketWatchSchedules(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 6, 27, 17, 46, 0, 0, time.UTC)
	for _, tc := range []struct {
		name     string
		schedule string
		want     string
	}{
		{name: "quarter-hour", schedule: "*/15 * * * *", want: "2026-06-27T18:00:00Z"},
		{name: "six-hourly", schedule: "0 */6 * * *", want: "2026-06-27T18:00:00Z"},
		{name: "fixed-daily", schedule: "30 8 * * *", want: "2026-06-28T08:30:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := computeNextRunAt(tc.schedule, "", base); got != tc.want {
				t.Fatalf("computeNextRunAt(%q)=%q, want %q", tc.schedule, got, tc.want)
			}
		})
	}
	if got := computeNextRunAt("not a cron", "", base); got != "" {
		t.Fatalf("invalid schedule should not produce next run, got %q", got)
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

func TestRunScheduledRecordsPartialFailureWithoutBlockingOtherWatches(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	failingQuery, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Scheduled failing watch",
		Keywords:      []string{"afx fail"},
		ProviderScope: []string{"ebay"},
		ScheduleCron:  "*/15 * * * *",
		Enabled:       true,
		MaxRetryCount: 0,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(failing) error = %v", err)
	}
	healthyQuery, err := svc.CreateQuerySetForProfile(context.Background(), "profile-a", QuerySet{
		Name:          "Scheduled healthy watch",
		Keywords:      []string{"afx ok"},
		ProviderScope: []string{"ebay"},
		ScheduleCron:  "*/15 * * * *",
		Enabled:       true,
		MaxRetryCount: 0,
	})
	if err != nil {
		t.Fatalf("CreateQuerySetForProfile(healthy) error = %v", err)
	}

	ran, err := svc.RunScheduledForProfile(context.Background(), "profile-a", &scheduledMixedProvider{})
	if err == nil {
		t.Fatal("expected partial scheduled failure error")
	}
	if ran != 1 {
		t.Fatalf("expected one healthy scheduled watch to run after one failed watch, got %d", ran)
	}

	var failedRuns, healthyRuns, healthyResults int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND trigger_type = 'scheduled' AND status = 'failed' AND error_category = 'provider_error' AND error_message = 'provider outage' AND retry_guidance <> ''`, "profile-a", failingQuery.ID).Scan(&failedRuns); err != nil {
		t.Fatalf("query failed scheduled runs: %v", err)
	}
	if failedRuns != 1 {
		t.Fatalf("expected durable failed scheduled run with retry guidance, got %d", failedRuns)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND trigger_type = 'scheduled' AND status = 'succeeded' AND result_count = 1 AND new_result_count = 1`, "profile-a", healthyQuery.ID).Scan(&healthyRuns); err != nil {
		t.Fatalf("query healthy scheduled runs: %v", err)
	}
	if healthyRuns != 1 {
		t.Fatalf("expected durable successful scheduled run after partial failure, got %d", healthyRuns)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND listing_id = 'SCHEDULED-OK-1'`, "profile-a", healthyQuery.ID).Scan(&healthyResults); err != nil {
		t.Fatalf("query healthy scheduled candidates: %v", err)
	}
	if healthyResults != 1 {
		t.Fatalf("expected healthy watch candidate to persist despite partial failure, got %d", healthyResults)
	}

	reloadedFailing, err := svc.GetQuerySetForProfile(context.Background(), "profile-a", failingQuery.ID)
	if err != nil {
		t.Fatalf("GetQuerySetForProfile(failing) error = %v", err)
	}
	if reloadedFailing.LastRunStatus != "failed" || reloadedFailing.LastRunMessage != "provider outage" {
		t.Fatalf("expected failed scheduled snapshot, got %+v", reloadedFailing)
	}
	reloadedHealthy, err := svc.GetQuerySetForProfile(context.Background(), "profile-a", healthyQuery.ID)
	if err != nil {
		t.Fatalf("GetQuerySetForProfile(healthy) error = %v", err)
	}
	if reloadedHealthy.LastRunStatus != "succeeded" || reloadedHealthy.LastCandidateCount != 1 {
		t.Fatalf("expected healthy scheduled snapshot to remain succeeded, got %+v", reloadedHealthy)
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
