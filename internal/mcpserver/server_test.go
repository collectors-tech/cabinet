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
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolsListDerivesFromEnabledExecutableAgentSkillRegistry(t *testing.T) {
	registry := agentskills.NewProfileRegistry("profile-main", []agentskills.Skill{
		{
			ID:              "local.archive.enabled_reader",
			Version:         "0.1.0",
			DisplayName:     "Enabled archive reader",
			Description:     "Read archived Cabinet metadata.",
			Category:        "testing",
			Status:          agentskills.StatusAvailable,
			SafetyLevel:     agentskills.SafetyReadOnly,
			RequiredContext: []string{"profile"},
			InputSchemaRefs: []string{"archive_query"},
			Enabled:         true,
		},
		{
			ID:          "local.archive.disabled_writer",
			Version:     "0.1.0",
			DisplayName: "Disabled archive writer",
			Description: "Write archived Cabinet metadata.",
			Category:    "testing",
			Status:      agentskills.StatusAvailable,
			SafetyLevel: agentskills.SafetyConfirmRequired,
			Enabled:     true,
		},
		{
			ID:          "local.archive.invalid_reader",
			Version:     "0.1.0",
			DisplayName: "Invalid archive reader",
			Description: "Invalid imported Cabinet metadata reader.",
			Category:    "testing",
			Status:      agentskills.StatusAvailable,
			SafetyLevel: agentskills.SafetyReadOnly,
			Enabled:     true,
		},
	}, []agentskills.InstalledSkillState{
		{ProfileID: "profile-main", SkillID: "local.archive.disabled_writer", Enabled: false},
		{ProfileID: "profile-main", SkillID: "local.archive.invalid_reader", Enabled: true, Status: agentskills.StatusInvalid},
	})

	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:tools-list",
		SessionIDSeed: "mcp-tools-list-test-session",
		SkillRegistry: registry,
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

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-tools-list-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	result := clientSession.InitializeResult()
	if result == nil || result.Capabilities.Tools == nil {
		t.Fatalf("initialize should advertise MCP tool list capability, got %#v", result)
	}

	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	byName := map[string]*mcp.Tool{}
	for _, tool := range tools.Tools {
		byName[tool.Name] = tool
	}
	for _, expected := range []string{
		"cabinet.inventory.search_items",
		"cabinet.inventory.create_item",
		"local.archive.enabled_reader",
	} {
		if byName[expected] == nil {
			t.Fatalf("tools/list omitted enabled executable skill %q; names=%v", expected, toolNames(tools.Tools))
		}
	}
	for _, omitted := range []string{
		"cabinet.guided.inventory.update_item",
		"local.archive.disabled_writer",
		"local.archive.invalid_reader",
	} {
		if byName[omitted] != nil {
			t.Fatalf("tools/list exposed disabled, unavailable, or unimplemented skill %q", omitted)
		}
	}
	inventorySearch := byName["cabinet.inventory.search_items"]
	if inventorySearch.Description == "" || inventorySearch.InputSchema == nil {
		t.Fatalf("inventory search tool should carry registry description and deterministic input schema, got %+v", inventorySearch)
	}
}

