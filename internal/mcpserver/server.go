package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ServerName    = "cabinet"
	ServerTitle   = "Cabinet"
	extensionName = "collectors-tech/cabinet-profile"
)

type Config struct {
	ProfileID     string
	ProfileLabel  string
	Version       string
	VersionDigest string
	SessionIDSeed string
	ReceiptSink   ReceiptSink

	SkillRegistry     AgentSkillRegistry
	AuthorityReviewer AuthorityReviewer
	SkillDispatcher   AgentSkillDispatcher
	ActionPreviewer   ActionPreviewer
	ActionConfirmer   ActionConfirmer
}

type AgentSkillRegistry interface {
	List() []agentskills.Skill
}

type AuthorityReviewer interface {
	ReviewAgentAuthority(ctx context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error)
}

type AuthorityReviewerFunc func(ctx context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error)

func (f AuthorityReviewerFunc) ReviewAgentAuthority(ctx context.Context, req agentskills.PreviewRequest) (agentskills.AgentAuthorityReview, error) {
	return f(ctx, req)
}

type AgentSkillDispatcher interface {
	ApplyAgentSkill(ctx context.Context, req agentskills.PreviewRequest) (map[string]any, string, error)
}

type AgentSkillDispatcherFunc func(ctx context.Context, req agentskills.PreviewRequest) (map[string]any, string, error)

func (f AgentSkillDispatcherFunc) ApplyAgentSkill(ctx context.Context, req agentskills.PreviewRequest) (map[string]any, string, error) {
	return f(ctx, req)
}

type ActionPreviewer interface {
	PreviewAction(ctx context.Context, in chat.PreviewActionInput) (chat.ActionPreview, error)
}

type ActionPreviewerFunc func(ctx context.Context, in chat.PreviewActionInput) (chat.ActionPreview, error)

func (f ActionPreviewerFunc) PreviewAction(ctx context.Context, in chat.PreviewActionInput) (chat.ActionPreview, error) {
	return f(ctx, in)
}

type ActionConfirmer interface {
	ApplyAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
	CancelAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
}

type ActionConfirmerFunc struct {
	Apply  func(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
	Cancel func(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
}

func (f ActionConfirmerFunc) ApplyAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error) {
	return f.Apply(ctx, in)
}

func (f ActionConfirmerFunc) CancelAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error) {
	return f.Cancel(ctx, in)
}

func NewServer(cfg Config) (*mcp.Server, error) {
	profileID := strings.TrimSpace(cfg.ProfileID)
	if profileID == "" {
		return nil, errors.New("mcp profile binding is required")
	}
	version := strings.TrimSpace(cfg.Version)
	if version == "" {
		version = "dev"
	}
	capabilities := &mcp.ServerCapabilities{}
	capabilities.AddExtension(extensionName, map[string]any{
		"profile_id":     profileID,
		"profile_label":  strings.TrimSpace(cfg.ProfileLabel),
		"version_digest": strings.TrimSpace(cfg.VersionDigest),
	})
	instructions := fmt.Sprintf("Cabinet MCP session bound to profile %s. No tools or resources are exposed by this foundation server.", profileID)
	if cfg.SkillRegistry != nil {
		instructions = fmt.Sprintf("Cabinet MCP session bound to profile %s. Agent Skill tools are governed by Cabinet authority, confirmation, and audit policy.", profileID)
	}
	options := &mcp.ServerOptions{
		Capabilities: capabilities,
		Instructions: instructions,
	}
	if seed := strings.TrimSpace(cfg.SessionIDSeed); seed != "" {
		options.GetSessionID = func() string { return seed }
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Title:   ServerTitle,
		Version: version,
	}, options)
	if cfg.AuthorityReviewer != nil {
		server.AddReceivingMiddleware(authorityMiddleware(cfg, profileID, version))
	}
	if cfg.ReceiptSink != nil {
		server.AddReceivingMiddleware(receiptMiddleware(cfg, profileID, version))
	}
	registerAgentSkillConfirmationTools(server, cfg, profileID, version)
	registerAgentSkillTools(server, cfg, profileID)
	return server, nil
}

