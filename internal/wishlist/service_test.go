package wishlist

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestWishlistCRUDAndBelowTarget(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c1','q1','L1','AFX P-1',35,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	svc := NewService(conn)
	created, err := svc.Create(context.Background(), Entry{
		ItemID:       "i1",
		TargetPrice:  50,
		Priority:     "high",
		Notes:        "want soon",
		HighlightHit: true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	all, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 1 || all[0].ID != created.ID {
		t.Fatalf("unexpected list: %+v", all)
	}
	if !all[0].BelowTargetNow {
		t.Fatal("expected below-target indicator true")
	}
	hits, err := svc.Hits(context.Background(), "i1")
	if err != nil {
		t.Fatalf("Hits() error = %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected scanner hits for wishlist item")
	}
	if err := svc.Update(context.Background(), Entry{
		ID:          created.ID,
		ItemID:      "i1",
		TargetPrice: 40,
		Priority:    "low",
		Notes:       "later",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	empty, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty wishlist, got %d", len(empty))
	}
}
