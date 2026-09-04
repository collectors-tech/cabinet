package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	browserAuthStatusTimeout = 8 * time.Second
	browserAuthLoginTimeout  = 10 * time.Minute
	browserAuthTurnTimeout   = 2 * time.Minute
)

var browserAuthModelPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,80}$`)

type BrowserAuthStatus struct {
	State         string `json:"state"`
	Authenticated bool   `json:"authenticated"`
	Method        string `json:"method"`
	Message       string `json:"message"`
}

type BrowserAuthTurnRequest struct {
	ProfileID string
	ThreadID  string
	Model     string
	Messages  []AssistantTurnMessage
	Context   map[string]any
}

type BrowserAuthRuntime interface {
	Status(ctx context.Context) (BrowserAuthStatus, error)
	StartLogin(ctx context.Context) (BrowserAuthStatus, error)
	RunAssistantTurn(ctx context.Context, request BrowserAuthTurnRequest) (string, error)
}

type CodexBrowserAuthRuntime struct {
	mu            sync.Mutex
	loginInFlight bool
}

func NewCodexBrowserAuthRuntime() *CodexBrowserAuthRuntime {
	return &CodexBrowserAuthRuntime{}
}

func (r *CodexBrowserAuthRuntime) Status(ctx context.Context) (BrowserAuthStatus, error) {
	executable, err := resolveCodexExecutable()
	if err != nil {
		return BrowserAuthStatus{
			State: "runtime_missing", Method: "chatgpt",
			Message: "Install the OpenAI Codex runtime to use ChatGPT browser sign-in.",
		}, nil
	}
	statusCtx, cancel := boundedBrowserAuthContext(ctx, browserAuthStatusTimeout)
	defer cancel()
	command := exec.CommandContext(statusCtx, executable, "login", "status")
	configureBackgroundCommand(command)
	output, runErr := command.CombinedOutput()
	text := strings.ToLower(strings.TrimSpace(string(output)))
	if runErr == nil && strings.Contains(text, "logged in using chatgpt") {
		return BrowserAuthStatus{
			State: "connected", Authenticated: true, Method: "chatgpt",
			Message: "ChatGPT sign-in is available on this PC.",
		}, nil
	}
	if errors.Is(statusCtx.Err(), context.DeadlineExceeded) {
		return BrowserAuthStatus{State: "unavailable", Method: "chatgpt", Message: "ChatGPT sign-in status timed out. Retry from this PC."}, nil
	}
	r.mu.Lock()
	inFlight := r.loginInFlight
	r.mu.Unlock()
	if inFlight {
		return BrowserAuthStatus{State: "signing_in", Method: "chatgpt", Message: "Finish signing in in the browser window."}, nil
	}
	return BrowserAuthStatus{State: "signed_out", Method: "chatgpt", Message: "Sign in with ChatGPT to use Cabinet Chat."}, nil
}

func (r *CodexBrowserAuthRuntime) StartLogin(ctx context.Context) (BrowserAuthStatus, error) {
	status, err := r.Status(ctx)
	if err != nil || status.Authenticated || status.State == "runtime_missing" {
		return status, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.loginInFlight {
		return BrowserAuthStatus{State: "signing_in", Method: "chatgpt", Message: "Finish signing in in the browser window."}, nil
	}
	executable, err := resolveCodexExecutable()
	if err != nil {
		return BrowserAuthStatus{State: "runtime_missing", Method: "chatgpt", Message: "Install the OpenAI Codex runtime to use ChatGPT browser sign-in."}, nil
	}
	loginCtx, cancel := context.WithTimeout(context.Background(), browserAuthLoginTimeout)
	command := exec.CommandContext(loginCtx, executable, "login")
	configureBackgroundCommand(command)
	command.Stdout = nil
	command.Stderr = nil
	command.Stdin = nil
	if err := command.Start(); err != nil {
		cancel()
		return BrowserAuthStatus{State: "unavailable", Method: "chatgpt", Message: "Cabinet could not open ChatGPT sign-in. Retry from this PC."}, fmt.Errorf("start codex browser login: %w", err)
	}
	r.loginInFlight = true
	go func() {
		_ = command.Wait()
		cancel()
		r.mu.Lock()
		r.loginInFlight = false
		r.mu.Unlock()
	}()
	return BrowserAuthStatus{State: "signing_in", Method: "chatgpt", Message: "Finish signing in in the browser window."}, nil
}

func (r *CodexBrowserAuthRuntime) RunAssistantTurn(ctx context.Context, request BrowserAuthTurnRequest) (string, error) {
	status, err := r.Status(ctx)
	if err != nil {
		return "", err
	}
	if !status.Authenticated {
		return "", assistantProviderClassifiedError{class: "missing_credentials", err: fmt.Errorf("chatgpt browser sign-in is required")}
	}
	model := strings.TrimSpace(request.Model)
	if !browserAuthModelPattern.MatchString(model) {
		return "", assistantProviderClassifiedError{class: "unsupported_model", err: fmt.Errorf("browser auth model is invalid")}
	}
	executable, err := resolveCodexExecutable()
	if err != nil {
		return "", assistantProviderClassifiedError{class: "unhealthy_provider", err: err}
	}
	tempDir, err := os.MkdirTemp("", "cabinet-openai-browser-auth-")
	if err != nil {
		return "", assistantProviderClassifiedError{class: "provider_failure", err: fmt.Errorf("create isolated browser auth workspace")}
	}
	defer os.RemoveAll(tempDir)

	schemaPath := filepath.Join(tempDir, "response.schema.json")
	resultPath := filepath.Join(tempDir, "response.json")
	schema := []byte(`{"type":"object","properties":{"text":{"type":"string","minLength":1}},"required":["text"],"additionalProperties":false}`)
	if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
		return "", assistantProviderClassifiedError{class: "provider_failure", err: fmt.Errorf("prepare browser auth response contract")}
	}
	prompt, err := browserAuthPrompt(request)
	if err != nil {
		return "", assistantProviderClassifiedError{class: "provider_failure", err: fmt.Errorf("prepare browser auth request")}
	}
	turnCtx, cancel := boundedBrowserAuthContext(ctx, browserAuthTurnTimeout)
	defer cancel()
	args := []string{
		"exec", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check",
		"--sandbox", "read-only", "--disable", "shell_tool", "--disable", "browser_use",
		"--disable", "apps", "--disable", "plugins", "--disable", "multi_agent",
		"--disable", "skill_search", "--model", model, "--cd", tempDir,
		"--output-schema", schemaPath, "--output-last-message", resultPath, "-",
	}
	command := exec.CommandContext(turnCtx, executable, args...)
	configureBackgroundCommand(command)
	command.Stdin = strings.NewReader(prompt)
	var diagnostic bytes.Buffer
	command.Stdout = &diagnostic
	command.Stderr = &diagnostic
	if err := command.Run(); err != nil {
		class := classifyAssistantProviderError(classifyAssistantProviderTransportError(err))
		if errors.Is(turnCtx.Err(), context.DeadlineExceeded) {
			class = "timeout"
		}
		return "", assistantProviderClassifiedError{class: class, err: fmt.Errorf("browser-authenticated OpenAI turn failed")}
	}
	raw, err := os.ReadFile(resultPath)
	if err != nil || len(raw) > 64*1024 {
		return "", assistantProviderClassifiedError{class: "provider_failure", err: fmt.Errorf("browser-authenticated OpenAI response unavailable")}
	}
	var response struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &response); err != nil || strings.TrimSpace(response.Text) == "" {
		return "", assistantProviderClassifiedError{class: "provider_failure", err: fmt.Errorf("browser-authenticated OpenAI response invalid")}
	}
	return strings.TrimSpace(response.Text), nil
}

func browserAuthPrompt(request BrowserAuthTurnRequest) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"profile_id": request.ProfileID,
		"thread_id":  request.ThreadID,
		"messages":   request.Messages,
		"context":    request.Context,
	})
	if err != nil {
		return "", err
	}
	plannerContract := ""
	for _, message := range request.Messages {
		if message.Role == "system" && strings.Contains(strings.ToLower(message.Content), "cabinet planner json") {
			plannerContract = " This is a governed planner request. The response schema's text string MUST itself contain only a valid JSON object with fields decision, skill_id, parameters, message, error_code, and next_action. " +
				"Use decision select_skill only with one exact skill_id from context.skills; otherwise use clarify or reject. Never claim you need direct dashboard or database access because Cabinet executes the selected skill after your response."
			break
		}
	}
	return "You are the language-provider transport for Cabinet Chat. Do not call tools, inspect files, browse, run commands, or modify data. " +
		"Cabinet owns all tools and mutations behind explicit preview and confirmation. Follow the supplied system and user messages exactly, using only the supplied conversation and context." + plannerContract +
		" Return JSON matching the outer response schema; its text field is the exact assistant response requested by the supplied messages, not a paraphrase of the task.\n\nCabinet request JSON:\n" + string(payload), nil
}

func resolveCodexExecutable() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("CABINET_CODEX_BINARY")); configured != "" {
		if info, err := os.Stat(configured); err == nil && !info.IsDir() {
			return configured, nil
		}
		return "", fmt.Errorf("configured Codex runtime is unavailable")
	}
	for _, candidate := range codexExecutableCandidates() {
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil
		}
	}
	for _, candidate := range codexInstalledPaths() {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("Codex runtime is not installed")
}

func boundedBrowserAuthContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
