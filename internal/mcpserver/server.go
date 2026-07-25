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

	AuthorityReviewer AuthorityReviewer
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
		server.AddReceivingMiddleware(authorityMiddleware(cfg.AuthorityReviewer, profileID))
	}
	if cfg.ReceiptSink != nil {
		server.AddReceivingMiddleware(receiptMiddleware(cfg, profileID, version))
	}
	return server, nil
}

func authorityMiddleware(reviewer AuthorityReviewer, profileID string) mcp.Middleware {
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
