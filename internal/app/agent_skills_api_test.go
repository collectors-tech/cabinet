package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
)

type apiSkillPayload struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Status          string   `json:"status"`
	SafetyLevel     string   `json:"safety_level"`
	RequiredContext []string `json:"required_context"`
	Capabilities    []string `json:"capabilities"`
	GuidedWorkflows []string `json:"guided_workflows"`
	BuiltIn         bool     `json:"built_in"`
	Removable       bool     `json:"removable"`
	Executable      bool     `json:"executable"`
	NextAction      string   `json:"next_action"`
	Permissions     struct {
		LocalWrite      bool `json:"local_write"`
		Destructive     bool `json:"destructive"`
		RequiresConfirm bool `json:"requires_confirm"`
	} `json:"permissions"`
}

func TestAgentSkillRegistryAPIExposesGovernedSkillMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/agent/skills?profile_id=test-profile", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("skills status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		ProfileID string            `json:"profile_id"`
		Skills    []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if payload.ProfileID != "test-profile" {
		t.Fatalf("expected profile echo, got %q", payload.ProfileID)
	}
	inventory := findAPISkill(payload.Skills, "cabinet.inventory.update_item")
	if inventory == nil {
		t.Fatalf("missing inventory update skill")
	}
	if inventory.Source != "built-in" || !inventory.BuiltIn || inventory.Removable {
		t.Fatalf("expected immutable built-in metadata, got %+v", inventory)
	}
	if inventory.SafetyLevel != "confirm-required" || !inventory.Permissions.LocalWrite || !inventory.Permissions.RequiresConfirm {
		t.Fatalf("expected confirm-required write safety, got %+v", inventory)
	}
	if !slices.Contains(inventory.Capabilities, "inventory.item.update") {
		t.Fatalf("expected capability binding, got %+v", inventory.Capabilities)
	}

	guided := findAPISkill(payload.Skills, "cabinet.guided.inventory.update_item")
	if guided == nil {
		t.Fatalf("missing guided inventory update skill")
	}
	if guided.Status != "requires-implementation" || guided.Executable {
		t.Fatalf("guided skill must remain non-executable before #1513, got %+v", guided)
	}
	if !slices.Contains(guided.GuidedWorkflows, "inventory.item.update") || guided.NextAction == "" {
		t.Fatalf("expected guided workflow binding and next action, got %+v", guided)
	}

	inboxSearch := findAPISkill(payload.Skills, "cabinet.inbox.search_notifications")
	if inboxSearch == nil {
		t.Fatalf("missing Inbox search skill")
	}
	if inboxSearch.SafetyLevel != "read-only" || inboxSearch.Permissions.LocalWrite || !inboxSearch.Executable {
		t.Fatalf("expected executable read-only Inbox search metadata, got %+v", inboxSearch)
	}

	inboxMutation := findAPISkill(payload.Skills, "cabinet.inbox.mark_handled")
	if inboxMutation == nil {
		t.Fatalf("missing Inbox mark handled skill")
	}
	if inboxMutation.Status != "available" || !inboxMutation.Executable || !inboxMutation.Permissions.RequiresConfirm {
		t.Fatalf("expected Inbox mutation to be executable and confirmation-gated, got %+v", inboxMutation)
	}

	userSearch := findAPISkill(payload.Skills, "cabinet.users.search")
	if userSearch == nil {
		t.Fatalf("missing Users search skill")
	}
	if userSearch.SafetyLevel != "read-only" || !slices.Contains(userSearch.RequiredContext, "admin_session") || !userSearch.Executable {
		t.Fatalf("expected executable read-only Users search metadata, got %+v", userSearch)
	}

	removeUser := findAPISkill(payload.Skills, "cabinet.users.remove_user")
	if removeUser == nil {
		t.Fatalf("missing remove user skill")
	}
	if removeUser.SafetyLevel != "destructive" || !removeUser.Permissions.Destructive || !removeUser.Executable {
		t.Fatalf("expected destructive executable remove user metadata, got %+v", removeUser)
	}

	runWatch := findAPISkill(payload.Skills, "cabinet.market_watch.run_watch")
	if runWatch == nil {
		t.Fatalf("missing Market Watch run skill")
	}
	if runWatch.SafetyLevel != "confirm-required" || !runWatch.Permissions.RequiresConfirm || !runWatch.Executable {
		t.Fatalf("expected confirmation-gated Market Watch run metadata, got %+v", runWatch)
	}

	addPurchaseLine := findAPISkill(payload.Skills, "cabinet.purchases.add_line_item")
	if addPurchaseLine == nil {
		t.Fatalf("missing Purchases add line item skill")
	}
	if addPurchaseLine.SafetyLevel != "confirm-required" || !addPurchaseLine.Permissions.RequiresConfirm || !slices.Contains(addPurchaseLine.RequiredContext, "target_order") {
		t.Fatalf("expected confirmation-gated purchase line item metadata, got %+v", addPurchaseLine)
	}

	wishlistSearch := findAPISkill(payload.Skills, "cabinet.wishlist.search_entries")
	if wishlistSearch == nil {
		t.Fatalf("missing Wishlist search skill")
	}
	if wishlistSearch.SafetyLevel != "read-only" || wishlistSearch.Permissions.LocalWrite || !wishlistSearch.Executable {
		t.Fatalf("expected executable read-only Wishlist search metadata, got %+v", wishlistSearch)
	}

	wishlistPurchased := findAPISkill(payload.Skills, "cabinet.wishlist.mark_purchased")
	if wishlistPurchased == nil {
		t.Fatalf("missing Wishlist mark purchased skill")
	}
	if wishlistPurchased.SafetyLevel != "confirm-required" || !wishlistPurchased.Permissions.RequiresConfirm || !slices.Contains(wishlistPurchased.RequiredContext, "purchase_details") {
		t.Fatalf("expected confirmation-gated Wishlist purchased metadata, got %+v", wishlistPurchased)
	}

	collectionsSearch := findAPISkill(payload.Skills, "cabinet.collections.search")
	if collectionsSearch == nil {
		t.Fatalf("missing Collections search skill")
	}
	if collectionsSearch.SafetyLevel != "read-only" || collectionsSearch.Permissions.LocalWrite || !collectionsSearch.Executable {
		t.Fatalf("expected executable read-only Collections search metadata, got %+v", collectionsSearch)
	}

	collectionDelete := findAPISkill(payload.Skills, "cabinet.collections.soft_delete")
	if collectionDelete == nil {
		t.Fatalf("missing Collections soft delete skill")
	}
	if collectionDelete.SafetyLevel != "confirm-required" || !collectionDelete.Permissions.RequiresConfirm || !slices.Contains(collectionDelete.RequiredContext, "collection") {
		t.Fatalf("expected confirmation-gated Collections delete metadata, got %+v", collectionDelete)
	}
}

