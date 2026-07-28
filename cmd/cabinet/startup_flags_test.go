package main

import (
	"os"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestParseStartupArgsBuildsEnvOverrides(t *testing.T) {
	t.Parallel()

	overrides, err := parseStartupArgs([]string{
		"--port", "19090",
		"--data-dir", "/tmp/cabinet-data",
		"--profile", "demo-profile",
		"--auth-mode", "zitadel",
		"--base-url", "http://127.0.0.1:19090",
		"--restart",
		"--allow-parallel",
		"--seed-sample-data",
		"--log-level", "debug",
	})
	if err != nil {
		t.Fatalf("parseStartupArgs returned error: %v", err)
	}

	if overrides.Env["CABINET_PORT"] != "19090" {
		t.Fatalf("expected CABINET_PORT=19090, got %q", overrides.Env["CABINET_PORT"])
	}
	if overrides.Env["CABINET_DATA_DIR"] != "/tmp/cabinet-data" {
		t.Fatalf("expected CABINET_DATA_DIR override")
	}
	if overrides.Env["CABINET_PROFILE"] != "demo-profile" {
		t.Fatalf("expected CABINET_PROFILE override")
	}
	if overrides.Env["CABINET_AUTH_MODE"] != "zitadel" {
		t.Fatalf("expected CABINET_AUTH_MODE override")
	}
	if overrides.Env["CABINET_AUTH_IDENTITY_MODE"] != "zitadel" {
		t.Fatalf("expected CABINET_AUTH_IDENTITY_MODE override")
	}
	if overrides.Env["CABINET_BASE_URL"] != "http://127.0.0.1:19090" {
		t.Fatalf("expected CABINET_BASE_URL override")
	}
	if overrides.Env["CABINET_RESTART"] != "true" {
		t.Fatalf("expected CABINET_RESTART=true")
	}
	if overrides.Env["CABINET_ALLOW_PARALLEL"] != "true" {
		t.Fatalf("expected CABINET_ALLOW_PARALLEL=true")
	}
	if overrides.Env["CABINET_SEED_SAMPLE_DATA"] != "true" {
		t.Fatalf("expected CABINET_SEED_SAMPLE_DATA=true")
	}
	if overrides.Env["CABINET_LOG_LEVEL"] != "debug" {
		t.Fatalf("expected CABINET_LOG_LEVEL=debug")
	}
}

func TestParseStartupArgsRejectsConflictingPortAndListen(t *testing.T) {
	t.Parallel()

	_, err := parseStartupArgs([]string{
		"--listen", "127.0.0.1:18080",
		"--port", "18081",
	})
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestParseStartupArgsRejectsInvalidAuthMode(t *testing.T) {
	t.Parallel()

	_, err := parseStartupArgs([]string{"--auth-mode", "saml"})
	if err == nil {
		t.Fatal("expected auth-mode validation error, got nil")
	}
	if !strings.Contains(err.Error(), "auth-mode") {
		t.Fatalf("expected auth-mode error, got %v", err)
	}
}

func TestParseStartupArgsRejectsClerkAuthMode(t *testing.T) {
	t.Parallel()

	_, err := parseStartupArgs([]string{"--auth-mode", "clerk"})
	if err == nil {
		t.Fatal("expected clerk auth-mode validation error, got nil")
	}
	if !strings.Contains(err.Error(), "expected local or zitadel") {
		t.Fatalf("expected local/zitadel guidance, got %v", err)
	}
}

func TestParseStartupArgsAcceptsZitadelAuthMode(t *testing.T) {
	overrides, err := parseStartupArgs([]string{"--auth-mode", "zitadel"})
	if err != nil {
		t.Fatalf("expected zitadel auth mode to parse: %v", err)
	}
	if overrides.Env["CABINET_AUTH_IDENTITY_MODE"] != "zitadel" {
		t.Fatalf("expected CABINET_AUTH_IDENTITY_MODE=zitadel")
	}
}

func TestApplyStartupOverridesTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("CABINET_PORT", "17880")
	t.Setenv("CABINET_DATA_DIR", "")

	overrides := startupOverrides{
		Env: map[string]string{
			"CABINET_PORT":     "19091",
			"CABINET_DATA_DIR": t.TempDir(),
		},
	}
	applyStartupOverrides(overrides)
	cfg := config.Load()
	if cfg.Port != 19091 {
		t.Fatalf("expected CLI override port 19091, got %d", cfg.Port)
	}
	if cfg.DataDir == "" {
		t.Fatalf("expected non-empty data dir from CLI override")
	}
}

func TestBuildEffectiveStartupConfigLineIncludesResolvedFields(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:    "127.0.0.1:19999",
		Host:    "127.0.0.1",
		Port:    19999,
		DataDir: "/tmp/cabinet",
	}
	line := buildEffectiveStartupConfigLine(cfg)
	for _, token := range []string{
		"CABINET_EFFECTIVE_CONFIG",
		"addr=127.0.0.1:19999",
		"host=127.0.0.1",
		"port=19999",
		"data_dir=/tmp/cabinet",
	} {
		if !strings.Contains(line, token) {
			t.Fatalf("expected token %q in line %q", token, line)
		}
	}
}

func TestApplyStartupOverridesSetsEnvironmentVariables(t *testing.T) {
	overrides := startupOverrides{
		Env: map[string]string{
			"CABINET_LOG_LEVEL": "info",
		},
	}
	applyStartupOverrides(overrides)
	if got := os.Getenv("CABINET_LOG_LEVEL"); got != "info" {
		t.Fatalf("expected CABINET_LOG_LEVEL=info, got %q", got)
	}
}

func TestValidateStartupOverridesRejectsRestartWithAllowParallel(t *testing.T) {
	t.Parallel()

	err := validateStartupOverrides(startupOverrides{
		Env: map[string]string{
			"CABINET_RESTART":        "true",
			"CABINET_ALLOW_PARALLEL": "true",
			"CABINET_PROFILE":        "demo",
		},
	})
	if err == nil {
		t.Fatal("expected restart/allow-parallel validation error, got nil")
	}
	if !strings.Contains(err.Error(), "restart") {
		t.Fatalf("expected restart validation error, got %v", err)
	}
}
