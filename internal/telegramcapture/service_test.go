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
				FileID:       "telegram-file-photo-1",
				FileUniqueID: "telegram-unique-photo-1",
				FileSize:     2048,
				Filename:     "front.jpg",
				MIMEType:     "image/jpeg",
				Kind:         "photo",
				Reader:       strings.NewReader("front-image-bytes"),
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
	messageMedia, ok := result.Message.Context["media"].([]any)
	if !ok || len(messageMedia) != 1 {
		t.Fatalf("message context did not preserve Telegram media audit details: %#v", result.Message.Context["media"])
	}
	messageMedia0, ok := messageMedia[0].(map[string]any)
	if !ok || messageMedia0["file_unique_id"] != "telegram-unique-photo-1" || !mediaFileSizeEquals(messageMedia0["file_size"], 2048) {
		t.Fatalf("message context did not preserve Telegram media source identifiers: %#v", result.Message.Context["media"])
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
	inboxMedia, ok := result.InboxItem.Metadata["media"].([]any)
	if !ok || len(inboxMedia) != 1 {
		t.Fatalf("inbox metadata did not preserve Telegram media audit details: %#v", result.InboxItem.Metadata["media"])
	}
	inboxMedia0, ok := inboxMedia[0].(map[string]any)
	if !ok || inboxMedia0["file_id"] != "telegram-file-photo-1" || inboxMedia0["file_unique_id"] != "telegram-unique-photo-1" || !mediaFileSizeEquals(inboxMedia0["file_size"], 2048) || inboxMedia0["filename"] != "front.jpg" || inboxMedia0["mime_type"] != "image/jpeg" {
		t.Fatalf("inbox metadata did not preserve Telegram media source fields: %#v", result.InboxItem.Metadata["media"])
	}
	inboxSourceMetadata, ok := result.InboxItem.Metadata["source_metadata"].(map[string]any)
	if !ok || inboxSourceMetadata["caption"] != "front and barcode" {
		t.Fatalf("inbox metadata did not preserve Telegram source metadata: %#v", result.InboxItem.Metadata["source_metadata"])
	}
	if result.TelegramReply.ConfirmationState != "preview_required" ||
		!strings.Contains(result.TelegramReply.Text, "Open Cabinet to confirm or cancel") ||
		!strings.Contains(result.TelegramReply.ReviewURL, result.Thread.ID) ||
		!strings.Contains(result.TelegramReply.ReviewURL, result.Preview.ID) {
		t.Fatalf("telegram reply did not expose a confirmation handoff: %+v", result.TelegramReply)
	}
	if len(result.TelegramReply.ActionButtons) != 3 ||
		result.TelegramReply.ActionButtons[0].Kind != "url" ||
		result.TelegramReply.ActionButtons[0].URL != result.TelegramReply.ReviewURL ||
		result.TelegramReply.ActionButtons[1].CallbackData != "cabinet:catalog_capture:confirm:"+result.Preview.ID ||
		result.TelegramReply.ActionButtons[2].CallbackData != "cabinet:catalog_capture:cancel:"+result.Preview.ID {
		t.Fatalf("telegram reply did not expose structured review/confirm/cancel buttons: %+v", result.TelegramReply.ActionButtons)
	}
	if result.InboxItem.Metadata["review_url"] != result.TelegramReply.ReviewURL {
		t.Fatalf("inbox metadata did not preserve Telegram review URL: %+v", result.InboxItem.Metadata)
	}
	replyMeta, ok := result.InboxItem.Metadata["telegram_reply"].(map[string]any)
	if !ok || replyMeta["review_url"] != result.TelegramReply.ReviewURL || len(replyMeta["actions"].([]any)) == 0 || len(replyMeta["action_buttons"].([]any)) != 3 {
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

func TestTelegramCapturePreservesResolvedLookupEvidence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profileID := "profile-lookup"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Lookup"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	svc := NewService(staticAuthorizer{"sender-lookup|chat-lookup": profileID}, chat.NewService(conn, filepath.Join(t.TempDir(), "attachments")))
	result, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID: "sender-lookup",
		ChatID:   "chat-lookup",
		Barcode:  "4904810900019",
		Draft: Draft{
			PartNumber:       "TOMY-001",
			Title:            "Lookup-backed Tomy release",
			Brand:            "Tomy",
			Category:         "Die-cast",
			LookupSource:     "barcode_local",
			LookupURL:        "/api/barcodes/4904810900019",
			LookupConfidence: "high",
		},
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() lookup-backed capture error = %v", err)
	}
	lookup, ok := result.Preview.Payload["lookup"].(map[string]any)
	if !ok || lookup["source"] != "barcode_local" || lookup["url"] != "/api/barcodes/4904810900019" || lookup["confidence"] != "high" {
		t.Fatalf("preview payload did not preserve lookup evidence: %+v", result.Preview.Payload)
	}
	messageLookup, ok := result.Message.Context["lookup"].(map[string]any)
	if !ok || messageLookup["source"] != "barcode_local" || messageLookup["url"] != "/api/barcodes/4904810900019" || messageLookup["confidence"] != "high" {
		t.Fatalf("message context did not preserve lookup audit evidence: %+v", result.Message.Context)
	}
	inboxLookup, ok := result.InboxItem.Metadata["lookup"].(map[string]any)
	if !ok || inboxLookup["source"] != "barcode_local" || inboxLookup["confidence"] != "high" {
		t.Fatalf("inbox metadata did not preserve lookup audit evidence: %+v", result.InboxItem.Metadata)
	}
	if result.Preview.Payload["part_number"] != "TOMY-001" || result.Preview.Payload["title"] != "Lookup-backed Tomy release" {
		t.Fatalf("lookup-backed draft fields were not used in preview: %+v", result.Preview.Payload)
	}
}

