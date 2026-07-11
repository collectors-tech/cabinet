package datamgmt

import (
	"context"
	"path/filepath"
	"strings"
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

	applySummary, err := svc.ApplyImport(context.Background(), snap, ApplyOptions{DefaultAction: "merge"})
	if err != nil {
		t.Fatalf("ApplyImport() error = %v", err)
	}
	if applySummary.TotalItems != 2 || applySummary.Created != 1 || applySummary.Merged != 1 || applySummary.Skipped != 0 || applySummary.Failed != 0 {
		t.Fatalf("unexpected apply summary: %#v", applySummary)
	}

	var itemCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM canonical_items`).Scan(&itemCount); err != nil {
		t.Fatalf("count canonical_items: %v", err)
	}
	if itemCount != 2 {
		t.Fatalf("expected 2 canonical items after import, got %d", itemCount)
	}
}

func TestDryRunAndApplyImportUseMatchingPartNumberNormalization(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('existing','AFX','Slot Car','P-400','Existing Car')`); err != nil {
		t.Fatalf("seed existing item: %v", err)
	}

	svc := NewService(conn)
	snap := Snapshot{
		SchemaVersion: 1,
		Items: []SnapshotItem{
			{Brand: "AFX", Category: "Slot Car", PartNumber: " P-400 ", Title: "Whitespace Conflict"},
		},
	}

	dryRun, err := svc.DryRunImport(context.Background(), snap)
	if err != nil {
		t.Fatalf("DryRunImport() error = %v", err)
	}
	if dryRun.TotalItems != 1 || dryRun.NewItems != 0 || dryRun.Conflicts != 1 {
		t.Fatalf("expected dry run to match apply conflict normalization, got %#v", dryRun)
	}
	if len(dryRun.ConflictDetails) != 1 || dryRun.ConflictDetails[0].PartNumber != "P-400" || dryRun.ConflictDetails[0].ExistingID != "existing" {
		t.Fatalf("unexpected normalized conflict detail: %#v", dryRun.ConflictDetails)
	}

	apply, err := svc.ApplyImport(context.Background(), snap, ApplyOptions{DefaultAction: "merge"})
	if err != nil {
		t.Fatalf("ApplyImport() error = %v", err)
	}
	if apply.TotalItems != 1 || apply.Created != 0 || apply.Merged != 1 || apply.Failed != 0 {
		t.Fatalf("expected apply to merge the same conflict reported by dry run, got %#v", apply)
	}
}

func TestExportSnapshotImportsIntoCleanDatabaseWithRelationships(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join(t.TempDir(), "source.db")
	sourceConn, err := db.OpenAndMigrate(context.Background(), sourcePath)
	if err != nil {
		t.Fatalf("open source database: %v", err)
	}
	t.Cleanup(func() { _ = sourceConn.Close() })

	if _, err := sourceConn.Exec(`
		INSERT INTO canonical_items (id, brand, category, part_number, title, tags_json)
		VALUES ('roundtrip-item','AFX','Slot Car','P-500','Relationship Roundtrip','["sealed","tracked"]')
	`); err != nil {
		t.Fatalf("seed source item: %v", err)
	}
	if _, err := sourceConn.Exec(`INSERT INTO item_barcodes (id, item_id, barcode) VALUES ('roundtrip-barcode', 'roundtrip-item', '0123456789012')`); err != nil {
		t.Fatalf("seed source barcode: %v", err)
	}
	if _, err := sourceConn.Exec(`
		INSERT INTO instances (id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES ('roundtrip-instance', 'roundtrip-item', 'mint', 'sealed', 2, 'Shelf A', 49.95, '2026-07-11', 'boxed pair')
	`); err != nil {
		t.Fatalf("seed source instance: %v", err)
	}
	if _, err := sourceConn.Exec(`
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order)
		VALUES ('roundtrip-photo', 'roundtrip-item', 'front.jpg', 'media/items/roundtrip/front.jpg', 'media/items/roundtrip/front-preview.jpg', 'media/items/roundtrip/front-thumb.jpg', 1, 3)
	`); err != nil {
		t.Fatalf("seed source photo: %v", err)
	}

	snapshot, err := NewService(sourceConn).ExportSnapshot(context.Background())
	if err != nil {
		t.Fatalf("export source snapshot: %v", err)
	}
	if len(snapshot.Items) != 1 || len(snapshot.Items[0].Barcodes) != 1 || len(snapshot.Items[0].Instances) != 1 || len(snapshot.Items[0].Photos) != 1 {
		t.Fatalf("export snapshot missing relationship evidence: %#v", snapshot.Items)
	}

	targetPath := filepath.Join(t.TempDir(), "target.db")
	targetConn, err := db.OpenAndMigrate(context.Background(), targetPath)
	if err != nil {
		t.Fatalf("open target database: %v", err)
	}
	t.Cleanup(func() { _ = targetConn.Close() })

	apply, err := NewService(targetConn).ApplyImport(context.Background(), snapshot, ApplyOptions{DefaultAction: "merge"})
	if err != nil {
		t.Fatalf("apply snapshot into clean target: %v", err)
	}
	if apply.TotalItems != 1 || apply.Created != 1 || apply.Merged != 0 || apply.Skipped != 0 || apply.Failed != 0 {
		t.Fatalf("unexpected clean target apply summary: %#v", apply)
	}

	roundTrip, err := NewService(targetConn).ExportSnapshot(context.Background())
	if err != nil {
		t.Fatalf("export round-trip snapshot: %v", err)
	}
	if len(roundTrip.Items) != 1 {
		t.Fatalf("expected one round-trip item, got %#v", roundTrip.Items)
	}
	got := roundTrip.Items[0]
	if got.PartNumber != "P-500" || len(got.Tags) != 2 || got.Tags[0] != "sealed" || len(got.Barcodes) != 1 || got.Barcodes[0] != "0123456789012" {
		t.Fatalf("round-trip item metadata/barcodes mismatch: %#v", got)
	}
	if len(got.Instances) != 1 {
		t.Fatalf("expected one round-trip instance, got %#v", got.Instances)
	}
	instance := got.Instances[0]
	if instance.Quantity != 2 || instance.Condition != "mint" || instance.Status != "sealed" || instance.StorageLocation != "Shelf A" || instance.AcquisitionPrice != 49.95 || instance.AcquisitionDate != "2026-07-11" || instance.Notes != "boxed pair" {
		t.Fatalf("round-trip instance relationship mismatch: %#v", instance)
	}
	if len(got.Photos) != 1 {
		t.Fatalf("expected one round-trip photo reference, got %#v", got.Photos)
	}
	photo := got.Photos[0]
	if photo.Filename != "front.jpg" || photo.OriginalPath != "media/items/roundtrip/front.jpg" || photo.PreviewPath != "media/items/roundtrip/front-preview.jpg" || photo.ThumbnailPath != "media/items/roundtrip/front-thumb.jpg" || !photo.IsPrimary || photo.DisplayOrder != 3 {
		t.Fatalf("round-trip photo reference mismatch: %#v", photo)
	}
}

