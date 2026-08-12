package auth

import (
	"testing"
	"time"
)

func TestUnlockedSessionProfileBindingFailsClosedForMissingFakeAndExpiredTokens(t *testing.T) {
	t.Parallel()

	service := &Service{
		unlockedSessions: map[string]unlockedSession{},
		autoLockTimeout:  20 * time.Millisecond,
	}
	for _, token := range []string{"", "fake-local-agent-session"} {
		if profileID, err := service.ValidateUnlockedSessionProfile(token); err == nil || profileID != "" {
			t.Fatalf("token %q unexpectedly resolved profile=%q err=%v", token, profileID, err)
		}
	}

	token, err := service.CreateUnlockedSession("profile-owner")
	if err != nil {
		t.Fatalf("create unlocked session: %v", err)
	}
	profileID, err := service.ValidateUnlockedSessionProfile(token)
	if err != nil || profileID != "profile-owner" {
		t.Fatalf("valid token profile=%q err=%v", profileID, err)
	}
	time.Sleep(40 * time.Millisecond)
	if profileID, err = service.ValidateUnlockedSessionProfile(token); err == nil || profileID != "" {
		t.Fatalf("expired token unexpectedly resolved profile=%q err=%v", profileID, err)
	}
}
