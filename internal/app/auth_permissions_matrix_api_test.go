package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/licensing"
	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestAuthPermissionsPlanCapabilityMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		plan     string
		features []string
	}{
		{plan: "mvp", features: []string{"collection_core"}},
		{plan: "creator", features: []string{"collection_core", "ai_assist", "scanner_automation"}},
		{plan: "teams", features: []string{"collection_core", "ai_assist", "price_tracking", "scanner_automation"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.plan, func(t *testing.T) {
			t.Parallel()

			a := newTestApp(t)
			token := signedLikeJWTForPlan("user_"+tc.plan, tc.plan)
			resp := doRequest(
				t,
				a,
				http.MethodPost,
				"/api/auth/cloud/session/bootstrap",
				strings.NewReader(`{"provider":"clerk","token":"`+token+`"}`),
				map[string]string{"Content-Type": "application/json"},
			)
			if resp.Code != http.StatusOK {
				t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
			}
			var payload struct {
				Plan     string   `json:"plan"`
				Features []string `json:"features"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload.Plan != tc.plan {
				t.Fatalf("expected plan %q, got %q", tc.plan, payload.Plan)
			}
			if strings.Join(payload.Features, ",") != strings.Join(tc.features, ",") {
				t.Fatalf("expected features %v, got %v", tc.features, payload.Features)
			}
		})
	}
}

func TestAuthPermissionsFeatureGateMatrixFromCloudPlan(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	licenseSvc := licensing.NewService(a.db, profile.NewRepository(a.db), "")

	setCloudPlan := func(plan string) {
		if _, err := a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('cloud.plan', ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, plan); err != nil {
			t.Fatalf("set cloud plan: %v", err)
		}
	}

	cfg := config.Config{}
	profileID := "profile-test"

	setCloudPlan("mvp")
	if hasProFeatureAccess(t.Context(), a.db, licenseSvc, cfg, profileID, "scanner_automation") {
		t.Fatalf("mvp plan must not allow scanner_automation")
	}

	setCloudPlan("creator")
	if !hasProFeatureAccess(t.Context(), a.db, licenseSvc, cfg, profileID, "scanner_automation") {
		t.Fatalf("creator plan must allow scanner_automation")
	}
	if hasProFeatureAccess(t.Context(), a.db, licenseSvc, cfg, profileID, "price_tracking") {
		t.Fatalf("creator plan must not allow price_tracking")
	}

	setCloudPlan("teams")
	if !hasProFeatureAccess(t.Context(), a.db, licenseSvc, cfg, profileID, "price_tracking") {
		t.Fatalf("teams plan must allow price_tracking")
	}
	if !hasProFeatureAccess(t.Context(), a.db, licenseSvc, cfg, profileID, "ai_assist") {
		t.Fatalf("teams plan must allow ai_assist")
	}
}

func TestAuthPermissionsEffectiveDiagnosticsContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	token := signedLikeJWTForPlan("user_teams", "teams")
	bootstrap := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"`+token+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if bootstrap.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrap.Code, bootstrap.Body.String())
	}

	resp := doRequest(t, a, http.MethodGet, "/api/auth/cloud/session/effective", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("effective permissions status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	for _, field := range []string{"provider", "user_id", "role", "plan", "features"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("effective diagnostics missing %q in %+v", field, payload)
		}
	}
}

func signedLikeJWTForPlan(userID, plan string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s","email":"%s@example.com","plan":"%s","role":"member"}`, userID, userID, plan)))
	signature := base64.RawURLEncoding.EncodeToString([]byte("sig"))
	return header + "." + payload + "." + signature
}