func registerAgentSkillConfirmationTools(server *mcp.Server, cfg Config, profileID string, version string) {
	if cfg.ActionConfirmer == nil {
		return
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cabinet.agent_skill.apply_preview",
		Description: "Apply a Cabinet Agent Skill preview after explicit confirmation.",
		InputSchema: mcpConfirmationInputSchema("confirm"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
		confirm, _ := args["confirm"].(bool)
		if !confirm {
			recordMCPConfirmationReceipt(ctx, cfg, req, "cabinet.agent_skill.apply_preview", profileID, version, "confirm_required", "confirm_required")
			return nil, nil, errors.New("confirm_required")
		}
		result, err := cfg.ActionConfirmer.ApplyAction(ctx, mcpApplyActionInput(profileID, args))
		recordMCPConfirmationReceipt(ctx, cfg, req, "cabinet.agent_skill.apply_preview", profileID, version, mcpConfirmationOutcome("confirmed", err), mcpConfirmationErrorClass(err))
		return nil, mcpActionConfirmationResult("confirmed", result), err
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cabinet.agent_skill.cancel_preview",
		Description: "Cancel a pending Cabinet Agent Skill preview without applying a mutation.",
		InputSchema: mcpConfirmationInputSchema("cancel"),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
		result, err := cfg.ActionConfirmer.CancelAction(ctx, mcpCancelActionInput(profileID, args))
		recordMCPConfirmationReceipt(ctx, cfg, req, "cabinet.agent_skill.cancel_preview", profileID, version, mcpConfirmationOutcome("cancelled", err), mcpConfirmationErrorClass(err))
		return nil, mcpActionConfirmationResult("cancelled", result), err
	})
}

func registerAgentSkillTools(server *mcp.Server, cfg Config, profileID string) {
	registry := cfg.SkillRegistry
	if registry == nil {
		return
	}
	for _, skill := range registry.List() {
		if !skill.Enabled || !skill.Executable {
			continue
		}
		skill := skill
		mcp.AddTool(server, &mcp.Tool{
			Name:        strings.TrimSpace(skill.ID),
			Description: strings.TrimSpace(skill.Description),
			InputSchema: mcpInputSchemaForAgentSkill(skill),
		}, func(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, map[string]any, error) {
			result, err := dispatchAgentSkillToolCall(ctx, cfg, profileID, skill, req, args)
			return nil, result, err
		})
	}
}

func dispatchAgentSkillToolCall(ctx context.Context, cfg Config, profileID string, skill agentskills.Skill, req *mcp.CallToolRequest, args map[string]any) (map[string]any, error) {
	if args == nil {
		args = map[string]any{}
	}
	params := make(map[string]any, len(args))
	for key, value := range args {
		params[key] = value
	}
	params = mcpAgentSkillBoundParameters(profileID, params)
	request := agentskills.PreviewRequest{
		SkillID:        strings.TrimSpace(skill.ID),
		ProfileID:      strings.TrimSpace(profileID),
		Confirm:        false,
		SourceSurface:  "mcp.tools.call",
		SourceChannel:  "mcp",
		SourceThreadID: sessionIDForReceipt(req.GetSession(), strings.TrimSpace(cfg.SessionIDSeed)),
		Parameters:     params,
	}
	if skill.Permissions.ExternalWrite || skill.Permissions.Destructive || skill.SafetyLevel == agentskills.SafetyExternalWrite || skill.SafetyLevel == agentskills.SafetyDestructive {
		return nil, fmt.Errorf("mcp agent skill dispatch requires Cabinet confirmation for high-risk skill: %s", skill.ID)
	}
	if skill.Permissions.LocalWrite || skill.Permissions.RequiresConfirm || skill.SafetyLevel == agentskills.SafetyPreviewOnly || skill.SafetyLevel == agentskills.SafetyConfirmRequired {
		if cfg.ActionPreviewer != nil {
			preview, err := cfg.ActionPreviewer.PreviewAction(ctx, chat.PreviewActionInput{
				ProfileID:    request.ProfileID,
				ThreadID:     request.SourceThreadID,
				CapabilityID: firstNonEmptyMCPReceipt(firstString(skill.RequiredActions), firstString(skill.Capabilities)),
				Payload:      request.Parameters,
			})
			if err != nil {
				return nil, fmt.Errorf("mcp agent skill preview failed: %w", err)
			}
			return mcpAgentSkillDurablePreviewResult(skill, request, preview), nil
		}
		return mcpAgentSkillPreviewResult(skill, request), nil
	}
	if cfg.SkillDispatcher == nil {
		return nil, fmt.Errorf("mcp agent skill dispatch is not configured: %s", skill.ID)
	}
	result, blocker, err := cfg.SkillDispatcher.ApplyAgentSkill(ctx, agentskills.PreviewRequest{
		SkillID:        request.SkillID,
		ProfileID:      request.ProfileID,
		Confirm:        request.Confirm,
		SourceSurface:  request.SourceSurface,
		SourceChannel:  request.SourceChannel,
		SourceThreadID: request.SourceThreadID,
		Parameters:     request.Parameters,
	})
	if err != nil {
		if strings.TrimSpace(blocker) != "" {
			return nil, fmt.Errorf("mcp agent skill dispatch failed: %s", blocker)
		}
		return nil, err
	}
	if result == nil {
		result = map[string]any{}
	}
	return result, nil
}

