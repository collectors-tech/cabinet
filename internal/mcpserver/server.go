package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

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
	if cfg.ReceiptSink != nil {
		server.AddReceivingMiddleware(receiptMiddleware(cfg, profileID, version))
	}
	return server, nil
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
