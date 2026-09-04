package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestResetE2EDatabaseSerializesConcurrentCalls(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	a.db.SetMaxOpenConns(8)
	a.db.SetMaxIdleConns(8)

	const callers = 12
	start := make(chan struct{})
	errs := make(chan error, callers)
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			ready.Done()
			<-start
			errs <- resetE2EDatabase(context.Background(), a.db)
		}()
	}
	ready.Wait()
	close(start)

	for range callers {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent reset failed: %v", err)
		}
	}

	connections := make([]*sql.Conn, 0, 8)
	defer func() {
		for _, conn := range connections {
			_ = conn.Close()
		}
	}()
	for range 8 {
		conn, err := a.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("acquire pooled connection: %v", err)
		}
		connections = append(connections, conn)
		var foreignKeys int
		if err := conn.QueryRowContext(context.Background(), `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
			t.Fatalf("read pooled foreign_keys state: %v", err)
		}
		if foreignKeys != 1 {
			t.Fatalf("pooled connection returned with foreign_keys=%d", foreignKeys)
		}
	}
}

func TestResetE2EDatabaseRecoversConnectionAfterBusyCommit(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	a.db.SetMaxOpenConns(2)
	a.db.SetMaxIdleConns(2)

	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	reader, err := a.db.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire reader connection: %v", err)
	}
	defer reader.Close()
	if _, err := reader.ExecContext(context.Background(), `BEGIN`); err != nil {
		t.Fatalf("begin reader transaction: %v", err)
	}
	defer func() { _, _ = reader.ExecContext(context.Background(), `ROLLBACK`) }()
	var itemCount int
	if err := reader.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM canonical_items`).Scan(&itemCount); err != nil {
		t.Fatalf("hold reader transaction: %v", err)
	}
	if itemCount == 0 {
		t.Fatal("expected bootstrap data before reset contention")
	}

	writer, err := a.db.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire writer connection: %v", err)
	}
	if _, err := writer.ExecContext(context.Background(), `PRAGMA busy_timeout = 10`); err != nil {
		writer.Close()
		t.Fatalf("shorten writer busy timeout: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("release writer connection: %v", err)
	}

	firstResetErr := resetE2EDatabase(context.Background(), a.db)
	if firstResetErr == nil {
		t.Fatal("expected held reader transaction to block the first reset commit")
	}
	class, operation := classifyE2EResetFailure(firstResetErr)
	t.Logf("first reset failed as class=%s operation=%s", class, operation)

	if _, err := reader.ExecContext(context.Background(), `ROLLBACK`); err != nil {
		t.Fatalf("release reader transaction: %v", err)
	}
	if err := resetE2EDatabase(context.Background(), a.db); err != nil {
		class, operation := classifyE2EResetFailure(err)
		t.Fatalf("reset did not recover after contention cleared: class=%s operation=%s err=%v", class, operation, err)
	}
}

