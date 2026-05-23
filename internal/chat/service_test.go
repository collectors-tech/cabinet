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
	stillPending, err := svc.GetActionPreview(ctx, profileID, preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview(after confirm-required) error = %v", err)
	}
	if stillPending.Status != "previewed" || stillPending.AppliedAt != "" {
		t.Fatalf("expected non-confirmed apply to keep preview unapplied, got %+v", stillPending)
	}
	var itemCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM canonical_items WHERE profile_id = ?`, profileID).Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items after confirm-required apply: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected non-confirmed apply to leave inventory unchanged, got %d items", itemCount)
	}
	msgs, err = svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages(after confirm-required) error = %v", err)
	}
	for _, msg := range msgs {
		if msg.Role == "assistant" && strings.Contains(msg.Content, "Applied create_item_stub") {
			t.Fatalf("non-confirmed apply must not record applied assistant outcome, got %+v", msg)
		}
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

func TestServiceActionPreviewRejectsCrossThreadApply(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "profile-cross-thread"
	if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, "Cross Thread"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	ownerThread, err := svc.CreateThread(ctx, profileID, "Owner Thread", map[string]any{
		"profile": map[string]any{"id": profileID},
	})
	if err != nil {
		t.Fatalf("CreateThread(owner) error = %v", err)
	}
	otherThread, err := svc.CreateThread(ctx, profileID, "Other Thread", map[string]any{
		"profile": map[string]any{"id": profileID},
	})
	if err != nil {
		t.Fatalf("CreateThread(other) error = %v", err)
	}
	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  ownerThread.ID,
		Action:    "create_inventory_item",
		Payload: map[string]any{
			"part_number": "THREAD-OWNER-ONLY",
			"title":       "Owner Thread Only",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(owner) error = %v", err)
	}

	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  otherThread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "preview not found") {
		t.Fatalf("expected wrong-thread preview apply to fail as not found, got %v", err)
	}

	var itemCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM canonical_items WHERE profile_id = ? AND part_number = 'THREAD-OWNER-ONLY'", profileID).Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items after rejected thread apply: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected rejected wrong-thread apply to leave inventory unchanged, got %d items", itemCount)
	}

	stillPending, err := svc.GetActionPreview(ctx, profileID, preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview(owner) error = %v", err)
	}
	if stillPending.Status != "previewed" || stillPending.ThreadID != ownerThread.ID || stillPending.AppliedAt != "" {
		t.Fatalf("expected owner preview to remain pending on original thread, got %+v", stillPending)
	}
	for _, threadID := range []string{ownerThread.ID, otherThread.ID} {
		msgs, err := svc.ListMessages(ctx, profileID, threadID)
		if err != nil {
			t.Fatalf("ListMessages(%s) error = %v", threadID, err)
		}
		for _, msg := range msgs {
			if msg.Role == "assistant" && strings.Contains(msg.Content, "Applied create_inventory_item") {
				t.Fatalf("wrong-thread apply must not record applied assistant outcome in thread %s, got %+v", threadID, msg)
			}
		}
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

func TestServiceCollectionAssignmentRejectsMissingTarget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "profile-collection-missing"
	if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, "Collection Missing"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	thread, err := svc.CreateThread(ctx, profileID, "Collection Missing Thread", map[string]any{
		"profile": map[string]any{"id": profileID},
	})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "assign_collection_item",
		Payload: map[string]any{
			"item_id":         "missing-collection-target",
			"collection_name": "Do Not Create",
			"part_number":     "MISSING-COLLECTION-001",
			"title":           "Missing Collection Target",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(assign_collection_item) error = %v", err)
	}

	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "collection assignment target not found") {
		t.Fatalf("expected missing collection target apply to fail, got %v", err)
	}

	var workspaceSettings int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND key = 'collections.workspace.v1'", profileID).Scan(&workspaceSettings); err != nil {
		t.Fatalf("count workspace collection settings after failed assign: %v", err)
	}
	if workspaceSettings != 0 {
		t.Fatalf("expected failed collection assignment to leave workspace collections unchanged, got %d settings", workspaceSettings)
	}

	stillPending, err := svc.GetActionPreview(ctx, profileID, preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview() error = %v", err)
	}
	if stillPending.Status != "previewed" {
		t.Fatalf("expected failed collection assignment to keep preview pending, got %+v", stillPending)
	}
	msgs, err := svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	for _, msg := range msgs {
		if msg.Role == "assistant" && strings.Contains(msg.Content, "Applied assign_collection_item") {
			t.Fatalf("failed collection assignment must not record applied assistant outcome, got %+v", msg)
		}
	}
}

func TestServiceCancelPreviewRecordsOutcomeWithoutMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "profile-cancel-preview"
	if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, "Cancel Preview"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	thread, err := svc.CreateThread(ctx, profileID, "Cancel Preview Thread", map[string]any{
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
			"item_id":     "cancel-target-item",
			"part_number": "CANCEL-SHOULD-NOT-APPLY",
			"title":       "Canceled Update Title",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(update_inventory_item) error = %v", err)
	}

	result, err := svc.CancelAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
	})
	if err != nil {
		t.Fatalf("CancelAction() error = %v", err)
	}
	if result.Applied || result.Action != "update_inventory_item" || result.ItemID != "cancel-target-item" {
		t.Fatalf("expected canceled update result with target and no mutation, got %+v", result)
	}
	var itemCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM canonical_items WHERE id = 'cancel-target-item' OR part_number = 'CANCEL-SHOULD-NOT-APPLY' OR title = 'Canceled Update Title'").Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items after cancel: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected cancel to leave inventory unchanged, got %d matching items", itemCount)
	}
	cancelled, err := svc.GetActionPreview(ctx, profileID, preview.ID)
	if err != nil {
		t.Fatalf("GetActionPreview() error = %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("expected preview status cancelled, got %+v", cancelled)
	}
	msgs, err := svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	last := msgs[len(msgs)-1]
	if last.Role != "assistant" ||
		!strings.Contains(last.Content, "Canceled update_inventory_item") ||
		!strings.Contains(last.Content, "no mutation applied") ||
		strings.Contains(last.Content, "Applied update_inventory_item") {
		t.Fatalf("expected canceled assistant outcome without applied wording, got %+v", last)
	}
	resultContext, _ := last.Context["action_result"].(map[string]any)
	if resultContext["confirmation"] != "cancelled" || resultContext["mutation_applied"] != false {
		t.Fatalf("expected canceled action_result context, got %+v", resultContext)
	}
	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "preview already applied") {
		t.Fatalf("expected canceled preview to reject later apply, got %v", err)
	}
}

func TestServiceCancelCollectionAssignmentRecordsTargetWithoutMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	profileID := "profile-cancel-collection"
	itemID := "collection-cancel-target"
	if _, err := conn.ExecContext(ctx, "INSERT INTO profiles(id, name) VALUES (?, ?)", profileID, "Cancel Collection"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "INSERT INTO canonical_items(id, profile_id, part_number, title, brand, category) VALUES (?, ?, ?, ?, ?, ?)", itemID, profileID, "COL-CANCEL-001", "Collection Cancel Item", "AFX", "General"); err != nil {
		t.Fatalf("insert inventory item: %v", err)
	}
	thread, err := svc.CreateThread(ctx, profileID, "Cancel Collection Thread", map[string]any{
		"profile": map[string]any{"id": profileID},
	})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}

	preview, err := svc.PreviewAction(ctx, PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "assign_collection_item",
		Payload: map[string]any{
			"item_id":         itemID,
			"collection_name": "Do Not Assign",
		},
	})
	if err != nil {
		t.Fatalf("PreviewAction(assign_collection_item) error = %v", err)
	}

	result, err := svc.CancelAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
	})
	if err != nil {
		t.Fatalf("CancelAction() error = %v", err)
	}
	if result.Applied || result.Action != "assign_collection_item" || result.ItemID != itemID || result.CollectionName != "Do Not Assign" {
		t.Fatalf("expected canceled collection assignment target with no mutation, got %+v", result)
	}

	var workspaceSettings int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND key = 'collections.workspace.v1' AND value LIKE '%Do Not Assign%'", profileID).Scan(&workspaceSettings); err != nil {
		t.Fatalf("count workspace collection settings: %v", err)
	}
	if workspaceSettings != 0 {
		t.Fatalf("expected cancel to leave collection membership unchanged, got %d matching workspace settings", workspaceSettings)
	}
	msgs, err := svc.ListMessages(ctx, profileID, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	last := msgs[len(msgs)-1]
	if last.Role != "assistant" ||
		!strings.Contains(last.Content, "Canceled assign_collection_item") ||
		!strings.Contains(last.Content, itemID) ||
		!strings.Contains(last.Content, "Do Not Assign") ||
		!strings.Contains(last.Content, "no mutation applied") {
		t.Fatalf("expected collection cancel assistant outcome with target evidence, got %+v", last)
	}
	resultContext, _ := last.Context["action_result"].(map[string]any)
	if resultContext["collection_name"] != "Do Not Assign" || resultContext["mutation_applied"] != false {
		t.Fatalf("expected canceled collection action_result context, got %+v", resultContext)
	}
	if _, err := svc.ApplyAction(ctx, ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		PreviewID: preview.ID,
		Confirm:   true,
	}); err == nil || !strings.Contains(err.Error(), "preview already applied") {
		t.Fatalf("expected canceled collection preview to reject later apply, got %v", err)
	}
}
