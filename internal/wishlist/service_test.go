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
		Currency:     "aud",
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
	if len(all) != 1 || all[0].ID != created.ID || all[0].Currency != "AUD" {
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
		Currency:    "nzd",
		Priority:    "low",
		Notes:       "later",
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	updated, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID() after currency update error = %v", err)
	}
	if updated.Currency != "NZD" {
		t.Fatalf("expected normalized updated currency NZD, got %+v", updated)
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
	deleted, err := svc.ListByProfileDeleted(context.Background(), "", true)
	if err != nil {
		t.Fatalf("ListByProfileDeleted() after delete error = %v", err)
	}
	if len(deleted) != 1 || deleted[0].ID != created.ID || !deleted[0].Deleted {
		t.Fatalf("expected soft-deleted wishlist row, got %+v", deleted)
	}
	var itemStatus string
	if err := conn.QueryRow(`SELECT status FROM canonical_items WHERE id = 'i1'`).Scan(&itemStatus); err != nil {
		t.Fatalf("load item status after soft delete: %v", err)
	}
	if itemStatus != "active" {
		t.Fatalf("expected soft delete to restore item status active, got %q", itemStatus)
	}
	if err := svc.RestoreForProfile(context.Background(), "", created.ID); err != nil {
		t.Fatalf("RestoreForProfile() error = %v", err)
	}
	restored, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() after restore error = %v", err)
	}
	if len(restored) != 1 || restored[0].ID != created.ID || restored[0].Deleted {
		t.Fatalf("expected restored wishlist row, got %+v", restored)
	}
	if err := conn.QueryRow(`SELECT status FROM canonical_items WHERE id = 'i1'`).Scan(&itemStatus); err != nil {
		t.Fatalf("load item status after restore: %v", err)
	}
	if itemStatus != "wishlist" {
		t.Fatalf("expected restore to set item status wishlist, got %q", itemStatus)
	}
	if _, err := conn.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, notes) VALUES ('inst1','i1','loose','loose',1,'keep inventory')`); err != nil {
		t.Fatalf("seed inventory instance: %v", err)
	}
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() before permanent delete error = %v", err)
	}
	if err := svc.PermanentDeleteForProfile(context.Background(), "", created.ID); err != nil {
		t.Fatalf("PermanentDeleteForProfile() error = %v", err)
	}
	deletedAfterPermanent, err := svc.ListByProfileDeleted(context.Background(), "", true)
	if err != nil {
		t.Fatalf("ListByProfileDeleted() after permanent delete error = %v", err)
	}
	if len(deletedAfterPermanent) != 0 {
		t.Fatalf("expected permanent delete to remove deleted row, got %+v", deletedAfterPermanent)
	}
	var instanceCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM instances WHERE item_id = 'i1'`).Scan(&instanceCount); err != nil {
		t.Fatalf("count inventory instance after permanent delete: %v", err)
	}
	if instanceCount != 1 {
		t.Fatalf("expected permanent wishlist delete to preserve inventory instances, got %d", instanceCount)
	}
}

func TestWishlistCurrencyDefaultsLegacyInputAndRejectsInvalidCode(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('legacy-item','AFX','Slot','LEGACY-CURRENCY','Legacy Currency')`); err != nil {
		t.Fatalf("seed legacy item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('invalid-item','AFX','Slot','INVALID-CURRENCY','Invalid Currency')`); err != nil {
		t.Fatalf("seed invalid item: %v", err)
	}

	svc := NewService(conn)
	legacy, err := svc.Create(context.Background(), Entry{ItemID: "legacy-item", TargetPrice: 25})
	if err != nil {
		t.Fatalf("Create() legacy-default currency error = %v", err)
	}
	if legacy.Currency != "USD" {
		t.Fatalf("expected backward-compatible USD default, got %+v", legacy)
	}

	if _, err := svc.Create(context.Background(), Entry{ItemID: "invalid-item", TargetPrice: 25, Currency: "dollars"}); err == nil {
		t.Fatal("expected invalid non-ISO currency to fail")
	}
}