func TestResetE2EDatabaseToleratesActiveReaderAfterPreflight(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	a.db.SetMaxOpenConns(2)
	a.db.SetMaxIdleConns(2)

	if err := resetE2EDatabase(context.Background(), a.db); err != nil {
		t.Fatalf("preflight reset failed: %v", err)
	}
	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	reader, err := a.db.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire reader connection: %v", err)
	}
	defer reader.Close()
	if _, err := reader.ExecContext(context.Background(), `BEGIN`); err != nil {
		t.Fatalf("begin reader transaction: %v", err)
	}
	defer func() { _, _ = reader.ExecContext(context.Background(), `ROLLBACK`) }()
	var itemCount int
	if err := reader.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM canonical_items`).Scan(&itemCount); err != nil {
		t.Fatalf("hold reader transaction: %v", err)
	}
	if itemCount == 0 {
		t.Fatal("expected bootstrap data before reset")
	}

	if err := resetE2EDatabase(context.Background(), a.db); err != nil {
		class, operation := classifyE2EResetFailure(err)
		t.Fatalf("reset failed with active reader: class=%s operation=%s err=%v", class, operation, err)
	}
}

func TestE2EResetEndpointEvictsForeignIdleWriterBeforeBegin(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	a.db.SetMaxOpenConns(2)
	a.db.SetMaxIdleConns(2)

	if err := resetE2EDatabase(context.Background(), a.db); err != nil {
		t.Fatalf("preflight reset failed: %v", err)
	}
	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	writer, err := a.db.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire foreign writer connection: %v", err)
	}
	clean, err := a.db.Conn(context.Background())
	if err != nil {
		writer.Close()
		t.Fatalf("acquire clean reset connection: %v", err)
	}
	if _, err := writer.ExecContext(context.Background(), `BEGIN IMMEDIATE`); err != nil {
		clean.Close()
		writer.Close()
		t.Fatalf("begin foreign writer transaction: %v", err)
	}
	if err := writer.Close(); err != nil {
		clean.Close()
		t.Fatalf("return foreign writer to idle pool: %v", err)
	}
	if err := clean.Close(); err != nil {
		t.Fatalf("return clean connection to idle pool: %v", err)
	}

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		resetErr := resetE2EDatabase(context.Background(), a.db)
		class, operation := classifyE2EResetFailure(resetErr)
		t.Fatalf(
			"reset status=%d body=%s diagnostic=%s/%s err=%v",
			reset.Code,
			reset.Body.String(),
			class,
			operation,
			resetErr,
		)
	}
}

func TestE2EResetDiagnosticRedactsUnderlyingStorageError(t *testing.T) {
	t.Parallel()

	underlying := errors.New("private SQL and row data must not reach runtime logs")
	failure := newE2EResetFailure("begin_transaction", underlying)
	diagnostic := e2eResetDiagnostic(failure)

	if diagnostic != "e2e reset failed: class=unexpected_storage operation=begin_transaction" {
		t.Fatalf("unexpected safe diagnostic: %q", diagnostic)
	}
	if strings.Contains(diagnostic, underlying.Error()) {
		t.Fatalf("diagnostic leaked underlying storage error: %q", diagnostic)
	}
}

func TestRunE2EResetWithRetryIsBoundedToStorageContention(t *testing.T) {
	t.Parallel()

	contentionAttempts := 0
	err := runE2EResetWithRetry(context.Background(), func() error {
		contentionAttempts++
		if contentionAttempts < e2eResetMaxAttempts {
			return &e2eResetFailure{
				class:     "storage_contention",
				operation: "begin_transaction",
				err:       errors.New("database contention"),
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("bounded contention retry failed: %v", err)
	}
	if contentionAttempts != e2eResetMaxAttempts {
		t.Fatalf("contention attempts=%d want=%d", contentionAttempts, e2eResetMaxAttempts)
	}

	unexpectedAttempts := 0
	err = runE2EResetWithRetry(context.Background(), func() error {
		unexpectedAttempts++
		return newE2EResetFailure("begin_transaction", errors.New("non-retryable storage failure"))
	})
	if err == nil {
		t.Fatal("expected non-retryable storage failure")
	}
	if unexpectedAttempts != 1 {
		t.Fatalf("unexpected storage failure attempts=%d want=1", unexpectedAttempts)
	}
}

func TestE2EAgentResponseStateFixturePersistsNormalizedLatestMessage(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}
	for _, state := range []string{
		"read_result", "clarification_required", "setup_required", "authority_blocked", "unsupported",
		"provider_unavailable", "retryable_failure", "preview_required", "preview_expired", "preview_failed",
		"preview_stale_target", "cancelled", "applied",
	} {
		payload := `{"profile_id":"e2e-profile-001","thread_id":"e2e-thread-001","state":"` + state + `","original_intent":"bounded ` + state + ` intent"}`
		seed := doRequest(t, a, http.MethodPost, "/api/test/chat/agent-response-state", strings.NewReader(payload), map[string]string{"Content-Type": "application/json"})
		if seed.Code != http.StatusOK {
			t.Fatalf("seed %s status=%d body=%s", state, seed.Code, seed.Body.String())
		}
	}

	messagesResponse := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id=e2e-profile-001&thread_id=e2e-thread-001", nil, nil)
	if messagesResponse.Code != http.StatusOK {
		t.Fatalf("messages status=%d body=%s", messagesResponse.Code, messagesResponse.Body.String())
	}
	var payload struct {
		Messages []struct {
			Context map[string]any `json:"context"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(messagesResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	if len(payload.Messages) != 15 {
		t.Fatalf("expected 2 bootstrap + 13 matrix messages, got %d", len(payload.Messages))
	}
	latest := payload.Messages[len(payload.Messages)-1].Context["agent_response"].(map[string]any)
	if latest["state"] != "applied" || latest["outcome"] != "applied" {
		t.Fatalf("latest response is not deterministic applied state: %+v", latest)
	}
	if skill := latest["skill"].(map[string]any); skill["name"] != "Cabinet Inventory Search" {
		t.Fatalf("latest response lost governed skill: %+v", skill)
	}
	if source := latest["source"].(map[string]any); source["surface"] != "chats.main" || source["channel"] != "in-app" {
		t.Fatalf("latest response lost source bounds: %+v", source)
	}

	ordinary := doRequest(t, a, http.MethodPost, "/api/test/chat/agent-response-state", strings.NewReader(`{"profile_id":"e2e-profile-001","thread_id":"e2e-thread-001","state":"ordinary_response","original_intent":"ordinary"}`), map[string]string{"Content-Type": "application/json"})
	if ordinary.Code != http.StatusOK {
		t.Fatalf("ordinary status=%d body=%s", ordinary.Code, ordinary.Body.String())
	}
	messagesResponse = doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id=e2e-profile-001&thread_id=e2e-thread-001", nil, nil)
	if err := json.Unmarshal(messagesResponse.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode messages after ordinary response: %v", err)
	}
	if latestContext := payload.Messages[len(payload.Messages)-1].Context; latestContext["agent_response"] != nil {
		t.Fatalf("ordinary assistant response retained stale agent state: %+v", latestContext)
	}
}