func TestAgentSkillPreviewAPIBlocksUnsafeAdminMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"test-profile",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{
			"target_user":"owner@example.test",
			"target_role_current":"owner"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		SkillID              string `json:"skill_id"`
		Allowed              bool   `json:"allowed"`
		PreviewOnly          bool   `json:"preview_only"`
		MutationApplied      bool   `json:"mutation_applied"`
		ConfirmationRequired bool   `json:"confirmation_required"`
		Blocker              string `json:"blocker"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if payload.SkillID != "cabinet.users.remove_user" || payload.Allowed || !payload.PreviewOnly || payload.MutationApplied {
		t.Fatalf("expected preview-only blocked admin mutation, got %+v", payload)
	}
	if !payload.ConfirmationRequired || payload.Blocker != "users_admin_protected_owner_remove_blocked" {
		t.Fatalf("expected protected owner confirmation blocker, got %+v", payload)
	}
}

func TestAgentSkillPreviewAPIBlocksWishlistAndCollectionMissingContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	wishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"test-profile",
		"skill_id":"cabinet.wishlist.mark_purchased",
		"parameters":{"purchase_url":"https://example.test/order"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlist.Code != http.StatusOK {
		t.Fatalf("wishlist preview status=%d body=%s", wishlist.Code, wishlist.Body.String())
	}
	if !strings.Contains(wishlist.Body.String(), `"blocker":"wishlist_entry_required"`) ||
		!strings.Contains(wishlist.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected missing wishlist entry blocker, body=%s", wishlist.Body.String())
	}

	allItems := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"test-profile",
		"skill_id":"cabinet.collections.soft_delete",
		"parameters":{"collection_name":"All Items"}
	}`), map[string]string{"Content-Type": "application/json"})
	if allItems.Code != http.StatusOK {
		t.Fatalf("collections All Items preview status=%d body=%s", allItems.Code, allItems.Body.String())
	}
	if !strings.Contains(allItems.Body.String(), `"blocker":"collections_all_items_protected"`) ||
		!strings.Contains(allItems.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected protected All Items blocker, body=%s", allItems.Body.String())
	}

	nonEmptyDelete := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"test-profile",
		"skill_id":"cabinet.collections.soft_delete",
		"parameters":{"collection_name":"Display Case","has_items":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if nonEmptyDelete.Code != http.StatusOK {
		t.Fatalf("collections non-empty delete preview status=%d body=%s", nonEmptyDelete.Code, nonEmptyDelete.Body.String())
	}
	if !strings.Contains(nonEmptyDelete.Body.String(), `"blocker":"collections_delete_destination_required"`) {
		t.Fatalf("expected missing destination blocker, body=%s", nonEmptyDelete.Body.String())
	}
}

