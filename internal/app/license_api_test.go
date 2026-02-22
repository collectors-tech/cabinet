package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/licensing"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestLicenseStatusAndFreeTierCap(t *testing.T) {
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
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	for i := 0; i < 150; i++ {
		_, _ = a.db.ExecContext(context.Background(), `INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES (?, 'AFX', 'Slot', ?, ?)`, "seed-"+strconv.Itoa(i), "P-"+strconv.Itoa(i), "T-"+strconv.Itoa(i))
	}
	limitResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"brand":"AFX","category":"Slot","part_number":"PX","title":"Item"}`), map[string]string{"Content-Type": "application/json"})
	if limitResp.Code != http.StatusPaymentRequired {
		t.Fatalf("expected free tier limit 402, got %d body=%s", limitResp.Code, limitResp.Body.String())
	}

	payload := licensing.Payload{
		ProfileID: p.ID,
		Tier:      "pro",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		Features:  []string{"scanner_automation", "price_tracking", "ai_assist"},
	}
	raw, _ := json.Marshal(payload)
	sig := ed25519.Sign(priv, raw)
	importReq := `{"profile_id":"` + p.ID + `","license":{"payload_base64":"` + base64.StdEncoding.EncodeToString(raw) + `","signature_base64":"` + base64.StdEncoding.EncodeToString(sig) + `"}}`
	importResp := doRequest(t, a, http.MethodPost, "/api/license/import", strings.NewReader(importReq), map[string]string{"Content-Type": "application/json"})
	if importResp.Code != http.StatusOK {
		t.Fatalf("license import status=%d body=%s", importResp.Code, importResp.Body.String())
	}
	statusResp := doRequest(t, a, http.MethodGet, "/api/license/status?profile_id="+p.ID, nil, nil)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("license status status=%d body=%s", statusResp.Code, statusResp.Body.String())
	}
}
