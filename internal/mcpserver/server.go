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
	options := &mcp.ServerOptions{
		Capabilities: capabilities,
		Instructions: fmt.Sprintf("Cabinet MCP session bound to profile %s. No tools or resources are exposed by this foundation server.", profileID),
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
	registerAgentSkillTools(server, cfg.SkillRegistry)
	return server, nil
}

func registerAgentSkillTools(server *mcp.Server, registry AgentSkillRegistry) {
	if registry == nil {
		return
	}
	for _, skill := range registry.List() {
		if !skill.Enabled || !skill.Executable {
			continue
		}
		skill := skill
		server.AddTool(&mcp.Tool{
			Name:        strings.TrimSpace(skill.ID),
			Description: strings.TrimSpace(skill.Description),
			InputSchema: mcpInputSchemaForAgentSkill(skill),
		}, func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return nil, fmt.Errorf("mcp agent skill dispatch is not implemented yet: %s", skill.ID)
		})
	}
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
			review, err := reviewer.ReviewAgentAuthority(ctx, agentskills.PreviewRequest{
				SkillID:        toolName,
				ProfileID:      profileID,
				Confirm:        true,
				SourceSurface:  "mcp.tools.call",
				SourceChannel:  "mcp",
				SourceThreadID: sessionIDForReceipt(req.GetSession(), ""),
				Parameters:     args,
			})
			if err != nil {
				return nil, err
			}
			recordMCPAuthorityReceipt(ctx, cfg, req, toolName, profileID, version, review)
			if !review.ApplyAllowed {
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
