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
