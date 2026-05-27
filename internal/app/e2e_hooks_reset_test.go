package app

import (
	"context"
	"net/http"
	"path/filepath"
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

func assertLegacyResetCleared(t *testing.T, a *App) {
	t.Helper()
	var remaining int
	if err := a.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM activity_logs`).Scan(&remaining); err != nil {
		t.Fatalf("count activity logs after reset: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected reset to clear present tables, got %d activity logs", remaining)
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
