package mcpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPTransportDisabledByDefault(t *testing.T) {
	handler, err := NewHTTPHandler(nil, HTTPTransportConfig{})
	if err != nil {
		t.Fatalf("NewHTTPHandler() disabled error = %v", err)
	}
	if handler != nil {
		t.Fatalf("NewHTTPHandler() disabled handler = %T, want nil", handler)
	}
}

func TestHTTPTransportRejectsNonLoopbackBinding(t *testing.T) {
	server := mustTestServer(t)
	for _, addr := range []string{"0.0.0.0:17890", "192.168.1.10:17890", "[2001:db8::1]:17890"} {
		t.Run(addr, func(t *testing.T) {
			_, err := NewHTTPHandler(server, HTTPTransportConfig{
				Enabled:    true,
				ListenAddr: addr,
				Credential: "local-secret",
			})
			if err == nil || !strings.Contains(err.Error(), "loopback") {
				t.Fatalf("NewHTTPHandler() error = %v, want loopback rejection", err)
			}
		})
	}
}

func TestHTTPTransportRequiresCredentialWhenEnabled(t *testing.T) {
	_, err := NewHTTPHandler(mustTestServer(t), HTTPTransportConfig{
		Enabled:    true,
		ListenAddr: "127.0.0.1:17890",
	})
	if err == nil || !strings.Contains(err.Error(), "credential") {
		t.Fatalf("NewHTTPHandler() error = %v, want credential rejection", err)
	}
}

func TestHTTPTransportRequiresBearerCredential(t *testing.T) {
	handler, err := NewHTTPHandler(mustTestServer(t), HTTPTransportConfig{
		Enabled:    true,
		ListenAddr: "127.0.0.1:17890",
		Credential: "local-secret",
	})
	if err != nil {
		t.Fatalf("NewHTTPHandler() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17890/mcp", strings.NewReader(`{}`))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing credential status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if strings.Contains(rec.Body.String(), "local-secret") {
		t.Fatalf("unauthorized response leaked credential: %q", rec.Body.String())
	}
}

func TestEnsureHTTPTransportCredentialGeneratesSecretWithoutSettingsLeak(t *testing.T) {
	t.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	t.Setenv("CABINET_FORCE_SECURESTORE_FAIL", "1")

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := profile.NewRepository(conn)
	p, err := repo.Create(context.Background(), "Main")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	credential, err := EnsureHTTPTransportCredential(context.Background(), repo, p.ID)
	if err != nil {
		t.Fatalf("EnsureHTTPTransportCredential() error = %v", err)
	}
	if len(credential) < 32 {
		t.Fatalf("generated credential is too short: %q", credential)
	}

	stored, err := repo.GetSecret(context.Background(), p.ID, HTTPTransportCredentialSecretKey)
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if stored != credential {
		t.Fatalf("stored credential = %q, want generated credential", stored)
	}

	again, err := EnsureHTTPTransportCredential(context.Background(), repo, p.ID)
	if err != nil {
		t.Fatalf("EnsureHTTPTransportCredential() second call error = %v", err)
	}
	if again != credential {
		t.Fatalf("EnsureHTTPTransportCredential() should reuse existing credential, got %q want %q", again, credential)
	}

	var settingsRows int
	if err := conn.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND value LIKE '%' || ? || '%'`, p.ID, credential).Scan(&settingsRows); err != nil {
		t.Fatalf("query profile_settings leak check: %v", err)
	}
	if settingsRows != 0 {
		t.Fatalf("generated MCP credential leaked into ordinary settings rows: %d", settingsRows)
	}
}

func TestHTTPTransportStatusDoesNotExposeCredential(t *testing.T) {
	t.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	t.Setenv("CABINET_FORCE_SECURESTORE_FAIL", "1")

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := profile.NewRepository(conn)
	p, err := repo.Create(context.Background(), "Main")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	disabled := HTTPTransportStatus(context.Background(), repo, p.ID, HTTPTransportConfig{})
	if disabled.State != "disabled" || disabled.Enabled || disabled.CredentialConfigured {
		t.Fatalf("unexpected disabled status: %+v", disabled)
	}

	missing := HTTPTransportStatus(context.Background(), repo, p.ID, HTTPTransportConfig{
		Enabled:    true,
		ListenAddr: "127.0.0.1:17890",
	})
	if missing.State != "misconfigured" || missing.CredentialConfigured || missing.RecoveryAction != "generate_mcp_http_credential" {
		t.Fatalf("unexpected missing-credential status: %+v", missing)
	}

	credential, err := EnsureHTTPTransportCredential(context.Background(), repo, p.ID)
	if err != nil {
		t.Fatalf("EnsureHTTPTransportCredential() error = %v", err)
	}
	ready := HTTPTransportStatus(context.Background(), repo, p.ID, HTTPTransportConfig{
		Enabled:    true,
		ListenAddr: "127.0.0.1:17890",
	})
	if ready.State != "ready" || !ready.CredentialConfigured || ready.Credential != "" {
		t.Fatalf("unexpected ready status: %+v", ready)
	}
	if strings.Contains(ready.Guidance, credential) || strings.Contains(ready.RecoveryAction, credential) {
		t.Fatalf("status leaked credential: %+v", ready)
	}
}

func mustTestServer(t *testing.T) *mcp.Server {
	t.Helper()
	server, err := NewServer(Config{
		ProfileID: "profile-main",
		Version:   "0.1.0-test",
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return server
}
