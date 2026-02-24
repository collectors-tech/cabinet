package app

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestLogsEndpointsAndDebugToggle(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(create.Body).Decode(&p)
	debug := doRequest(t, a, http.MethodPost, "/api/logs/debug", strings.NewReader(`{"profile_id":"`+p.ID+`","enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if debug.Code != http.StatusOK {
		t.Fatalf("debug toggle status=%d body=%s", debug.Code, debug.Body.String())
	}
	activity := doRequest(t, a, http.MethodGet, "/api/logs/activity?limit=10", nil, nil)
	if activity.Code != http.StatusOK {
		t.Fatalf("activity logs status=%d body=%s", activity.Code, activity.Body.String())
	}
	export := doRequest(t, a, http.MethodGet, "/api/logs/export", nil, nil)
	if export.Code != http.StatusOK {
		t.Fatalf("export logs status=%d body=%s", export.Code, export.Body.String())
	}
}

func TestRuntimeRecoveryPromptAfterUncleanShutdown(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	conn, err := db.OpenAndMigrate(context.Background(), cfg.DBPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	_, _ = conn.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('clean_shutdown','0',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='0', updated_at=CURRENT_TIMESTAMP`)
	_ = conn.Close()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = a.close() })
	resp := doRequest(t, a, http.MethodGet, "/api/runtime/recovery", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("runtime recovery status=%d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"recovery_required":true`) {
		t.Fatalf("expected recovery_required true, got %s", resp.Body.String())
	}
}

func TestRuntimeEndpointIncludesBuildMetadata(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)

	resp := doRequest(t, a, http.MethodGet, "/api/runtime", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("runtime status=%d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"app_version"`) {
		t.Fatalf("expected app_version in runtime payload, got %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"build_date"`) {
		t.Fatalf("expected build_date in runtime payload, got %s", resp.Body.String())
	}
}