func mcpAgentSkillPreviewResult(skill agentskills.Skill, req agentskills.PreviewRequest) map[string]any {
	return map[string]any{
		"skill_id":              strings.TrimSpace(skill.ID),
		"status":                strings.TrimSpace(string(skill.Status)),
		"safety_level":          strings.TrimSpace(string(skill.SafetyLevel)),
		"profile_id":            strings.TrimSpace(req.ProfileID),
		"source_channel":        strings.TrimSpace(req.SourceChannel),
		"source_surface":        strings.TrimSpace(req.SourceSurface),
		"preview_only":          true,
		"mutation_applied":      false,
		"confirmation_required": true,
		"confirmation_state":    "preview_required",
		"next_action":           firstNonEmptyMCPReceipt(strings.TrimSpace(skill.NextAction), "Review the preview in Cabinet and explicitly confirm before applying this Agent Skill."),
	}
}

func mcpAgentSkillDurablePreviewResult(skill agentskills.Skill, req agentskills.PreviewRequest, preview chat.ActionPreview) map[string]any {
	result := mcpAgentSkillPreviewResult(skill, req)
	result["preview_id"] = strings.TrimSpace(preview.ID)
	result["thread_id"] = strings.TrimSpace(preview.ThreadID)
	result["action"] = strings.TrimSpace(preview.Action)
	result["capability_id"] = strings.TrimSpace(preview.CapabilityID)
	result["status"] = strings.TrimSpace(preview.Status)
	result["confirmation_state"] = "pending"
	result["apply_tool"] = "cabinet.agent_skill.apply_preview"
	result["cancel_tool"] = "cabinet.agent_skill.cancel_preview"
	result["next_action"] = "Call cabinet.agent_skill.apply_preview with confirm=true, or cabinet.agent_skill.cancel_preview to cancel this preview."
	return result
}

func mcpApplyActionInput(profileID string, args map[string]any) chat.ApplyActionInput {
	if args == nil {
		args = map[string]any{}
	}
	confirm, _ := args["confirm"].(bool)
	return chat.ApplyActionInput{
		ProfileID: strings.TrimSpace(profileID),
		ThreadID:  strings.TrimSpace(fmt.Sprint(args["thread_id"])),
		PreviewID: strings.TrimSpace(fmt.Sprint(args["preview_id"])),
		Confirm:   confirm,
	}
}

func mcpCancelActionInput(profileID string, args map[string]any) chat.ApplyActionInput {
	input := mcpApplyActionInput(profileID, args)
	input.Confirm = false
	return input
}

func mcpActionConfirmationResult(state string, result chat.ApplyActionResult) map[string]any {
	return map[string]any{
		"preview_id":         strings.TrimSpace(result.PreviewID),
		"action":             strings.TrimSpace(result.Action),
		"confirmation_state": strings.TrimSpace(state),
		"mutation_applied":   result.Applied,
		"applied":            result.Applied,
		"item_id":            strings.TrimSpace(result.ItemID),
		"wishlist_id":        strings.TrimSpace(result.WishlistID),
		"collection_name":    strings.TrimSpace(result.CollectionName),
		"part_number":        strings.TrimSpace(result.PartNumber),
		"title":              strings.TrimSpace(result.Title),
	}
}