func TestReadOnlyToolsCallDispatchesThroughGovernedAgentSkillPath(t *testing.T) {
	registry := agentskills.NewProfileRegistry("profile-main", nil, nil)
	var reviewed agentskills.PreviewRequest
	var dispatched agentskills.PreviewRequest
	var receipts []OperationReceipt
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:read-only-call",
		SessionIDSeed: "mcp-read-only-call-test-session",
		SkillRegistry: registry,
		ReceiptSink: ReceiptSinkFunc(func(_ context.Context, receipt OperationReceipt) {
			receipts = append(receipts, receipt)
		}),
		AuthorityReviewer: AuthorityReviewerFunc(func(_ context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
			reviewed = req
			return registry.ReviewAuthority(req, agentskills.AgentAuthorityPolicy{
				ProfileID:  "profile-main",
				Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
				EntryPoint: "mcp",
			})
		}),
		SkillDispatcher: AgentSkillDispatcherFunc(func(_ context.Context, req agentskills.PreviewRequest) (map[string]any, string, error) {
			dispatched = req
			return map[string]any{
				"operation":  "inventory.item.search",
				"profile_id": req.ProfileID,
				"read_only":  true,
				"query":      req.Parameters["query"],
				"items": []map[string]any{{
					"item_id": "item-main-1",
					"title":   "Main profile Charizard",
				}},
				"total": 1,
			}, "", nil
		}),
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

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-read-only-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.inventory.search_items",
		Arguments: map[string]any{
			"query":      "Charizard",
			"profile_id": "profile-other",
		},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	payload, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("StructuredContent = %T, want map[string]any: %#v", result.StructuredContent, result.StructuredContent)
	}
	if payload["profile_id"] != "profile-main" || payload["operation"] != "inventory.item.search" || payload["read_only"] != true {
		t.Fatalf("tool result should be grounded to the bound profile and read-only operation, got %#v", payload)
	}
	if reviewed.SkillID != "cabinet.inventory.search_items" ||
		reviewed.ProfileID != "profile-main" ||
		reviewed.SourceChannel != "mcp" ||
		reviewed.SourceSurface != "mcp.tools.call" ||
		reviewed.Parameters["query"] != "Charizard" {
		t.Fatalf("unexpected authority request: %+v", reviewed)
	}
	if dispatched.SkillID != "cabinet.inventory.search_items" ||
		dispatched.ProfileID != "profile-main" ||
		dispatched.SourceChannel != "mcp" ||
		dispatched.SourceSurface != "mcp.tools.call" ||
		dispatched.Parameters["query"] != "Charizard" {
		t.Fatalf("unexpected governed dispatch request: %+v", dispatched)
	}
	if dispatched.Parameters["profile_id"] != "profile-other" {
		t.Fatalf("dispatcher should receive original arguments while binding execution to cfg profile, got %+v", dispatched.Parameters)
	}
	var applyReceipt *OperationReceipt
	for i := range receipts {
		if receipts[i].Method == "tools/call" &&
			receipts[i].Capability == "tool:cabinet.inventory.search_items" &&
			receipts[i].Outcome == "apply_allowed" {
			applyReceipt = &receipts[i]
			break
		}
	}
	if applyReceipt == nil || applyReceipt.ProfileID != "profile-main" || applyReceipt.VersionDigest != "git:read-only-call" {
		t.Fatalf("expected redacted allowed authority receipt for read-only call, got %+v", receipts)
	}
	body, err := json.Marshal(receipts)
	if err != nil {
		t.Fatalf("marshal receipts: %v", err)
	}
	for _, forbidden := range []string{"Charizard", "profile-other"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("read-only MCP receipt leaked tool argument %q: %s", forbidden, body)
		}
	}
}

