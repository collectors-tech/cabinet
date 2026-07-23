package mcpserver

import (
	"errors"
	"fmt"
	"strings"

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
	return mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Title:   ServerTitle,
		Version: version,
	}, options), nil
}
