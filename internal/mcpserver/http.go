package mcpserver

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type HTTPTransportConfig struct {
	Enabled    bool
	ListenAddr string
	Credential string
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
