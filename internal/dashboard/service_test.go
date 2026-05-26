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
	s, err := svc.Summary(context.Background(), "")
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

	foundRecentlyAdded := false
	for _, card := range s.Cards {
		if card.Title != "Recently Added" {
			continue
		}
		foundRecentlyAdded = true
		if card.Link != "/collections" {
			t.Fatalf("expected Recently Added card to target /collections, got %q", card.Link)
		}
	}
	if !foundRecentlyAdded {
		t.Fatalf("expected Recently Added dashboard card, got %+v", s.Cards)
	}
}

func TestSummaryScopesSignalsToProfile(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO profiles(id, name) VALUES ('profile-a','Profile A'),('profile-b','Profile B')`); err != nil {
		t.Fatalf("seed profiles: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, created_at)
		VALUES
			('item-a','profile-a','AFX','Slot','PA-1','Profile A Camaro','2026-05-01T10:00:00Z'),
			('item-b','profile-b','AFX','Slot','PB-1','Profile B Porsche','2026-05-02T10:00:00Z')
	`); err != nil {
		t.Fatalf("seed items: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES
			('inst-a','item-a','used','loose',2,'shelf',10,'',''),
			('inst-b','item-b','used','loose',7,'case',20,'','')
	`); err != nil {
		t.Fatalf("seed instances: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json)
		VALUES ('query-a','profile-a','A','["pa"]','[]'),('query-b','profile-b','B','["pb"]','[]')
	`); err != nil {
		t.Fatalf("seed query sets: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES
			('cand-a','profile-a','query-a','LA','Profile A PA-1',20,0,'http://a','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2),
			('cand-b','profile-b','query-b','LB','Profile B PB-1',20,0,'http://b','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2)
	`); err != nil {
		t.Fatalf("seed candidates: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('cand-a','','not_in_collection',0,1,'PA-1',CURRENT_TIMESTAMP),
			('cand-b','','not_in_collection',0,1,'PB-1',CURRENT_TIMESTAMP)
	`); err != nil {
		t.Fatalf("seed matches: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit)
		VALUES
			('wish-a','profile-a','item-a',30,'high','',1),
			('wish-b','profile-b','item-b',30,'high','',1)
	`); err != nil {
		t.Fatalf("seed wishlist: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count)
		VALUES
			('price-a1','item-a','2026-02-20','ebay',15,15,15,0),
			('price-a2','item-a','2026-02-21','ebay',12,12,12,4),
			('price-b1','item-b','2026-02-20','ebay',25,25,25,0),
			('price-b2','item-b','2026-02-21','ebay',22,22,22,6)
	`); err != nil {
		t.Fatalf("seed price snapshots: %v", err)
	}

	svc := NewService(conn)
	s, err := svc.Summary(context.Background(), "profile-a")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}

	if s.Collection.TotalItems != 1 || s.Collection.TotalInstances != 2 || s.Collection.EstimatedValue != 20 {
		t.Fatalf("expected profile-a collection stats only, got %+v", s.Collection)
	}
	if s.NewDiscoveries != 1 || s.LowStockDiscoveries != 1 || s.WishlistHits != 1 || s.PriceDrops != 1 || s.Restocks != 1 {
		t.Fatalf("expected profile-a action signals only, got %+v", s)
	}
	if len(s.RecentlyAdded) != 1 || s.RecentlyAdded[0] != "Profile A Camaro" {
		t.Fatalf("expected profile-a recent item only, got %+v", s.RecentlyAdded)
	}
}
