package collection

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestCreateAndListItemsAndInstances(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	item, err := repo.CreateItem(context.Background(), Item{
		Brand:      "AFX",
		Category:   "Slot Car",
		PartNumber: "AFX-1001",
		Title:      "Speed Demon",
		Tags:       []string{"blue", "limited"},
	})
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}

	items, err := repo.ListItems(context.Background())
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PartNumber != "AFX-1001" {
		t.Fatalf("unexpected part number: %q", items[0].PartNumber)
	}

	instance, err := repo.CreateInstance(context.Background(), Instance{
		ItemID:   item.ID,
		Status:   "sealed",
		Quantity: 2,
	})
	if err != nil {
		t.Fatalf("CreateInstance() error = %v", err)
	}
	if instance.ID == "" {
		t.Fatal("expected instance id")
	}

	instances, err := repo.ListInstancesByItemID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("ListInstancesByItemID() error = %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].Status != "sealed" {
		t.Fatalf("unexpected instance status: %q", instances[0].Status)
	}

	if _, err := repo.CreateInstance(context.Background(), Instance{
		ItemID: item.ID,
		Status: "invalid_status",
	}); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateItemSupportsMinimumQuickAddFields(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	item, err := repo.CreateItem(context.Background(), Item{
		PartNumber: "QA-100",
		Title:      "Quick Add Item",
	})
	if err != nil {
		t.Fatalf("CreateItem() with minimum fields error = %v", err)
	}
	if item.PartNumber != "QA-100" || item.Title != "Quick Add Item" {
		t.Fatalf("unexpected item payload: %#v", item)
	}
	if item.Brand == "" {
		t.Fatal("expected default brand to be populated")
	}
	if item.Category == "" {
		t.Fatal("expected default category to be populated")
	}
}

func TestUpdateItemAndBulkEditItems(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	first, err := repo.CreateItem(context.Background(), Item{
		PartNumber: "B-1",
		Title:      "First",
		Brand:      "AFX",
		Category:   "Cars",
	})
	if err != nil {
		t.Fatalf("CreateItem(first) error = %v", err)
	}
	second, err := repo.CreateItem(context.Background(), Item{
		PartNumber: "B-2",
		Title:      "Second",
		Brand:      "Tyco",
		Category:   "Cars",
	})
	if err != nil {
		t.Fatalf("CreateItem(second) error = %v", err)
	}

	updated, err := repo.UpdateItem(context.Background(), first.ID, Item{Title: "First Updated", Brand: "Mega G+"})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if updated.Title != "First Updated" || updated.Brand != "Mega G+" {
		t.Fatalf("unexpected updated item: %#v", updated)
	}

	bulk, err := repo.BulkEditItems(context.Background(), []string{first.ID, second.ID}, Item{Category: "Updated Category"})
	if err != nil {
		t.Fatalf("BulkEditItems() error = %v", err)
	}
	if bulk.UpdatedCount != 2 {
		t.Fatalf("expected updated count 2, got %d", bulk.UpdatedCount)
	}

	reloadedFirst, err := repo.GetItemByID(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("GetItemByID(first) error = %v", err)
	}
	if reloadedFirst.Category != "Updated Category" {
		t.Fatalf("expected first category updated, got %q", reloadedFirst.Category)
	}
	reloadedSecond, err := repo.GetItemByID(context.Background(), second.ID)
	if err != nil {
		t.Fatalf("GetItemByID(second) error = %v", err)
	}
	if reloadedSecond.Category != "Updated Category" {
		t.Fatalf("expected second category updated, got %q", reloadedSecond.Category)
	}
}
