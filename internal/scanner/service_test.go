package scanner

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

type testProvider struct {
	failures int
	calls    int
	items    []CandidateInput
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
