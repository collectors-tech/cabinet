package chat

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestServiceThreadMessagePreviewApplyLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "p1"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Default"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	thread, err := svc.CreateThread(ctx, profileID, "Main", map[string]any{
		"provider": "openai",
		"model":    "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.ID == "" {
		t.Fatalf("expected thread ID")
	}

	if _, err := svc.CreateMessage(ctx, profileID, thread.ID, "user", "hello", map[string]any{
		"route":   map[string]any{"pathname": "/inventory"},
		"profile": map[string]any{"id": profileID},
	}); err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	msgs, err := svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "create_item_stub",
		Payload: map[string]any{
			"part_number": "CHAT-001",
			"title":       "Chat Item",
			"brand":       "AFX",
			"category":    "General",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction() error = %v", err)
	}

	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   false,
	}); err == nil {
		t.Fatalf("expected confirm-required error")
	}

	applied, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	})
	if err != nil {
		t.Fatalf("ApplyAction() error = %v", err)
	}
	if !applied.Applied {
		t.Fatalf("expected applied=true")
	}
	if applied.ItemID == "" {
		t.Fatalf("expected created item id")
	}

	wishlistPreview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "create_wishlist_entry",
		Payload: map[string]any{
			"part_number": "CHAT-WISH-001",
			"title":       "Wishlist Item",
			"brand":       "AFX",
			"category":    "General",
			"priority":    "medium",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(create_wishlist_entry) error = %v", err)
	}
	wishlistResult, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: wishlistPreview.ID,
		Confirm:   true,
	})
	if err != nil {
		t.Fatalf("ApplyAction(create_wishlist_entry) error = %v", err)
	}
	if wishlistResult.WishlistID == "" || wishlistResult.ItemID == "" {
		t.Fatalf("expected wishlist and item ids, got wishlist=%q item=%q", wishlistResult.WishlistID, wishlistResult.ItemID)
	}
	var wishlistItemID string
	if err := conn.QueryRowContext(ctx, `SELECT item_id FROM wishlist_entries WHERE id = ? AND profile_id = ?`, wishlistResult.WishlistID, profileID).Scan(&wishlistItemID); err != nil {
		t.Fatalf("load created wishlist entry: %v", err)
	}
	if wishlistItemID != wishlistResult.ItemID {
		t.Fatalf("expected wishlist entry to reference created item %q, got %q", wishlistResult.ItemID, wishlistItemID)
	}
	msgs, err = svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages(after wishlist) error = %v", err)
	}
	wishlistMessage := msgs[len(msgs)-1]
	if wishlistMessage.Role != "assistant" ||
		!strings.Contains(wishlistMessage.Content, "Applied create_wishlist_entry to wishlist") ||
		!strings.Contains(wishlistMessage.Content, wishlistResult.WishlistID) ||
		!strings.Contains(wishlistMessage.Content, wishlistResult.ItemID) {
		t.Fatalf("expected wishlist assistant outcome with entry and item ids, got %+v", wishlistMessage)
	}

	updatePreview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "update_inventory_item",
		Payload: map[string]any{
			"item_id":     applied.ItemID,
			"title":       "Updated via Chat",
			"part_number": "CHAT-001-UPDATED",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(update_inventory_item) error = %v", err)
	}
	updateResult, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: updatePreview.ID,
		Confirm:   true,
	})
	if err != nil {
		t.Fatalf("ApplyAction(update_inventory_item) error = %v", err)
	}
	if updateResult.ItemID != applied.ItemID {
		t.Fatalf("expected update item id %q, got %q", applied.ItemID, updateResult.ItemID)
	}
	if updateResult.PartNumber != "CHAT-001-UPDATED" || updateResult.Title != "Updated via Chat" {
		t.Fatalf("expected update result field evidence, got %+v", updateResult)
	}

	var updatedTitle string
	if err := conn.QueryRowContext(ctx, `SELECT title FROM canonical_items WHERE id = ?`, applied.ItemID).Scan(&updatedTitle); err != nil {
		t.Fatalf("load updated canonical item: %v", err)
	}
	if updatedTitle != "Updated via Chat" {
		t.Fatalf("expected updated title, got %q", updatedTitle)
	}
	msgs, err = svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages(after update) error = %v", err)
	}
	updateMessage := msgs[len(msgs)-1]
	if updateMessage.Role != "assistant" ||
		!strings.Contains(updateMessage.Content, "Applied update_inventory_item") ||
		!strings.Contains(updateMessage.Content, "part_number=CHAT-001-UPDATED") ||
		!strings.Contains(updateMessage.Content, "title=Updated via Chat") {
		t.Fatalf("expected update assistant outcome with changed field evidence, got %+v", updateMessage)
	}
	resultContext, _ := updateMessage.Context["action_result"].(map[string]any)
	if resultContext["part_number"] != "CHAT-001-UPDATED" || resultContext["title"] != "Updated via Chat" {
		t.Fatalf("expected update action_result context field evidence, got %+v", resultContext)
	}

	assignPreview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "assign_collection_item",
		Payload: map[string]any{
			"item_id":         applied.ItemID,
			"part_number":     "CHAT-001-UPDATED",
			"title":           "Updated via Chat",
			"collection_name": "Store 1",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(assign_collection_item) error = %v", err)
	}
	assignResult, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: assignPreview.ID,
		Confirm:   true,
	})
	if err != nil {
		t.Fatalf("ApplyAction(assign_collection_item) error = %v", err)
	}
	if assignResult.ItemID != applied.ItemID || assignResult.CollectionName != "Store 1" {
		t.Fatalf("expected collection assignment result for %q Store 1, got %+v", applied.ItemID, assignResult)
	}
	var workspaceState string
	if err := conn.QueryRowContext(ctx, `SELECT value FROM profile_settings WHERE profile_id = ? AND key = 'collections.workspace.v1'`, profileID).Scan(&workspaceState); err != nil {
		t.Fatalf("load workspace collection setting: %v", err)
	}
	if !strings.Contains(workspaceState, applied.ItemID) || !strings.Contains(workspaceState, "Store 1") {
		t.Fatalf("expected persisted workspace assignment, got %s", workspaceState)
	}
	msgs, err = svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages(after assign) error = %v", err)
	}
	last := msgs[len(msgs)-1]
	if last.Role != "assistant" || !strings.Contains(last.Content, "Applied assign_collection_item") {
		t.Fatalf("expected assistant outcome message, got %+v", last)
	}
}