func TestAgentSkillAPIPropagatesInvocationSourceContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Source Context"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	record := doRequest(t, a, http.MethodPost, "/api/chat/inbox", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"records":[{
			"local_history_id":"agent-skill-source-context",
			"title":"Agent skill source context",
			"summary":"Needs sourced review"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if record.Code != http.StatusCreated {
		t.Fatalf("create inbox record status=%d body=%s", record.Code, record.Body.String())
	}
	var recordPayload struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(record.Body).Decode(&recordPayload); err != nil {
		t.Fatalf("decode inbox record: %v", err)
	}
	if len(recordPayload.Items) != 1 {
		t.Fatalf("expected one inbox item, got %+v", recordPayload.Items)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"source_surface":"inbox.notification.card",
		"source_channel":"in-app",
		"source_thread_id":"thread-source-context",
		"source_message_id":"message-source-context",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview source context status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(preview.Body.String(), `"source_channel":"in-app"`) ||
		!strings.Contains(preview.Body.String(), `"source_thread_id":"thread-source-context"`) ||
		!strings.Contains(preview.Body.String(), `"source_message_id":"message-source-context"`) ||
		!strings.Contains(preview.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected preview to retain source context without mutation, body=%s", preview.Body.String())
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"confirm":true,
		"source_surface":"inbox.notification.card",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-42",
		"source_message_id":"tg-message-99",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply source context status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(apply.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(apply.Body.String(), `"source_thread_id":"tg-chat-42"`) ||
		!strings.Contains(apply.Body.String(), `"source_message_id":"tg-message-99"`) ||
		!strings.Contains(apply.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(apply.Body.String(), `"status":"read"`) {
		t.Fatalf("expected confirmed apply to retain channel source context, body=%s", apply.Body.String())
	}
}

func TestAgentSkillAPIPropagatesExternalChannelContextForMarketWatchAndPurchases(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill External Channel Context"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	marketWatchPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"source_surface":"telegram.market_watch.result",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-1710",
		"source_message_id":"tg-message-market-watch-1710",
		"parameters":{"provider_id":"ebay","result_id":"result-telegram-1710","destination":"wishlist","source_url":"https://example.test/listing/telegram-1710"}
	}`), map[string]string{"Content-Type": "application/json"})
	if marketWatchPreview.Code != http.StatusOK {
		t.Fatalf("market watch external preview status=%d body=%s", marketWatchPreview.Code, marketWatchPreview.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.market_watch.handoff_result"`,
		`"source_surface":"telegram.market_watch.result"`,
		`"source_channel":"telegram"`,
		`"source_thread_id":"tg-chat-1710"`,
		`"source_message_id":"tg-message-market-watch-1710"`,
		`"mutation_applied":false`,
		`"confirmation_required":true`,
	} {
		if !strings.Contains(marketWatchPreview.Body.String(), want) {
			t.Fatalf("market watch external preview missing %s: body=%s", want, marketWatchPreview.Body.String())
		}
	}

	purchasesApply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.create_order",
		"confirm":true,
		"source_surface":"telegram.purchase.capture",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-1710",
		"source_message_id":"tg-message-purchase-1710",
		"parameters":{
			"purchase_source":"telegram",
			"order_id":"telegram-order-1710",
			"title":"Telegram purchase order item",
			"part_number":"TG-1710",
			"source_url":"https://example.test/orders/telegram-1710",
			"quantity":1,
			"amount":88,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if purchasesApply.Code != http.StatusOK {
		t.Fatalf("purchases external apply status=%d body=%s", purchasesApply.Code, purchasesApply.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.purchases.create_order"`,
		`"source_surface":"telegram.purchase.capture"`,
		`"source_channel":"telegram"`,
		`"source_thread_id":"tg-chat-1710"`,
		`"source_message_id":"tg-message-purchase-1710"`,
		`"mutation_applied":true`,
		`"operation":"purchases.order.create"`,
		`"order_id":"telegram-order-1710"`,
		`"purchase_persisted":true`,
		`"provenance_preserved":true`,
	} {
		if !strings.Contains(purchasesApply.Body.String(), want) {
			t.Fatalf("purchases external apply missing %s: body=%s", want, purchasesApply.Body.String())
		}
	}
	var purchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND external_ref = 'telegram-order-1710' AND source = 'telegram'`, p.ID).Scan(&purchaseCount); err != nil {
		t.Fatalf("count external purchase order evidence: %v", err)
	}
	if purchaseCount != 1 {
		t.Fatalf("expected one persisted external purchase order, got %d", purchaseCount)
	}
}

func TestAgentSkillApplyAPIHandlesMediaSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Media"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES ('media-item-1', ?, 'AFX', 'Slot Cars', 'MEDIA-1', 'Agent media target item', 'active');
		INSERT INTO chat_threads(id, profile_id, title)
		VALUES ('media-thread-1', ?, 'Agent media uploads');
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-attach-1', ?, 'media-thread-1', 'loose-reference.jpg', 'image/jpeg', 123, 'https://example.test/media/loose-reference.jpg');
	`, p.ID, p.ID, p.ID); err != nil {
		t.Fatalf("seed media skill data: %v", err)
	}

	search := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.search",
		"parameters":{"query":"loose-reference"}
	}`), map[string]string{"Content-Type": "application/json"})
	if search.Code != http.StatusOK {
		t.Fatalf("media search status=%d body=%s", search.Code, search.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"media.search"`,
		`"read_only":true`,
		`"filename":"loose-reference.jpg"`,
	} {
		if !strings.Contains(search.Body.String(), want) {
			t.Fatalf("media search response missing %s: body=%s", want, search.Body.String())
		}
	}

	upload := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.upload_or_import",
		"confirm":true,
		"source_surface":"telegram.media.upload",
		"source_channel":"telegram",
		"source_thread_id":"tg-media-thread-1709",
		"source_message_id":"tg-media-message-1709",
		"parameters":{
			"source_url":"https://example.test/media/imported-reference.jpg",
			"filename":"imported-reference.jpg",
			"title":"Imported agent reference",
			"notes":"imported by agent skill"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if upload.Code != http.StatusOK {
		t.Fatalf("media upload/import status=%d body=%s", upload.Code, upload.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"media.upload_or_import"`,
		`"media_persisted":true`,
		`"provenance_preserved":true`,
		`"source_url":"https://example.test/media/imported-reference.jpg"`,
		`"source_channel":"telegram"`,
	} {
		if !strings.Contains(upload.Body.String(), want) {
			t.Fatalf("media upload response missing %s: body=%s", want, upload.Body.String())
		}
	}
	var uploadPayload struct {
		Target struct {
			MediaID string `json:"media_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(upload.Body).Decode(&uploadPayload); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if uploadPayload.Target.MediaID == "" {
		t.Fatalf("expected uploaded media id, body=%s", upload.Body.String())
	}
	var importedCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM chat_attachments WHERE profile_id = ? AND id = ? AND stored_path = 'https://example.test/media/imported-reference.jpg'`, p.ID, uploadPayload.Target.MediaID).Scan(&importedCount); err != nil {
		t.Fatalf("count imported media attachment: %v", err)
	}
	if importedCount != 1 {
		t.Fatalf("expected one persisted imported media attachment, got %d", importedCount)
	}

	attach := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.attach_to_item",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","item_id":"media-item-1"}
	}`), map[string]string{"Content-Type": "application/json"})
	if attach.Code != http.StatusOK {
		t.Fatalf("media attach status=%d body=%s", attach.Code, attach.Body.String())
	}
	if !strings.Contains(attach.Body.String(), `"operation":"media.attach_to_item"`) ||
		!strings.Contains(attach.Body.String(), `"attachment_persisted":true`) ||
		!strings.Contains(attach.Body.String(), `"provenance_preserved":true`) {
		t.Fatalf("expected media attach persistence evidence, body=%s", attach.Body.String())
	}
	var linkCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = 'media-item-1'`, p.ID, uploadPayload.Target.MediaID).Scan(&linkCount); err != nil {
		t.Fatalf("count media attachment link: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("expected one persisted media link, got %d", linkCount)
	}

	updateNotes := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.update_notes",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","notes":"agent skill updated notes"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateNotes.Code != http.StatusOK {
		t.Fatalf("media update notes status=%d body=%s", updateNotes.Code, updateNotes.Body.String())
	}
	if !strings.Contains(updateNotes.Body.String(), `"operation":"media.update_notes"`) ||
		!strings.Contains(updateNotes.Body.String(), `"metadata_persisted":true`) ||
		!strings.Contains(updateNotes.Body.String(), `"notes":"agent skill updated notes"`) {
		t.Fatalf("expected media notes persistence evidence, body=%s", updateNotes.Body.String())
	}

	detach := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.detach_from_item",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","item_id":"media-item-1"}
	}`), map[string]string{"Content-Type": "application/json"})
	if detach.Code != http.StatusOK {
		t.Fatalf("media detach status=%d body=%s", detach.Code, detach.Body.String())
	}
	if !strings.Contains(detach.Body.String(), `"operation":"media.detach_from_item"`) ||
		!strings.Contains(detach.Body.String(), `"detachment_persisted":true`) {
		t.Fatalf("expected media detach persistence evidence, body=%s", detach.Body.String())
	}
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = 'media-item-1'`, p.ID, uploadPayload.Target.MediaID).Scan(&linkCount); err != nil {
		t.Fatalf("count media link after detach: %v", err)
	}
	if linkCount != 0 {
		t.Fatalf("expected media link to be detached, got %d", linkCount)
	}
}

func TestAgentSkillApplyAPIHandlesDiscoveriesSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Discoveries"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES ('discoveries-item-1', ?, 'AFX', 'Slot Cars', 'DISC-ITEM-1', 'Agent discovery wishlist target', 'active');
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json)
		VALUES ('discoveries-q1', ?, 'Agent discoveries saved search', '["afx"]', '[]', '["ebay"]');
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES
			('disc-review', ?, 'discoveries-q1', 'LIST-REVIEW', 'Agent review discovery', 14, 0, 'https://example.test/review', '', 'seller-review', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 3),
			('disc-dismiss', ?, 'discoveries-q1', 'LIST-DISMISS', 'Agent dismiss discovery', 15, 0, 'https://example.test/dismiss', '', 'seller-dismiss', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 2),
			('disc-wishlist', ?, 'discoveries-q1', 'LIST-WISH', 'Agent wishlist discovery', 16, 0, 'https://example.test/wish', '', 'seller-wish', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 1),
			('disc-purchase', ?, 'discoveries-q1', 'LIST-PURCHASE', 'Agent purchase discovery', 17, 0, 'https://example.test/purchase', '', 'seller-purchase', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 5),
			('disc-inventory', ?, 'discoveries-q1', 'LIST-INVENTORY', 'Agent inventory discovery', 18, 0, 'https://example.test/inventory', '', 'seller-inventory', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 6);
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('disc-review', '', 'not_in_collection', 0.8, 1, 'DISC-REVIEW', CURRENT_TIMESTAMP),
			('disc-dismiss', '', 'not_in_collection', 0.8, 1, 'DISC-DISMISS', CURRENT_TIMESTAMP),
			('disc-wishlist', 'discoveries-item-1', 'not_in_collection', 0.9, 0, 'DISC-WISH', CURRENT_TIMESTAMP),
			('disc-purchase', '', 'not_in_collection', 0.9, 0, 'DISC-PURCHASE', CURRENT_TIMESTAMP),
			('disc-inventory', '', 'not_in_collection', 0.9, 0, 'DISC-INVENTORY', CURRENT_TIMESTAMP);
	`, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID); err != nil {
		t.Fatalf("seed discovery skill data: %v", err)
	}

	search := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.search",
		"parameters":{"provider_id":"ebay","query":"wishlist"}
	}`), map[string]string{"Content-Type": "application/json"})
	if search.Code != http.StatusOK {
		t.Fatalf("discoveries search status=%d body=%s", search.Code, search.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"discoveries.search"`,
		`"read_only":true`,
		`"candidate_id":"disc-wishlist"`,
	} {
		if !strings.Contains(search.Body.String(), want) {
			t.Fatalf("discoveries search response missing %s: body=%s", want, search.Body.String())
		}
	}

	review := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.review_result",
		"parameters":{"provider_id":"ebay","result_id":"disc-review"}
	}`), map[string]string{"Content-Type": "application/json"})
	if review.Code != http.StatusOK {
		t.Fatalf("discoveries review status=%d body=%s", review.Code, review.Body.String())
	}
	if !strings.Contains(review.Body.String(), `"operation":"discoveries.review_result"`) ||
		!strings.Contains(review.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(review.Body.String(), `"source_result_url":"https://example.test/review"`) {
		t.Fatalf("expected read-only discovery review evidence, body=%s", review.Body.String())
	}

	dismiss := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.dismiss_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-dismiss","notes":"not relevant for collection"}
	}`), map[string]string{"Content-Type": "application/json"})
	if dismiss.Code != http.StatusOK {
		t.Fatalf("discoveries dismiss status=%d body=%s", dismiss.Code, dismiss.Body.String())
	}
	if !strings.Contains(dismiss.Body.String(), `"operation":"discoveries.dismiss_result"`) ||
		!strings.Contains(dismiss.Body.String(), `"action":"ignore"`) ||
		!strings.Contains(dismiss.Body.String(), `"discovery_persisted":true`) ||
		!strings.Contains(dismiss.Body.String(), `"provenance_preserved":true`) {
		t.Fatalf("expected discovery dismiss persistence evidence, body=%s", dismiss.Body.String())
	}
	var dismissedCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM ignored_candidates WHERE candidate_id = 'disc-dismiss'`).Scan(&dismissedCount); err != nil {
		t.Fatalf("count dismissed discovery: %v", err)
	}
	if dismissedCount != 1 {
		t.Fatalf("expected one ignored discovery marker, got %d", dismissedCount)
	}

	wishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.send_to_wishlist",
		"confirm":true,
		"source_surface":"telegram.discoveries.result",
		"source_channel":"telegram",
		"source_thread_id":"tg-discovery-thread-1709",
		"source_message_id":"tg-discovery-message-1709",
		"parameters":{"provider_id":"ebay","result_id":"disc-wishlist","notes":"promote from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlist.Code != http.StatusOK {
		t.Fatalf("discoveries wishlist status=%d body=%s", wishlist.Code, wishlist.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"discoveries.send_to_wishlist"`,
		`"action":"add_to_wishlist"`,
		`"source_channel":"telegram"`,
		`"discovery_persisted":true`,
		`"provenance_preserved":true`,
	} {
		if !strings.Contains(wishlist.Body.String(), want) {
			t.Fatalf("discoveries wishlist response missing %s: body=%s", want, wishlist.Body.String())
		}
	}
	var wishlistCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM wishlist_entries WHERE profile_id = ? AND item_id = 'discoveries-item-1' AND highlight_hit = 1`, p.ID).Scan(&wishlistCount); err != nil {
		t.Fatalf("count discovery wishlist handoff: %v", err)
	}
	if wishlistCount != 1 {
		t.Fatalf("expected one persisted wishlist handoff, got %d", wishlistCount)
	}

	purchase := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.create_purchase",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-purchase","quantity":2,"notes":"purchase candidate from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if purchase.Code != http.StatusOK {
		t.Fatalf("discoveries purchase status=%d body=%s", purchase.Code, purchase.Body.String())
	}
	if !strings.Contains(purchase.Body.String(), `"operation":"discoveries.create_purchase"`) ||
		!strings.Contains(purchase.Body.String(), `"action":"mark_purchased"`) ||
		!strings.Contains(purchase.Body.String(), `"discovery_persisted":true`) {
		t.Fatalf("expected discovery purchase persistence evidence, body=%s", purchase.Body.String())
	}
	var purchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND source = 'market_watch' AND external_ref = 'LIST-PURCHASE'`, p.ID).Scan(&purchaseCount); err != nil {
		t.Fatalf("count discovery purchase handoff: %v", err)
	}
	if purchaseCount != 1 {
		t.Fatalf("expected one persisted purchase handoff, got %d", purchaseCount)
	}

	inventory := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.create_or_update_inventory_candidate",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-inventory","notes":"inventory candidate from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if inventory.Code != http.StatusOK {
		t.Fatalf("discoveries inventory status=%d body=%s", inventory.Code, inventory.Body.String())
	}
	if !strings.Contains(inventory.Body.String(), `"operation":"discoveries.create_or_update_inventory_candidate"`) ||
		!strings.Contains(inventory.Body.String(), `"action":"create_item"`) ||
		!strings.Contains(inventory.Body.String(), `"discovery_persisted":true`) {
		t.Fatalf("expected discovery inventory persistence evidence, body=%s", inventory.Body.String())
	}
	var inventoryCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = 'DISC-INVENTORY' AND title = 'Agent inventory discovery'`, p.ID).Scan(&inventoryCount); err != nil {
		t.Fatalf("count discovery inventory candidate: %v", err)
	}
	if inventoryCount != 1 {
		t.Fatalf("expected one persisted inventory candidate item, got %d", inventoryCount)
	}
}

func TestAgentSkillApplyAPIConfirmsInboxMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Inbox Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	record := doRequest(t, a, http.MethodPost, "/api/chat/inbox", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"records":[{
			"local_history_id":"agent-skill-apply-1",
			"title":"Agent skill handoff",
			"summary":"Needs review"
		},{
			"local_history_id":"agent-skill-apply-2",
			"title":"Agent skill archive",
			"summary":"Can be archived"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if record.Code != http.StatusCreated {
		t.Fatalf("create inbox record status=%d body=%s", record.Code, record.Body.String())
	}
	var recordPayload struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(record.Body).Decode(&recordPayload); err != nil {
		t.Fatalf("decode inbox record: %v", err)
	}
	if len(recordPayload.Items) != 2 {
		t.Fatalf("expected two inbox items, got %+v", recordPayload.Items)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview inbox status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"mutation_applied":false`) || !strings.Contains(preview.Body.String(), `"blocker":"confirmation_required"`) {
		t.Fatalf("preview must stay non-mutating and confirmation-gated, body=%s", preview.Body.String())
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"confirm":true,
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply inbox status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"mutation_applied":true`) || !strings.Contains(apply.Body.String(), `"status":"read"`) {
		t.Fatalf("expected confirmed apply to mark inbox item read, body=%s", apply.Body.String())
	}

	archive := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.archive_or_hide",
		"confirm":true,
		"parameters":{"notification_id":"`+recordPayload.Items[1].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if archive.Code != http.StatusOK {
		t.Fatalf("archive inbox status=%d body=%s", archive.Code, archive.Body.String())
	}
	if !strings.Contains(archive.Body.String(), `"mutation_applied":true`) || !strings.Contains(archive.Body.String(), `"status":"archived"`) {
		t.Fatalf("expected confirmed apply to archive inbox item, body=%s", archive.Body.String())
	}
}

