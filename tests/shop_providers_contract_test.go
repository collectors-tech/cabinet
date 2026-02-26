package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/app"
	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestEbayProviderResponseContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		serverStatus         int
		serverBody           string
		token                string
		expectErrContains    string
		expectCandidateCount int
	}{
		{
			name:                 "success returns normalized candidate fields",
			serverStatus:         http.StatusOK,
			serverBody:           `{"itemSummaries":[{"itemId":"v1|123|0","title":"AFX Mega G+ Camaro","price":{"value":"45.00"},"itemWebUrl":"https://ebay/item/123","image":{"imageUrl":"https://img/123.jpg"},"seller":{"username":"slot-seller"}}]}`,
			token:                "token-ok",
			expectCandidateCount: 1,
		},
		{
			name:              "non-2xx status returns explicit error",
			serverStatus:      http.StatusServiceUnavailable,
			serverBody:        `{"error":"service_unavailable"}`,
			token:             "token-ok",
			expectErrContains: "ebay search status: 503",
		},
		{
			name:              "missing token is validation failure",
			serverStatus:      http.StatusOK,
			serverBody:        `{"itemSummaries":[]}`,
			token:             "",
			expectErrContains: "missing ebay bearer token",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Accept"); got != "application/json" {
					t.Fatalf("expected Accept application/json, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				_, _ = w.Write([]byte(tt.serverBody))
			}))
			t.Cleanup(srv.Close)

			provider := ebay.NewProvider(ebay.ProviderConfig{
				BaseURL:     srv.URL,
				BearerToken: tt.token,
				Marketplace: "EBAY_AU",
			})

			out, err := provider.Search(context.Background(), scanner.QuerySet{
				Name:       "eBay Contract",
				Keywords:   []string{"afx", "camaro"},
				MaxPrice:   100,
				Exclusions: []string{"damaged"},
				Region:     "AU",
			})
			if tt.expectErrContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectErrContains)
				}
				if !strings.Contains(err.Error(), tt.expectErrContains) {
					t.Fatalf("expected error containing %q, got %q", tt.expectErrContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected provider error: %v", err)
			}
			if len(out) != tt.expectCandidateCount {
				t.Fatalf("expected %d candidates, got %d", tt.expectCandidateCount, len(out))
			}
			if len(out) == 0 {
				return
			}
			got := out[0]
			if got.ListingID == "" || got.Title == "" || got.URL == "" || got.Seller == "" {
				t.Fatalf("required fields missing in normalized candidate: %+v", got)
			}
			if got.Price != 45.00 {
				t.Fatalf("expected normalized numeric price=45.00, got %.2f", got.Price)
			}
			if got.Source != "ebay" {
				t.Fatalf("expected source=ebay, got %q", got.Source)
			}
		})
	}
}

func TestAmazonDisabledModeReturns409ContractEnvelope(t *testing.T) {
	t.Parallel()

	// API-contract harness for disabled provider mode using scanner registry.
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reg := scanner.NewProviderRegistry()
		provider, err := reg.Provider("amazon")
		if err != nil {
			http.Error(w, `{"error_code":"PROVIDER_RESOLUTION_FAILED"}`, http.StatusInternalServerError)
			return
		}
		_, searchErr := provider.Search(context.Background(), scanner.QuerySet{
			Name:     "Amazon disabled test",
			Keywords: []string{"slot car"},
		})
		if searchErr != nil && strings.Contains(strings.ToLower(searchErr.Error()), "disabled") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code":  "PROVIDER_DISABLED",
				"provider":    "amazon",
				"message":     "Amazon provider is disabled for this profile",
				"next_action": "enable_provider_or_choose_supported_source",
			})
			return
		}
		http.Error(w, `{"error_code":"UNEXPECTED_PROVIDER_STATE"}`, http.StatusInternalServerError)
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/providers/amazon/scan")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode 409 payload: %v", err)
	}
	required := []string{"error_code", "provider", "message", "next_action"}
	for _, k := range required {
		v, ok := payload[k]
		if !ok {
			t.Fatalf("missing required error field %q in payload: %#v", k, payload)
		}
		if _, ok := v.(string); !ok || strings.TrimSpace(v.(string)) == "" {
			t.Fatalf("field %q must be non-empty string, got %#v", k, v)
		}
	}
	if payload["error_code"] != "PROVIDER_DISABLED" {
		t.Fatalf("expected error_code PROVIDER_DISABLED, got %#v", payload["error_code"])
	}
	if payload["provider"] != "amazon" {
		t.Fatalf("expected provider amazon, got %#v", payload["provider"])
	}
}