func TestServiceActionPreviewRejectsCrossProfileApply(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	for _, profileID := range []string{"profile-a", "profile-b"} {
		if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, profileID); err != nil {
			t.Fatalf("insert profile %s: %v", profileID, err)
		}
	}

	ownerThread, err := svc.CreateThread(ctx, "profile-a", "Owner Thread", map[string]any{
		"profile": map[string]any{"id": "profile-a"},
	})
	if err != nil {
		t.Fatalf("CreateThread(owner) error = %v", err)
	}
	otherThread, err := svc.CreateThread(ctx, "profile-b", "Other Thread", map[string]any{
		"profile": map[string]any{"id": "profile-b"},
	})
	if err != nil {
		t.Fatalf("CreateThread(other) error = %v", err)
	}
	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: "profile-a",
		ThreadID:  ownerThread.ID,
		Action:    "create_inventory_item",
		Payload: map[string]any{
			"part_number": "PROFILE-A-ONLY",
			"title":       "Profile A Only",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(owner) error = %v", err)
	}

	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: "profile-b",
		ThreadID:  otherThread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "preview not found") {
		t.Fatalf("expected wrong-profile preview lookup to fail as not found, got %v", err)
	}

	var itemCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM canonical_items WHERE part_number = 'PROFILE-A-ONLY'").Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items after rejected apply: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected rejected wrong-profile apply to leave inventory unchanged, got %d items", itemCount)
	}

	stillPending, err := svc.GetActionPreview(ctx, "profile-a", preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview(owner) error = %v", err)
	}
	if stillPending.Status != "previewed" || stillPending.ThreadID != ownerThread.ID {
		t.Fatalf("expected owner preview to remain pending on original thread, got %+v", stillPending)
	}
}

func TestServiceUpdatePreviewApplyRejectsMissingTarget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "profile-update-missing"
	if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, "Update Missing"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	thread, err := svc.CreateThread(ctx, profileID, "Update Missing Thread", map[string]any{
		"profile": map[string]any{"id": profileID},
	})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "update_inventory_item",
		Payload: map[string]any{
			"item_id":     "missing-item-id",
			"title":       "Should Not Apply",
			"part_number": "MISSING-UPDATE-001",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(update_inventory_item) error = %v", err)
	}

	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "update item target not found") {
		t.Fatalf("expected missing update target apply to fail, got %v", err)
	}

	var itemCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM canonical_items WHERE part_number = 'MISSING-UPDATE-001' OR title = 'Should Not Apply'").Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items after failed update apply: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected failed update apply to leave inventory unchanged, got %d matching items", itemCount)
	}

	stillPending, err := svc.GetActionPreview(ctx, profileID, preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview() error = %v", err)
	}
	if stillPending.Status != "previewed" {
		t.Fatalf("expected failed update apply to keep preview pending, got %+v", stillPending)
	}
	msgs, err := svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	for _, msg := range msgs {
		if msg.Role == "assistant" && strings.Contains(msg.Content, "Applied update_inventory_item") {
			t.Fatalf("failed update apply must not record applied assistant outcome, got %+v", msg)
		}
	}
}
