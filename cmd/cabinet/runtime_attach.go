package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/launcher"
)

type runtimeAttachDecision struct {
	Attach bool
	URL    string
	PID    int
	Reason string
}

type runtimeSetupAttachMetadata struct {
	Runtime struct {
		ResolvedURL string `json:"resolvedUrl"`
	} `json:"runtime"`
	Meta struct {
		CurrentURL string `json:"currentUrl"`
	} `json:"meta"`
}

func resolveRunningRuntimeAttach(
	dataDir string,
	pidAlive func(pid int) bool,
	urlHealthy func(runtimeURL string) bool,
) (runtimeAttachDecision, error) {
	pidPath := filepath.Join(strings.TrimSpace(dataDir), "cabinet.pid")
	rawPID, err := os.ReadFile(pidPath)
	if errors.Is(err, os.ErrNotExist) {
		return runtimeAttachDecision{}, nil
	}
	if err != nil {
		return runtimeAttachDecision{}, fmt.Errorf("read pid file: %w", err)
	}

	pid, err := parseRuntimePID(rawPID)
	if err != nil {
		_ = os.Remove(pidPath)
		return runtimeAttachDecision{}, nil
	}

	if !pidAlive(pid) {
		_ = os.Remove(pidPath)
		return runtimeAttachDecision{}, nil
	}

	setupURL, err := readRuntimeURLFromSetupMetadata(dataDir)
	if err != nil {
		_ = os.Remove(pidPath)
		return runtimeAttachDecision{}, nil
	}
	if strings.TrimSpace(setupURL) == "" {
		_ = os.Remove(pidPath)
		return runtimeAttachDecision{}, nil
	}
	if !urlHealthy(setupURL) {
		_ = os.Remove(pidPath)
		return runtimeAttachDecision{}, nil
	}

	return runtimeAttachDecision{
		Attach: true,
		URL:    setupURL,
		PID:    pid,
	}, nil
}

func parseRuntimePID(raw []byte) (int, error) {
	pidRaw := strings.TrimSpace(string(raw))
	if pidRaw == "" {
		return 0, fmt.Errorf("empty pid")
	}
	pid, err := strconv.Atoi(pidRaw)
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid pid")
	}
	return pid, nil
}

func readRuntimeURLFromSetupMetadata(dataDir string) (string, error) {
	setupPath := filepath.Join(strings.TrimSpace(dataDir), "cabinet.json")
	raw, err := os.ReadFile(setupPath)
	if err != nil {
		return "", err
	}
	var payload runtimeSetupAttachMetadata
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	resolved := strings.TrimSpace(payload.Runtime.ResolvedURL)
	if resolved != "" {
		return resolved, nil
	}
	return strings.TrimSpace(payload.Meta.CurrentURL), nil
}

func isRuntimeProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").CombinedOutput()
		if err != nil {
			return false
		}
		text := strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(text), "no tasks are running") {
			return false
		}
		return strings.Contains(text, fmt.Sprintf("\"%d\"", pid))
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func isRuntimeEndpointHealthy(runtimeURL string) bool {
	baseURL, err := url.Parse(strings.TrimSpace(runtimeURL))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return false
	}
	baseURL.Path = "/healthz"
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	req, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 350 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func runtimeAttachLogLine(decision runtimeAttachDecision, dataDir string) string {
	resolvedPort := 0
	if parsed, err := url.Parse(strings.TrimSpace(decision.URL)); err == nil {
		if port, err := strconv.Atoi(strings.TrimSpace(parsed.Port())); err == nil {
			resolvedPort = port
		}
	}
	return fmt.Sprintf(
		"CABINET_RUNTIME_ATTACH url=%s pid=%d data_dir=%s resolved_port=%d reason=%s",
		strings.TrimSpace(decision.URL),
		decision.PID,
		strings.TrimSpace(dataDir),
		resolvedPort,
		attachReasonOrDefault(decision.Reason),
	)
}

func resolveRequestedRuntimeAttach(
	cfg config.Config,
	allowParallel bool,
	urlHealthy func(runtimeURL string) bool,
) runtimeAttachDecision {
	if allowParallel {
		return runtimeAttachDecision{}
	}

	requestedURL := launcher.StartupURLFromAddr(cfg.Addr)
	if strings.TrimSpace(requestedURL) == "" {
		return runtimeAttachDecision{}
	}
	if !urlHealthy(requestedURL) {
		return runtimeAttachDecision{}
	}

	return runtimeAttachDecision{
		Attach: true,
		URL:    requestedURL,
		PID:    0,
		Reason: "requested_endpoint_healthy",
	}
}

func attachReasonOrDefault(reason string) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return "same_data_dir"
	}
	return trimmed
}
