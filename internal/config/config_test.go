package config

import "testing"

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
