package telegrambotconnector

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/telegrambotadapter"
	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

func TestClientGetMeValidatesBotWithoutLeakingToken(t *testing.T) {
	t.Parallel()
	const token = "123456:super-secret-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot"+token+"/getMe" {
			t.Fatalf("getMe path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"result": map[string]any{"id": 42, "is_bot": true, "username": "cabinet_test_bot"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, token, server.Client())
	bot, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if bot.ID != 42 || !bot.IsBot || bot.Username != "cabinet_test_bot" {
		t.Fatalf("unexpected bot identity: %+v", bot)
	}
	if strings.Contains(client.SafeEndpoint(), token) {
		t.Fatalf("SafeEndpoint() leaked token: %q", client.SafeEndpoint())
	}
}

func TestClientDetectsWebhookConflictAndDeletesOnlyExplicitly(t *testing.T) {
	t.Parallel()
	const token = "bot-token"
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, strings.TrimPrefix(r.URL.Path, "/bot"+token+"/"))
		switch methods[len(methods)-1] {
		case "getWebhookInfo":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{"url": "https://old.example/telegram"}})
		case "deleteWebhook":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": true})
		default:
			t.Fatalf("unexpected Bot API method: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, token, server.Client())
	info, err := client.GetWebhookInfo(context.Background())
	if err != nil {
		t.Fatalf("GetWebhookInfo() error = %v", err)
	}
	if !info.ConflictsWithPolling() {
		t.Fatalf("expected webhook conflict: %+v", info)
	}
	if !reflect.DeepEqual(methods, []string{"getWebhookInfo"}) {
		t.Fatalf("conflict detection mutated webhook: %v", methods)
	}
	if err := client.DeleteWebhook(context.Background(), false); err != nil {
		t.Fatalf("DeleteWebhook() error = %v", err)
	}
	if !reflect.DeepEqual(methods, []string{"getWebhookInfo", "deleteWebhook"}) {
		t.Fatalf("explicit resolution methods = %v", methods)
	}
}

func TestClientGetUpdatesSendsOffsetTimeoutAndRestrictedUpdateTypes(t *testing.T) {
	t.Parallel()
	const token = "bot-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bot"+token+"/getUpdates" {
			t.Fatalf("getUpdates path = %q", r.URL.Path)
		}
		var body struct {
			Offset         int64    `json:"offset"`
			Timeout        int      `json:"timeout"`
			AllowedUpdates []string `json:"allowed_updates"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode getUpdates: %v", err)
		}
		if body.Offset != 91 || body.Timeout != 25 || !reflect.DeepEqual(body.AllowedUpdates, RequiredUpdateTypes()) {
			t.Fatalf("unexpected getUpdates body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": []any{}})
	}))
	defer server.Close()

	client := NewClient(server.URL, token, server.Client())
	if _, err := client.GetUpdates(context.Background(), 91, 25, RequiredUpdateTypes()); err != nil {
		t.Fatalf("GetUpdates() error = %v", err)
	}
}

func TestClientAPIErrorHonorsRetryAfterAndRedactsToken(t *testing.T) {
	t.Parallel()
	const token = "123456:must-never-leak"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          false,
			"error_code":  429,
			"description": "retry token " + token,
			"parameters":  map[string]any{"retry_after": 7},
		})
	}))
	defer server.Close()

	_, err := NewClient(server.URL, token, server.Client()).GetUpdates(context.Background(), 0, 25, RequiredUpdateTypes())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusTooManyRequests || apiErr.RetryAfter != 7*time.Second {
		t.Fatalf("unexpected API error: %#v", err)
	}
	if strings.Contains(err.Error(), token) {
		t.Fatalf("API error leaked token: %v", err)
	}
}

func TestPollerAdvancesOffsetAfterProcessingAndRestartSkipsDuplicates(t *testing.T) {
	t.Parallel()
	store := &memoryOffsetStore{}
	api := &scriptedBotAPI{updates: [][]telegrambotadapter.Update{{
		{UpdateID: 10, Message: telegramMessage(10, 100, 100, "first")},
		{UpdateID: 11, Message: telegramMessage(11, 100, 100, "second")},
	}}}
	var processed []int64
	poller := NewPoller("profile-1", api, store, func(_ context.Context, update telegrambotadapter.Update) error {
		processed = append(processed, update.UpdateID)
		return nil
	})
	if _, err := poller.PollOnce(context.Background()); err != nil {
		t.Fatalf("PollOnce() error = %v", err)
	}
	if got := store.Offset("profile-1"); got != 12 {
		t.Fatalf("stored offset = %d, want 12", got)
	}
	if !reflect.DeepEqual(processed, []int64{10, 11}) {
		t.Fatalf("processed updates = %v", processed)
	}

	restartedAPI := &scriptedBotAPI{updates: [][]telegrambotadapter.Update{{
		{UpdateID: 11, Message: telegramMessage(11, 100, 100, "duplicate")},
		{UpdateID: 12, Message: telegramMessage(12, 100, 100, "after restart")},
	}}}
	restarted := NewPoller("profile-1", restartedAPI, store, func(_ context.Context, update telegrambotadapter.Update) error {
		processed = append(processed, update.UpdateID)
		return nil
	})
	if _, err := restarted.PollOnce(context.Background()); err != nil {
		t.Fatalf("restarted PollOnce() error = %v", err)
	}
	if !reflect.DeepEqual(processed, []int64{10, 11, 12}) {
		t.Fatalf("restart replayed a duplicate: %v", processed)
	}
	if got := restartedAPI.offsets[0]; got != 12 {
		t.Fatalf("getUpdates restart offset = %d, want 12", got)
	}
	if !reflect.DeepEqual(restartedAPI.allowed[0], RequiredUpdateTypes()) {
		t.Fatalf("allowed_updates = %v", restartedAPI.allowed[0])
	}
}

func TestPollerDoesNotAdvancePastFailedGovernedProcessing(t *testing.T) {
	t.Parallel()
	store := &memoryOffsetStore{}
	api := &scriptedBotAPI{updates: [][]telegrambotadapter.Update{{
		{UpdateID: 20, Message: telegramMessage(20, 100, 100, "ok")},
		{UpdateID: 21, Message: telegramMessage(21, 100, 100, "fail")},
		{UpdateID: 22, Message: telegramMessage(22, 100, 100, "must wait")},
	}}}
	poller := NewPoller("profile-1", api, store, func(_ context.Context, update telegrambotadapter.Update) error {
		if update.UpdateID == 21 {
			return errors.New("governed processing failed")
		}
		return nil
	})
	if _, err := poller.PollOnce(context.Background()); err == nil {
		t.Fatal("expected governed processing failure")
	}
	if got := store.Offset("profile-1"); got != 21 {
		t.Fatalf("offset advanced past failed update: %d", got)
	}
}

func TestPollerAdvancesAfterGovernedSuccessWhenReplyDeliveryFails(t *testing.T) {
	t.Parallel()
	store := &memoryOffsetStore{}
	api := &scriptedBotAPI{updates: [][]telegrambotadapter.Update{{
		{UpdateID: 23, Message: telegramMessage(23, 100, 100, "processed")},
	}}}
	poller := NewPoller("profile-1", api, store, func(_ context.Context, _ telegrambotadapter.Update) error {
		return MarkProcessed(errors.New("Telegram reply delivery failed"))
	})
	if _, err := poller.PollOnce(context.Background()); err == nil {
		t.Fatal("expected visible reply delivery failure")
	}
	if got := store.Offset("profile-1"); got != 24 {
		t.Fatalf("processed update was left eligible for duplicate apply: offset=%d", got)
	}
}

func TestPollerRejectsMalformedUpdateWithoutAdvancingOffset(t *testing.T) {
	t.Parallel()
	store := &memoryOffsetStore{}
	api := &scriptedBotAPI{updates: [][]telegrambotadapter.Update{{
		{UpdateID: 0},
	}}}
	processed := false
	poller := NewPoller("profile-1", api, store, func(_ context.Context, _ telegrambotadapter.Update) error {
		processed = true
		return nil
	})
	if _, err := poller.PollOnce(context.Background()); err == nil {
		t.Fatal("expected malformed update failure")
	}
	if processed {
		t.Fatal("malformed update reached governed processing")
	}
	if got := store.Offset("profile-1"); got != 0 {
		t.Fatalf("malformed update advanced offset: %d", got)
	}
}

func TestRunUsesBoundedBackoffAndTelegramRetryAfter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api := &scriptedBotAPI{errs: []error{
		&APIError{Code: 429, Description: "Too Many Requests", RetryAfter: 1500 * time.Millisecond},
		errors.New("offline"),
		&APIError{Code: 429, Description: "Too Many Requests", RetryAfter: 30 * time.Second},
	}}
	store := &memoryOffsetStore{}
	var waits []time.Duration
	poller := NewPoller("profile-1", api, store, func(context.Context, telegrambotadapter.Update) error { return nil })
	poller.SetBackoff(250*time.Millisecond, 2*time.Second)
	poller.SetWait(func(_ context.Context, d time.Duration) error {
		waits = append(waits, d)
		if len(waits) == 3 {
			cancel()
		}
		return nil
	})
	_ = poller.Run(ctx)
	if len(waits) < 3 || waits[0] != 1500*time.Millisecond || waits[1] != 500*time.Millisecond || waits[2] != 2*time.Second {
		t.Fatalf("unexpected retry schedule: %v", waits)
	}
	for _, wait := range waits {
		if wait > 2*time.Second {
			t.Fatalf("backoff exceeded bound: %v", waits)
		}
	}
}

func TestManagerRetryDelayHonorsTelegramRetryAfterWithinRuntimeBound(t *testing.T) {
	t.Parallel()
	if got := managerRetryDelay(&APIError{Code: 429, RetryAfter: 3 * time.Second}, 500*time.Millisecond, 30*time.Second); got != 3*time.Second {
		t.Fatalf("retry_after delay = %v, want 3s", got)
	}
	if got := managerRetryDelay(&APIError{Code: 429, RetryAfter: 2 * time.Minute}, time.Second, 30*time.Second); got != 30*time.Second {
		t.Fatalf("bounded retry_after delay = %v, want 30s", got)
	}
	if got := managerRetryDelay(errors.New("offline"), 4*time.Second, 30*time.Second); got != 4*time.Second {
		t.Fatalf("network retry delay = %v, want 4s", got)
	}
}

func TestPairingIsShortLivedSingleUseProfileScopedAndPrivate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC)
	store := NewMemoryPairingStore()
	service := NewPairingService(store)
	service.SetClock(func() time.Time { return now })
	service.SetCodeGenerator(func() (string, error) { return "CAB-ABCD-1234", nil })
	created, err := service.Create(context.Background(), "profile-1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Code != "CAB-ABCD-1234" || created.ProfileID != "profile-1" || !created.ExpiresAt.Equal(now.Add(5*time.Minute)) {
		t.Fatalf("unexpected pairing request: %+v", created)
	}
	if store.ContainsPlaintext("CAB-ABCD-1234") {
		t.Fatal("pairing store retained plaintext code")
	}

	group := telegrambotadapter.Update{UpdateID: 1, Message: telegramMessageWithType(1, 100, -1001, "group", "/start CAB-ABCD-1234")}
	if _, err := service.Consume(context.Background(), group); !errors.Is(err, ErrPrivateChatRequired) {
		t.Fatalf("group pairing error = %v", err)
	}
	differentSender := telegrambotadapter.Update{UpdateID: 2, Message: telegramMessageWithType(2, 101, 100, "private", "/start CAB-ABCD-1234")}
	if _, err := service.Consume(context.Background(), differentSender); !errors.Is(err, ErrPrivateChatRequired) {
		t.Fatalf("different sender pairing error = %v", err)
	}
	paired, err := service.Consume(context.Background(), telegrambotadapter.Update{UpdateID: 3, Message: telegramMessageWithType(3, 100, 100, "private", "/start CAB-ABCD-1234")})
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if paired.ProfileID != "profile-1" || paired.SenderID != "100" || paired.ChatID != "100" {
		t.Fatalf("unexpected pairing result: %+v", paired)
	}
	if _, err := service.Consume(context.Background(), telegrambotadapter.Update{UpdateID: 4, Message: telegramMessageWithType(4, 100, 100, "private", "/start CAB-ABCD-1234")}); !errors.Is(err, ErrPairingCodeUsed) {
		t.Fatalf("second consume error = %v", err)
	}

	service.SetCodeGenerator(func() (string, error) { return "CAB-EXPR-0001", nil })
	if _, err := service.Create(context.Background(), "profile-2", time.Minute); err != nil {
		t.Fatalf("Create(expiring) error = %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := service.Consume(context.Background(), telegrambotadapter.Update{UpdateID: 5, Message: telegramMessageWithType(5, 200, 200, "private", "/start CAB-EXPR-0001")}); !errors.Is(err, ErrPairingCodeExpired) {
		t.Fatalf("expired consume error = %v", err)
	}
}

func TestSQLStorePersistsOffsetAcrossRuntimeRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(ctx, path)
	if err != nil {
		t.Fatalf("OpenAndMigrate(first) error = %v", err)
	}
	created, err := profile.NewRepository(conn).Create(ctx, "Telegram restart")
	if err != nil {
		t.Fatalf("Create(profile) error = %v", err)
	}
	if err := NewSQLStore(conn).SaveOffset(ctx, created.ID, 2086); err != nil {
		t.Fatalf("SaveOffset() error = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("Close(first) error = %v", err)
	}

	restarted, err := db.OpenAndMigrate(ctx, path)
	if err != nil {
		t.Fatalf("OpenAndMigrate(restart) error = %v", err)
	}
	defer restarted.Close()
	offset, err := NewSQLStore(restarted).LoadOffset(ctx, created.ID)
	if err != nil {
		t.Fatalf("LoadOffset(restart) error = %v", err)
	}
	if offset != 2086 {
		t.Fatalf("restart offset = %d, want 2086", offset)
	}
}

func TestManagerPollsConfiguredProfileWhenAnotherProfileIsActive(t *testing.T) {
	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	defer conn.Close()
	profiles := profile.NewRepository(conn)
	active, err := profiles.Create(ctx, "Active without Telegram")
	if err != nil {
		t.Fatalf("Create(active) error = %v", err)
	}
	configured, err := profiles.Create(ctx, "Telegram profile")
	if err != nil {
		t.Fatalf("Create(configured) error = %v", err)
	}
	if err := profiles.SetActiveProfile(ctx, active.ID); err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}
	if err := profiles.PutSettings(ctx, configured.ID, map[string]string{"telegram.polling.enabled": "true"}); err != nil {
		t.Fatalf("PutSettings() error = %v", err)
	}
	manager := NewManager(conn, profiles, nil, "", nil)
	ids, err := manager.enabledProfileIDs(ctx)
	if err != nil {
		t.Fatalf("enabledProfileIDs() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != configured.ID {
		t.Fatalf("enabled profile IDs = %v, want [%s]", ids, configured.ID)
	}
}

func telegramMessage(messageID, senderID, chatID int64, text string) *telegrambotadapter.WebhookMessage {
	return telegramMessageWithType(messageID, senderID, chatID, "private", text)
}

func telegramMessageWithType(messageID, senderID, chatID int64, chatType, text string) *telegrambotadapter.WebhookMessage {
	return &telegrambotadapter.WebhookMessage{
		MessageID: messageID,
		From:      telegramcapture.WebhookUser{ID: senderID},
		Chat:      telegramcapture.WebhookChat{ID: chatID, Type: chatType},
		Text:      text,
	}
}

type scriptedBotAPI struct {
	mu      sync.Mutex
	updates [][]telegrambotadapter.Update
	errs    []error
	offsets []int64
	allowed [][]string
}

func (s *scriptedBotAPI) GetUpdates(_ context.Context, offset int64, _ int, allowed []string) ([]telegrambotadapter.Update, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offsets = append(s.offsets, offset)
	s.allowed = append(s.allowed, append([]string(nil), allowed...))
	if len(s.errs) > 0 {
		err := s.errs[0]
		s.errs = s.errs[1:]
		return nil, err
	}
	if len(s.updates) == 0 {
		return nil, nil
	}
	out := s.updates[0]
	s.updates = s.updates[1:]
	return out, nil
}

type memoryOffsetStore struct {
	mu      sync.Mutex
	offsets map[string]int64
}

func (s *memoryOffsetStore) LoadOffset(_ context.Context, profileID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Offset(profileID), nil
}

func (s *memoryOffsetStore) SaveOffset(_ context.Context, profileID string, offset int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.offsets == nil {
		s.offsets = map[string]int64{}
	}
	s.offsets[profileID] = offset
	return nil
}

func (s *memoryOffsetStore) Offset(profileID string) int64 {
	if s.offsets == nil {
		return 0
	}
	return s.offsets[profileID]
}