func TestTelegramCaptureCallbackConfirmsAndCancelsPendingPreview(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profileID := "profile-callback"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Callback"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	chatSvc := chat.NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	svc := NewService(staticAuthorizer{"sender-cb|chat-cb": profileID}, chatSvc)
	result, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID: "sender-cb",
		ChatID:   "chat-cb",
		Barcode:  "4904810900017",
		Draft:    Draft{Title: "Callback Draft", Brand: "AFX", Category: "Slot"},
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() error = %v", err)
	}
	confirmed, err := svc.HandleCatalogCaptureCallback(ctx, CallbackInput{
		SenderID:     "sender-cb",
		ChatID:       "chat-cb",
		CallbackData: "cabinet:catalog_capture:confirm:" + result.Preview.ID,
	})
	if err != nil {
		t.Fatalf("HandleCatalogCaptureCallback(confirm) error = %v", err)
	}
	if !confirmed.ApplyResult.Applied || confirmed.ApplyResult.ItemID == "" || confirmed.TelegramReply.ConfirmationState != "confirmed" {
		t.Fatalf("expected callback confirmation to apply item and return Telegram confirmation: %+v", confirmed)
	}
	var createdTitle string
	if err := conn.QueryRowContext(ctx, `SELECT title FROM canonical_items WHERE id = ? AND profile_id = ?`, confirmed.ApplyResult.ItemID, profileID).Scan(&createdTitle); err != nil {
		t.Fatalf("confirmed callback did not create catalog item: %v", err)
	}
	if createdTitle != "Callback Draft" {
		t.Fatalf("confirmed callback created wrong item title %q", createdTitle)
	}

	cancelResult, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID: "sender-cb",
		ChatID:   "chat-cb",
		Barcode:  "4904810900018",
		Draft:    Draft{Title: "Cancel Draft"},
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() cancel fixture error = %v", err)
	}
	cancelled, err := svc.HandleCatalogCaptureCallback(ctx, CallbackInput{
		SenderID:     "sender-cb",
		ChatID:       "chat-cb",
		CallbackData: "cabinet:catalog_capture:cancel:" + cancelResult.Preview.ID,
	})
	if err != nil {
		t.Fatalf("HandleCatalogCaptureCallback(cancel) error = %v", err)
	}
	if cancelled.ApplyResult.Applied || cancelled.TelegramReply.ConfirmationState != "cancelled" {
		t.Fatalf("expected callback cancellation to leave preview unapplied: %+v", cancelled)
	}
	var cancelledCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM canonical_items WHERE part_number = ?`, "4904810900018").Scan(&cancelledCount); err != nil {
		t.Fatalf("count cancelled item: %v", err)
	}
	if cancelledCount != 0 {
		t.Fatalf("cancelled callback must not create a catalog item, count=%d", cancelledCount)
	}
}

func TestTelegramCaptureCallbackRejectsDifferentAuthorizedSender(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ownerProfileID := "profile-callback-owner"
	otherProfileID := "profile-callback-other"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?), (?, ?)`, ownerProfileID, "Telegram Callback Owner", otherProfileID, "Telegram Callback Other"); err != nil {
		t.Fatalf("insert profiles: %v", err)
	}
	chatSvc := chat.NewService(conn, filepath.Join(t.TempDir(), "attachments"))
	svc := NewService(staticAuthorizer{
		"sender-owner|chat-owner": ownerProfileID,
		"sender-other|chat-other": otherProfileID,
	}, chatSvc)
	result, err := svc.IngestCatalogCapture(ctx, CaptureInput{
		SenderID: "sender-owner",
		ChatID:   "chat-owner",
		Barcode:  "4904810900020",
		Draft:    Draft{Title: "Owner Draft"},
	})
	if err != nil {
		t.Fatalf("IngestCatalogCapture() error = %v", err)
	}

	_, err = svc.HandleCatalogCaptureCallback(ctx, CallbackInput{
		SenderID:     "sender-other",
		ChatID:       "chat-other",
		CallbackData: "cabinet:catalog_capture:confirm:" + result.Preview.ID,
	})
	if err == nil {
		t.Fatalf("expected different authorized sender/chat to be rejected before applying preview")
	}
	var itemCount int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM canonical_items WHERE profile_id IN (?, ?)`, ownerProfileID, otherProfileID).Scan(&itemCount); err != nil {
		t.Fatalf("count catalog items: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("cross-sender callback must not create a catalog item, count=%d", itemCount)
	}
}

func TestTelegramCaptureAmbiguousTextRequiresFollowUp(t *testing.T) {
	t.Parallel()
	svc := NewService(staticAuthorizer{"sender-3|chat-3": "profile-follow-up"}, nilChatService{})
	_, err := svc.IngestCatalogCapture(context.Background(), CaptureInput{
		SenderID: "sender-3",
		ChatID:   "chat-3",
		Text:     "This is the blue boxed one from the bench.",
	})
	if !errors.Is(err, ErrDraftNeedsFollowUp) {
		t.Fatalf("expected ErrDraftNeedsFollowUp, got %v", err)
	}
	var followUp DraftNeedsFollowUpError
	if !errors.As(err, &followUp) {
		t.Fatalf("expected DraftNeedsFollowUpError, got %T", err)
	}
	if followUp.Reply.ConfirmationState != "follow_up_required" ||
		!strings.Contains(followUp.Reply.Text, "barcode, part number, or clearer item title") ||
		len(followUp.MissingFields) == 0 {
		t.Fatalf("follow-up prompt did not explain the missing draft identity: %+v", followUp)
	}
	if len(followUp.Reply.ActionButtons) != 3 ||
		followUp.Reply.ActionButtons[0].Kind != "reply" ||
		followUp.Reply.ActionButtons[0].Action != "reply_with_barcode" {
		t.Fatalf("follow-up prompt did not expose structured reply actions: %+v", followUp.Reply.ActionButtons)
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
	if len(input.Media) != 1 || input.Media[0].FileID != "large-photo" || input.Media[0].FileUniqueID != "large" || input.Media[0].FileSize != 2048 || input.Media[0].Filename != "large.jpg" || input.Media[0].Kind != "photo" {
		t.Fatalf("largest photo was not mapped to Cabinet media input: %+v", input.Media)
	}
}

func TestGroupCaptureInputsCombinesTelegramAlbumMedia(t *testing.T) {
	t.Parallel()
	first, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2001,
		Message: &WebhookMessage{
			MessageID:    61,
			Date:         1716170600,
			From:         WebhookUser{ID: 12345},
			Chat:         WebhookChat{ID: -5235769556},
			Caption:      "Front and barcode 4904810900016",
			MediaGroupID: "album-42",
			Photo: []WebhookPhotoSize{
				{FileID: "front-small", FileUniqueID: "front-small", Width: 90, Height: 90, FileSize: 512},
				{FileID: "front-large", FileUniqueID: "front", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{})
	if err != nil {
		t.Fatalf("first CaptureInputFromWebhookUpdate() error = %v", err)
	}
	second, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2002,
		Message: &WebhookMessage{
			MessageID:    62,
			Date:         1716170601,
			From:         WebhookUser{ID: 12345},
			Chat:         WebhookChat{ID: -5235769556},
			Caption:      "Back of the same boxed item",
			MediaGroupID: "album-42",
			Photo: []WebhookPhotoSize{
				{FileID: "back-large", FileUniqueID: "back", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{Title: "Album grouped draft", Brand: "Tomy"})
	if err != nil {
		t.Fatalf("second CaptureInputFromWebhookUpdate() error = %v", err)
	}
	other, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2003,
		Message: &WebhookMessage{
			MessageID:    63,
			From:         WebhookUser{ID: 99999},
			Chat:         WebhookChat{ID: -5235769556},
			Caption:      "Different sender in a same-named album",
			MediaGroupID: "album-42",
			Photo: []WebhookPhotoSize{
				{FileID: "other-large", FileUniqueID: "other", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{Title: "Other sender draft"})
	if err != nil {
		t.Fatalf("other CaptureInputFromWebhookUpdate() error = %v", err)
	}
	third, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2004,
		Message: &WebhookMessage{
			MessageID:    64,
			Date:         1716170602,
			From:         WebhookUser{ID: 12345},
			Chat:         WebhookChat{ID: -5235769556},
			Caption:      "Side detail for the same boxed item",
			MediaGroupID: "album-42",
			Photo: []WebhookPhotoSize{
				{FileID: "side-large", FileUniqueID: "side", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{})
	if err != nil {
		t.Fatalf("third CaptureInputFromWebhookUpdate() error = %v", err)
	}

	grouped := GroupCaptureInputs([]CaptureInput{first, second, other, third})
	if len(grouped) != 2 {
		t.Fatalf("expected album inputs to group by sender/chat/media group, got %d: %+v", len(grouped), grouped)
	}
	album := grouped[0]
	if album.SenderID != "12345" || album.ChatID != "-5235769556" || album.GroupingHint != "album-42" {
		t.Fatalf("grouped album identity was not preserved: %+v", album)
	}
	if len(album.Media) != 3 || album.Media[0].FileID != "front-large" || album.Media[1].FileID != "back-large" || album.Media[2].FileID != "side-large" {
		t.Fatalf("grouped album media was not preserved in order: %+v", album.Media)
	}
	if album.Barcode != "4904810900016" || album.Draft.Title != "Album grouped draft" || album.Draft.Brand != "Tomy" {
		t.Fatalf("grouped album did not preserve barcode and downstream draft fields: %+v", album)
	}
	if !strings.Contains(album.Text, "Front and barcode") || !strings.Contains(album.Text, "Back of the same boxed item") || !strings.Contains(album.Text, "Side detail") {
		t.Fatalf("grouped album text did not preserve captions: %q", album.Text)
	}
	messageIDs, ok := album.SourceMetadata["grouped_message_ids"].([]string)
	if !ok || len(messageIDs) != 3 || messageIDs[0] != "61" || messageIDs[1] != "62" || messageIDs[2] != "64" {
		t.Fatalf("grouped metadata did not preserve message ids: %+v", album.SourceMetadata)
	}
	if album.SourceMetadata["message_count"] != 3 || !strings.Contains(album.SourceMetadata["payload_type"].(string), "telegram_album") {
		t.Fatalf("grouped metadata did not classify the album: %+v", album.SourceMetadata)
	}
	updateIDs, ok := album.SourceMetadata["grouped_update_ids"].([]any)
	if !ok || len(updateIDs) != 3 || updateIDs[0] != int64(2001) || updateIDs[1] != int64(2002) || updateIDs[2] != int64(2004) {
		t.Fatalf("grouped metadata did not preserve distinct update ids in order: %+v", album.SourceMetadata)
	}
	if grouped[1].SenderID != "99999" || len(grouped[1].Media) != 1 {
		t.Fatalf("same media_group_id from another sender must remain separate: %+v", grouped[1])
	}
}

func TestGroupCaptureInputsKeepsUngroupedPhotosSeparate(t *testing.T) {
	t.Parallel()
	first, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2101,
		Message: &WebhookMessage{
			MessageID: 71,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "Loose front photo 4904810900016",
			Photo: []WebhookPhotoSize{
				{FileID: "loose-front", FileUniqueID: "loose-front", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{Title: "Loose front draft"})
	if err != nil {
		t.Fatalf("first CaptureInputFromWebhookUpdate() error = %v", err)
	}
	second, err := CaptureInputFromWebhookUpdate(WebhookUpdate{
		UpdateID: 2102,
		Message: &WebhookMessage{
			MessageID: 72,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "Loose back photo 4904810900017",
			Photo: []WebhookPhotoSize{
				{FileID: "loose-back", FileUniqueID: "loose-back", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	}, Draft{Title: "Loose back draft"})
	if err != nil {
		t.Fatalf("second CaptureInputFromWebhookUpdate() error = %v", err)
	}

	grouped := GroupCaptureInputs([]CaptureInput{first, second})
	if len(grouped) != 2 {
		t.Fatalf("independent photo messages without media_group_id must remain separate, got %d: %+v", len(grouped), grouped)
	}
	if grouped[0].MessageID != "71" || grouped[1].MessageID != "72" {
		t.Fatalf("separate capture message ids were not preserved: %+v", grouped)
	}
	if grouped[0].GroupingHint != "" || grouped[1].GroupingHint != "" {
		t.Fatalf("ungrouped photo messages must not gain grouping hints: %+v", grouped)
	}
	if len(grouped[0].Media) != 1 || grouped[0].Media[0].FileID != "loose-front" ||
		len(grouped[1].Media) != 1 || grouped[1].Media[0].FileID != "loose-back" {
		t.Fatalf("separate photo media were not preserved independently: %+v", grouped)
	}
	if grouped[0].Barcode != "4904810900016" || grouped[1].Barcode != "4904810900017" ||
		grouped[0].Draft.Title != "Loose front draft" || grouped[1].Draft.Title != "Loose back draft" {
		t.Fatalf("separate photo capture context was not preserved independently: %+v", grouped)
	}
}

func TestCaptureInputFromWebhookUpdateWithMediaResolvesTelegramPhotoBytes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	input, err := CaptureInputFromWebhookUpdateWithMedia(ctx, WebhookUpdate{
		UpdateID: 101,
		Message: &WebhookMessage{
			MessageID: 55,
			Date:      1716170500,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "Captured box barcode 9312345678901",
			Photo: []WebhookPhotoSize{
				{FileID: "small-photo", FileUniqueID: "small", Width: 90, Height: 90, FileSize: 1024},
				{FileID: "large-photo", FileUniqueID: "large", Width: 1280, Height: 720, FileSize: 2048},
			},
		},
	}, Draft{Title: "Resolved Photo Draft"}, staticMediaResolver{
		"large-photo": {
			FileUniqueID: "resolved-large",
			FileSize:     4096,
			Filename:     "telegram-large-photo.jpg",
			MIMEType:     "image/jpeg",
			Reader:       strings.NewReader("resolved-photo-bytes"),
		},
	})
	if err != nil {
		t.Fatalf("CaptureInputFromWebhookUpdateWithMedia() error = %v", err)
	}
	if len(input.Media) != 1 || input.Media[0].FileID != "large-photo" || input.Media[0].FileUniqueID != "resolved-large" || input.Media[0].FileSize != 4096 || input.Media[0].Filename != "telegram-large-photo.jpg" || input.Media[0].Reader == nil {
		t.Fatalf("resolved media did not preserve Telegram file id and attach bytes: %+v", input.Media)
	}

	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profileID := "profile-resolved-media"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Resolved Media"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	svc := NewService(staticAuthorizer{"12345|-5235769556": profileID}, chat.NewService(conn, filepath.Join(t.TempDir(), "attachments")))
	result, err := svc.IngestCatalogCapture(ctx, input)
	if err != nil {
		t.Fatalf("IngestCatalogCapture() resolved media error = %v", err)
	}
	if len(result.Attachments) != 1 || result.Attachments[0].Filename != "telegram-large-photo.jpg" || result.Attachments[0].MimeType != "image/jpeg" {
		t.Fatalf("resolved attachment metadata not preserved: %+v", result.Attachments)
	}
	if result.Attachments[0].SizeBytes != int64(len("resolved-photo-bytes")) {
		t.Fatalf("resolved media bytes were not persisted, attachment=%+v", result.Attachments[0])
	}
}

func TestCaptureInputFromWebhookUpdateWithMediaResolvesTelegramImageDocumentBytes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	input, err := CaptureInputFromWebhookUpdateWithMedia(ctx, WebhookUpdate{
		UpdateID: 102,
		Message: &WebhookMessage{
			MessageID: 56,
			Date:      1716170501,
			From:      WebhookUser{ID: 23456},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "Box art document barcode 4904810900021",
			Document: &WebhookDocument{
				FileID:       "telegram-doc-file",
				FileUniqueID: "telegram-doc-unique",
				FileName:     "box-art.png",
				MIMEType:     "image/png",
				FileSize:     4096,
			},
		},
	}, Draft{Title: "Resolved Document Draft"}, staticMediaResolver{
		"telegram-doc-file": {
			FileUniqueID: "resolved-doc-unique",
			FileSize:     8192,
			Filename:     "box-art.png",
			MIMEType:     "image/png",
			Reader:       strings.NewReader("resolved-document-bytes"),
		},
	})
	if err != nil {
		t.Fatalf("CaptureInputFromWebhookUpdateWithMedia() document error = %v", err)
	}
	if len(input.Media) != 1 || input.Media[0].FileID != "telegram-doc-file" || input.Media[0].FileUniqueID != "resolved-doc-unique" || input.Media[0].FileSize != 8192 || input.Media[0].Kind != "document_image" || input.Media[0].Reader == nil {
		t.Fatalf("resolved image document did not preserve Telegram document media: %+v", input.Media)
	}
	if input.SourceMetadata["payload_type"] != "caption+document" || input.Barcode != "4904810900021" {
		t.Fatalf("image document capture did not preserve payload type and barcode context: %+v", input)
	}

	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profileID := "profile-resolved-document"
	if _, err := conn.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, "Telegram Resolved Document"); err != nil {
		t.Fatalf("insert profile: %v", err)
	}
	svc := NewService(staticAuthorizer{"23456|-5235769556": profileID}, chat.NewService(conn, filepath.Join(t.TempDir(), "attachments")))
	result, err := svc.IngestCatalogCapture(ctx, input)
	if err != nil {
		t.Fatalf("IngestCatalogCapture() resolved document error = %v", err)
	}
	if len(result.Attachments) != 1 || result.Attachments[0].Filename != "box-art.png" || result.Attachments[0].MimeType != "image/png" {
		t.Fatalf("resolved document attachment metadata not preserved: %+v", result.Attachments)
	}
	media, ok := result.InboxItem.Metadata["media"].([]any)
	if !ok || len(media) != 1 {
		t.Fatalf("inbox metadata did not preserve image document media: %+v", result.InboxItem.Metadata)
	}
	media0, ok := media[0].(map[string]any)
	if !ok || media0["kind"] != "document_image" || media0["file_id"] != "telegram-doc-file" || media0["file_unique_id"] != "resolved-doc-unique" || !mediaFileSizeEquals(media0["file_size"], 8192) {
		t.Fatalf("inbox metadata did not preserve Telegram document audit fields: %+v", media)
	}
	if result.Attachments[0].SizeBytes != int64(len("resolved-document-bytes")) {
		t.Fatalf("resolved image document bytes were not persisted, attachment=%+v", result.Attachments[0])
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
func (nilChatService) GetActionPreview(context.Context, string, string) (chat.ActionPreview, error) {
	return chat.ActionPreview{}, nil
}
func (nilChatService) ApplyAction(context.Context, chat.ApplyActionInput) (chat.ApplyActionResult, error) {
	return chat.ApplyActionResult{}, nil
}
func (nilChatService) CancelAction(context.Context, chat.ApplyActionInput) (chat.ApplyActionResult, error) {
	return chat.ApplyActionResult{}, nil
}
func (nilChatService) CreateWorkflowRun(context.Context, chat.CreateWorkflowRunInput) (chat.WorkflowRun, error) {
	return chat.WorkflowRun{}, nil
}
func (nilChatService) UpdateWorkflowRun(context.Context, chat.UpdateWorkflowRunInput) (chat.WorkflowRun, error) {
	return chat.WorkflowRun{}, nil
}
func (nilChatService) CreateInboxItem(context.Context, chat.InboxItem) (chat.InboxItem, error) {
	return chat.InboxItem{}, nil
}

type staticMediaResolver map[string]MediaInput

func (r staticMediaResolver) ResolveTelegramMedia(_ context.Context, media MediaInput) (MediaInput, error) {
	resolved, ok := r[strings.TrimSpace(media.FileID)]
	if !ok {
		return MediaInput{}, errors.New("missing telegram media fixture")
	}
	return resolved, nil
}

func mediaFileSizeEquals(value any, expected int) bool {
	switch typed := value.(type) {
	case int:
		return typed == expected
	case int64:
		return typed == int64(expected)
	case float64:
		return typed == float64(expected)
	default:
		return false
	}
}