func TestApplyImportReportsSkippedAndInvalidActionFailures(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('existing','AFX','Slot Car','P-300','Existing Car')`); err != nil {
		t.Fatalf("seed existing item: %v", err)
	}

	svc := NewService(conn)
	snap := Snapshot{
		SchemaVersion: 1,
		Items: []SnapshotItem{
			{Brand: "AFX", Category: "Slot Car", PartNumber: "P-300", Title: "Skip Existing"},
			{Brand: "AFX", Category: "Slot Car", PartNumber: "P-301", Title: "Create New"},
		},
	}

	summary, err := svc.ApplyImport(context.Background(), snap, ApplyOptions{
		DefaultAction: "merge",
		Overrides:     map[string]string{"P-300": "skip"},
	})
	if err != nil {
		t.Fatalf("ApplyImport() with skip override error = %v", err)
	}
	if summary.TotalItems != 2 || summary.Created != 1 || summary.Merged != 0 || summary.Skipped != 1 || summary.Failed != 0 {
		t.Fatalf("unexpected skip apply summary: %#v", summary)
	}

	failingSummary, err := svc.ApplyImport(context.Background(), snap, ApplyOptions{DefaultAction: "delete"})
	if err == nil {
		t.Fatal("expected invalid action error")
	}
	if failingSummary.TotalItems != 2 || failingSummary.Failed != 2 {
		t.Fatalf("expected failed count for rejected apply, got %#v", failingSummary)
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

func TestParseCSVToSnapshotWithMappingAndMissingOptionalColumns(t *testing.T) {
	t.Parallel()

	svc := NewService(nil)
	csvInput := "maker,kind,pn,name\nAFX,Slot Car,P-201,Mapped Car\n"
	snap, err := svc.ParseCSVToSnapshot(CSVImportRequest{
		CSV: csvInput,
		Mapping: map[string]string{
			"brand":       "maker",
			"category":    "kind",
			"part_number": "pn",
			"title":       "name",
		},
	})
	if err != nil {
		t.Fatalf("ParseCSVToSnapshot() with mapping error = %v", err)
	}
	if len(snap.Items) != 1 {
		t.Fatalf("expected 1 item from mapped csv, got %d", len(snap.Items))
	}
	if snap.Items[0].Make != "" || snap.Items[0].Description != "" {
		t.Fatalf("optional fields should be empty when not present: %#v", snap.Items[0])
	}
}

func TestMaintenanceSummaries(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn)
	reindex, err := svc.Reindex(context.Background())
	if err != nil {
		t.Fatalf("Reindex() error = %v", err)
	}
	if !reindex.OK || reindex.Operation != "reindex_search" || !reindex.RebuiltSearchIndex || strings.TrimSpace(reindex.CompletedAt) == "" {
		t.Fatalf("unexpected reindex summary: %#v", reindex)
	}

	repair, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("Repair() error = %v", err)
	}
	if !repair.OK || repair.Operation != "integrity_check" || repair.IntegrityCheck != "ok" || strings.TrimSpace(repair.CompletedAt) == "" {
		t.Fatalf("unexpected repair summary: %#v", repair)
	}
}
