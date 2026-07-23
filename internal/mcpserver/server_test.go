package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestInitializeAdvertisesCabinetIdentityAndProfileBinding(t *testing.T) {
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-test-session",
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-test-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	result := clientSession.InitializeResult()
	if result == nil {
		t.Fatal("InitializeResult() is nil")
	}
	if result.ServerInfo == nil || result.ServerInfo.Name != ServerName || result.ServerInfo.Title != ServerTitle || result.ServerInfo.Version != "0.1.0-test" {
		t.Fatalf("unexpected server info: %#v", result.ServerInfo)
	}
	if !strings.Contains(result.Instructions, "profile-main") {
		t.Fatalf("initialize instructions should include the explicit profile binding, got %q", result.Instructions)
	}
	extension := result.Capabilities.Extensions[extensionName].(map[string]any)
	if extension["profile_id"] != "profile-main" || extension["profile_label"] != "Main collection" || extension["version_digest"] != "git:abc123" {
		t.Fatalf("unexpected Cabinet profile extension: %#v", extension)
	}
	if result.Capabilities.Tools != nil || result.Capabilities.Resources != nil || result.Capabilities.Prompts != nil {
		t.Fatalf("foundation server should not advertise tools/resources/prompts yet: %#v", result.Capabilities)
	}
}

func TestNewServerRejectsMissingProfileBinding(t *testing.T) {
	if _, err := NewServer(Config{Version: "0.1.0-test"}); err == nil {
		t.Fatal("NewServer() should reject missing profile binding")
	}
}
