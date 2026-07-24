package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/collectors-tech/cabinet/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type launcherConfig struct {
	ProfileID     string
	ProfileLabel  string
	Version       string
	VersionDigest string
}

func main() {
	cfg, err := parseLauncherArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatalf("invalid MCP launcher args: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runLauncher(ctx, cfg); err != nil {
		log.Fatalf("MCP launcher failed: %v", err)
	}
}

func parseLauncherArgs(args []string) (launcherConfig, error) {
	fs := flag.NewFlagSet("cabinet-mcp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var cfg launcherConfig
	fs.StringVar(&cfg.ProfileID, "profile-id", "", "Explicit Cabinet profile ID to bind this MCP session to.")
	fs.StringVar(&cfg.ProfileLabel, "profile-label", "", "Human-readable label for the bound Cabinet profile.")
	fs.StringVar(&cfg.Version, "version", "", "Cabinet version advertised during MCP initialization.")
	fs.StringVar(&cfg.VersionDigest, "version-digest", "", "Cabinet build/version digest advertised in redacted metadata.")

	if err := fs.Parse(args); err != nil {
		return launcherConfig{}, err
	}
	if extra := fs.Args(); len(extra) > 0 {
		return launcherConfig{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(extra, " "))
	}
	cfg.ProfileID = strings.TrimSpace(cfg.ProfileID)
	cfg.ProfileLabel = strings.TrimSpace(cfg.ProfileLabel)
	cfg.Version = strings.TrimSpace(cfg.Version)
	cfg.VersionDigest = strings.TrimSpace(cfg.VersionDigest)
	if cfg.ProfileID == "" {
		return launcherConfig{}, errors.New("explicit --profile-id is required")
	}
	return cfg, nil
}

func runLauncher(ctx context.Context, cfg launcherConfig) error {
	server, err := mcpserver.NewServer(mcpserver.Config{
		ProfileID:     cfg.ProfileID,
		ProfileLabel:  cfg.ProfileLabel,
		Version:       cfg.Version,
		VersionDigest: cfg.VersionDigest,
	})
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}