func TestLocalWriteToolsCallReturnsPreviewWithoutDispatchingMutation(t *testing.T) {
	registry := agentskills.NewProfileRegistry("profile-main", nil, nil)
	var reviewed agentskills.PreviewRequest
	dispatched := false
	var receipts []OperationReceipt
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:local-write-preview",
		SessionIDSeed: "mcp-local-write-preview-test-session",
		SkillRegistry: registry,
		ReceiptSink: ReceiptSinkFunc(func(_ context.Context, receipt OperationReceipt) {
			receipts = append(receipts, receipt)
		}),
		AuthorityReviewer: AuthorityReviewerFunc(func(_ context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
			reviewed = req
			return registry.ReviewAuthority(req, agentskills.AgentAuthorityPolicy{
				ProfileID:  "profile-main",
				Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
				EntryPoint: "mcp",
			})
		}),
		SkillDispatcher: AgentSkillDispatcherFunc(func(_ context.Context, req agentskills.PreviewRequest) (map[string]any, string, error) {
			dispatched = true
			return map[string]any{"unexpected": req.SkillID}, "", nil
		}),
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

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-local-write-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.inventory.create_item",
		Arguments: map[string]any{
			"title":       "Preview Only MCP Item",
			"part_number": "MCP-PREVIEW-1",
		},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	payload, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("StructuredContent = %T, want map[string]any: %#v", result.StructuredContent, result.StructuredContent)
	}
	if payload["skill_id"] != "cabinet.inventory.create_item" ||
		payload["profile_id"] != "profile-main" ||
		payload["preview_only"] != true ||
		payload["mutation_applied"] != false ||
		payload["confirmation_required"] != true ||
		payload["confirmation_state"] != "preview_required" {
		t.Fatalf("local-write MCP call should return a confirmation-gated preview, got %#v", payload)
	}
	if dispatched {
		t.Fatal("local-write MCP preview must not dispatch the mutation apply path")
	}
	if reviewed.SkillID != "cabinet.inventory.create_item" ||
		reviewed.ProfileID != "profile-main" ||
		reviewed.Confirm ||
		reviewed.SourceChannel != "mcp" ||
		reviewed.SourceSurface != "mcp.tools.call" {
		t.Fatalf("unexpected local-write authority request: %+v", reviewed)
	}
	if reviewed.Parameters["title"] != "Preview Only MCP Item" || reviewed.Parameters["part_number"] != "MCP-PREVIEW-1" {
		t.Fatalf("expected tool arguments to reach authority preview review, got %+v", reviewed.Parameters)
	}
	var previewReceipt *OperationReceipt
	for i := range receipts {
		if receipts[i].Method == "tools/call" &&
			receipts[i].Capability == "tool:cabinet.inventory.create_item" &&
			receipts[i].Outcome == "preview_allowed" {
			previewReceipt = &receipts[i]
			break
		}
	}
	if previewReceipt == nil || previewReceipt.ProfileID != "profile-main" || previewReceipt.VersionDigest != "git:local-write-preview" {
		t.Fatalf("expected redacted preview authority receipt for local-write call, got %+v", receipts)
	}
	body, err := json.Marshal(receipts)
	if err != nil {
		t.Fatalf("marshal receipts: %v", err)
	}
	for _, forbidden := range []string{"Preview Only MCP Item", "MCP-PREVIEW-1"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("local-write MCP receipt leaked tool argument %q: %s", forbidden, body)
		}
	}
}

