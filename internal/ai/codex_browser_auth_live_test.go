package ai

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCodexBrowserAuthRuntimeLiveChatGPTTurn(t *testing.T) {
	if os.Getenv("CABINET_TEST_OPENAI_BROWSER_AUTH_LIVE") != "1" {
		t.Skip("set CABINET_TEST_OPENAI_BROWSER_AUTH_LIVE=1 for the user-present ChatGPT proof")
	}
	runtime := NewCodexBrowserAuthRuntime()
	status, err := runtime.Status(context.Background())
	if err != nil {
		t.Fatalf("read ChatGPT browser-auth status: %v", err)
	}
	if !status.Authenticated || status.Method != "chatgpt" {
		t.Fatalf("expected authenticated ChatGPT browser login, got %#v", status)
	}
	text, err := runtime.RunAssistantTurn(context.Background(), BrowserAuthTurnRequest{
		ProfileID: "live-browser-auth-proof",
		ThreadID:  "live-browser-auth-proof",
		Model:     "gpt-5.6-luna",
		Messages: []AssistantTurnMessage{{
			Role: "user", Content: "Reply with exactly: CABINET_CHATGPT_BROWSER_AUTH_OK",
		}},
		Context: map[string]any{"purpose": "live_provider_test", "actions_allowed": false},
	})
	if err != nil {
		t.Fatalf("run browser-authenticated ChatGPT turn: %v", err)
	}
	if !strings.Contains(text, "CABINET_CHATGPT_BROWSER_AUTH_OK") {
		t.Fatalf("unexpected browser-auth response %q", text)
	}
}

func TestCodexBrowserAuthRuntimeLiveGovernedPlannerTurn(t *testing.T) {
	if os.Getenv("CABINET_TEST_OPENAI_BROWSER_AUTH_LIVE") != "1" {
		t.Skip("set CABINET_TEST_OPENAI_BROWSER_AUTH_LIVE=1 for the user-present ChatGPT proof")
	}
	runtime := NewCodexBrowserAuthRuntime()
	text, err := runtime.RunAssistantTurn(context.Background(), BrowserAuthTurnRequest{
		ProfileID: "live-browser-auth-planner-proof",
		ThreadID:  "live-browser-auth-planner-proof",
		Model:     "gpt-5.6-luna",
		Messages: []AssistantTurnMessage{
			{Role: "system", Content: "Return only structured Cabinet planner JSON. Cabinet owns all tool dispatch and mutation authority."},
			{Role: "user", Content: "Summarize my dashboard."},
		},
		Context: map[string]any{
			"skills": []map[string]any{{
				"skill_id":       "cabinet.dashboard.summarise_activity",
				"category":       "dashboard",
				"classification": "read_only",
				"parameters":     map[string]any{},
			}},
		},
	})
	if err != nil {
		t.Fatalf("run browser-authenticated planner turn: %v", err)
	}
	var selection struct {
		Decision string `json:"decision"`
		SkillID  string `json:"skill_id"`
	}
	if err := json.Unmarshal([]byte(text), &selection); err != nil {
		t.Fatalf("planner text is not structured selection JSON: %v; text=%q", err, text)
	}
	if selection.Decision != "select_skill" || selection.SkillID != "cabinet.dashboard.summarise_activity" {
		t.Fatalf("unexpected governed planner selection: %#v", selection)
	}
}
