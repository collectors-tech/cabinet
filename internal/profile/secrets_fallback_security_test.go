package profile

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestSecretFallbackDoesNotPersistPlaintextAPIKey(t *testing.T) {
	t.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	t.Setenv("CABINET_FORCE_SECURESTORE_FAIL", "1")

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	p, err := repo.Create(context.Background(), "Wave2 Security")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	const secretValue = "sk-wave2-plaintext-check"
	if err := repo.PutSecret(context.Background(), p.ID, "openai_api_key", secretValue); err != nil {
		t.Fatalf("put secret: %v", err)
	}

	var stored string
	if err := conn.QueryRowContext(context.Background(), `SELECT value FROM profile_secrets WHERE profile_id = ? AND key = ?`, p.ID, "openai_api_key").Scan(&stored); err != nil {
		t.Fatalf("query stored fallback secret: %v", err)
	}
	if strings.Contains(stored, secretValue) {
		t.Fatalf("fallback storage leaked plaintext secret: %q", stored)
	}
	if !strings.HasPrefix(stored, "enc:v1:") {
		t.Fatalf("expected encrypted fallback prefix enc:v1:, got %q", stored)
	}

	got, err := repo.GetSecret(context.Background(), p.ID, "openai_api_key")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got != secretValue {
		t.Fatalf("expected roundtrip secret value, got %q", got)
	}
}