func TestAgentSkillApplyAPIConfirmsUsersMutationAndProtectsOwner(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Users Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	invite := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.invite_user",
		"confirm":true,
		"parameters":{"target_email":"agent_skill_invite@example.test","target_role":"view"}
	}`), map[string]string{"Content-Type": "application/json"})
	if invite.Code != http.StatusOK {
		t.Fatalf("invite apply status=%d body=%s", invite.Code, invite.Body.String())
	}
	if !strings.Contains(invite.Body.String(), `"mutation_applied":true`) || !strings.Contains(invite.Body.String(), `"email":"agent_skill_invite@example.test"`) || !strings.Contains(invite.Body.String(), `"status":"invited"`) {
		t.Fatalf("expected invite apply result, body=%s", invite.Body.String())
	}

	users, err := listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	var ownerID string
	for _, user := range users {
		if strings.HasPrefix(user.Username, "owner_") {
			ownerID = user.ID
		}
	}
	if ownerID == "" {
		t.Fatalf("expected seeded owner user, got %+v", users)
	}
	var invitedID string
	for _, user := range users {
		if user.Email == "agent_skill_invite@example.test" {
			invitedID = user.ID
		}
	}
	if invitedID == "" {
		t.Fatalf("expected invited user, got %+v", users)
	}

	updateRole := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.update_role",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`","target_role":"admin"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateRole.Code != http.StatusOK {
		t.Fatalf("update role apply status=%d body=%s", updateRole.Code, updateRole.Body.String())
	}
	if !strings.Contains(updateRole.Body.String(), `"mutation_applied":true`) || !strings.Contains(updateRole.Body.String(), `"role":"admin"`) {
		t.Fatalf("expected confirmed role update result, body=%s", updateRole.Body.String())
	}

	deactivate := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.activate_or_deactivate",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`","target_status":"inactive"}
	}`), map[string]string{"Content-Type": "application/json"})
	if deactivate.Code != http.StatusOK {
		t.Fatalf("deactivate apply status=%d body=%s", deactivate.Code, deactivate.Body.String())
	}
	if !strings.Contains(deactivate.Body.String(), `"mutation_applied":true`) || !strings.Contains(deactivate.Body.String(), `"status":"inactive"`) {
		t.Fatalf("expected confirmed deactivate result, body=%s", deactivate.Body.String())
	}

	removeOwner := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{"target_user":"`+ownerID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if removeOwner.Code != http.StatusBadRequest {
		t.Fatalf("remove protected owner status=%d body=%s", removeOwner.Code, removeOwner.Body.String())
	}
	if !strings.Contains(removeOwner.Body.String(), `"mutation_applied":false`) || !strings.Contains(removeOwner.Body.String(), `"blocker":"users_admin_protected_owner_change_blocked"`) {
		t.Fatalf("expected protected owner blocker, body=%s", removeOwner.Body.String())
	}

	removeInvited := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if removeInvited.Code != http.StatusOK {
		t.Fatalf("remove invited user status=%d body=%s", removeInvited.Code, removeInvited.Body.String())
	}
	if !strings.Contains(removeInvited.Body.String(), `"mutation_applied":true`) || !strings.Contains(removeInvited.Body.String(), `"removed_user_id":"`+invitedID+`"`) {
		t.Fatalf("expected confirmed remove user result, body=%s", removeInvited.Body.String())
	}
	users, err = listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users after remove: %v", err)
	}
	for _, user := range users {
		if user.ID == invitedID {
			t.Fatalf("expected invited user to be removed, got %+v", users)
		}
	}
}

