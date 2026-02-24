package chat

import (
	"context"
	"path/filepath"
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

	thread, err := svc.CreateThread(ctx, profileID, "Main")
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.ID == "" {
		t.Fatalf("expected thread ID")
	}

	if _, err := svc.CreateMessage(ctx, profileID, thread.ID, "user", "hello"); err != nil {
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
}
