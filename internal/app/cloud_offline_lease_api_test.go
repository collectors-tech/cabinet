package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCloudOfflineLeaseIssueAndValidate(t *testing.T) {
	a := newTestApp(t)
	token := "e30.eyJzdWIiOiJ1c2VyX2xlYXNlIiwiZW1haWwiOiJsZWFzZUBleGFtcGxlLmNvbSIsInBsYW4iOiJwcm8ifQ.e30"

	issue := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/lease/issue",
		strings.NewReader(`{"provider":"zitadel","token":"`+token+`","duration_seconds":120}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if issue.Code != http.StatusOK {
		t.Fatalf("issue status=%d body=%s", issue.Code, issue.Body.String())
	}

	var issued map[string]any
	if err := json.Unmarshal(issue.Body.Bytes(), &issued); err != nil {
		t.Fatalf("decode issued lease: %v", err)
	}
	leaseToken, _ := issued["lease_token"].(string)
	if strings.TrimSpace(leaseToken) == "" {
		t.Fatalf("expected lease token in response: %s", issue.Body.String())
	}

	validate := doRequest(
		t,
		a,
		http.MethodGet,
		"/api/auth/cloud/lease/validate?lease_token="+leaseToken,
		nil,
		nil,
	)
	if validate.Code != http.StatusOK {
		t.Fatalf("validate status=%d body=%s", validate.Code, validate.Body.String())
	}
	if !strings.Contains(validate.Body.String(), `"valid":true`) {
		t.Fatalf("expected valid lease response, got %s", validate.Body.String())
	}
}

func TestCloudOfflineLeaseIssueRejectsRetiredClerkProvider(t *testing.T) {
	a := newTestApp(t)
	token := "e30.eyJzdWIiOiJ1c2VyX2xlYXNlIiwiZW1haWwiOiJsZWFzZUBleGFtcGxlLmNvbSIsInBsYW4iOiJwcm8ifQ.e30"

	issue := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/lease/issue",
		strings.NewReader(`{"provider":"clerk","token":"`+token+`","duration_seconds":120}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if issue.Code != http.StatusBadRequest {
		t.Fatalf("retired clerk provider expected 400, got %d body=%s", issue.Code, issue.Body.String())
	}
	if !strings.Contains(issue.Body.String(), `"error":"unsupported_provider"`) {
		t.Fatalf("expected unsupported_provider error, got %s", issue.Body.String())
	}
}

func TestCloudOfflineLeaseExpiredBlocksProtectedFeature(t *testing.T) {
	a := newTestApp(t)
	token := "e30.eyJzdWIiOiJ1c2VyX2V4cGlyZWQiLCJlbWFpbCI6ImV4cGlyZWRAZXhhbXBsZS5jb20iLCJwbGFuIjoiZnJlZSJ9.e30"

	issue := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/lease/issue",
		strings.NewReader(`{"provider":"zitadel","token":"`+token+`","duration_seconds":1}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if issue.Code != http.StatusOK {
		t.Fatalf("issue status=%d body=%s", issue.Code, issue.Body.String())
	}

	var issued map[string]any
	if err := json.Unmarshal(issue.Body.Bytes(), &issued); err != nil {
		t.Fatalf("decode issued lease: %v", err)
	}
	leaseToken, _ := issued["lease_token"].(string)
	if strings.TrimSpace(leaseToken) == "" {
		t.Fatalf("expected lease token in response: %s", issue.Body.String())
	}

	time.Sleep(2 * time.Second)

	protected := doRequest(
		t,
		a,
		http.MethodGet,
		"/api/auth/cloud/offline/protected",
		nil,
		map[string]string{"X-Cabinet-Lease": leaseToken},
	)
	if protected.Code != http.StatusLocked {
		t.Fatalf("expected protected endpoint to block expired lease (423), got %d body=%s", protected.Code, protected.Body.String())
	}
}

func TestCloudOfflineLeaseRenewalFlow(t *testing.T) {
	a := newTestApp(t)
	token := "e30.eyJzdWIiOiJ1c2VyX3JlbmV3IiwiZW1haWwiOiJyZW5ld0BleGFtcGxlLmNvbSIsInBsYW4iOiJwcm8ifQ.e30"

	issue := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/lease/issue",
		strings.NewReader(`{"provider":"zitadel","token":"`+token+`","duration_seconds":30}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if issue.Code != http.StatusOK {
		t.Fatalf("issue status=%d body=%s", issue.Code, issue.Body.String())
	}

	var issued map[string]any
	if err := json.Unmarshal(issue.Body.Bytes(), &issued); err != nil {
		t.Fatalf("decode issued lease: %v", err)
	}
	leaseToken, _ := issued["lease_token"].(string)
	if strings.TrimSpace(leaseToken) == "" {
		t.Fatalf("expected lease token in response: %s", issue.Body.String())
	}
	initialExpiry, _ := issued["expires_at"].(string)

	renew := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/lease/renew",
		strings.NewReader(`{"lease_token":"`+leaseToken+`","duration_seconds":120}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if renew.Code != http.StatusOK {
		t.Fatalf("renew status=%d body=%s", renew.Code, renew.Body.String())
	}
	var renewed map[string]any
	if err := json.Unmarshal(renew.Body.Bytes(), &renewed); err != nil {
		t.Fatalf("decode renewed lease: %v", err)
	}
	renewedExpiry, _ := renewed["expires_at"].(string)
	if renewedExpiry == "" || renewedExpiry == initialExpiry {
		t.Fatalf("expected renewed expiry to change, initial=%q renewed=%q", initialExpiry, renewedExpiry)
	}
}
