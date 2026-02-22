package datamgmt

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestDryRunAndApplyImport(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('existing','AFX','Slot Car','P-100','Existing Car')`); err != nil {
		t.Fatalf("seed existing item: %v", err)
	}

	svc := NewService(conn)
	snap := Snapshot{
		SchemaVersion: 1,
		Items: []SnapshotItem{
			{
				Brand:      "AFX",
				Category:   "Slot Car",
				PartNumber: "P-100",
				Title:      "Conflict Item",
				Barcodes:   []string{"123"},
				Instances:  []SnapshotInstance{{Status: "sealed", Quantity: 1}},
			},
			{
				Brand:      "AFX",
				Category:   "Slot Car",
				PartNumber: "P-101",
				Title:      "New Item",
				Barcodes:   []string{"124"},
				Instances:  []SnapshotInstance{{Status: "loose", Quantity: 2}},
			},
		},
	}

	sum, err := svc.DryRunImport(context.Background(), snap)
	if err != nil {
		t.Fatalf("DryRunImport() error = %v", err)
	}
	if sum.TotalItems != 2 || sum.NewItems != 1 || sum.Conflicts != 1 {
		t.Fatalf("unexpected dry run summary: %#v", sum)
	}

	if err := svc.ApplyImport(context.Background(), snap, ApplyOptions{DefaultAction: "merge"}); err != nil {
		t.Fatalf("ApplyImport() error = %v", err)
	}

	var itemCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM canonical_items`).Scan(&itemCount); err != nil {
		t.Fatalf("count canonical_items: %v", err)
	}
	if itemCount != 2 {
		t.Fatalf("expected 2 canonical items after import, got %d", itemCount)
	}
}

func TestParseCSVToSnapshot(t *testing.T) {
	t.Parallel()

	svc := NewService(nil)
	csvInput := "brand,category,part_number,title,make,model,year,scale,series,description\nAFX,Slot Car,P-200,CSV Car,Tomy,Turbo,1990,HO,Series A,Imported\n"
	snap, err := svc.ParseCSVToSnapshot(CSVImportRequest{
		CSV: csvInput,
	})
	if err != nil {
		t.Fatalf("ParseCSVToSnapshot() error = %v", err)
	}
	if snap.SchemaVersion != 1 {
		t.Fatalf("expected schema version 1, got %d", snap.SchemaVersion)
	}
	if len(snap.Items) != 1 {
		t.Fatalf("expected 1 item from csv, got %d", len(snap.Items))
	}
	if snap.Items[0].PartNumber != "P-200" {
		t.Fatalf("unexpected part number: %q", snap.Items[0].PartNumber)
	}
}
