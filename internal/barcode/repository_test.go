package barcode

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestAddListLookup(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-a','AFX','Slot Car','A-1','Car A'), ('item-b','AFX','Slot Car','A-2','Car B')`); err != nil {
		t.Fatalf("seed items: %v", err)
	}

	repo := NewRepository(conn)
	if _, err := repo.Add(context.Background(), "item-a", ""); err == nil {
		t.Fatal("expected validation error for empty barcode")
	}
	if _, err := repo.Add(context.Background(), "item-a", "12345"); err != nil {
		t.Fatalf("Add item-a error = %v", err)
	}
	if _, err := repo.Add(context.Background(), "item-b", "12345"); err != nil {
		t.Fatalf("Add item-b error = %v", err)
	}

	aList, err := repo.ListByItem(context.Background(), "item-a")
	if err != nil {
		t.Fatalf("ListByItem error = %v", err)
	}
	if len(aList) != 1 {
		t.Fatalf("expected 1 barcode for item-a, got %d", len(aList))
	}

	lookup, err := repo.Lookup(context.Background(), "12345")
	if err != nil {
		t.Fatalf("Lookup error = %v", err)
	}
	if len(lookup) != 2 {
		t.Fatalf("expected duplicate barcode records 2, got %d", len(lookup))
	}
}
