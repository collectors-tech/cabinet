package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestZitadelBillingWebhookRejectsInvalidSignature(t *testing.T) {
	_ = os.Setenv("CABINET_ZITADEL_WEBHOOK_SECRET", "test-secret")
	t.Cleanup(func() { _ = os.Unsetenv("CABINET_ZITADEL_WEBHOOK_SECRET") })
	a := newTestApp(t)

	body := `{"type":"subscription.updated","data":{"user_id":"user_sig","plan":"pro"}}`
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/zitadel/webhook",
		strings.NewReader(body),
		map[string]string{
			"Content-Type":                "application/json",
			"X-Cabinet-Webhook-Signature": "invalid-signature",
		},
	)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid signature, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestZitadelBillingWebhookRejectsRetiredClerkRoute(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("CABINET_ZITADEL_WEBHOOK_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("CABINET_ZITADEL_WEBHOOK_SECRET") })
	a := newTestApp(t)

	body := `{"type":"subscription.updated","data":{"user_id":"user_sig","plan":"pro"}}`
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/clerk/webhook",
		strings.NewReader(body),
		map[string]string{
			"Content-Type":                "application/json",
			"X-Cabinet-Webhook-Signature": hmacSignature(secret, body),
		},
	)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("retired clerk webhook expected 404, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestZitadelBillingWebhookAppliesPlanTransitions(t *testing.T) {
	secret := "test-secret"
	_ = os.Setenv("CABINET_ZITADEL_WEBHOOK_SECRET", secret)
	t.Cleanup(func() { _ = os.Unsetenv("CABINET_ZITADEL_WEBHOOK_SECRET") })
	a := newTestApp(t)

	bootstrapToken := "e30.eyJzdWIiOiJ1c2VyX2JpbGxpbmciLCJlbWFpbCI6ImJpbGxpbmdAZXhhbXBsZS5jb20iLCJwbGFuIjoiZnJlZSJ9.e30"
	before := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"`+bootstrapToken+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if before.Code != http.StatusOK || !strings.Contains(before.Body.String(), `"plan":"free"`) {
		t.Fatalf("expected free plan before webhook, got status=%d body=%s", before.Code, before.Body.String())
	}

	upgradeBody := `{"type":"subscription.updated","data":{"user_id":"user_billing","plan":"pro"}}`
	upgrade := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/zitadel/webhook",
		strings.NewReader(upgradeBody),
		map[string]string{
			"Content-Type":                "application/json",
			"X-Cabinet-Webhook-Signature": hmacSignature(secret, upgradeBody),
		},
	)
	if upgrade.Code != http.StatusOK {
		t.Fatalf("upgrade webhook expected 200, got %d body=%s", upgrade.Code, upgrade.Body.String())
	}

	afterUpgrade := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"`+bootstrapToken+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if afterUpgrade.Code != http.StatusOK || !strings.Contains(afterUpgrade.Body.String(), `"plan":"pro"`) {
		t.Fatalf("expected pro plan after upgrade, got status=%d body=%s", afterUpgrade.Code, afterUpgrade.Body.String())
	}

	downgradeBody := `{"type":"subscription.updated","data":{"user_id":"user_billing","plan":"free"}}`
	downgrade := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/zitadel/webhook",
		strings.NewReader(downgradeBody),
		map[string]string{
			"Content-Type":                "application/json",
			"X-Cabinet-Webhook-Signature": hmacSignature(secret, downgradeBody),
		},
	)
	if downgrade.Code != http.StatusOK {
		t.Fatalf("downgrade webhook expected 200, got %d body=%s", downgrade.Code, downgrade.Body.String())
	}

	afterDowngrade := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"`+bootstrapToken+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if afterDowngrade.Code != http.StatusOK || !strings.Contains(afterDowngrade.Body.String(), `"plan":"free"`) {
		t.Fatalf("expected free plan after downgrade, got status=%d body=%s", afterDowngrade.Code, afterDowngrade.Body.String())
	}
}

func hmacSignature(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
