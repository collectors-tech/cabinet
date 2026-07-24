package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
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

func TestRawProtocolUnknownMethodReturnsStructuredErrorAndKeepsSessionAlive(t *testing.T) {
	conn, closeConn := rawProtocolConnection(t)
	defer closeConn()

	writeRequest(t, conn, "unknown-1", "cabinet/unknown", `{}`)
	resp := readResponse(t, conn)
	if got := resp.ID.Raw(); got != "unknown-1" {
		t.Fatalf("unknown-method response ID = %v, want unknown-1", got)
	}
	assertStructuredProtocolErrorDoesNotLeakProfile(t, resp, "unknown-method")

	writeRequest(t, conn, "ping-1", "ping", `{}`)
	resp = readResponse(t, conn)
	if resp.Error != nil {
		t.Fatalf("ping after unknown method returned error: %v", resp.Error)
	}
	if got := strings.TrimSpace(string(resp.Result)); got != "{}" {
		t.Fatalf("ping result = %s, want {}", got)
	}
}

func TestRawProtocolInvalidNonPingMethodBeforeInitializeReturnsStructuredErrorThenInitializes(t *testing.T) {
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-invalid-init-test-session",
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
	conn, err := clientTransport.Connect(ctx)
	if err != nil {
		t.Fatalf("client transport Connect() error = %v", err)
	}
	defer conn.Close()

	writeRequest(t, conn, "early-tools", "tools/list", `{}`)
	resp := readResponse(t, conn)
	if got := resp.ID.Raw(); got != "early-tools" {
		t.Fatalf("pre-initialize invalid method response ID = %v, want early-tools", got)
	}
	assertStructuredProtocolErrorDoesNotLeakProfile(t, resp, "pre-initialize invalid method")

	writeRequest(t, conn, "init-1", "initialize", `{"clientInfo":{"name":"cabinet-raw-test","version":"0.1.0"},"protocolVersion":"2025-06-18","capabilities":{}}`)
	resp = readResponse(t, conn)
	if resp.Error != nil {
		t.Fatalf("initialize after invalid method returned error: %v", resp.Error)
	}
}

func TestToolCallTimeoutCancelsOnlyInFlightOperationAndKeepsSessionAlive(t *testing.T) {
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-timeout-test-session",
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	started := make(chan struct{})
	cancelled := make(chan error, 1)
	mcp.AddTool(server, &mcp.Tool{Name: "cabinet.test.hang"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		close(started)
		<-ctx.Done()
		cancelled <- ctx.Err()
		return nil, nil, ctx.Err()
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-timeout-test-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	callCtx, cancelCall := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelCall()
	_, err = clientSession.CallTool(callCtx, &mcp.CallToolParams{Name: "cabinet.test.hang"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("CallTool() error = %v, want context deadline exceeded", err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for test tool to start")
	}
	select {
	case err := <-cancelled:
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("server-side cancellation error = %v, want context cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for in-flight tool cancellation")
	}

	if err := clientSession.Ping(ctx, nil); err != nil {
		t.Fatalf("Ping() after timed-out tool call error = %v", err)
	}
}

func rawProtocolConnection(t *testing.T) (mcp.Connection, func()) {
	t.Helper()
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-raw-test-session",
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		cancel()
		t.Fatalf("server.Connect() error = %v", err)
	}
	conn, err := clientTransport.Connect(ctx)
	if err != nil {
		cancel()
		serverSession.Close()
		t.Fatalf("client transport Connect() error = %v", err)
	}

	writeRequest(t, conn, "init-1", "initialize", `{"clientInfo":{"name":"cabinet-raw-test","version":"0.1.0"},"protocolVersion":"2025-06-18","capabilities":{}}`)
	if resp := readResponse(t, conn); resp.Error != nil {
		t.Fatalf("initialize returned error: %v", resp.Error)
	}
	writeNotification(t, conn, "notifications/initialized", `{}`)

	cleanup := func() {
		conn.Close()
		serverSession.Close()
		cancel()
	}
	return conn, cleanup
}

func assertStructuredProtocolErrorDoesNotLeakProfile(t *testing.T, resp *jsonrpc.Response, label string) {
	t.Helper()
	var rpcErr *jsonrpc.Error
	if !errors.As(resp.Error, &rpcErr) {
		t.Fatalf("%s response should contain JSON-RPC error, got %T %[1]v", label, resp.Error)
	}
	if strings.TrimSpace(rpcErr.Message) == "" {
		t.Fatalf("%s JSON-RPC error should include a message: %#v", label, rpcErr)
	}
	if strings.Contains(strings.ToLower(rpcErr.Message), "profile-main") {
		t.Fatalf("%s leaked profile details: %q", label, rpcErr.Message)
	}
}

func writeRequest(t *testing.T, conn mcp.Connection, idValue string, method string, params string) {
	t.Helper()
	id, err := jsonrpc.MakeID(idValue)
	if err != nil {
		t.Fatalf("MakeID(%q) error = %v", idValue, err)
	}
	if err := conn.Write(context.Background(), &jsonrpc.Request{
		ID:     id,
		Method: method,
		Params: json.RawMessage(params),
	}); err != nil {
		t.Fatalf("write %s request error = %v", method, err)
	}
}

func writeNotification(t *testing.T, conn mcp.Connection, method string, params string) {
	t.Helper()
	if err := conn.Write(context.Background(), &jsonrpc.Request{
		Method: method,
		Params: json.RawMessage(params),
	}); err != nil {
		t.Fatalf("write %s notification error = %v", method, err)
	}
}

func readResponse(t *testing.T, conn mcp.Connection) *jsonrpc.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	resp, ok := msg.(*jsonrpc.Response)
	if !ok {
		t.Fatalf("Read() returned %T, want *jsonrpc.Response", msg)
	}
	return resp
}
