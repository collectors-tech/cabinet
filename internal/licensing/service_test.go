package licensing

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestImportValidateAndGating(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profiles := profile.NewRepository(conn)
	p, err := profiles.Create(context.Background(), "Default")
	if err != nil {
		t.Fatalf("Create profile error: %v", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey error: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pub)
	payload := Payload{
		ProfileID: p.ID,
		Tier:      "pro",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		Features:  []string{"scanner_automation", "price_tracking", "ai_assist"},
	}
	raw, _ := json.Marshal(payload)
	sig := ed25519.Sign(priv, raw)
	lic := File{
		PayloadBase64:   base64.StdEncoding.EncodeToString(raw),
		SignatureBase64: base64.StdEncoding.EncodeToString(sig),
	}

	svc := NewService(conn, profiles, pubB64)
	if err := svc.Import(context.Background(), p.ID, lic); err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	status, err := svc.Status(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.State != "valid" || status.Tier != "pro" {
		t.Fatalf("unexpected license status: %+v", status)
	}
	ok, err := svc.Allow(context.Background(), p.ID, "ai_assist")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !ok {
		t.Fatal("expected pro feature to be allowed")
	}
}
