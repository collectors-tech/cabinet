package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
