package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/mcpserver"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type launcherConfig struct {
	ProfileID     string
	ProfileLabel  string
	Version       string
	VersionDigest string
	DataDir       string
	DBPath        string
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
	fs.StringVar(&cfg.DataDir, "data-dir", "", "Cabinet data directory used to verify the profile binding.")
	fs.StringVar(&cfg.DBPath, "db-path", "", "Cabinet SQLite database path used to verify the profile binding.")

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
	cfg.DataDir = strings.TrimSpace(cfg.DataDir)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	if cfg.ProfileID == "" {
		return launcherConfig{}, errors.New("explicit --profile-id is required")
	}
	if cfg.DataDir != "" && cfg.DBPath != "" {
		return launcherConfig{}, errors.New("use either --data-dir or --db-path, not both")
	}
	return cfg, nil
}

func runLauncher(ctx context.Context, cfg launcherConfig) error {
	verifiedCfg, err := verifyProfileAuthority(ctx, cfg)
	if err != nil {
		return err
	}
	server, err := mcpserver.NewServer(mcpserver.Config{
		ProfileID:     verifiedCfg.ProfileID,
		ProfileLabel:  verifiedCfg.ProfileLabel,
		Version:       verifiedCfg.Version,
		VersionDigest: verifiedCfg.VersionDigest,
	})
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}

func verifyProfileAuthority(ctx context.Context, cfg launcherConfig) (launcherConfig, error) {
	dbPath := strings.TrimSpace(cfg.DBPath)
	if dbPath == "" && strings.TrimSpace(cfg.DataDir) != "" {
		dbPath = filepath.Join(strings.TrimSpace(cfg.DataDir), "cabinet.db")
	}
	if dbPath == "" {
		return cfg, nil
	}

	conn, err := db.OpenAndMigrate(ctx, dbPath)
	if err != nil {
		return launcherConfig{}, fmt.Errorf("open Cabinet profile store: %w", err)
	}
	defer conn.Close()

	p, err := profile.NewRepository(conn).GetByID(ctx, strings.TrimSpace(cfg.ProfileID))
	if err != nil {
		return launcherConfig{}, fmt.Errorf("mcp profile binding rejected: %w", err)
	}
	if strings.TrimSpace(cfg.ProfileLabel) == "" {
		cfg.ProfileLabel = p.Name
	}
	return cfg, nil
}
