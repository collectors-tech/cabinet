package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFutureHooksFieldsDisabledByDefault(t *testing.T) {
	t.Parallel()
	conn, err := OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	var forSale int
	var offers string
	if err := conn.QueryRow(`SELECT for_sale, structured_offers_json FROM canonical_items WHERE id = 'i1'`).Scan(&forSale, &offers); err != nil {
		t.Fatalf("query future hook fields error: %v", err)
	}
	if forSale != 0 || offers != "[]" {
		t.Fatalf("expected disabled defaults, got for_sale=%d offers=%s", forSale, offers)
	}
}
