package dashboard

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestSummaryIncludesCoreSignals(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes) VALUES ('in1','i1','used','loose',2,'shelf',10,'','')`); err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('c1','q1','L1','AFX P-9',20,0,'http://x','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2)`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c1','','not_in_collection',0,1,'P-9',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO wishlist_entries(id, item_id, target_price, priority, notes, highlight_hit) VALUES ('w1','i1',30,'high','',1)`); err != nil {
		t.Fatalf("seed wishlist: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count) VALUES ('p1','i1','2026-02-20','ebay',15,15,15,0),('p2','i1','2026-02-21','ebay',12,12,12,4)`); err != nil {
		t.Fatalf("seed snapshots: %v", err)
	}

	svc := NewService(conn)
	s, err := svc.Summary(context.Background())
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if s.Collection.TotalItems != 1 || s.Collection.TotalInstances != 2 {
		t.Fatalf("unexpected collection stats: %+v", s.Collection)
	}
	if s.NewDiscoveries < 1 || s.PriceDrops < 1 {
		t.Fatalf("expected discoveries and price drops, got %+v", s)
	}
	if s.LowStockDiscoveries < 1 || s.Restocks < 1 {
		t.Fatalf("expected low stock and restock signals, got %+v", s)
	}
	if len(s.Cards) == 0 || s.Cards[0].Link == "" {
		t.Fatalf("expected dashboard deep-link cards, got %+v", s.Cards)
	}
}
