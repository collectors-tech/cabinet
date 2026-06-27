package app

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/update"
)

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