func TestAgentSkillApplyAPIHandlesIntegrationsAndSettingsSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Integrations Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', 'error', 'credentials expired; refresh token required', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}

	testConnection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.test_connection",
		"parameters":{"provider_id":"ebay","provider_secret":"must-not-leak"}
	}`), map[string]string{"Content-Type": "application/json"})
	if testConnection.Code != http.StatusOK {
		t.Fatalf("test connection apply status=%d body=%s", testConnection.Code, testConnection.Body.String())
	}
	if !strings.Contains(testConnection.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(testConnection.Body.String(), `"operation":"integrations.provider.test_connection"`) ||
		!strings.Contains(testConnection.Body.String(), `"connection_status":"needs_reauthentication"`) ||
		!strings.Contains(testConnection.Body.String(), `"provider_health"`) ||
		!strings.Contains(testConnection.Body.String(), `"next_action":"check_provider_health_and_credentials"`) ||
		strings.Contains(testConnection.Body.String(), "must-not-leak") {
		t.Fatalf("expected non-mutating provider health test without secret leak, body=%s", testConnection.Body.String())
	}

	configure := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"confirm":true,
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"telegram",
		"source_thread_id":"integration-thread-1",
		"source_message_id":"integration-message-1",
		"parameters":{"provider_id":"ebay","provider_secret":"must-not-leak","setup_step":"oauth"}
	}`), map[string]string{"Content-Type": "application/json"})
	if configure.Code != http.StatusOK {
		t.Fatalf("configure provider apply status=%d body=%s", configure.Code, configure.Body.String())
	}
	if !strings.Contains(configure.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(configure.Body.String(), `"source_surface":"settings.integrations.provider.card"`) ||
		!strings.Contains(configure.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(configure.Body.String(), `"source_thread_id":"integration-thread-1"`) ||
		!strings.Contains(configure.Body.String(), `"source_message_id":"integration-message-1"`) ||
		!strings.Contains(configure.Body.String(), `"operation":"integrations.provider.configure"`) ||
		!strings.Contains(configure.Body.String(), `"secret_redacted":true`) ||
		!strings.Contains(configure.Body.String(), `"secret_persisted":true`) ||
		strings.Contains(configure.Body.String(), "must-not-leak") {
		t.Fatalf("expected confirmed provider configure result without secret leak, body=%s", configure.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.ebay.enabled", "true")
	assertProfileSetting(t, a, p.ID, "integration.ebay.setup_step", "oauth")

	repair := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.repair_provider",
		"confirm":true,
		"parameters":{"provider_name":"ebay"}
	}`), map[string]string{"Content-Type": "application/json"})
	if repair.Code != http.StatusOK {
		t.Fatalf("repair provider apply status=%d body=%s", repair.Code, repair.Body.String())
	}
	if !strings.Contains(repair.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(repair.Body.String(), `"operation":"integrations.provider.repair"`) ||
		!strings.Contains(repair.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected confirmed provider repair result without external write claim, body=%s", repair.Body.String())
	}

	disable := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.disable_provider",
		"confirm":true,
		"parameters":{"provider_id":"ebay"}
	}`), map[string]string{"Content-Type": "application/json"})
	if disable.Code != http.StatusOK {
		t.Fatalf("disable provider apply status=%d body=%s", disable.Code, disable.Body.String())
	}
	if !strings.Contains(disable.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(disable.Body.String(), `"operation":"integrations.provider.disable"`) ||
		!strings.Contains(disable.Body.String(), `"external_write_claimed":false`) ||
		!strings.Contains(disable.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed provider disable result without external write claim, body=%s", disable.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.ebay.enabled", "false")

	appearance := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_appearance",
		"confirm":true,
		"parameters":{"setting_key":"theme","setting_scope":"appearance","setting_value":"dark"}
	}`), map[string]string{"Content-Type": "application/json"})
	if appearance.Code != http.StatusOK {
		t.Fatalf("appearance apply status=%d body=%s", appearance.Code, appearance.Body.String())
	}
	if !strings.Contains(appearance.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(appearance.Body.String(), `"operation":"settings.appearance.update"`) ||
		!strings.Contains(appearance.Body.String(), `"setting_key":"theme"`) ||
		!strings.Contains(appearance.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed appearance setting result, body=%s", appearance.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "theme", "dark")

	storageStatus := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.storage.show_status"
	}`), map[string]string{"Content-Type": "application/json"})
	if storageStatus.Code != http.StatusOK {
		t.Fatalf("storage status apply status=%d body=%s", storageStatus.Code, storageStatus.Body.String())
	}
	if !strings.Contains(storageStatus.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(storageStatus.Body.String(), `"operation":"storage.status.show"`) ||
		!strings.Contains(storageStatus.Body.String(), `"read_only":true`) {
		t.Fatalf("expected read-only storage status result, body=%s", storageStatus.Body.String())
	}

	configureBackup := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.storage.configure_backup",
		"confirm":true,
		"parameters":{"backup_target":"backups/nightly"}
	}`), map[string]string{"Content-Type": "application/json"})
	if configureBackup.Code != http.StatusOK {
		t.Fatalf("configure backup apply status=%d body=%s", configureBackup.Code, configureBackup.Body.String())
	}
	if !strings.Contains(configureBackup.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(configureBackup.Body.String(), `"operation":"storage.backup.configure"`) ||
		!strings.Contains(configureBackup.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed backup settings persistence result, body=%s", configureBackup.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "storage.backup_target", "backups/nightly")

	exportBundle := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.export_bundle",
		"parameters":{"export_scope":"profile"}
	}`), map[string]string{"Content-Type": "application/json"})
	if exportBundle.Code != http.StatusOK {
		t.Fatalf("export bundle apply status=%d body=%s", exportBundle.Code, exportBundle.Body.String())
	}
	if !strings.Contains(exportBundle.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(exportBundle.Body.String(), `"operation":"data.export.bundle"`) ||
		!strings.Contains(exportBundle.Body.String(), `"read_only":true`) {
		t.Fatalf("expected read-only export bundle result, body=%s", exportBundle.Body.String())
	}

	importFile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.import_file",
		"confirm":true,
		"parameters":{"file_path":"imports/profile.json"}
	}`), map[string]string{"Content-Type": "application/json"})
	if importFile.Code != http.StatusOK {
		t.Fatalf("import file apply status=%d body=%s", importFile.Code, importFile.Body.String())
	}
	if !strings.Contains(importFile.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(importFile.Body.String(), `"operation":"data.import.file"`) ||
		!strings.Contains(importFile.Body.String(), `"impact":"import_preview_confirmed"`) {
		t.Fatalf("expected confirmed import preview result, body=%s", importFile.Body.String())
	}

	restore := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.restore_backup",
		"confirm":true,
		"parameters":{"backup_path":"backups/cabinet-backup.zip"}
	}`), map[string]string{"Content-Type": "application/json"})
	if restore.Code != http.StatusOK {
		t.Fatalf("restore backup apply status=%d body=%s", restore.Code, restore.Body.String())
	}
	if !strings.Contains(restore.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(restore.Body.String(), `"operation":"data.backup.restore"`) ||
		!strings.Contains(restore.Body.String(), `"destructive_confirmation":true`) {
		t.Fatalf("expected destructive restore confirmation result, body=%s", restore.Body.String())
	}
}

