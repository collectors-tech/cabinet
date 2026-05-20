package telegramcapture

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/db"
)

type staticAuthorizer map[string]string

func (a staticAuthorizer) AuthorizeTelegramCapture(_ context.Context, senderID, chatID string) (AuthorizedProfile, error) {
	profileID := a[strings.TrimSpace(senderID)+"|"+strings.TrimSpace(chatID)]
	return AuthorizedProfile{ProfileID: profileID}, nil
}

func TestTelegramCaptureRejectsUnauthorizedSender(t *testing.T) {
	t.Parallel()
	svc := NewService(staticAuthorizer{}, nilChatService{})
	_, err := svc.IngestCatalogCapture(context.Background(), CaptureInput{
		SenderID: "telegram-user-1",
		ChatID:   "telegram-chat-1",
		Draft:    Draft{PartNumber: "TG-001", Title: "Telegram Draft"},
	})
	if !errors.Is(err, ErrUnauthorizedSender) {
		t.Fatalf("expected ErrUnauthorizedSender, got %v", err)
	}
}

func TestTelegramCaptureCreatesPreviewThreadAndInboxWithoutApplying(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	profileID := "profile-telegram"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Capture"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}

	chatSvc := chat.NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	svc := NewService(staticAuthorizer{"sender-1|chat-1": profileID}, chatSvc)
	result, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID:     "sender-1",
		ChatID:       "chat-1",
		MessageID:    "message-42",
		Text:         "Box says limited release.",
		Barcode:      "9312345678901",
		GroupingHint: "same-item",
		Draft: Draft{
			Title:    "Limited Release Model",
			Brand:    "AFX",
			Category: "Die-cast",
		},
		Media: []MediaInput{
			{
				FileID:   "telegram-file-photo-1",
				Filename: "front.jpg",
				MIMEType: "image/jpeg",
				Kind:     "photo",
				Reader:   strings.NewReader("front-image-bytes"),
			},
		},
		SourceMetadata: map[string]any{"caption": "front and barcode"},
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() error = %v", err)
	}
	if result.ProfileID != profileID {
		t.Fatalf("expected profile %q, got %q", profileID, result.ProfileID)
	}
	if result.Thread.ID == "" || result.Message.ID == "" || result.Preview.ID == "" || result.InboxItem.ID == "" {
		t.Fatalf("expected persisted thread/message/preview/inbox IDs, got %+v", result)
	}
	if result.Thread.Metadata["source_channel"] != "telegram" || result.Thread.Metadata["source_message_id"] != "message-42" {
		t.Fatalf("thread metadata did not preserve Telegram source: %+v", result.Thread.Metadata)
	}
	if len(result.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(result.Attachments))
	}
	if result.Attachments[0].Filename != "front.jpg" || result.Attachments[0].MimeType != "image/jpeg" || result.Attachments[0].SizeBytes == 0 {
		t.Fatalf("attachment metadata not preserved: %+v", result.Attachments[0])
	}
	if result.Message.Context["source_channel"] != "telegram" || result.Message.Context["barcode"] != "9312345678901" {
		t.Fatalf("message context did not preserve Telegram/barcode context: %+v", result.Message.Context)
	}
	if result.Preview.Action != "create_inventory_item" || result.Preview.Status != "previewed" {
		t.Fatalf("expected previewed create_inventory_item action, got %+v", result.Preview)
	}
	if result.Preview.Payload["part_number"] != "9312345678901" || result.Preview.Payload["title"] != "Limited Release Model" {
		t.Fatalf("preview payload did not map barcode/text draft: %+v", result.Preview.Payload)
	}
	if result.InboxItem.Source != "telegram_catalog_capture" || result.InboxItem.Metadata["confirmation_state"] != "preview_required" {
		t.Fatalf("inbox item did not record confirmation-required source: %+v", result.InboxItem)
	}
	if result.TelegramReply.ConfirmationState != "preview_required" ||
		!strings.Contains(result.TelegramReply.Text, "Open Cabinet to confirm or cancel") ||
		!strings.Contains(result.TelegramReply.ReviewURL, result.Thread.ID) ||
		!strings.Contains(result.TelegramReply.ReviewURL, result.Preview.ID) {
		t.Fatalf("telegram reply did not expose a confirmation handoff: %+v", result.TelegramReply)
	}
	if result.InboxItem.Metadata["review_url"] != result.TelegramReply.ReviewURL {
		t.Fatalf("inbox metadata did not preserve Telegram review URL: %+v", result.InboxItem.Metadata)
	}
	replyMeta, ok := result.InboxItem.Metadata["telegram_reply"].(map[string]any)
	if !ok || replyMeta["review_url"] != result.TelegramReply.ReviewURL || len(replyMeta["actions"].([]any)) == 0 {
		t.Fatalf("inbox metadata did not preserve Telegram reply controls: %#v", result.InboxItem.Metadata["telegram_reply"])
	}

	var itemCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM canonical_items WHERE profile_id = ?`, profileID).Scan(&itemCount); err != nil {
		t.Fatalf("count canonical items: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("telegram capture must not apply without confirmation; got %d items", itemCount)
	}
	if _, err := chatSvc.ApplyAction(ctx, chat.ApplyActionInput{
		ProfileID: profileID,
		ThreadID:  result.Thread.ID,
		PreviewID: result.Preview.ID,
		Confirm:   false,
	}); err == nil {
		t.Fatalf("expected confirmation-required apply error")
	}
}

func TestTelegramCaptureDerivesDraftFromBarcodeOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profileID := "profile-barcode"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Barcode"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	svc := NewService(staticAuthorizer{"sender-2|chat-2": profileID}, chat.NewService(conn, filepath.Join(t.TempDir(), "attachments")))
	result, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID: "sender-2",
		ChatID:   "chat-2",
		Barcode:  "4904810900016",
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() barcode-only error = %v", err)
	}
	if result.Preview.Payload["part_number"] != "4904810900016" || result.Preview.Payload["title"] != "Barcode 4904810900016" {
		t.Fatalf("barcode-only capture did not create manual draft path: %+v", result.Preview.Payload)
	}
}

func TestCaptureInputFromWebhookUpdateNormalizesMixedPhotoCaption(t *testing.T) {
	t.Parallel()
	input, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 99,
		Message: &WebhookMessage{
			MessageID:    42,
			Date:         1716170000,
			From:         WebhookUser{ID: 12345},
			Chat:         WebhookChat{ID: -5235769556},
			Caption:      "Front photo with barcode 9312345678901",
			MediaGroupID: "album-1",
			Photo: []WebhookPhotoSize{
				{FileID: "small-photo", FileUniqueID: "small", Width: 90, Height: 90, FileSize: 1024},
				{FileID: "large-photo", FileUniqueID: "large", Width: 1280, Height: 720, FileSize: 2048},
			},
		},
	}, Draft{Title: "Caption Draft"})
	if err != nil {
		t.Fatalf("CaptureInputFromWebhookUpdate() error = %v", err)
	}
	if input.SenderID != "12345" || input.ChatID != "-5235769556" || input.MessageID != "42" {
		t.Fatalf("telegram IDs were not normalized: %+v", input)
	}
	if input.Text != "Front photo with barcode 9312345678901" || input.Barcode != "9312345678901" {
		t.Fatalf("caption/barcode were not normalized: text=%q barcode=%q", input.Text, input.Barcode)
	}
	if input.GroupingHint != "album-1" || input.SourceMetadata["payload_type"] != "caption+photo" {
		t.Fatalf("source metadata did not preserve Telegram grouping/type: %+v", input.SourceMetadata)
	}
	if len(input.Media) != 1 || input.Media[0].FileID != "large-photo" || input.Media[0].Filename != "large.jpg" || input.Media[0].Kind != "photo" {
		t.Fatalf("largest photo was not mapped to Cabinet media input: %+v", input.Media)
	}
}

func TestCaptureInputFromWebhookUpdateNormalizesTextOnlyBarcode(t *testing.T) {
	t.Parallel()
	input, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 100,
		Message: &WebhookMessage{
			MessageID: 7,
			From:      WebhookUser{ID: 2468},
			Chat:      WebhookChat{ID: 1357},
			Text:      "Please draft this barcode: 4904810900016",
		},
	}, Draft{})
	if err != nil {
		t.Fatalf("CaptureInputFromWebhookUpdate() text-only error = %v", err)
	}
	if input.Text != "Please draft this barcode: 4904810900016" || input.Barcode != "4904810900016" {
		t.Fatalf("text-only barcode was not normalized: %+v", input)
	}
	if len(input.Media) != 0 || input.SourceMetadata["payload_type"] != "text" {
		t.Fatalf("text-only webhook should not create media and should preserve type: %+v", input)
	}
}

type nilChatService struct{}

func (nilChatService) CreateThread(context.Context, string, string, map[string]any) (chat.Thread, error) {
	return chat.Thread{}, nil
}
func (nilChatService) CreateMessage(context.Context, string, string, string, string, map[string]any) (chat.Message, error) {
	return chat.Message{}, nil
}
func (nilChatService) SaveAttachment(context.Context, string, string, string, string, io.Reader) (chat.Attachment, error) {
	return chat.Attachment{}, nil
}
func (nilChatService) PreviewAction(context.Context, chat.PreviewActionInput) (chat.ActionPreview, error) {
	return chat.ActionPreview{}, nil
}
func (nilChatService) CreateInboxItem(context.Context, chat.InboxItem) (chat.InboxItem, error) {
	return chat.InboxItem{}, nil
}
