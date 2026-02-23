package config

import (
	"os"
	"path/filepath"
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
