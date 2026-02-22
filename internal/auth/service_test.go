package auth

import (
	"testing"
	"time"
)

func TestRandomTokenUniqueAndNonEmpty(t *testing.T) {
	t.Parallel()

	a, err := randomToken()
	if err != nil {
		t.Fatalf("randomToken() error = %v", err)
	}
	b, err := randomToken()
	if err != nil {
		t.Fatalf("randomToken() error = %v", err)
	}
	if a == "" || b == "" {
		t.Fatal("expected non-empty tokens")
	}
	if a == b {
		t.Fatal("expected distinct tokens")
	}
}

func TestSessionLifecycle(t *testing.T) {
	t.Parallel()

	s := &Service{sessions: map[string]pendingSession{}}
	s.storeSession("abc", pendingSession{
		ProfileID: "p1",
		Kind:      "registration",
		ExpiresAt: time.Now().Add(1 * time.Minute),
	})

	got, err := s.popSession("abc", "registration")
	if err != nil {
		t.Fatalf("popSession() error = %v", err)
	}
	if got.ProfileID != "p1" {
		t.Fatalf("expected profile p1, got %q", got.ProfileID)
	}

	if _, err := s.popSession("abc", "registration"); err == nil {
		t.Fatal("expected missing session error after pop")
	}
}