func mcpConfirmationInputSchema(action string) map[string]any {
	properties := map[string]any{
		"preview_id": map[string]any{
			"type":        "string",
			"description": "Cabinet Agent Skill preview token.",
		},
		"thread_id": map[string]any{
			"type":        "string",
			"description": "Cabinet Chat thread that owns the preview token.",
		},
	}
	if action == "confirm" {
		properties["confirm"] = map[string]any{
			"type":        "boolean",
			"description": "Must be true to apply the preview.",
		}
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"preview_id", "thread_id"},
		"properties":           properties,
	}
}

func mcpAgentSkillBoundParameters(profileID string, args map[string]any) map[string]any {
	out := make(map[string]any, len(args)+1)
	for key, value := range args {
		out[key] = value
	}
	if _, ok := out["workspace_id"]; !ok {
		out["workspace_id"] = "profile:" + strings.TrimSpace(profileID)
	}
	return out
}

func mcpInputSchemaForAgentSkill(skill agentskills.Skill) map[string]any {
	properties := map[string]any{}
	for _, ref := range skill.InputSchemaRefs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		properties[ref] = map[string]any{
			"type":        "string",
			"description": "Cabinet Agent Skill input field " + ref + ".",
		}
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"properties":           properties,
		"x-cabinet-skill": map[string]any{
			"id":               strings.TrimSpace(skill.ID),
			"version":          strings.TrimSpace(skill.Version),
			"safety_level":     strings.TrimSpace(string(skill.SafetyLevel)),
			"status":           strings.TrimSpace(string(skill.Status)),
			"requires_confirm": skill.Permissions.RequiresConfirm,
		},
	}
}

func authorityMiddleware(cfg Config, profileID string, version string) mcp.Middleware {
	reviewer := cfg.AuthorityReviewer
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method != "tools/call" || reviewer == nil {
				return next(ctx, method, req)
			}
			toolName, args := mcpToolCallAuthorityInput(req.GetParams())
			if strings.TrimSpace(toolName) == "" {
				return next(ctx, method, req)
			}
			if mcpAgentSkillConfirmationTool(toolName) {
				return next(ctx, method, req)
			}
			args = mcpAgentSkillBoundParameters(profileID, args)
			review, err := reviewer.ReviewAgentAuthority(ctx, agentskills.PreviewRequest{
				SkillID:        toolName,
				ProfileID:      profileID,
				Confirm:        false,
				SourceSurface:  "mcp.tools.call",
				SourceChannel:  "mcp",
				SourceThreadID: sessionIDForReceipt(req.GetSession(), ""),
				Parameters:     args,
			})
			if err != nil {
				return nil, err
			}
			recordMCPAuthorityReceipt(ctx, cfg, req, toolName, profileID, version, review)
			if !review.PreviewAllowed && !review.ApplyAllowed {
				blocker := strings.TrimSpace(review.Blocker)
				if blocker == "" {
					blocker = "agent_authority_blocked"
				}
				return nil, fmt.Errorf("mcp agent authority blocked: %s", blocker)
			}
			return next(ctx, method, req)
		}
	}
}

func mcpAgentSkillConfirmationTool(toolName string) bool {
	switch strings.TrimSpace(toolName) {
	case "cabinet.agent_skill.apply_preview", "cabinet.agent_skill.cancel_preview":
		return true
	default:
		return false
	}
}

