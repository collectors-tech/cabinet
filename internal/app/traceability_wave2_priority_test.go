package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestWave2CloudBootstrapRequiresVerifiedTokenInStrictMode(t *testing.T) {
	t.Setenv("CABINET_CLOUD_AUTH_ENFORCE_SIGNED_TOKENS", "1")
	t.Setenv("CABINET_CLOUD_AUTH_HS256_SECRET", "wave2-secret")
	a := newTestApp(t)

	unsigned := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"e30.eyJzdWIiOiJ1c2VyX3N0cmljdCIsImVtYWlsIjoic3RyaWN0QGV4YW1wbGUuY29tIiwicGxhbiI6InBybyJ9.e30"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if unsigned.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned token expected 401 in strict mode, got %d body=%s", unsigned.Code, unsigned.Body.String())
	}

	signed := mustHS256JWT(t, "wave2-secret", map[string]any{
		"sub":   "user_strict",
		"email": "strict@example.com",
		"plan":  "pro",
	})
	verified := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"`+signed+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if verified.Code != http.StatusOK {
		t.Fatalf("signed token expected 200 in strict mode, got %d body=%s", verified.Code, verified.Body.String())
	}
}

func TestWave2DiagnosticsOptInAndLocalOnlyContracts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	getConfig := doRequest(t, a, http.MethodGet, "/api/diagnostics/config", nil, nil)
	if getConfig.Code != http.StatusOK {
		t.Fatalf("diagnostics config expected 200, got %d body=%s", getConfig.Code, getConfig.Body.String())
	}

	localOnly := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/diagnostics/event",
		strings.NewReader(`{"type":"error","category":"provider","message":"provider timeout","session_id":"s-local"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if localOnly.Code != http.StatusOK {
		t.Fatalf("diagnostics event local-only expected 200, got %d body=%s", localOnly.Code, localOnly.Body.String())
	}
	if !strings.Contains(localOnly.Body.String(), `"remote_sent":false`) {
		t.Fatalf("expected remote_sent=false for opt-out mode, got %s", localOnly.Body.String())
	}
}

func TestWave2DiagnosticsSentryCompatibleEnvelopeWhenOptedIn(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		receivedRaw []byte
	)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		receivedRaw = append([]byte(nil), body...)
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer remote.Close()

	a := newTestApp(t)

	enable := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/diagnostics/config",
		strings.NewReader(`{"remote_opt_in":true,"provider":"sentry","remote_url":"`+remote.URL+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if enable.Code != http.StatusOK {
		t.Fatalf("enable diagnostics expected 200, got %d body=%s", enable.Code, enable.Body.String())
	}

	send := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/diagnostics/event",
		strings.NewReader(`{"type":"error","category":"provider","message":"ebay timeout","session_id":"sess-wave2"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if send.Code != http.StatusOK {
		t.Fatalf("send diagnostics expected 200, got %d body=%s", send.Code, send.Body.String())
	}
	if !strings.Contains(send.Body.String(), `"remote_sent":true`) {
		t.Fatalf("expected remote_sent=true, got %s", send.Body.String())
	}

	mu.Lock()
	payload := append([]byte(nil), receivedRaw...)
	mu.Unlock()
	if len(payload) == 0 {
		t.Fatal("expected remote diagnostics payload")
	}
	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("decode sentry envelope: %v body=%s", err, string(payload))
	}
	for _, required := range []string{"event_id", "timestamp", "message", "level"} {
		if _, ok := envelope[required]; !ok {
			t.Fatalf("missing sentry-compatible field %q in envelope: %+v", required, envelope)
		}
	}
}

func TestWave2ErrorTaxonomyClassificationContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/errors/classify",
		strings.NewReader(`{"error_code":"provider_timeout","message":"ebay request timed out"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("error classify expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode classify payload: %v", err)
	}
	if payload["category"] != "provider" {
		t.Fatalf("expected provider category, got %+v", payload)
	}
	if strings.TrimSpace(stringifyAny(payload["next_action"])) == "" {
		t.Fatalf("expected deterministic next_action in payload: %+v", payload)
	}
}

func mustHS256JWT(t *testing.T, secret string, payload map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	enc := base64.RawURLEncoding
	unsigned := enc.EncodeToString(hb) + "." + enc.EncodeToString(pb)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	sig := enc.EncodeToString(mac.Sum(nil))
	return unsigned + "." + sig
}