func TestMCPAgentSkillPreviewApplyCancelTokensUseChatConfirmationBoundary(t *testing.T) {
	registry := agentskills.NewProfileRegistry("profile-main", nil, nil)
	var previewInput chat.PreviewActionInput
	var applyInputs []chat.ApplyActionInput
	var cancelInput chat.ApplyActionInput
	applyCount := 0
	appliedTokens := map[string]chat.ApplyActionResult{}
	server, err := NewServer(Config{
		ProfileID:     "profile-main",
		ProfileLabel:  "Main collection",
		Version:       "0.1.0-test",
		VersionDigest: "git:mcp-confirmation",
		SessionIDSeed: "mcp-confirmation-test-session",
		SkillRegistry: registry,
		AuthorityReviewer: AuthorityReviewerFunc(func(_ context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
			return registry.ReviewAuthority(req, agentskills.AgentAuthorityPolicy{
				ProfileID:  "profile-main",
				Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
				EntryPoint: "mcp",
			})
		}),
		ActionPreviewer: ActionPreviewerFunc(func(_ context.Context, in chat.PreviewActionInput) (chat.ActionPreview, error) {
			previewInput = in
			return chat.ActionPreview{
				ID:           "preview-mcp-1",
				ProfileID:    in.ProfileID,
				ThreadID:     in.ThreadID,
				CapabilityID: in.CapabilityID,
				Action:       "create_inventory_item",
				Status:       "previewed",
				Payload:      in.Payload,
			}, nil
		}),
		ActionConfirmer: ActionConfirmerFunc{
			Apply: func(_ context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error) {
				applyInputs = append(applyInputs, in)
				if !in.Confirm {
					return chat.ApplyActionResult{}, errors.New("confirm_required")
				}
				if result, ok := appliedTokens[in.PreviewID]; ok {
					return result, nil
				}
				applyCount++
				result := chat.ApplyActionResult{
					Applied:   true,
					Action:    "create_inventory_item",
					ItemID:    "item-mcp-1",
					PreviewID: in.PreviewID,
					Title:     "Confirmed MCP Item",
				}
				appliedTokens[in.PreviewID] = result
				return result, nil
			},
			Cancel: func(_ context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error) {
				cancelInput = in
				return chat.ApplyActionResult{
					Applied:   false,
					Action:    "create_inventory_item",
					PreviewID: in.PreviewID,
				}, nil
			},
		},
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

	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-confirmation-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer clientSession.Close()

	previewResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.inventory.create_item",
		Arguments: map[string]any{
			"title":       "Confirmed MCP Item",
			"part_number": "MCP-CONFIRM-1",
		},
	})
	if err != nil {
		t.Fatalf("preview CallTool() error = %v", err)
	}
	previewPayload, ok := previewResult.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("preview StructuredContent = %T, want map[string]any", previewResult.StructuredContent)
	}
	if previewPayload["preview_id"] != "preview-mcp-1" ||
		previewPayload["confirmation_state"] != "pending" ||
		previewPayload["apply_tool"] != "cabinet.agent_skill.apply_preview" ||
		previewPayload["cancel_tool"] != "cabinet.agent_skill.cancel_preview" ||
		previewPayload["mutation_applied"] != false {
		t.Fatalf("preview should expose durable apply/cancel token guidance, got %#v", previewPayload)
	}
	if previewInput.ProfileID != "profile-main" ||
		previewInput.ThreadID != "mcp-confirmation-test-session" ||
		previewInput.CapabilityID != "inventory.item.create" ||
		previewInput.Payload["workspace_id"] != "profile:profile-main" {
		t.Fatalf("preview input should be profile/thread bound to Chat action boundary, got %+v", previewInput)
	}

	confirmMissingResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.agent_skill.apply_preview",
		Arguments: map[string]any{
			"preview_id": "preview-mcp-1",
			"thread_id":  "mcp-confirmation-test-session",
			"confirm":    false,
		},
	})
	if err != nil {
		t.Fatalf("apply without confirm=true should return a tool error result, got transport error %v", err)
	}
	if confirmMissingResult == nil || !confirmMissingResult.IsError {
		t.Fatalf("apply without confirm=true should be rejected as a tool error result, got %#v", confirmMissingResult)
	}
	if len(applyInputs) != 0 {
		t.Fatalf("apply without confirm=true must not reach Chat confirmation boundary, got %+v", applyInputs)
	}

	applyResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.agent_skill.apply_preview",
		Arguments: map[string]any{
			"preview_id": "preview-mcp-1",
			"thread_id":  "mcp-confirmation-test-session",
			"confirm":    true,
		},
	})
	if err != nil {
		t.Fatalf("apply CallTool() error = %v", err)
	}
	applyPayload, ok := applyResult.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("apply StructuredContent = %T, want map[string]any", applyResult.StructuredContent)
	}
	if applyPayload["preview_id"] != "preview-mcp-1" ||
		applyPayload["confirmation_state"] != "confirmed" ||
		applyPayload["mutation_applied"] != true ||
		applyPayload["item_id"] != "item-mcp-1" {
		t.Fatalf("apply result should expose confirmed mutation evidence, got %#v", applyPayload)
	}
	if len(applyInputs) != 1 ||
		applyInputs[0].ProfileID != "profile-main" ||
		applyInputs[0].ThreadID != "mcp-confirmation-test-session" ||
		applyInputs[0].PreviewID != "preview-mcp-1" ||
		!applyInputs[0].Confirm {
		t.Fatalf("apply input should be bound to configured profile with explicit confirmation, got %+v", applyInputs)
	}

	replayResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.agent_skill.apply_preview",
		Arguments: map[string]any{
			"preview_id": "preview-mcp-1",
			"thread_id":  "mcp-confirmation-test-session",
			"confirm":    true,
		},
	})
	if err != nil {
		t.Fatalf("replay CallTool() error = %v", err)
	}
	replayPayload, _ := replayResult.StructuredContent.(map[string]any)
	if replayPayload["item_id"] != "item-mcp-1" || applyCount != 1 {
		t.Fatalf("replay should return the idempotent applied result without a second mutation, payload=%#v applyCount=%d", replayPayload, applyCount)
	}

	cancelResult, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "cabinet.agent_skill.cancel_preview",
		Arguments: map[string]any{
			"preview_id": "preview-mcp-2",
			"thread_id":  "mcp-confirmation-test-session",
		},
	})
	if err != nil {
		t.Fatalf("cancel CallTool() error = %v", err)
	}
	cancelPayload, ok := cancelResult.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("cancel StructuredContent = %T, want map[string]any", cancelResult.StructuredContent)
	}
	if cancelPayload["preview_id"] != "preview-mcp-2" ||
		cancelPayload["confirmation_state"] != "cancelled" ||
		cancelPayload["mutation_applied"] != false {
		t.Fatalf("cancel result should expose non-mutating cancellation evidence, got %#v", cancelPayload)
	}
	if cancelInput.ProfileID != "profile-main" ||
		cancelInput.ThreadID != "mcp-confirmation-test-session" ||
		cancelInput.PreviewID != "preview-mcp-2" ||
		cancelInput.Confirm {
		t.Fatalf("cancel input should be bound to configured profile without confirmation apply, got %+v", cancelInput)
	}
}

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
	var receipts []OperationReceipt
	called := false
	server, err := NewServer(Config{
		ProfileID:     "profile-read-only",
		ProfileLabel:  "Read only",
		Version:       "0.1.0-test",
		VersionDigest: "git:authority",
		SessionIDSeed: "mcp-authority-test-session",
		ReceiptSink: ReceiptSinkFunc(func(_ context.Context, receipt OperationReceipt) {
			receipts = append(receipts, receipt)
		}),
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
	var authorityReceipt *OperationReceipt
	for i := range receipts {
		if receipts[i].Method == "tools/call" &&
			receipts[i].Capability == "tool:cabinet.inventory.create_item" &&
			receipts[i].Outcome == "blocked" {
			authorityReceipt = &receipts[i]
			break
		}
	}
	if authorityReceipt == nil {
		t.Fatalf("expected blocked MCP authority receipt, got %+v", receipts)
	}
	if authorityReceipt.ProfileID != "profile-read-only" ||
		authorityReceipt.InputClass != "tool_arguments" ||
		authorityReceipt.ErrorClass != "agent_authority_read_only" ||
		authorityReceipt.VersionDigest != "git:authority" {
		t.Fatalf("blocked authority receipt missing redacted decision metadata: %+v", authorityReceipt)
	}
	body, err := json.Marshal(receipts)
	if err != nil {
		t.Fatalf("marshal MCP authority receipts: %v", err)
	}
	for _, forbidden := range []string{"Blocked MCP Item", "MCP-AUTH-1"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("MCP authority receipt leaked tool argument value %q: %s", forbidden, body)
		}
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
		AuthorityReviewer: AuthorityReviewerFunc(func(_ context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
			return agentskills.AgentAuthorityReview{
				ProfileID:    req.ProfileID,
				EntryPoint:   "mcp",
				SkillID:      req.SkillID,
				Decision:     "allowed",
				ApplyAllowed: true,
			}, nil
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
	var allowedAuthorityReceipt *OperationReceipt
	for i := range receipts {
		if receipts[i].Method == "tools/call" &&
			receipts[i].Capability == "tool:cabinet.test.receipt" &&
			receipts[i].Outcome == "apply_allowed" {
			allowedAuthorityReceipt = &receipts[i]
			break
		}
	}
	if allowedAuthorityReceipt == nil {
		t.Fatalf("expected allowed MCP authority receipt, got %+v", receipts)
	}
	if allowedAuthorityReceipt.ProfileID != "profile-main" ||
		allowedAuthorityReceipt.InputClass != "tool_arguments" ||
		allowedAuthorityReceipt.ErrorClass != "" {
		t.Fatalf("allowed authority receipt missing redacted decision metadata: %+v", allowedAuthorityReceipt)
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

func toolNames(tools []*mcp.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}
