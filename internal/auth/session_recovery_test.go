package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestUnlockSessionAutoLockByInactivity(t *testing.T) {
	t.Parallel()

	s := &Service{
		unlockedSessions: map[string]unlockedSession{},
		autoLockTimeout:  50 * time.Millisecond,
	}
	token, err := s.CreateUnlockedSession("p1")
	if err != nil {
		t.Fatalf("CreateUnlockedSession() error = %v", err)
	}
	if err := s.ValidateUnlockedSession(token); err != nil {
		t.Fatalf("ValidateUnlockedSession() immediate error = %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	if err := s.ValidateUnlockedSession(token); err == nil {
		t.Fatal("expected session to auto-lock after inactivity timeout")
	}
}

func TestRecoveryPassphraseRoundTrip(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	profiles := profile.NewRepository(conn)
	p, err := profiles.Create(context.Background(), "Default")
	if err != nil {
		t.Fatalf("Create() profile error = %v", err)
	}

	svc, err := NewService(config.Config{
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
	}, conn, profiles)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if err := svc.SetRecoveryPassphrase(context.Background(), p.ID, "hunter2"); err != nil {
		t.Fatalf("SetRecoveryPassphrase() error = %v", err)
	}
	ok, err := svc.VerifyRecoveryPassphrase(context.Background(), p.ID, "hunter2")
	if err != nil {
		t.Fatalf("VerifyRecoveryPassphrase() error = %v", err)
	}
	if !ok {
		t.Fatal("expected recovery passphrase verification to succeed")
	}
	notOK, err := svc.VerifyRecoveryPassphrase(context.Background(), p.ID, "wrong")
	if err != nil {
		t.Fatalf("VerifyRecoveryPassphrase() wrong error = %v", err)
	}
	if notOK {
		t.Fatal("expected wrong recovery passphrase verification to fail")
	}
}
