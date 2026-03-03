package collection

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestAuditMetadataAndHistoryForItemLifecycle(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	created, err := repo.CreateItem(context.Background(), Item{
		PartNumber: "AUD-001",
		Title:      "Audit Item",
		Brand:      "AFX",
		Category:   "Cars",
	})
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if strings.TrimSpace(created.CreatedBy) == "" {
		t.Fatal("expected created_by metadata to be populated")
	}
	if strings.TrimSpace(created.UpdatedBy) == "" {
		t.Fatal("expected updated_by metadata to be populated")
	}

	updated, err := repo.UpdateItem(context.Background(), created.ID, Item{
		Title: "Audit Item Updated",
	})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if strings.TrimSpace(updated.UpdatedBy) == "" {
		t.Fatal("expected updated_by metadata to remain populated")
	}

	deleted, err := repo.SetItemLifecycleStatus(context.Background(), created.ID, "deleted")
	if err != nil {
		t.Fatalf("SetItemLifecycleStatus(deleted) error = %v", err)
	}
	if strings.TrimSpace(deleted.DeletedAt) == "" {
		t.Fatal("expected deleted_at metadata to be populated on soft delete")
	}
	if strings.TrimSpace(deleted.DeletedBy) == "" {
		t.Fatal("expected deleted_by metadata to be populated on soft delete")
	}

	restored, err := repo.SetItemLifecycleStatus(context.Background(), created.ID, "active")
	if err != nil {
		t.Fatalf("SetItemLifecycleStatus(active) error = %v", err)
	}
	if strings.TrimSpace(restored.DeletedAt) != "" {
		t.Fatal("expected deleted_at metadata cleared on restore")
	}
	if strings.TrimSpace(restored.DeletedBy) != "" {
		t.Fatal("expected deleted_by metadata cleared on restore")
	}

	events, err := repo.ListAuditEventsByEntity(context.Background(), "canonical_item", created.ID)
	if err != nil {
		t.Fatalf("ListAuditEventsByEntity() error = %v", err)
	}
	if len(events) < 4 {
		t.Fatalf("expected at least 4 audit events (create/update/delete/restore), got %d", len(events))
	}
	for i := range events {
		if strings.TrimSpace(events[i].ID) == "" {
			t.Fatalf("audit event at index %d missing id", i)
		}
		if strings.TrimSpace(events[i].Actor) == "" {
			t.Fatalf("audit event at index %d missing actor", i)
		}
		if strings.TrimSpace(events[i].Action) == "" {
			t.Fatalf("audit event at index %d missing action", i)
		}
		if strings.TrimSpace(events[i].CreatedAt) == "" {
			t.Fatalf("audit event at index %d missing created_at", i)
		}
		if i > 0 {
			prev, prevErr := time.Parse(time.RFC3339, events[i-1].CreatedAt)
			cur, curErr := time.Parse(time.RFC3339, events[i].CreatedAt)
			if prevErr == nil && curErr == nil && cur.Before(prev) {
				t.Fatalf("expected timeline ordering ascending, got %s before %s", events[i].CreatedAt, events[i-1].CreatedAt)
			}
		}
	}

	var sawTrackedDiff bool
	for _, event := range events {
		if event.Action == "update" {
			beforeTitle, _ := event.Before["title"].(string)
			afterTitle, _ := event.After["title"].(string)
			if beforeTitle != "" && afterTitle == "Audit Item Updated" {
				sawTrackedDiff = true
				break
			}
		}
	}
	if !sawTrackedDiff {
		t.Fatal("expected update audit event to contain before/after tracked field diff for title")
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
