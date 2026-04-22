package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestResolveRequestedRuntimeProbeCabinetVsOccupiedVsFree(t *testing.T) {
	t.Parallel()

	cfg := config.Config{Addr: "127.0.0.1:17882", Host: "127.0.0.1", Port: 17882}

	cabinetProbe := resolveRequestedRuntimeProbe(
		cfg,
		func(runtimeURL string) (runtimeEndpointProbe, error) {
			return runtimeEndpointProbe{URL: runtimeURL, PID: 4242, DataDir: "C:/cabinet/data"}, nil
		},
		func(string) bool { return true },
	)
	if cabinetProbe.Status != "cabinet" || !cabinetProbe.PortInUse || cabinetProbe.PID != 4242 {
		t.Fatalf("expected cabinet probe, got %+v", cabinetProbe)
	}

	occupiedProbe := resolveRequestedRuntimeProbe(
		cfg,
		func(string) (runtimeEndpointProbe, error) {
			return runtimeEndpointProbe{}, http.ErrServerClosed
		},
		func(string) bool { return true },
	)
	if occupiedProbe.Status != "occupied" || !occupiedProbe.PortInUse {
		t.Fatalf("expected occupied probe, got %+v", occupiedProbe)
	}

	freeProbe := resolveRequestedRuntimeProbe(
		cfg,
		func(string) (runtimeEndpointProbe, error) {
			return runtimeEndpointProbe{}, http.ErrServerClosed
		},
		func(string) bool { return false },
	)
	if freeProbe.Status != "free" || freeProbe.PortInUse {
		t.Fatalf("expected free probe, got %+v", freeProbe)
	}
}

func TestFetchRuntimeEndpointProbeReadsRuntimeMetadata(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/runtime", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"app_version":  "rev-test",
			"runtime_port": 19991,
			"pid":          5511,
			"data_dir":     "C:/cabinet/runtime",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	probe, err := fetchRuntimeEndpointProbe(server.URL)
	if err != nil {
		t.Fatalf("fetchRuntimeEndpointProbe error: %v", err)
	}
	if probe.Status != "cabinet" || probe.PID != 5511 || probe.DataDir != "C:/cabinet/runtime" {
		t.Fatalf("unexpected probe: %+v", probe)
	}
}

func TestRestartRequestedRuntimeFallsBackToForcedTermination(t *testing.T) {
	t.Parallel()

	probe := runtimeEndpointProbe{
		URL: "http://127.0.0.1:17882/",
		PID: 4040,
	}
	shutdownCalls := 0
	terminateCalls := 0
	waitCalls := 0

	result, err := restartRequestedRuntime(
		probe,
		"127.0.0.1:17882",
		func(string, string) (runtimeShutdownResponse, error) {
			shutdownCalls++
			return runtimeShutdownResponse{OK: true, PID: 4040}, nil
		},
		func(string, string, time.Duration) bool {
			waitCalls++
			return waitCalls > 1
		},
		func(pid int) bool { return pid == 4040 },
		func(pid int) error {
			terminateCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("restartRequestedRuntime error: %v", err)
	}
	if shutdownCalls != 1 || terminateCalls != 1 {
		t.Fatalf("expected 1 shutdown call and 1 terminate call, got shutdown=%d terminate=%d", shutdownCalls, terminateCalls)
	}
	if !result.Forced || !result.Restarted || !result.EndpointDown {
		t.Fatalf("expected forced successful restart result, got %+v", result)
	}
}

func TestRuntimeEndpointStatusLogLineIncludesPidAndStatus(t *testing.T) {
	t.Parallel()

	line := runtimeEndpointStatusLogLine(runtimeEndpointProbe{
		URL:       "http://127.0.0.1:17882/",
		Status:    "cabinet",
		PortInUse: true,
		PID:       1234,
		DataDir:   "C:/cabinet/data",
	})
	for _, token := range []string{
		"CABINET_RUNTIME_ENDPOINT_STATUS",
		"status=cabinet",
		"port_in_use=true",
		"pid=1234",
		"data_dir=C:/cabinet/data",
	} {
		if !strings.Contains(line, token) {
			t.Fatalf("expected token %q in line %q", token, line)
		}
	}
}
