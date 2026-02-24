package discovery

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestListNotInCollectionAndActions(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('c1','q1','L1','AFX P-2',10,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2)`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c1','','not_in_collection',0,1,'P-2',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	svc := NewService(conn)
	items, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected stock fields in discovery item, got %+v", items[0])
	}

	if err := svc.ApplyAction(context.Background(), Action{
		CandidateID: "c1",
		Type:        ActionIgnore,
	}); err != nil {
		t.Fatalf("ApplyAction(ignore) error = %v", err)
	}
	afterIgnore, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() after ignore error = %v", err)
	}
	if len(afterIgnore) != 0 {
		t.Fatalf("expected 0 items after ignore, got %d", len(afterIgnore))
	}
	if err := svc.ResetIgnored(context.Background()); err != nil {
		t.Fatalf("ResetIgnored() error = %v", err)
	}
	afterReset, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() after reset error = %v", err)
	}
	if len(afterReset) != 1 {
		t.Fatalf("expected 1 item after reset, got %d", len(afterReset))
	}
}