func TestResetE2EDatabaseSkipsMissingLegacyTables(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	seedLegacyResetState(t, a)

	if err := resetE2EDatabase(context.Background(), a.db); err != nil {
		t.Fatalf("reset with missing legacy table: %v", err)
	}
	assertLegacyResetCleared(t, a)
}

func TestE2EResetEndpointSkipsMissingLegacyTables(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	seedLegacyResetState(t, a)

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	assertLegacyResetCleared(t, a)
}

func TestE2EResetEndpointClearsWishlistPurchaseDeliveryState(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)
	seedWishlistPurchaseDeliveryResetState(t, a)

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	assertTableEmpty(t, a, "expected_arrivals")
	assertTableEmpty(t, a, "commerce_lifecycle_entries")
	assertTableEmpty(t, a, "wishlist_entries")
	assertTableEmpty(t, a, "instances")
	assertTableEmpty(t, a, "canonical_items")
}

func TestE2EResetEndpointClearsProfileSwitcherSeedState(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("initial reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	primary := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Primary DB"}`), map[string]string{"Content-Type": "application/json"})
	if primary.Code != http.StatusCreated {
		t.Fatalf("primary profile status=%d body=%s", primary.Code, primary.Body.String())
	}
	showcase := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Showcase DB"}`), map[string]string{"Content-Type": "application/json"})
	if showcase.Code != http.StatusCreated {
		t.Fatalf("showcase profile status=%d body=%s", showcase.Code, showcase.Body.String())
	}
	var primaryPayload struct {
		ID string `json:"id"`
	}
	var showcasePayload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(primary.Body.Bytes(), &primaryPayload); err != nil {
		t.Fatalf("decode primary profile: %v", err)
	}
	if err := json.Unmarshal(showcase.Body.Bytes(), &showcasePayload); err != nil {
		t.Fatalf("decode showcase profile: %v", err)
	}

	activatePrimary := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+primaryPayload.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activatePrimary.Code != http.StatusOK {
		t.Fatalf("activate primary status=%d body=%s", activatePrimary.Code, activatePrimary.Body.String())
	}
	primaryItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"PRI-001","title":"Primary Item","brand":"AFX","category":"Cars"}`), map[string]string{"Content-Type": "application/json"})
	if primaryItem.Code != http.StatusCreated {
		t.Fatalf("primary item status=%d body=%s", primaryItem.Code, primaryItem.Body.String())
	}
	activateShowcase := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+showcasePayload.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activateShowcase.Code != http.StatusOK {
		t.Fatalf("activate showcase status=%d body=%s", activateShowcase.Code, activateShowcase.Body.String())
	}
	showcaseItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"SHW-001","title":"Showcase Item","brand":"AFX","category":"Cars"}`), map[string]string{"Content-Type": "application/json"})
	if showcaseItem.Code != http.StatusCreated {
		t.Fatalf("showcase item status=%d body=%s", showcaseItem.Code, showcaseItem.Body.String())
	}

	reset = doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	assertTableEmpty(t, a, "canonical_items")
	assertTableEmpty(t, a, "profiles")
}