func TestAgentSkillApplyAPICapturesStubbedProviderWritePathEvidence(t *testing.T) {
	t.Parallel()

	const providerSecret = "issue-1780-provider-secret-must-not-leak"

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Provider Evidence"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('openai', 'auth_missing', 'missing credential: configure provider API key', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}

	testConnection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.test_connection",
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"in-app",
		"parameters":{"provider_id":"openai","provider_secret":"`+providerSecret+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if testConnection.Code != http.StatusOK {
		t.Fatalf("test connection status=%d body=%s", testConnection.Code, testConnection.Body.String())
	}
	requireBodyOmitsSecret(t, testConnection.Body.String(), providerSecret)
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"integrations.provider.test_connection"`,
		`"connection_status":"needs_setup"`,
		`"guidance":"Save provider credentials and marketplace setup before running Market Watch."`,
		`"next_action":"review_provider_status"`,
	} {
		if !strings.Contains(testConnection.Body.String(), want) {
			t.Fatalf("provider readiness response missing %s: body=%s", want, testConnection.Body.String())
		}
	}

	configure := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"confirm":true,
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"in-app",
		"parameters":{
			"provider_id":"openai",
			"provider_secret":"`+providerSecret+`",
			"setup_step":"api-key",
			"base_url":"https://api.openai.test",
			"marketplace":"global",
			"items_per_page":"25"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if configure.Code != http.StatusOK {
		t.Fatalf("configure provider status=%d body=%s", configure.Code, configure.Body.String())
	}
	configureBody := configure.Body.String()
	requireBodyOmitsSecret(t, configureBody, providerSecret)
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"integrations.provider.configure"`,
		`"secret_redacted":true`,
		`"secret_persisted":true`,
		`"external_write_claimed":false`,
		`"source_surface":"settings.integrations.provider.card"`,
		`"source_channel":"in-app"`,
		`"next_action":"Run provider health validation from Integrations before routing live provider workflows."`,
	} {
		if !strings.Contains(configureBody, want) {
			t.Fatalf("configure provider response missing %s: body=%s", want, configureBody)
		}
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.enabled", "true")
	assertProfileSetting(t, a, p.ID, "integration.openai.setup_step", "api-key")
	assertProfileSetting(t, a, p.ID, "integration.openai.base_url", "https://api.openai.test")
	assertProfileSetting(t, a, p.ID, "integration.openai.marketplace", "global")
	assertProfileSetting(t, a, p.ID, "integration.openai.items_per_page", "25")
	assertProfileSecret(t, a, p.ID, "integration.openai.token", providerSecret)

	repair := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.repair_provider",
		"confirm":true,
		"parameters":{"provider_id":"openai"}
	}`), map[string]string{"Content-Type": "application/json"})
	if repair.Code != http.StatusOK {
		t.Fatalf("repair provider status=%d body=%s", repair.Code, repair.Body.String())
	}
	requireBodyOmitsSecret(t, repair.Body.String(), providerSecret)
	if !strings.Contains(repair.Body.String(), `"operation":"integrations.provider.repair"`) ||
		!strings.Contains(repair.Body.String(), `"next_action":"Run a provider health check after reviewing repaired setup steps."`) ||
		!strings.Contains(repair.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected actionable repair result without external write claim, body=%s", repair.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.repair_status", "reviewed")

	disable := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.disable_provider",
		"confirm":true,
		"parameters":{"provider_id":"openai"}
	}`), map[string]string{"Content-Type": "application/json"})
	if disable.Code != http.StatusOK {
		t.Fatalf("disable provider status=%d body=%s", disable.Code, disable.Body.String())
	}
	requireBodyOmitsSecret(t, disable.Body.String(), providerSecret)
	if !strings.Contains(disable.Body.String(), `"operation":"integrations.provider.disable"`) ||
		!strings.Contains(disable.Body.String(), `"next_action":"Confirm provider disabled state in Integrations before routing provider-backed workflows."`) ||
		!strings.Contains(disable.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected actionable disable result without external write claim, body=%s", disable.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.enabled", "false")
}

