package mcpserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/agentskills"
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

func TestRawProtocolInvalidInitializeParamsReturnsStructuredErrorThenInitializes(t *testing.T) {
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-invalid-params-test-session",
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

	writeRequest(t, conn, "bad-init", "initialize", `{"clientInfo":"not-an-object","protocolVersion":18,"capabilities":[]}`)
	resp := readResponse(t, conn)
	if got := resp.ID.Raw(); got != "bad-init" {
		t.Fatalf("invalid-params response ID = %v, want bad-init", got)
	}
	assertStructuredProtocolErrorDoesNotLeakProfile(t, resp, "invalid initialize params")

	writeRequest(t, conn, "init-1", "initialize", `{"clientInfo":{"name":"cabinet-raw-test","version":"0.1.0"},"protocolVersion":"2025-06-18","capabilities":{}}`)
	resp = readResponse(t, conn)
	if resp.Error != nil {
		t.Fatalf("initialize after invalid params returned error: %v", resp.Error)
	}
}

func TestRawIOProtocolMalformedInitializeMessageReturnsStructuredErrorThenInitializes(t *testing.T) {
	writer, reader, closeConn := rawIOProtocolConnection(t)
	defer closeConn()

	writeRawLine(t, writer, `{"jsonrpc":"2.0","id":"malformed-1","method":"initialize","params":{"clientInfo":{"name":["not","a","string"]},"protocolVersion":false,"capabilities":"not-an-object"}}`)
	resp := readRawLineResponse(t, reader)
	if got := resp.ID.Raw(); got != "malformed-1" {
		t.Fatalf("malformed-message response ID = %v, want malformed-1", got)
	}
	assertStructuredProtocolErrorDoesNotLeakProfile(t, resp, "malformed initialize message")

	writeRawLine(t, writer, `{"jsonrpc":"2.0","id":"init-1","method":"initialize","params":{"clientInfo":{"name":"cabinet-raw-io-test","version":"0.1.0"},"protocolVersion":"2025-06-18","capabilities":{}}}`)
	resp = readRawLineResponse(t, reader)
	if resp.Error != nil {
		t.Fatalf("initialize after malformed message returned error: %v", resp.Error)
	}

	writeRawLine(t, writer, `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`)
	writeRawLine(t, writer, `{"jsonrpc":"2.0","id":"ping-1","method":"ping","params":{}}`)
	resp = readRawLineResponse(t, reader)
	if resp.Error != nil {
		t.Fatalf("ping after malformed message returned error: %v", resp.Error)
	}
	if got := strings.TrimSpace(string(resp.Result)); got != "{}" {
		t.Fatalf("ping result = %s, want {}", got)
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

func TestMCPToolCallAuthorityBlocksReadOnlyMutationBeforeHandler(t *testing.T) {
	var reviewed agentskills.PreviewRequest
	called := false
	server, err := NewServer(Config{
		ProfileID:     "profile-read-only",
		ProfileLabel:  "Read only",
		Version:       "0.1.0-test",
		VersionDigest: "git:authority",
		SessionIDSeed: "mcp-authority-test-session",
		AuthorityReviewer: AuthorityReviewerFunc(func(_ context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
			reviewed = req
			return agentskills.AgentAuthorityReview{
				ProfileID:            req.ProfileID,
				EntryPoint:           "mcp",
				SkillID:              req.SkillID,
				Decision:             "blocked",
				Blocker:              "agent_authority_read_only",
				ConfirmationRequired: true,
			}, nil
		}),
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cabinet.inventory.create_item"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		called = true
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "mutated"}}}, nil, nil
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-authority-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.inventory.create_item",
		Arguments: map[string]any{
			"title":       "Blocked MCP Item",
			"part_number": "MCP-AUTH-1",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "agent_authority_read_only") {
		t.Fatalf("expected read-only authority error, got %v", err)
	}
	if called {
		t.Fatal("MCP tool handler ran despite read-only authority blocker")
	}
	if reviewed.SkillID != "cabinet.inventory.create_item" ||
		reviewed.ProfileID != "profile-read-only" ||
		reviewed.SourceChannel != "mcp" ||
		reviewed.SourceSurface != "mcp.tools.call" {
		t.Fatalf("unexpected authority request: %+v", reviewed)
	}
	if reviewed.Parameters["title"] != "Blocked MCP Item" || reviewed.Parameters["part_number"] != "MCP-AUTH-1" {
		t.Fatalf("expected tool arguments to reach authority review, got %+v", reviewed.Parameters)
	}
}

func TestServerRecordsRedactedDiagnosticReceiptsForMaterialOperations(t *testing.T) {
	var receipts []OperationReceipt
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-receipt-test-session",
		ReceiptSink: ReceiptSinkFunc(func(_ context.Context, receipt OperationReceipt) {
			receipts = append(receipts, receipt)
		}),
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cabinet.test.receipt"}, func(ctx context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, nil, nil
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-receipt-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.test.receipt",
		Arguments: map[string]any{
			"query":   "Nintendo",
			"token":   "mcp-super-secret",
			"api_key": "provider-secret",
		},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if len(receipts) < 2 {
		t.Fatalf("expected initialize and tool receipts, got %+v", receipts)
	}
	initReceipt := receipts[0]
	if initReceipt.Method != "initialize" || initReceipt.Capability != "session.initialize" || initReceipt.Outcome != "ok" {
		t.Fatalf("unexpected initialize receipt: %+v", initReceipt)
	}
	if initReceipt.ProfileID != "profile-main" || initReceipt.ProfileLabel != "Main collection" || initReceipt.ClientName != "cabinet-receipt-client" || initReceipt.VersionDigest != "git:abc123" {
		t.Fatalf("initialize receipt missing trust-boundary fields: %+v", initReceipt)
	}

	toolReceipt := receipts[len(receipts)-1]
	if toolReceipt.Method != "tools/call" || toolReceipt.Capability != "tool:cabinet.test.receipt" || toolReceipt.InputClass != "tool_arguments" || toolReceipt.Outcome != "ok" {
		t.Fatalf("unexpected tool receipt: %+v", toolReceipt)
	}
	body, err := json.Marshal(receipts)
	if err != nil {
		t.Fatalf("marshal receipts: %v", err)
	}
	for _, forbidden := range []string{"mcp-super-secret", "provider-secret", "Nintendo"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("receipt leaked sensitive input %q: %s", forbidden, body)
		}
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

func rawIOProtocolConnection(t *testing.T) (*io.PipeWriter, *bufio.Reader, func()) {
	t.Helper()
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:abc123",
		SessionIDSeed: "mcp-raw-io-test-session",
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	clientToServerReader, clientToServerWriter := io.Pipe()
	serverToClientReader, serverToClientWriter := io.Pipe()
	transport := &mcp.IOTransport{
		Reader: clientToServerReader,
		Writer: serverToClientWriter,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	serverSession, err := server.Connect(ctx, transport, nil)
	if err != nil {
		cancel()
		t.Fatalf("server.Connect() error = %v", err)
	}

	cleanup := func() {
		clientToServerWriter.Close()
		serverToClientReader.Close()
		serverSession.Close()
		cancel()
	}
	return clientToServerWriter, bufio.NewReader(serverToClientReader), cleanup
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

func writeRawLine(t *testing.T, writer *io.PipeWriter, raw string) {
	t.Helper()
	if _, err := writer.Write([]byte(raw + "\n")); err != nil {
		t.Fatalf("write raw protocol line error = %v", err)
	}
}

func readRawLineResponse(t *testing.T, reader *bufio.Reader) *jsonrpc.Response {
	t.Helper()
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read raw protocol response error = %v", err)
	}
	msg, err := jsonrpc.DecodeMessage(bytes.TrimSpace(line))
	if err != nil {
		t.Fatalf("decode raw protocol response %q error = %v", line, err)
	}
	resp, ok := msg.(*jsonrpc.Response)
	if !ok {
		t.Fatalf("raw protocol response decoded to %T, want *jsonrpc.Response", msg)
	}
	return resp
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
