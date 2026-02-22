package search

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestSearchAndSavedFilters(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO profiles (id, name) VALUES ('p1','Default')`); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO canonical_items (id, brand, category, part_number, title, description, tags_json)
		VALUES
		('i1','AFX','Slot Car','A-100','Blue Runner','Fast car','["blue"]'),
		('i2','AFX','Slot Car','A-101','Red Runner','Track ready','["red"]')
	`); err != nil {
		t.Fatalf("seed items: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO instances (id, item_id, condition, status, quantity, storage_location, acquisition_price, notes)
		VALUES
		('s1','i1','mint','sealed',1,'Shelf A',20.0,'Blue notes'),
		('s2','i2','used','loose',1,'Shelf B',10.0,'Red notes')
	`); err != nil {
		t.Fatalf("seed instances: %v", err)
	}

	r := NewRepository(conn)
	items, err := r.SearchItems(context.Background(), Query{Text: "Blue", Limit: 10})
	if err != nil {
		t.Fatalf("SearchItems error = %v", err)
	}
	if len(items) != 1 || items[0].ID != "i1" {
		t.Fatalf("expected i1 only, got %#v", items)
	}
	filtered, err := r.SearchItems(context.Background(), Query{Condition: "used", Status: "loose", Tags: "red", Scale: "", SortBy: "price"})
	if err != nil {
		t.Fatalf("SearchItems filtered error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "i2" {
		t.Fatalf("expected i2 filtered result, got %#v", filtered)
	}

	sf, err := r.SaveFilter(context.Background(), "p1", "Blue Cars", Query{Text: "Blue", Brand: "AFX"})
	if err != nil {
		t.Fatalf("SaveFilter error = %v", err)
	}
	if sf.ID == "" {
		t.Fatal("expected saved filter id")
	}

	saved, err := r.ListSavedFilters(context.Background(), "p1")
	if err != nil {
		t.Fatalf("ListSavedFilters error = %v", err)
	}
	if len(saved) != 1 || saved[0].Name != "Blue Cars" {
		t.Fatalf("unexpected saved filters: %#v", saved)
	}

	updated, err := r.UpdateFilter(context.Background(), sf.ID, "p1", "Blue Cars Updated", Query{Text: "Runner", SortBy: "part_number"})
	if err != nil {
		t.Fatalf("UpdateFilter error = %v", err)
	}
	if updated.Name != "Blue Cars Updated" {
		t.Fatalf("unexpected updated name: %q", updated.Name)
	}

	if err := r.DeleteFilter(context.Background(), sf.ID, "p1"); err != nil {
		t.Fatalf("DeleteFilter error = %v", err)
	}
	savedAfterDelete, err := r.ListSavedFilters(context.Background(), "p1")
	if err != nil {
		t.Fatalf("ListSavedFilters after delete error = %v", err)
	}
	if len(savedAfterDelete) != 0 {
		t.Fatalf("expected no saved filters after delete, got %#v", savedAfterDelete)
	}
}
