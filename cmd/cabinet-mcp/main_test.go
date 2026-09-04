package main

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/mcpserver"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const cabinetMCPTestHelperReady = "CABINET_MCP_TEST_HELPER_READY"

type connectedMCPTransport struct {
	connection mcp.Connection
}

func (t *connectedMCPTransport) Connect(context.Context) (mcp.Connection, error) {
	return t.connection, nil
}

func TestParseLauncherArgsRequiresExplicitProfile(t *testing.T) {
	if _, err := parseLauncherArgs([]string{}); err == nil {
		t.Fatal("parseLauncherArgs() should reject missing profile binding")
	}
}

func TestParseLauncherArgsBuildsProfileBoundConfig(t *testing.T) {
	cfg, err := parseLauncherArgs([]string{
		"--profile-id", "profile-main",
		"--profile-label", "Main collection",
		"--version", "1.2.3",
		"--version-digest", "git:abc123",
		"--data-dir", "C:/Cabinet/Data",
	})
	if err != nil {
		t.Fatalf("parseLauncherArgs() error = %v", err)
	}
	if cfg.ProfileID != "profile-main" || cfg.ProfileLabel != "Main collection" || cfg.Version != "1.2.3" || cfg.VersionDigest != "git:abc123" || cfg.DataDir != "C:/Cabinet/Data" {
		t.Fatalf("unexpected launcher config: %#v", cfg)
	}
}

func TestParseLauncherArgsRejectsConflictingProfileStores(t *testing.T) {
	_, err := parseLauncherArgs([]string{
		"--profile-id", "profile-main",
		"--data-dir", "C:/Cabinet/Data",
		"--db-path", "C:/Cabinet/Data/cabinet.db",
	})
	if err == nil || !strings.Contains(err.Error(), "either --data-dir or --db-path") {
		t.Fatalf("parseLauncherArgs() error = %v, want profile store conflict", err)
	}
}

func TestRunLauncherRejectsMissingProfileBeforeTransport(t *testing.T) {
	err := runLauncher(context.Background(), launcherConfig{})
	if err == nil {
		t.Fatal("runLauncher() should reject missing profile binding")
	}
}

func TestVerifyProfileAuthorityRejectsUnknownProfileInDataDir(t *testing.T) {
	dataDir := t.TempDir()
	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(dataDir, "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	repo := profile.NewRepository(conn)
	p, err := repo.Create(context.Background(), "Main collection")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	verified, err := verifyProfileAuthority(context.Background(), launcherConfig{ProfileID: p.ID, DataDir: dataDir})
	if err != nil {
		t.Fatalf("verifyProfileAuthority() known profile error = %v", err)
	}
	if verified.ProfileLabel != "Main collection" {
		t.Fatalf("verifyProfileAuthority() hydrated profile label = %q, want Main collection", verified.ProfileLabel)
	}

	_, err = verifyProfileAuthority(context.Background(), launcherConfig{ProfileID: "profile-missing", DataDir: dataDir})
	if err == nil || !strings.Contains(err.Error(), "profile not found") {
		t.Fatalf("verifyProfileAuthority() unknown profile error = %v, want profile not found", err)
	}
}

func TestLauncherStdioInitializeSmoke(t *testing.T) {
	testLauncherStdioInitialize(t, 0)
}

func TestLauncherStdioInitializeSeparatesProcessStartupFromProtocolDeadline(t *testing.T) {
	startupElapsed, protocolElapsed := testLauncherStdioInitialize(t, 6*time.Second)
	if startupElapsed < 6*time.Second {
		t.Fatalf("helper startup elapsed = %v, want at least 6s", startupElapsed)
	}
	if protocolElapsed >= 5*time.Second {
		t.Fatalf("MCP protocol initialization elapsed = %v, want less than 5s", protocolElapsed)
	}
}

func testLauncherStdioInitialize(t *testing.T, startupDelay time.Duration) (time.Duration, time.Duration) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestLauncherStdioHelperProcess", "--",
		"--profile-id", "profile-main",
		"--profile-label", "Main collection",
		"--version", "0.1.0-smoke",
		"--version-digest", "git:stdio-smoke",
	)
	cmd.Env = append(os.Environ(),
		"CABINET_MCP_TEST_HELPER_PROCESS=1",
		"CABINET_MCP_TEST_HELPER_STARTUP_DELAY="+startupDelay.String(),
	)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("helper stderr pipe: %v", err)
	}

	commandTransport := &mcp.CommandTransport{Command: cmd, TerminateDuration: time.Second}
	startupStarted := time.Now()
	connection, err := commandTransport.Connect(context.Background())
	if err != nil {
		t.Fatalf("start helper command transport: %v", err)
	}

	type readinessResult struct {
		line string
		err  error
	}
	readiness := make(chan readinessResult, 1)
	go func() {
		line, err := bufio.NewReader(stderr).ReadString('\n')
		readiness <- readinessResult{line: strings.TrimSpace(line), err: err}
	}()
	startupTimer := time.NewTimer(15 * time.Second)
	defer startupTimer.Stop()
	select {
	case result := <-readiness:
		if result.err != nil {
			_ = connection.Close()
			t.Fatalf("wait for helper readiness: %v", result.err)
		}
		if result.line != cabinetMCPTestHelperReady {
			_ = connection.Close()
			t.Fatalf("helper readiness = %q, want %q", result.line, cabinetMCPTestHelperReady)
		}
	case <-startupTimer.C:
		_ = connection.Close()
		t.Fatal("helper process did not become ready within 15s")
	}
	startupElapsed := time.Since(startupStarted)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := mcp.NewClient(&mcp.Implementation{Name: "cabinet-stdio-smoke", Version: "0.1.0"}, nil)
	protocolStarted := time.Now()
	session, err := client.Connect(ctx, &connectedMCPTransport{connection: connection}, nil)
	protocolElapsed := time.Since(protocolStarted)
	if err != nil {
		_ = connection.Close()
		t.Fatalf("client.Connect() error = %v", err)
	}
	defer session.Close()

	result := session.InitializeResult()
	if result == nil {
		t.Fatal("InitializeResult() is nil")
	}
	if result.ServerInfo == nil || result.ServerInfo.Name != mcpserver.ServerName || result.ServerInfo.Version != "0.1.0-smoke" {
		t.Fatalf("unexpected server info from stdio launcher: %#v", result.ServerInfo)
	}
	if !strings.Contains(result.Instructions, "profile-main") {
		t.Fatalf("stdio launcher initialize instructions should include profile binding, got %q", result.Instructions)
	}
	return startupElapsed, protocolElapsed
}

func TestLauncherStdioHelperProcess(t *testing.T) {
	if os.Getenv("CABINET_MCP_TEST_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	cfg, err := parseLauncherArgs(args)
	if err != nil {
		t.Fatalf("parseLauncherArgs() error = %v", err)
	}
	startupDelay, err := time.ParseDuration(os.Getenv("CABINET_MCP_TEST_HELPER_STARTUP_DELAY"))
	if err != nil {
		t.Fatalf("parse helper startup delay: %v", err)
	}
	if startupDelay > 0 {
		time.Sleep(startupDelay)
	}
	if _, err := os.Stderr.WriteString(cabinetMCPTestHelperReady + "\n"); err != nil {
		t.Fatalf("signal helper readiness: %v", err)
	}
	if err := runLauncher(context.Background(), cfg); err != nil {
		t.Fatalf("runLauncher() error = %v", err)
	}
}