func TestAgentSkillApplyAPIHandlesMarketWatchAndPurchasesSkills(t *testing.T) {
	t.Parallel()

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer agent-skill-token" {
			t.Fatalf("expected bearer token header, got %q", got)
		}
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_AU" {
			t.Fatalf("expected EBAY_AU marketplace header, got %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "boxed kit" {
			t.Fatalf("expected query q=boxed kit, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"agent-skill-ebay-1","title":"Agent skill eBay live provider result","price":{"value":"88.50","currency":"AUD"},"itemWebUrl":"https://www.ebay.com.au/itm/agent-skill-ebay-1","image":{"imageUrl":"https://example.test/agent-skill-ebay-1.jpg"},"seller":{"username":"agent-seller"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK","estimatedAvailableQuantity":4}]}]}`))
	}))
	defer ebayStub.Close()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Market Watch Purchases"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if err := profile.NewRepository(a.db).PutSettings(context.Background(), p.ID, map[string]string{
		"ebay_base_url":                   ebayStub.URL,
		"ebay_bearer_token":               "agent-skill-token",
		"ebay_marketplace":                "EBAY_AU",
		"integration.ebay.items_per_page": "11",
	}); err != nil {
		t.Fatalf("save ebay provider settings: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', 'auth_missing', 'provider credentials required before live watch run', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}

	missingProvider := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.run_watch",
		"parameters":{"watch_id":"watch-42"}
	}`), map[string]string{"Content-Type": "application/json"})
	if missingProvider.Code != http.StatusOK {
		t.Fatalf("market watch preview status=%d body=%s", missingProvider.Code, missingProvider.Body.String())
	}
	if !strings.Contains(missingProvider.Body.String(), `"blocker":"market_watch_provider_required"`) ||
		!strings.Contains(missingProvider.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected provider blocker without mutation, body=%s", missingProvider.Body.String())
	}

	createWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.create_saved_watch",
		"confirm":true,
		"parameters":{"provider_id":"ebay","watch_name":"Agent boxed kits","watch_query":"boxed model kit","region":"AU","enabled":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if createWatch.Code != http.StatusOK {
		t.Fatalf("create saved watch status=%d body=%s", createWatch.Code, createWatch.Body.String())
	}
	for _, want := range []string{
		`"operation":"market_watch.watch.create"`,
		`"watch_persisted":true`,
		`"watch_query":"boxed model kit"`,
		`"provider_scope":["ebay"]`,
	} {
		if !strings.Contains(createWatch.Body.String(), want) {
			t.Fatalf("create watch response missing %s: body=%s", want, createWatch.Body.String())
		}
	}
	var watchID string
	if err := a.db.QueryRow(`SELECT id FROM scanner_query_sets WHERE profile_id = ? AND name = 'Agent boxed kits'`, p.ID).Scan(&watchID); err != nil {
		t.Fatalf("load created saved watch: %v", err)
	}

	updateWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.update_saved_watch",
		"confirm":true,
		"parameters":{"provider_id":"ebay","watch_id":"`+watchID+`","watch_name":"Agent boxed kits under 100","keywords":["boxed","kit"],"max_price":100}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateWatch.Code != http.StatusOK {
		t.Fatalf("update saved watch status=%d body=%s", updateWatch.Code, updateWatch.Body.String())
	}
	if !strings.Contains(updateWatch.Body.String(), `"operation":"market_watch.watch.update"`) ||
		!strings.Contains(updateWatch.Body.String(), `"watch_persisted":true`) ||
		!strings.Contains(updateWatch.Body.String(), `"max_price":100`) {
		t.Fatalf("expected persisted saved watch update evidence, body=%s", updateWatch.Body.String())
	}

	searchWatches := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.search_watches",
		"parameters":{"provider_id":"ebay","query":"under 100"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchWatches.Code != http.StatusOK {
		t.Fatalf("search saved watches status=%d body=%s", searchWatches.Code, searchWatches.Body.String())
	}
	if !strings.Contains(searchWatches.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(searchWatches.Body.String(), `"operation":"market_watch.watch.search"`) ||
		!strings.Contains(searchWatches.Body.String(), `"name":"Agent boxed kits under 100"`) {
		t.Fatalf("expected read-only saved watch reload evidence, body=%s", searchWatches.Body.String())
	}
	if _, err := a.db.Exec(`UPDATE provider_health SET status = 'ok', message = '', retry_after_seconds = 0, updated_at = CURRENT_TIMESTAMP WHERE provider = 'ebay'`); err != nil {
		t.Fatalf("mark provider health ready: %v", err)
	}

	runWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.run_watch",
		"confirm":true,
		"source_surface":"market_watch.saved_watch.row",
		"source_channel":"in-app",
		"parameters":{"provider_id":"ebay","watch_id":"`+watchID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if runWatch.Code != http.StatusOK {
		t.Fatalf("run watch status=%d body=%s", runWatch.Code, runWatch.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"market_watch.watch.run"`,
		`"provider_id":"ebay"`,
		`"watch_id":"` + watchID + `"`,
		`"provider_health"`,
		`"saved_watch"`,
		`"external_write_claimed":false`,
		`"live_provider_dispatched":true`,
		`"status":"confirmed_provider_run"`,
		`"candidate_count":1`,
		`"title":"Agent skill eBay live provider result"`,
		`"items_per_page_requested":11`,
		`"source_surface":"market_watch.saved_watch.row"`,
	} {
		if !strings.Contains(runWatch.Body.String(), want) {
			t.Fatalf("run watch response missing %s: body=%s", want, runWatch.Body.String())
		}
	}
	var providerRunCount, providerCandidateCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND provider = 'ebay' AND status = 'succeeded'`, p.ID, watchID).Scan(&providerRunCount); err != nil {
		t.Fatalf("count agent skill provider runs: %v", err)
	}
	if providerRunCount != 1 {
		t.Fatalf("expected one persisted provider run, got %d", providerRunCount)
	}
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND source = 'ebay' AND listing_id = 'agent-skill-ebay-1'`, p.ID, watchID).Scan(&providerCandidateCount); err != nil {
		t.Fatalf("count agent skill provider candidates: %v", err)
	}
	if providerCandidateCount != 1 {
		t.Fatalf("expected one persisted provider candidate, got %d", providerCandidateCount)
	}

	handoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-9","destination":"purchases","source_url":"https://example.test/listing/9"}
	}`), map[string]string{"Content-Type": "application/json"})
	if handoff.Code != http.StatusOK {
		t.Fatalf("handoff status=%d body=%s", handoff.Code, handoff.Body.String())
	}
	if !strings.Contains(handoff.Body.String(), `"provenance_preserved":true`) ||
		!strings.Contains(handoff.Body.String(), `"result_id":"result-9"`) ||
		!strings.Contains(handoff.Body.String(), `"destination":"purchases"`) ||
		!strings.Contains(handoff.Body.String(), `"handoff_persisted":true`) ||
		!strings.Contains(handoff.Body.String(), `"lifecycle_entry_id":"`) ||
		!strings.Contains(handoff.Body.String(), `"expected_arrival_id":"`) {
		t.Fatalf("expected provenance-preserving handoff result, body=%s", handoff.Body.String())
	}
	var handoffPurchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND source = 'market_watch' AND external_ref = 'result-9'`, p.ID).Scan(&handoffPurchaseCount); err != nil {
		t.Fatalf("count market watch purchase handoff: %v", err)
	}
	if handoffPurchaseCount != 1 {
		t.Fatalf("expected one persisted purchase handoff, got %d", handoffPurchaseCount)
	}

	wishlistHandoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-10","destination":"wishlist","title":"Wishlist handoff result","source_url":"https://example.test/listing/10","target_price":55,"priority":"high"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlistHandoff.Code != http.StatusOK {
		t.Fatalf("wishlist handoff status=%d body=%s", wishlistHandoff.Code, wishlistHandoff.Body.String())
	}
	if !strings.Contains(wishlistHandoff.Body.String(), `"destination_applied":"wishlist"`) ||
		!strings.Contains(wishlistHandoff.Body.String(), `"wishlist_entry_id":"`) ||
		!strings.Contains(wishlistHandoff.Body.String(), `"handoff_persisted":true`) {
		t.Fatalf("expected wishlist handoff persistence evidence, body=%s", wishlistHandoff.Body.String())
	}
	var wishlistHandoffCount int
	if err := a.db.QueryRow(`
		SELECT COUNT(1)
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		WHERE w.profile_id = ? AND i.title = 'Wishlist handoff result' AND w.priority = 'high'
	`, p.ID).Scan(&wishlistHandoffCount); err != nil {
		t.Fatalf("count market watch wishlist handoff: %v", err)
	}
	if wishlistHandoffCount != 1 {
		t.Fatalf("expected one persisted wishlist handoff, got %d", wishlistHandoffCount)
	}

	inventoryHandoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-11","destination":"inventory","title":"Inventory handoff result","source_url":"https://example.test/listing/11","condition":"sealed","quantity":3}
	}`), map[string]string{"Content-Type": "application/json"})
	if inventoryHandoff.Code != http.StatusOK {
		t.Fatalf("inventory handoff status=%d body=%s", inventoryHandoff.Code, inventoryHandoff.Body.String())
	}
	if !strings.Contains(inventoryHandoff.Body.String(), `"destination_applied":"inventory"`) ||
		!strings.Contains(inventoryHandoff.Body.String(), `"instance_id":"`) ||
		!strings.Contains(inventoryHandoff.Body.String(), `"handoff_persisted":true`) {
		t.Fatalf("expected inventory handoff persistence evidence, body=%s", inventoryHandoff.Body.String())
	}
	var inventoryHandoffCount int
	if err := a.db.QueryRow(`
		SELECT COUNT(1)
		FROM instances inst
		JOIN canonical_items i ON i.id = inst.item_id
		WHERE i.profile_id = ? AND i.title = 'Inventory handoff result' AND inst.condition = 'sealed' AND inst.quantity = 3
	`, p.ID).Scan(&inventoryHandoffCount); err != nil {
		t.Fatalf("count market watch inventory handoff: %v", err)
	}
	if inventoryHandoffCount != 1 {
		t.Fatalf("expected one persisted inventory handoff, got %d", inventoryHandoffCount)
	}

	missingOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.add_line_item",
		"parameters":{"item_id":"item-99"}
	}`), map[string]string{"Content-Type": "application/json"})
	if missingOrder.Code != http.StatusOK {
		t.Fatalf("purchase preview status=%d body=%s", missingOrder.Code, missingOrder.Body.String())
	}
	if !strings.Contains(missingOrder.Body.String(), `"blocker":"purchases_order_required"`) {
		t.Fatalf("expected missing order blocker, body=%s", missingOrder.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status) VALUES ('item-99', ?, 'AFX', 'Slot Cars', 'AGENT-PURCHASE-99', 'Agent purchase line item', 'active')`, p.ID); err != nil {
		t.Fatalf("seed purchase target item: %v", err)
	}

	createOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.create_order",
		"confirm":true,
		"parameters":{
			"purchase_source":"agent_manual",
			"order_id":"agent-order-1",
			"title":"Agent created purchase order item",
			"part_number":"AGENT-CREATE-1",
			"source_url":"https://example.test/order/1",
			"quantity":1,
			"amount":77,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if createOrder.Code != http.StatusOK {
		t.Fatalf("create order status=%d body=%s", createOrder.Code, createOrder.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.order.create"`,
		`"order_id":"agent-order-1"`,
		`"created_item":true`,
		`"purchase_persisted":true`,
		`"provenance_preserved":true`,
		`"expected_arrival_id":"`,
	} {
		if !strings.Contains(createOrder.Body.String(), want) {
			t.Fatalf("create order response missing %s: body=%s", want, createOrder.Body.String())
		}
	}

	addLine := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.add_line_item",
		"confirm":true,
		"parameters":{
			"order_id":"order-1",
			"item_id":"item-99",
			"source":"market_watch",
			"result_id":"result-9",
			"source_url":"https://example.test/listing/9",
			"seller":"seller-one",
			"tracking":"TRACK-99",
			"quantity":2,
			"amount":42.5,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if addLine.Code != http.StatusOK {
		t.Fatalf("add line item status=%d body=%s", addLine.Code, addLine.Body.String())
	}
	if !strings.Contains(addLine.Body.String(), `"operation":"purchases.order.add_line_item"`) ||
		!strings.Contains(addLine.Body.String(), `"order_id":"order-1"`) ||
		!strings.Contains(addLine.Body.String(), `"item_id":"item-99"`) ||
		!strings.Contains(addLine.Body.String(), `"purchase_persisted":true`) ||
		!strings.Contains(addLine.Body.String(), `"provenance_preserved":true`) ||
		!strings.Contains(addLine.Body.String(), `"lifecycle_entry_id":"`) ||
		!strings.Contains(addLine.Body.String(), `"expected_arrival_id":"`) {
		t.Fatalf("expected purchase add line item preview/apply evidence, body=%s", addLine.Body.String())
	}

	searchOrders := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.search_orders",
		"parameters":{"query":"order-1","status":"all"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchOrders.Code != http.StatusOK {
		t.Fatalf("search purchase orders status=%d body=%s", searchOrders.Code, searchOrders.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"purchases.orders.search"`,
		`"order_id":"order-1"`,
		`"source":"market_watch"`,
		`"seller":"seller-one"`,
		`"tracking":"TRACK-99"`,
		`"line_item_count":1`,
		`"expected_arrival_id":"`,
	} {
		if !strings.Contains(searchOrders.Body.String(), want) {
			t.Fatalf("purchase order search response missing %s: body=%s", want, searchOrders.Body.String())
		}
	}

	receiveLine := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.receive_line_item",
		"confirm":true,
		"parameters":{"order_id":"order-1","line_item_id":"item-99","delivered_on":"2026-07-05","notes":"received by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if receiveLine.Code != http.StatusOK {
		t.Fatalf("receive line item status=%d body=%s", receiveLine.Code, receiveLine.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.line_item.receive"`,
		`"purchase_persisted":true`,
		`"arrival_status":"delivered"`,
		`"delivered_on":"2026-07-05"`,
	} {
		if !strings.Contains(receiveLine.Body.String(), want) {
			t.Fatalf("receive line response missing %s: body=%s", want, receiveLine.Body.String())
		}
	}

	reconcile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.reconcile_item",
		"confirm":true,
		"parameters":{"order_id":"order-1","item_id":"item-99","instance_id":"instance-99","notes":"reconciled by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if reconcile.Code != http.StatusOK {
		t.Fatalf("reconcile item status=%d body=%s", reconcile.Code, reconcile.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.item.reconcile"`,
		`"purchase_persisted":true`,
		`"reconciliation_persisted":true`,
		`"arrival_status":"reconciled"`,
		`"reconciled_instance_id":"instance-99"`,
	} {
		if !strings.Contains(reconcile.Body.String(), want) {
			t.Fatalf("reconcile item response missing %s: body=%s", want, reconcile.Body.String())
		}
	}

	searchReceived := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.search_orders",
		"parameters":{"query":"order-1","status":"received"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchReceived.Code != http.StatusOK {
		t.Fatalf("search received purchase orders status=%d body=%s", searchReceived.Code, searchReceived.Body.String())
	}
	if !strings.Contains(searchReceived.Body.String(), `"status":"received"`) ||
		!strings.Contains(searchReceived.Body.String(), `"received_count":1`) ||
		!strings.Contains(searchReceived.Body.String(), `"status":"reconciled"`) {
		t.Fatalf("expected received/reconciled purchase order search evidence, body=%s", searchReceived.Body.String())
	}

	receiveOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.receive_order",
		"confirm":true,
		"parameters":{"order_id":"agent-order-1","delivered_on":"2026-07-06","notes":"bulk receive by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if receiveOrder.Code != http.StatusOK {
		t.Fatalf("receive order status=%d body=%s", receiveOrder.Code, receiveOrder.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.order.receive"`,
		`"order_id":"agent-order-1"`,
		`"purchase_persisted":true`,
		`"received_count":1`,
		`"status":"delivered"`,
	} {
		if !strings.Contains(receiveOrder.Body.String(), want) {
			t.Fatalf("receive order response missing %s: body=%s", want, receiveOrder.Body.String())
		}
	}
}

func TestAgentSkillApplyAPIRequiresConfirmationAndRejectsUnknownSkill(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Apply Guard"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	cancel := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.invite_user",
		"confirm":false,
		"parameters":{"target_email":"agent_skill_cancel@example.test","target_role":"view"}
	}`), map[string]string{"Content-Type": "application/json"})
	if cancel.Code != http.StatusConflict {
		t.Fatalf("cancelled apply status=%d body=%s", cancel.Code, cancel.Body.String())
	}
	if !strings.Contains(cancel.Body.String(), `"error":"confirmation_required"`) {
		t.Fatalf("expected confirmation_required on cancelled apply, body=%s", cancel.Body.String())
	}
	users, err := listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users after cancelled apply: %v", err)
	}
	for _, user := range users {
		if user.Email == "agent_skill_cancel@example.test" {
			t.Fatalf("cancelled apply must not create user, got %+v", users)
		}
	}

	unknown := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.unsupported",
		"confirm":true,
		"parameters":{"target_user":"missing"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown apply status=%d body=%s", unknown.Code, unknown.Body.String())
	}
	if !strings.Contains(unknown.Body.String(), `"error":"skill_not_found"`) {
		t.Fatalf("expected skill_not_found on unknown skill, body=%s", unknown.Body.String())
	}
}

func findAPISkill(skills []apiSkillPayload, id string) *apiSkillPayload {
	for i := range skills {
		if skills[i].ID == id {
			return &skills[i]
		}
	}
	return nil
}

func assertProfileSetting(t *testing.T, a *App, profileID, key, want string) {
	t.Helper()
	var got string
	if err := a.db.QueryRow(`SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, profileID, key).Scan(&got); err != nil {
		t.Fatalf("read profile setting %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("profile setting %s = %q, want %q", key, got, want)
	}
}

func assertProfileSecret(t *testing.T, a *App, profileID, key, want string) {
	t.Helper()
	got, err := profile.NewRepository(a.db).GetSecret(context.Background(), profileID, key)
	if err != nil {
		t.Fatalf("read profile secret %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("profile secret %s = %q, want %q", key, got, want)
	}
}

func requireBodyOmitsSecret(t *testing.T, body, secret string) {
	t.Helper()
	if strings.Contains(body, secret) {
		t.Fatalf("response leaked secret %q: body=%s", secret, body)
	}
}