func TestE2EResetEndpointClearsCollectionWorkspaceSnapshotState(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("initial reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}
	var bootstrapPayload e2eBootstrapResponse
	if err := json.Unmarshal(bootstrap.Body.Bytes(), &bootstrapPayload); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+bootstrapPayload.ProfileID+"/settings", strings.NewReader(`{"settings":{"inventory.folder-tree.v2":"[{\"id\":\"quick-create-shelf\",\"name\":\"Quick Create Shelf\",\"children\":[]}]"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	reset = doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	assertTableEmpty(t, a, "profile_settings")
	assertTableEmpty(t, a, "profiles")
}

func TestE2EBootstrapEndpointSeedsWishlistFixtures(t *testing.T) {
	t.Parallel()

	a := newE2ETestApp(t)

	reset := doRequest(t, a, http.MethodPost, "/api/test/reset", nil, nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	setupStatus := doRequest(t, a, http.MethodPost, "/api/test/runtime/setup-status", strings.NewReader(`{"state":"present"}`), map[string]string{"Content-Type": "application/json"})
	if setupStatus.Code != http.StatusOK {
		t.Fatalf("setup status=%d body=%s", setupStatus.Code, setupStatus.Body.String())
	}
	bootstrap := doRequest(t, a, http.MethodPost, "/api/test/bootstrap", nil, nil)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	var wishlistCount int
	if err := a.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM wishlist_entries WHERE profile_id = 'e2e-profile-001'`).Scan(&wishlistCount); err != nil {
		t.Fatalf("count wishlist fixtures: %v", err)
	}
	if wishlistCount != 3 {
		t.Fatalf("expected 3 seeded wishlist entries, got %d", wishlistCount)
	}
}

func seedLegacyResetState(t *testing.T, a *App) {
	t.Helper()
	ctx := context.Background()
	if _, err := a.db.ExecContext(ctx, `INSERT INTO activity_logs(id, level, action, details) VALUES ('legacy-reset-log', 'info', 'seed', '{}')`); err != nil {
		t.Fatalf("seed activity log: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `DROP TABLE pokemon_graded_overrides`); err != nil {
		t.Fatalf("drop legacy-missing table: %v", err)
	}
}

func seedWishlistPurchaseDeliveryResetState(t *testing.T, a *App) {
	t.Helper()
	ctx := context.Background()
	if _, err := a.db.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES ('reset-profile', 'Reset Profile')`); err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, part_number, title, status, category)
		VALUES ('reset-item', 'reset-profile', 'Reset Brand', 'RESET-001', 'Reset Item', 'wishlist', 'Trading Cards')
	`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, owned, delivered)
		VALUES ('reset-wish', 'reset-profile', 'reset-item', 1, 1)
	`); err != nil {
		t.Fatalf("seed wishlist: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO commerce_lifecycle_entries(id, profile_id, item_id, state, source, external_ref, quantity)
		VALUES ('reset-life', 'reset-profile', 'reset-item', 'purchase', 'wishlist', 'reset-wish', 1)
	`); err != nil {
		t.Fatalf("seed lifecycle: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO instances(id, item_id, status, quantity)
		VALUES ('reset-instance', 'reset-item', 'sealed', 1)
	`); err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO expected_arrivals(id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, status, reconciled_instance_id)
		VALUES ('reset-arrival', 'reset-profile', 'reset-item', 'reset-life', 'wishlist', 'reset-wish', 1, 'delivered', 'reset-instance')
	`); err != nil {
		t.Fatalf("seed arrival: %v", err)
	}
}

func assertLegacyResetCleared(t *testing.T, a *App) {
	t.Helper()
	assertTableEmpty(t, a, "activity_logs")
}

func assertTableEmpty(t *testing.T, a *App, table string) {
	t.Helper()
	var remaining int
	if err := a.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM `+table).Scan(&remaining); err != nil {
		t.Fatalf("count %s after reset: %v", table, err)
	}
	if remaining != 0 {
		t.Fatalf("expected reset to clear %s, got %d rows", table, remaining)
	}
}

func newE2ETestApp(t *testing.T) *App {
	t.Helper()
	base := t.TempDir()
	return newTestAppWithConfig(t, config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
		EnableE2EHooks: true,
	})
}
