package mcpserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const HTTPTransportCredentialSecretKey = "mcp_http_transport_token"

type HTTPTransportConfig struct {
	Enabled        bool
	ListenAddr     string
	Credential     string
	LastDiagnostic *DiagnosticOutcome
}

type HTTPTransportCredentialStore interface {
	GetSecret(ctx context.Context, profileID, key string) (string, error)
	PutSecret(ctx context.Context, profileID, key, value string) error
}

type HTTPTransportStatusReport struct {
	Enabled               bool               `json:"enabled"`
	State                 string             `json:"state"`
	ListenAddr            string             `json:"listen_addr,omitempty"`
	CredentialConfigured  bool               `json:"credential_configured"`
	Credential            string             `json:"-"`
	Guidance              string             `json:"guidance,omitempty"`
	RecoveryAction        string             `json:"recovery_action,omitempty"`
	LastDiagnosticOutcome *DiagnosticOutcome `json:"last_diagnostic_outcome,omitempty"`
}

func EnsureHTTPTransportCredential(ctx context.Context, store HTTPTransportCredentialStore, profileID string) (string, error) {
	if store == nil {
		return "", errors.New("mcp HTTP transport credential store is required")
	}
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return "", errors.New("mcp HTTP transport profile binding is required")
	}
	if existing, err := store.GetSecret(ctx, profileID, HTTPTransportCredentialSecretKey); err == nil && strings.TrimSpace(existing) != "" {
		return existing, nil
	}
	credential, err := generateHTTPTransportCredential()
	if err != nil {
		return "", err
	}
	if err := store.PutSecret(ctx, profileID, HTTPTransportCredentialSecretKey, credential); err != nil {
		return "", fmt.Errorf("store mcp HTTP transport credential: %w", err)
	}
	return credential, nil
}

func HTTPTransportStatus(ctx context.Context, store HTTPTransportCredentialStore, profileID string, cfg HTTPTransportConfig) HTTPTransportStatusReport {
	report := HTTPTransportStatusReport{
		Enabled:               cfg.Enabled,
		ListenAddr:            strings.TrimSpace(cfg.ListenAddr),
		LastDiagnosticOutcome: redactedDiagnosticOutcome(cfg.LastDiagnostic),
	}
	if !cfg.Enabled {
		report.State = "disabled"
		report.Guidance = "Enable the local MCP HTTP transport before configuring loopback clients."
		report.RecoveryAction = "enable_mcp_http_transport"
		return report
	}
	if err := validateLoopbackListenAddr(cfg.ListenAddr); err != nil {
		report.State = "misconfigured"
		report.Guidance = "Use a loopback listen address for the local MCP HTTP transport."
		report.RecoveryAction = "set_loopback_listen_address"
		return report
	}
	credential := strings.TrimSpace(cfg.Credential)
	if credential == "" && store != nil && strings.TrimSpace(profileID) != "" {
		if stored, err := store.GetSecret(ctx, strings.TrimSpace(profileID), HTTPTransportCredentialSecretKey); err == nil {
			credential = strings.TrimSpace(stored)
		}
	}
	if credential == "" {
		report.State = "misconfigured"
		report.Guidance = "Generate a local MCP HTTP credential before accepting client sessions."
		report.RecoveryAction = "generate_mcp_http_credential"
		return report
	}
	report.State = "ready"
	report.CredentialConfigured = true
	report.Guidance = "Local MCP HTTP transport is configured for loopback clients."
	return report
}

func redactedDiagnosticOutcome(outcome *DiagnosticOutcome) *DiagnosticOutcome {
	if outcome == nil {
		return nil
	}
	redacted := &DiagnosticOutcome{
		OperationID: strings.TrimSpace(outcome.OperationID),
		Capability:  strings.TrimSpace(outcome.Capability),
		Method:      strings.TrimSpace(outcome.Method),
		InputClass:  strings.TrimSpace(outcome.InputClass),
		Outcome:     strings.TrimSpace(outcome.Outcome),
		ErrorClass:  strings.TrimSpace(outcome.ErrorClass),
	}
	if redacted.OperationID == "" && redacted.Capability == "" && redacted.Method == "" && redacted.InputClass == "" && redacted.Outcome == "" && redacted.ErrorClass == "" {
		return nil
	}
	return redacted
}

func NewHTTPHandler(server *mcp.Server, cfg HTTPTransportConfig) (http.Handler, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if server == nil {
		return nil, errors.New("mcp HTTP transport requires a server")
	}
	if err := validateLoopbackListenAddr(cfg.ListenAddr); err != nil {
		return nil, err
	}
	credential := strings.TrimSpace(cfg.Credential)
	if credential == "" {
		return nil, errors.New("mcp HTTP transport credential is required")
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})
	return requireBearerCredential(handler, credential), nil
}

func generateHTTPTransportCredential() (string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", fmt.Errorf("generate mcp HTTP transport credential: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func validateLoopbackListenAddr(addr string) error {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return fmt.Errorf("invalid mcp HTTP listen address: %w", err)
	}
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("mcp HTTP transport must bind to a loopback address, got %q", host)
	}
	return nil
}

func requireBearerCredential(next http.Handler, credential string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer "+credential {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})
}
