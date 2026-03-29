package app

import (
  "encoding/json"
  "net/http"
  "strings"
  "testing"
)

func TestFutureHooksListDisabledByDefault(t *testing.T) {
  t.Parallel()

  a := newTestApp(t)
  resp := doRequest(t, a, http.MethodGet, "/api/future-hooks", nil, nil)
  if resp.Code != http.StatusOK {
    t.Fatalf("future hooks list status=%d body=%s", resp.Code, resp.Body.String())
  }

  var payload struct {
    Hooks []struct {
      ID        string `json:"id"`
      Status    string `json:"status"`
      Active    bool   `json:"active"`
      Operative bool   `json:"operative"`
    } `json:"hooks"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
    t.Fatalf("decode future hooks list: %v", err)
  }
  if len(payload.Hooks) < 5 {
    t.Fatalf("expected scaffolded future hooks, got %d", len(payload.Hooks))
  }
  for _, hook := range payload.Hooks {
    if hook.Status != "disabled" || hook.Active || hook.Operative {
      t.Fatalf("expected disabled non-operative hook, got %+v", hook)
    }
  }
}

func TestFutureHooksInvokeReturnsExplicitNotActive(t *testing.T) {
  t.Parallel()

  a := newTestApp(t)
  resp := doRequest(t, a, http.MethodPost, "/api/future-hooks/invoke", strings.NewReader(`{"hook_id":"structured-offers"}`), map[string]string{"Content-Type": "application/json"})
  if resp.Code != http.StatusConflict {
    t.Fatalf("future hook invoke status=%d body=%s", resp.Code, resp.Body.String())
  }

  var payload struct {
    Error  string `json:"error"`
    HookID string `json:"hook_id"`
    Status string `json:"status"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
    t.Fatalf("decode future hook invoke payload: %v", err)
  }
  if payload.Error != "hook_not_active" || payload.HookID != "structured-offers" || payload.Status != "disabled" {
    t.Fatalf("unexpected future hook invoke payload: %+v", payload)
  }
}