func TestAmazonProviderHealthAppRouterPathFeasibility(t *testing.T) {
	t.Parallel()

	baseURL, shutdown := startTestRuntimeApp(t)
	t.Cleanup(shutdown)

	resp, err := http.Get(baseURL + "/api/provider/health?provider=amazon")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 from app router provider health path, got %d body=%s", resp.StatusCode, string(b))
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode provider health payload: %v", err)
	}
	for _, k := range []string{"provider", "status", "message"} {
		v, ok := payload[k]
		if !ok {
			t.Fatalf("missing %q in provider health payload: %#v", k, payload)
		}
		if _, ok := v.(string); !ok {
			t.Fatalf("expected string %q, got %#v", k, v)
		}
	}
	if payload["provider"] != "amazon" {
		t.Fatalf("expected provider amazon, got %#v", payload["provider"])
	}
}

func TestAUWebshopThrottlingConformanceOPS001_RegionAU(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("open and migrate db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	svc := scanner.NewService(conn)
	qs, err := svc.CreateQuerySet(context.Background(), scanner.QuerySet{
		Name:          "AU Webshop Throttle",
		Keywords:      []string{"afx", "mega g"},
		Region:        "AU",
		Enabled:       true,
		RateLimitRPS:  1000, // service enforces a deterministic 100ms minimum backoff.
		MaxRetryCount: 2,
	})
	if err != nil {
		t.Fatalf("create query set: %v", err)
	}

	provider := &alwaysFailProvider{err: errors.New("429 too many requests")}
	start := time.Now()
	run, runErr := svc.RunNow(context.Background(), qs.ID, provider)
	elapsed := time.Since(start)

	if runErr == nil {
		t.Fatal("expected run error due to throttled provider")
	}
	if !strings.Contains(strings.ToLower(runErr.Error()), "run failed") {
		t.Fatalf("expected wrapped run failure error, got %v", runErr)
	}
	if provider.calls != 3 {
		t.Fatalf("expected 3 provider attempts (1 + max_retry_count), got %d", provider.calls)
	}
	if run.Attempts != 3 {
		t.Fatalf("expected attempts=3, got %+v", run)
	}
	// Retry sleeps are deterministic: 100ms*1 + 100ms*2 = >= 300ms total.
	if elapsed < 280*time.Millisecond {
		t.Fatalf("expected deterministic throttling backoff >=280ms, got %s", elapsed)
	}

	health, err := svc.ProviderHealth(context.Background(), "ebay")
	if err != nil {
		t.Fatalf("provider health: %v", err)
	}
	if health["status"] != "error" {
		t.Fatalf("expected degraded/error provider state after repeated throttling failures, got %+v", health)
	}
	if strings.TrimSpace(health["message"]) == "" {
		t.Fatalf("expected provider health message populated after throttling failures, got %+v", health)
	}

	failures, err := svc.ListFailures(context.Background())
	if err != nil {
		t.Fatalf("list failures: %v", err)
	}
	if len(failures) != 3 {
		t.Fatalf("expected 3 failure records (one per attempt), got %d", len(failures))
	}
	for i, f := range failures {
		if !strings.Contains(strings.ToLower(f["message"]), "429") {
			t.Fatalf("failure[%d] message must preserve throttling signal, got %#v", i, f)
		}
	}
}

func startTestRuntimeApp(t *testing.T) (string, func()) {
	t.Helper()

	_ = os.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	base := t.TempDir()
	addr := allocateLoopbackAddr(t)
	cfg := config.Config{
		Addr:           addr,
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://" + addr,
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	runtimeApp, err := app.New(cfg)
	if err != nil {
		t.Fatalf("app.New() failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runtimeApp.Run(ctx)
	}()
	baseURL := "http://" + addr
	waitForHealthz(t, baseURL)
	return baseURL, func() {
		cancel()
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("app shutdown returned error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timed out waiting for app shutdown")
		}
	}
}

func allocateLoopbackAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate port: %v", err)
	}
	defer l.Close()
	return l.Addr().String()
}

func waitForHealthz(t *testing.T, baseURL string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("healthz did not become ready at %s", baseURL)
}

type alwaysFailProvider struct {
	calls int
	err   error
}

func (p *alwaysFailProvider) Search(context.Context, scanner.QuerySet) ([]scanner.CandidateInput, error) {
	p.calls++
	return nil, fmt.Errorf("provider_throttle: %w", p.err)
}