func recordMCPConfirmationReceipt(ctx context.Context, cfg Config, req *mcp.CallToolRequest, toolName string, profileID string, version string, outcome string, errorClass string) {
	if cfg.ReceiptSink == nil || req == nil {
		return
	}
	sessionID := sessionIDForReceipt(req.GetSession(), strings.TrimSpace(cfg.SessionIDSeed))
	receipt := OperationReceipt{
		OperationID:   fmt.Sprintf("%s:confirmation:%s", operationIDSeed(sessionID, cfg.SessionIDSeed), strings.TrimSpace(toolName)),
		SessionID:     sessionID,
		ProfileID:     strings.TrimSpace(profileID),
		ProfileLabel:  strings.TrimSpace(cfg.ProfileLabel),
		Capability:    "tool:" + firstNonEmptyMCPReceipt(strings.TrimSpace(toolName), "unknown"),
		Method:        "tools/call",
		Version:       strings.TrimSpace(version),
		VersionDigest: strings.TrimSpace(cfg.VersionDigest),
		InputClass:    "confirmation_token",
		Outcome:       firstNonEmptyMCPReceipt(strings.TrimSpace(outcome), "ok"),
		ErrorClass:    strings.TrimSpace(errorClass),
	}
	cfg.ReceiptSink.RecordMCPReceipt(ctx, receipt)
}

func mcpConfirmationOutcome(success string, err error) string {
	if err != nil {
		return "error"
	}
	return strings.TrimSpace(success)
}

func mcpConfirmationErrorClass(err error) string {
	if err == nil {
		return ""
	}
	return errorClassForReceipt(err)
}

func recordMCPAuthorityReceipt(ctx context.Context, cfg Config, req mcp.Request, toolName string, profileID string, version string, review agentskills.AgentAuthorityReview) {
	if cfg.ReceiptSink == nil {
		return
	}
	sessionID := sessionIDForReceipt(req.GetSession(), strings.TrimSpace(cfg.SessionIDSeed))
	blocker := strings.TrimSpace(review.Blocker)
	receipt := OperationReceipt{
		OperationID:   fmt.Sprintf("%s:authority:%s", operationIDSeed(sessionID, cfg.SessionIDSeed), strings.TrimSpace(toolName)),
		SessionID:     sessionID,
		ProfileID:     strings.TrimSpace(profileID),
		ProfileLabel:  strings.TrimSpace(cfg.ProfileLabel),
		Capability:    "tool:" + firstNonEmptyMCPReceipt(strings.TrimSpace(toolName), "unknown"),
		Method:        "tools/call",
		Version:       strings.TrimSpace(version),
		VersionDigest: strings.TrimSpace(cfg.VersionDigest),
		InputClass:    "tool_arguments",
		Outcome:       mcpAuthorityReceiptOutcome(review),
		ErrorClass:    blocker,
	}
	cfg.ReceiptSink.RecordMCPReceipt(ctx, receipt)
}

func mcpAuthorityReceiptOutcome(review agentskills.AgentAuthorityReview) string {
	if review.ApplyAllowed {
		return "apply_allowed"
	}
	if review.PreviewAllowed {
		return "preview_allowed"
	}
	return "blocked"
}

func firstNonEmptyMCPReceipt(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mcpToolCallAuthorityInput(params mcp.Params) (string, map[string]any) {
	switch p := params.(type) {
	case *mcp.CallToolParams:
		return strings.TrimSpace(p.Name), normalizeToolArguments(p.Arguments)
	case *mcp.CallToolParamsRaw:
		return strings.TrimSpace(p.Name), normalizeRawToolArguments(p.Arguments)
	default:
		return "", map[string]any{}
	}
}

func normalizeToolArguments(args any) map[string]any {
	if args == nil {
		return map[string]any{}
	}
	if typed, ok := args.(map[string]any); ok {
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = value
		}
		return out
	}
	raw, err := json.Marshal(args)
	if err != nil {
		return map[string]any{}
	}
	return normalizeRawToolArguments(raw)
}

func normalizeRawToolArguments(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func receiptMiddleware(cfg Config, profileID string, version string) mcp.Middleware {
	profileLabel := strings.TrimSpace(cfg.ProfileLabel)
	versionDigest := strings.TrimSpace(cfg.VersionDigest)
	seed := strings.TrimSpace(cfg.SessionIDSeed)
	var sequence atomic.Uint64
	var clients sync.Map

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if receipt, ok := buildReceipt(req, method, err, profileID, profileLabel, version, versionDigest, seed, sequence.Add(1), &clients); ok {
				cfg.ReceiptSink.RecordMCPReceipt(ctx, receipt)
			}
			return result, err
		}
	}
}
