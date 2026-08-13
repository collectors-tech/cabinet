package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestProfileActivationClassifiesValidationAndStorageFailures(t *testing.T) {
	a := newTestApp(t)
	profiles := profile.NewRepository(a.db)
	created, err := profiles.Create(context.Background(), "Activation classification")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	current, err := profiles.Create(context.Background(), "Current active profile")
	if err != nil {
		t.Fatalf("create current profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), current.ID); err != nil {
		t.Fatalf("set current active profile: %v", err)
	}

	t.Run("missing profile remains a stable validation failure", func(t *testing.T) {
		response := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"missing-profile"}`), map[string]string{"Content-Type": "application/json"})
		if response.Code != http.StatusBadRequest || response.Body.String() != "{\"error\":\"invalid_profile_id\"}\n" {
			t.Fatalf("missing profile status=%d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("sqlite contention is retryable without leaking storage details", func(t *testing.T) {
		lock, err := a.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("reserve lock connection: %v", err)
		}
		defer lock.Close()
		if _, err := lock.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
			t.Fatalf("begin write lock: %v", err)
		}
		defer lock.ExecContext(context.Background(), "ROLLBACK") //nolint:errcheck -- best-effort test cleanup

		response := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+created.ID+`"}`), map[string]string{"Content-Type": "application/json"})
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("contended activation status=%d body=%s", response.Code, response.Body.String())
		}
		if response.Header().Get("Retry-After") != "1" {
			t.Fatalf("expected bounded Retry-After=1, got %q", response.Header().Get("Retry-After"))
		}
		if response.Body.String() != "{\"error\":\"profile_activation_unavailable\",\"retryable\":true,\"retry_after_seconds\":1}\n" {
			t.Fatalf("unexpected retryable body: %s", response.Body.String())
		}
		for _, leaked := range []string{"SQLITE", "database", a.cfg.DBPath, created.ID} {
			if strings.Contains(response.Body.String(), leaked) {
				t.Fatalf("retryable response leaked %q: %s", leaked, response.Body.String())
			}
		}
		active, err := profiles.GetActiveProfile(context.Background())
		if err != nil {
			t.Fatalf("load active profile after contention: %v", err)
		}
		if active.ID != current.ID {
			t.Fatalf("contention changed active profile: got %q want %q", active.ID, current.ID)
		}
	})

	t.Run("unexpected storage failure is fail closed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req := httptest.NewRequest(http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+created.ID+`"}`)).WithContext(ctx)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		a.srv.Handler.ServeHTTP(response, req)

		if response.Code != http.StatusInternalServerError || response.Body.String() != "{\"error\":\"profile_activation_failed\"}\n" {
			t.Fatalf("unexpected storage failure status=%d body=%s", response.Code, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "context canceled") || strings.Contains(response.Body.String(), created.ID) {
			t.Fatalf("unexpected storage response leaked internals: %s", response.Body.String())
		}
	})
}

func TestProfileActivationDiagnosticDoesNotIncludeStorageErrorDetails(t *testing.T) {
	storageErr := errors.New(`write C:\Users\private\cabinet.db: SQL statement failed`)
	failure := classifyProfileActivationFailure(storageErr)
	diagnostic := profileActivationDiagnostic(failure)
	if diagnostic != "profile activation failed: class=unexpected_storage" {
		t.Fatalf("unexpected diagnostic: %q", diagnostic)
	}
	for _, leaked := range []string{storageErr.Error(), "cabinet.db", "SQL statement"} {
		if strings.Contains(diagnostic, leaked) {
			t.Fatalf("diagnostic leaked %q: %q", leaked, diagnostic)
		}
	}
}
