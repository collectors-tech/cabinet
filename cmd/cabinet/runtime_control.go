package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/launcher"
)

type runtimeEndpointProbe struct {
	URL       string
	Status    string
	PortInUse bool
	PID       int
	DataDir   string
}

type runtimeShutdownResponse struct {
	OK     bool   `json:"ok"`
	PID    int    `json:"pid"`
	Reason string `json:"reason"`
}

type runtimeRestartResult struct {
	URL          string
	PID          int
	Forced       bool
	Restarted    bool
	EndpointDown bool
}

func startupWantsRestart() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CABINET_RESTART"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func resolveRequestedRuntimeProbe(
	cfg config.Config,
	inspect func(string) (runtimeEndpointProbe, error),
	portInUse func(string) bool,
) runtimeEndpointProbe {
	requestedURL := launcher.StartupURLFromAddr(cfg.Addr)
	probe := runtimeEndpointProbe{
		URL: requestedURL,
	}

	if inspected, err := inspect(requestedURL); err == nil {
		inspected.URL = requestedURL
		inspected.Status = "cabinet"
		inspected.PortInUse = true
		return inspected
	}

	if portInUse(cfg.Addr) {
		probe.Status = "occupied"
		probe.PortInUse = true
		return probe
	}

	probe.Status = "free"
	return probe
}

func fetchRuntimeEndpointProbe(runtimeURL string) (runtimeEndpointProbe, error) {
	if !isRuntimeEndpointHealthy(runtimeURL) {
		return runtimeEndpointProbe{}, fmt.Errorf("runtime endpoint unhealthy")
	}

	baseURL, err := url.Parse(strings.TrimSpace(runtimeURL))
	if err != nil {
		return runtimeEndpointProbe{}, fmt.Errorf("parse runtime url: %w", err)
	}
	baseURL.Path = "/api/runtime"
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	req, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return runtimeEndpointProbe{}, fmt.Errorf("new runtime request: %w", err)
	}
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return runtimeEndpointProbe{}, fmt.Errorf("get runtime endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return runtimeEndpointProbe{}, fmt.Errorf("runtime status %d", resp.StatusCode)
	}

	var payload struct {
		AppVersion  string `json:"app_version"`
		RuntimePort int    `json:"runtime_port"`
		PID         int    `json:"pid"`
		DataDir     string `json:"data_dir"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return runtimeEndpointProbe{}, fmt.Errorf("decode runtime payload: %w", err)
	}
	if payload.RuntimePort <= 0 {
		return runtimeEndpointProbe{}, fmt.Errorf("runtime payload missing port")
	}

	return runtimeEndpointProbe{
		URL:       strings.TrimSpace(runtimeURL),
		Status:    "cabinet",
		PortInUse: true,
		PID:       payload.PID,
		DataDir:   strings.TrimSpace(payload.DataDir),
	}, nil
}

func isRuntimeAddrInUse(addr string) bool {
	dialAddr := normalizeStartupDialAddr(addr)
	conn, err := net.DialTimeout("tcp", dialAddr, 250*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func normalizeStartupDialAddr(addr string) string {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return net.JoinHostPort("127.0.0.1", "17880")
	}
	if strings.TrimSpace(host) == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	if strings.TrimSpace(port) == "" {
		port = "17880"
	}
	return net.JoinHostPort(host, port)
}

func runtimeEndpointStatusLogLine(probe runtimeEndpointProbe) string {
	return fmt.Sprintf(
		"CABINET_RUNTIME_ENDPOINT_STATUS requested_url=%s status=%s port_in_use=%t pid=%d data_dir=%s",
		strings.TrimSpace(probe.URL),
		strings.TrimSpace(defaultRuntimeEndpointStatus(probe.Status)),
		probe.PortInUse,
		probe.PID,
		strings.TrimSpace(probe.DataDir),
	)
}

func runtimeRestartLogLine(result runtimeRestartResult, dataDir string) string {
	return fmt.Sprintf(
		"CABINET_RUNTIME_RESTART url=%s pid=%d forced=%t restarted=%t endpoint_down=%t data_dir=%s",
		strings.TrimSpace(result.URL),
		result.PID,
		result.Forced,
		result.Restarted,
		result.EndpointDown,
		strings.TrimSpace(dataDir),
	)
}

func defaultRuntimeEndpointStatus(status string) string {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}

func requestRuntimeShutdown(runtimeURL, reason string) (runtimeShutdownResponse, error) {
	baseURL, err := url.Parse(strings.TrimSpace(runtimeURL))
	if err != nil {
		return runtimeShutdownResponse{}, fmt.Errorf("parse runtime url: %w", err)
	}
	baseURL.Path = "/api/runtime/shutdown"
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	body, err := json.Marshal(map[string]string{
		"reason": strings.TrimSpace(reason),
	})
	if err != nil {
		return runtimeShutdownResponse{}, fmt.Errorf("marshal shutdown request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL.String(), bytes.NewReader(body))
	if err != nil {
		return runtimeShutdownResponse{}, fmt.Errorf("new shutdown request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return runtimeShutdownResponse{}, fmt.Errorf("post runtime shutdown: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return runtimeShutdownResponse{}, fmt.Errorf("runtime shutdown status %d", resp.StatusCode)
	}

	var payload runtimeShutdownResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return runtimeShutdownResponse{}, fmt.Errorf("decode shutdown response: %w", err)
	}
	return payload, nil
}

func restartRequestedRuntime(
	probe runtimeEndpointProbe,
	requestedAddr string,
	shutdownRequester func(string, string) (runtimeShutdownResponse, error),
	waitForReady func(string, string, time.Duration) bool,
	pidAlive func(int) bool,
	terminate func(int) error,
) (runtimeRestartResult, error) {
	result := runtimeRestartResult{
		URL: probe.URL,
		PID: probe.PID,
	}

	shutdownResp, shutdownErr := shutdownRequester(probe.URL, "restart")
	if shutdownResp.PID > 0 {
		result.PID = shutdownResp.PID
	}
	if shutdownErr == nil {
		if waitForReady(probe.URL, requestedAddr, 5*time.Second) {
			result.Restarted = true
			result.EndpointDown = true
			return result, nil
		}
	}

	if result.PID <= 0 || !pidAlive(result.PID) {
		return result, fmt.Errorf("restart handoff did not release endpoint cleanly")
	}
	if err := terminate(result.PID); err != nil {
		return result, fmt.Errorf("terminate runtime pid %d: %w", result.PID, err)
	}
	result.Forced = true
	if !waitForReady(probe.URL, requestedAddr, 5*time.Second) {
		return result, fmt.Errorf("restarted runtime endpoint did not clear after forced termination")
	}
	result.Restarted = true
	result.EndpointDown = true
	return result, nil
}

func waitForRuntimeRestartReady(runtimeURL, requestedAddr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isRuntimeEndpointHealthy(runtimeURL) && !isRuntimeAddrInUse(requestedAddr) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func terminateRuntimeProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid")
	}
	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F").Run()
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGKILL)
}
