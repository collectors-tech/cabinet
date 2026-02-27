package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestWave3ProGatedMutationEndpointsReturnDeterministic403AndNoSideEffects(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave3 Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('w3-item-1', ?, 'AFX','Slot','W3-1','Wave3 Item')`, p.ID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, enabled) VALUES ('w3-qs-1', ?, 'W3 Q', '[\"afx\"]', '[]', 1)`, p.ID); err != nil {
		t.Fatalf("seed query set: %v", err)
	}

	freeToken := "e30.eyJzdWIiOiJ1c2VyX3czIiwiZW1haWwiOiJ3M0BleGFtcGxlLmNvbSIsInBsYW4iOiJmcmVlIn0.e30"
	bootstrap := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"`+freeToken+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap free status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	assertForbidden := func(feature string, respCode int, body string) {
		t.Helper()
		if respCode != http.StatusForbidden {
			t.Fatalf("feature %s expected 403, got %d body=%s", feature, respCode, body)
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			t.Fatalf("decode forbidden payload for %s: %v body=%s", feature, err, body)
		}
		if payload["error"] != "forbidden" {
			t.Fatalf("feature %s expected error=forbidden got %+v", feature, payload)
		}
		if payload["error_code"] != "PRO_FEATURE_REQUIRED" {
			t.Fatalf("feature %s expected error_code PRO_FEATURE_REQUIRED got %+v", feature, payload)
		}
		if payload["feature"] != feature {
			t.Fatalf("feature %s expected payload feature %s got %+v", feature, feature, payload)
		}
		if strings.TrimSpace(stringifyAny(payload["message"])) == "" {
			t.Fatalf("feature %s expected non-empty message got %+v", feature, payload)
		}
	}

	denyTrack := doRequest(t, a, http.MethodPost, "/api/pricing/track", strings.NewReader(`{"item_id":"w3-item-1"}`), map[string]string{"Content-Type": "application/json"})
	assertForbidden("price_tracking", denyTrack.Code, denyTrack.Body.String())

	var trackedCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM tracked_items WHERE item_id = 'w3-item-1'`).Scan(&trackedCount); err != nil {
		t.Fatalf("count tracked item: %v", err)
	}
	if trackedCount != 0 {
		t.Fatalf("expected no tracked item side effect, got %d", trackedCount)
	}

	denyScanner := doRequest(t, a, http.MethodPost, "/api/scanner/run/scheduled", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	assertForbidden("scanner_automation", denyScanner.Code, denyScanner.Body.String())

	denyAI := doRequest(t, a, http.MethodPost, "/api/ai/test", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	assertForbidden("ai_assist", denyAI.Code, denyAI.Body.String())
}

func TestWave3RuntimeUpdateVerificationHarness(t *testing.T) {
	t.Parallel()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
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

	payload := []byte(`{"version":"1.2.3","notes":"wave3 update contract"}`)
	sig := ed25519.Sign(priv, payload)
	payloadB64 := base64.StdEncoding.EncodeToString(payload)
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	okResp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/update/install",
		strings.NewReader(`{"payload_base64":"`+payloadB64+`","signature_base64":"`+sigB64+`","manifest_channel":"stable"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if okResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid signed update install, got %d body=%s", okResp.Code, okResp.Body.String())
	}

	invalidSig := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/update/install",
		strings.NewReader(`{"payload_base64":"`+payloadB64+`","signature_base64":"invalid","manifest_channel":"stable"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if invalidSig.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid signature, got %d body=%s", invalidSig.Code, invalidSig.Body.String())
	}
	if !strings.Contains(invalidSig.Body.String(), `"error_code":"INVALID_UPDATE_SIGNATURE"`) {
		t.Fatalf("expected INVALID_UPDATE_SIGNATURE envelope, got %s", invalidSig.Body.String())
	}

	channelMismatch := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/update/install",
		strings.NewReader(`{"payload_base64":"`+payloadB64+`","signature_base64":"`+sigB64+`","manifest_channel":"beta"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if channelMismatch.Code != http.StatusConflict {
		t.Fatalf("expected 409 for channel mismatch, got %d body=%s", channelMismatch.Code, channelMismatch.Body.String())
	}
	if !strings.Contains(channelMismatch.Body.String(), `"error_code":"UPDATE_CHANNEL_MISMATCH"`) {
		t.Fatalf("expected UPDATE_CHANNEL_MISMATCH envelope, got %s", channelMismatch.Body.String())
	}
}
