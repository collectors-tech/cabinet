package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/licensing"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestAIAssistEndpoints(t *testing.T) {
	t.Parallel()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey error: %v", err)
	}
	cfg := config.Config{
		Addr:            "127.0.0.1:0",
		DataDir:         t.TempDir(),
		DBPath:          filepath.Join(t.TempDir(), "cabinet.db"),
		UpdateChannel:   update.ChannelStable,
		UpdatePublicKey: base64.StdEncoding.EncodeToString(pub),
		WebAuthnRPID:    "127.0.0.1",
		WebAuthnOrigin:  "http://127.0.0.1:8080",
		WebAuthnName:    "Cabinet Test",
		BackupInterval:  60,
	}
	a := newTestAppWithConfig(t, cfg)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"part_number\":\"P-1\",\"brand\":\"AFX\",\"title\":\"AFX P-1\",\"confidence\":0.9}"}}]}`))
	}))
	defer aiServer.Close()

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{"settings":{"openai_base_url":"`+aiServer.URL+`","ai_enabled":"true"}}`), map[string]string{"Content-Type": "application/json"})
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-test"}`), map[string]string{"Content-Type": "application/json"})

	payload := licensing.Payload{
		ProfileID: p.ID,
		Tier:      "pro",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		Features:  []string{"ai_assist"},
	}
	raw, _ := json.Marshal(payload)
	sig := ed25519.Sign(priv, raw)
	importReq := `{"profile_id":"` + p.ID + `","license":{"payload_base64":"` + base64.StdEncoding.EncodeToString(raw) + `","signature_base64":"` + base64.StdEncoding.EncodeToString(sig) + `"}}`
	importResp := doRequest(t, a, http.MethodPost, "/api/license/import", strings.NewReader(importReq), map[string]string{"Content-Type": "application/json"})
	if importResp.Code != http.StatusOK {
		t.Fatalf("license import status=%d body=%s", importResp.Code, importResp.Body.String())
	}

	aiTest := doRequest(t, a, http.MethodPost, "/api/ai/test", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if aiTest.Code != http.StatusOK {
		t.Fatalf("ai test status=%d body=%s", aiTest.Code, aiTest.Body.String())
	}
	aiSuggest := doRequest(t, a, http.MethodPost, "/api/ai/suggest/title", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"AFX P-1"}`), map[string]string{"Content-Type": "application/json"})
	if aiSuggest.Code != http.StatusOK {
		t.Fatalf("ai suggest status=%d body=%s", aiSuggest.Code, aiSuggest.Body.String())
	}
}

func TestAIAssistApplyRequiresExplicitConfirmation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	withoutConfirm := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/ai/suggestions/apply",
		strings.NewReader(`{"profile_id":"p1","suggestion_id":"sug_123","draft_id":"draft_42"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if withoutConfirm.Code != http.StatusConflict {
		t.Fatalf("expected 409 without confirm token, got %d body=%s", withoutConfirm.Code, withoutConfirm.Body.String())
	}
	var blocked map[string]any
	if err := json.Unmarshal(withoutConfirm.Body.Bytes(), &blocked); err != nil {
		t.Fatalf("decode blocked payload: %v", err)
	}
	if blocked["error_code"] != "AI_CONFIRM_REQUIRED" {
		t.Fatalf("expected AI_CONFIRM_REQUIRED, got %v", blocked["error_code"])
	}

	withConfirm := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/ai/suggestions/apply",
		strings.NewReader(`{"profile_id":"p1","suggestion_id":"sug_123","draft_id":"draft_42","confirm_token":"confirm_abc"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if withConfirm.Code != http.StatusOK {
		t.Fatalf("expected 200 with confirm token, got %d body=%s", withConfirm.Code, withConfirm.Body.String())
	}
}
