package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsAndChannelFallback(t *testing.T) {
	t.Setenv("CABINET_ADDR", "")
	t.Setenv("CABINET_DATA_DIR", "")
	t.Setenv("CABINET_DB_PATH", "")
	t.Setenv("CABINET_UPDATE_CHANNEL", "invalid")
	t.Setenv("CABINET_WEBAUTHN_ORIGIN", "")

	cfg := Load()
	if cfg.Addr != "127.0.0.1:17880" {
		t.Fatalf("expected default address 127.0.0.1:17880, got %q", cfg.Addr)
	}
	if cfg.DataDir == "" {
		t.Fatal("expected default data dir")
	}
	if cfg.DBPath == "" {
		t.Fatal("expected default db path")
	}
	if cfg.WebAuthnOrigin != "http://127.0.0.1:17880" {
		t.Fatalf("expected default WebAuthn origin http://127.0.0.1:17880, got %q", cfg.WebAuthnOrigin)
	}
	if string(cfg.UpdateChannel) != "stable" {
		t.Fatalf("expected stable fallback, got %q", cfg.UpdateChannel)
	}
}

func TestLoadFromDotEnvWhenProcessEnvUnset(t *testing.T) {
	t.Setenv("CABINET_ADDR", "")
	t.Setenv("CABINET_WEBAUTHN_ORIGIN", "")
	t.Setenv("CABINET_DATA_DIR", "")
	t.Setenv("CABINET_DB_PATH", "")
	t.Setenv("CABINET_UPDATE_CHANNEL", "")

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("CABINET_ADDR=127.0.0.1:19090\nCABINET_WEBAUTHN_ORIGIN=http://127.0.0.1:19090\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	cfg := Load()
	if cfg.Addr != "127.0.0.1:19090" {
		t.Fatalf("expected address from .env, got %q", cfg.Addr)
	}
	if cfg.WebAuthnOrigin != "http://127.0.0.1:19090" {
		t.Fatalf("expected WebAuthn origin from .env, got %q", cfg.WebAuthnOrigin)
	}
}

func TestLoadHostAndPortFromEnvironment(t *testing.T) {
	t.Setenv("CABINET_ADDR", "")
	t.Setenv("CABINET_HOST", "0.0.0.0")
	t.Setenv("CABINET_PORT", "18888")

	cfg := Load()
	if cfg.Addr != "0.0.0.0:18888" {
		t.Fatalf("expected address from CABINET_HOST/CABINET_PORT, got %q", cfg.Addr)
	}
	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected host from CABINET_HOST, got %q", cfg.Host)
	}
	if cfg.Port != 18888 {
		t.Fatalf("expected port 18888 from CABINET_PORT, got %d", cfg.Port)
	}
}

func TestLoadInvalidPortReturnsValidationError(t *testing.T) {
	t.Setenv("CABINET_ADDR", "")
	t.Setenv("CABINET_HOST", "127.0.0.1")
	t.Setenv("CABINET_PORT", "abc")

	cfg := Load()
	if cfg.ValidationError == "" {
		t.Fatalf("expected validation error for invalid CABINET_PORT")
	}
	if got := cfg.ValidationError; got == "" || !containsAll(got, []string{"CABINET_PORT", "invalid"}) {
		t.Fatalf("expected CABINET_PORT validation error, got %q", got)
	}
}

func containsAll(in string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(strings.ToLower(in), strings.ToLower(part)) {
			return false
		}
	}
	return true
}
