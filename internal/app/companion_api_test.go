package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/companion"
	"github.com/collectors-tech/cabinet/internal/profile"
)

const companionAPIOrigin = "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestCompanionPairingAPIRequiresApprovalAndSupportsSessionLifecycle(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := prepareCompanionAPIProfile(t, a)
	receipt := requestCompanionPairing(t, a, []string{
		companion.CapabilityModulesRead,
		companion.CapabilityCapturesSubmit,
		companion.CapabilityMediaSubmit,
		companion.CapabilitySessionManage,
	})

	unapproved := exchangeCompanionPairing(t, a, receipt, false)
	if unapproved.Code != http.StatusBadRequest || !strings.Contains(unapproved.Body.String(), "companion_pairing_exchange_invalid") {
		t.Fatalf("unapproved exchange status=%d body=%s", unapproved.Code, unapproved.Body.String())
	}

	approvalBody := `{"request_id":"` + receipt.RequestID + `","profile_id":"` + profileID + `"}`
	approved := doCompanionManagementRequest(t, a, http.MethodPost, "/api/companion/pairing/approvals", strings.NewReader(approvalBody), nil)
	if approved.Code != http.StatusOK {
		t.Fatalf("approval status=%d body=%s", approved.Code, approved.Body.String())
	}

	exchanged := exchangeCompanionPairing(t, a, receipt, true)
	var credential companion.CredentialResponse
	if err := json.NewDecoder(exchanged.Body).Decode(&credential); err != nil {
		t.Fatalf("decode exchange: %v", err)
	}
	if !strings.HasPrefix(credential.Credential, "cabcmp_") || credential.Session.ProfileID != profileID {
		t.Fatalf("unexpected exchange %+v", credential)
	}

	authorization := "Bearer " + credential.Credential
	modules := doCompanionExtensionRequest(t, a, http.MethodGet, "/api/companion/modules", nil, map[string]string{"Authorization": authorization})
	if modules.Code != http.StatusOK || !strings.Contains(modules.Body.String(), `"integration_instance_id"`) || strings.Contains(modules.Body.String(), "must-not-project") {
		t.Fatalf("modules status=%d body=%s", modules.Code, modules.Body.String())
	}

	payloadBody := `{"profile_id":"` + profileID + `","module_id":"ebay-purchase-capture","url":"https://www.ebay.com/itm/123","payload_type":"purchase_order","passive":true,"confidence_score":0.82}`
	payload := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(payloadBody), map[string]string{
		"Authorization": authorization, "Content-Type": "application/json",
	})
	if payload.Code != http.StatusAccepted || !strings.Contains(payload.Body.String(), `"remote_write":false`) {
		t.Fatalf("payload status=%d body=%s", payload.Code, payload.Body.String())
	}

	legacy := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(payloadBody), map[string]string{
		"Authorization": "Bearer companion:" + profileID, "Content-Type": "application/json",
	})
	if legacy.Code != http.StatusUnauthorized {
		t.Fatalf("legacy predictable bearer status=%d body=%s", legacy.Code, legacy.Body.String())
	}

	rotated := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/session/rotate", strings.NewReader(`{}`), map[string]string{
		"Authorization": authorization, "Content-Type": "application/json",
	})
	if rotated.Code != http.StatusOK {
		t.Fatalf("rotation status=%d body=%s", rotated.Code, rotated.Body.String())
	}
	var next companion.CredentialResponse
	if err := json.NewDecoder(rotated.Body).Decode(&next); err != nil {
		t.Fatalf("decode rotation: %v", err)
	}
	oldStatus := doCompanionExtensionRequest(t, a, http.MethodGet, "/api/companion/session", nil, map[string]string{"Authorization": authorization})
	if oldStatus.Code != http.StatusUnauthorized {
		t.Fatalf("old rotated credential status=%d body=%s", oldStatus.Code, oldStatus.Body.String())
	}

	sessions := doCompanionManagementRequest(t, a, http.MethodGet, "/api/companion/sessions", nil, nil)
	if sessions.Code != http.StatusOK || strings.Contains(sessions.Body.String(), credential.Credential) || strings.Contains(sessions.Body.String(), next.Credential) {
		t.Fatalf("session list status=%d body=%s", sessions.Code, sessions.Body.String())
	}
	revoked := doCompanionManagementRequest(t, a, http.MethodDelete, "/api/companion/sessions?all=true", nil, nil)
	if revoked.Code != http.StatusOK || !strings.Contains(revoked.Body.String(), `"revoked_count":1`) {
		t.Fatalf("managed revoke status=%d body=%s", revoked.Code, revoked.Body.String())
	}
	nextStatus := doCompanionExtensionRequest(t, a, http.MethodGet, "/api/companion/session", nil, map[string]string{"Authorization": "Bearer " + next.Credential})
	if nextStatus.Code != http.StatusUnauthorized {
		t.Fatalf("revoked credential status=%d body=%s", nextStatus.Code, nextStatus.Body.String())
	}
}

func TestCompanionAPILoopbackOriginBodyAndMediaBoundaries(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := prepareCompanionAPIProfile(t, a)

	body := `{"device_id":"device-a","device_name":"Chrome","protocol_version":"1"}`
	nonLoopback := doRawCompanionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(body), map[string]string{
		"Content-Type": "application/json", "Origin": companionAPIOrigin,
	}, "cabinet.example:17880", "203.0.113.10:1234")
	if nonLoopback.Code != http.StatusForbidden || !strings.Contains(nonLoopback.Body.String(), "companion_loopback_required") {
		t.Fatalf("non-loopback status=%d body=%s", nonLoopback.Code, nonLoopback.Body.String())
	}

	badOrigin := doRawCompanionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(body), map[string]string{
		"Content-Type": "application/json", "Origin": "https://attacker.example",
	}, "127.0.0.1:17880", "127.0.0.1:1234")
	if badOrigin.Code != http.StatusForbidden || !strings.Contains(badOrigin.Body.String(), "companion_origin_rejected") {
		t.Fatalf("bad origin status=%d body=%s", badOrigin.Code, badOrigin.Body.String())
	}

	wrongContentType := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(body), nil)
	if wrongContentType.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("content type status=%d body=%s", wrongContentType.Code, wrongContentType.Body.String())
	}
	deviceMismatchBody := `{"device_id":"another-device","device_name":"Chrome","protocol_version":"1","capabilities":["modules:read"]}`
	deviceMismatch := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(deviceMismatchBody), map[string]string{"Content-Type": "application/json"})
	if deviceMismatch.Code != http.StatusBadRequest || !strings.Contains(deviceMismatch.Body.String(), "companion_device_mismatch") {
		t.Fatalf("pairing device mismatch status=%d body=%s", deviceMismatch.Code, deviceMismatch.Body.String())
	}

	oversized := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(`{"device_id":"`+strings.Repeat("a", companionSmallBodyLimit)+`"}`), map[string]string{"Content-Type": "application/json"})
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized status=%d body=%s", oversized.Code, oversized.Body.String())
	}

	preflight := doRawCompanionRequest(t, a, http.MethodOptions, "/api/companion/pairing/requests", nil, map[string]string{"Origin": companionAPIOrigin}, "127.0.0.1:17880", "127.0.0.1:1234")
	if preflight.Code != http.StatusNoContent || preflight.Header().Get("Access-Control-Allow-Origin") != companionAPIOrigin {
		t.Fatalf("preflight status=%d headers=%v body=%s", preflight.Code, preflight.Header(), preflight.Body.String())
	}
	invalidExtensionPreflight := doRawCompanionRequest(t, a, http.MethodOptions, "/api/companion/pairing/requests", nil, map[string]string{
		"Origin": "chrome-extension://invalid",
	}, "127.0.0.1:17880", "127.0.0.1:1234")
	if invalidExtensionPreflight.Code != http.StatusForbidden || invalidExtensionPreflight.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("invalid extension preflight status=%d headers=%v", invalidExtensionPreflight.Code, invalidExtensionPreflight.Header())
	}

	trailingJSON := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(body+` {}`), map[string]string{"Content-Type": "application/json"})
	if trailingJSON.Code != http.StatusBadRequest || !strings.Contains(trailingJSON.Body.String(), "companion_invalid_json") {
		t.Fatalf("trailing JSON status=%d body=%s", trailingJSON.Code, trailingJSON.Body.String())
	}

	hostileLoopbackOrigin := doRawCompanionRequest(t, a, http.MethodGet, "/api/companion/sessions", nil, map[string]string{
		"Origin": "http://127.0.0.1:9999",
	}, "127.0.0.1:17880", "127.0.0.1:1234")
	if hostileLoopbackOrigin.Code != http.StatusForbidden || !strings.Contains(hostileLoopbackOrigin.Body.String(), "companion_origin_rejected") {
		t.Fatalf("hostile loopback origin status=%d body=%s", hostileLoopbackOrigin.Code, hostileLoopbackOrigin.Body.String())
	}

	receipt := requestCompanionPairing(t, a, []string{companion.CapabilityMediaSubmit})
	approvalBody := `{"request_id":"` + receipt.RequestID + `","profile_id":"` + profileID + `"}`
	approved := doCompanionManagementRequest(t, a, http.MethodPost, "/api/companion/pairing/approvals", strings.NewReader(approvalBody), nil)
	if approved.Code != http.StatusOK {
		t.Fatalf("media pairing approval status=%d body=%s", approved.Code, approved.Body.String())
	}
	exchanged := exchangeCompanionPairing(t, a, receipt, true)
	var credential companion.CredentialResponse
	if err := json.NewDecoder(exchanged.Body).Decode(&credential); err != nil {
		t.Fatalf("decode media credential: %v", err)
	}
	mediaDigest := sha256.Sum256([]byte("png-data"))
	mediaHeaders := map[string]string{
		"Authorization":             "Bearer " + credential.Credential,
		"Content-Type":              "image/png",
		"X-Cabinet-Profile":         profileID,
		"X-Cabinet-Idempotency-Key": "media-1",
		"X-Cabinet-Media-SHA256":    hex.EncodeToString(mediaDigest[:]),
	}
	validMedia := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/media-submissions", strings.NewReader("png-data"), mediaHeaders)
	if validMedia.Code != http.StatusNotImplemented || !strings.Contains(validMedia.Body.String(), "companion_media_persistence_not_implemented") {
		t.Fatalf("valid media transport status=%d body=%s", validMedia.Code, validMedia.Body.String())
	}
	var mediaAuditCount int
	if err := a.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM companion_audit_events WHERE profile_id = ? AND action = 'media.transport.validated' AND result_code = 'persistence_pending'`, profileID).Scan(&mediaAuditCount); err != nil || mediaAuditCount != 1 {
		t.Fatalf("media transport audit count=%d err=%v", mediaAuditCount, err)
	}

	invalidMediaHeaders := make(map[string]string, len(mediaHeaders))
	for key, value := range mediaHeaders {
		invalidMediaHeaders[key] = value
	}
	invalidMediaHeaders["X-Cabinet-Media-SHA256"] = "not-a-sha256"
	invalidMedia := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/media-submissions", strings.NewReader("png-data"), invalidMediaHeaders)
	if invalidMedia.Code != http.StatusBadRequest || !strings.Contains(invalidMedia.Body.String(), "companion_media_metadata_invalid") {
		t.Fatalf("invalid media metadata status=%d body=%s", invalidMedia.Code, invalidMedia.Body.String())
	}
	checksumMismatchHeaders := make(map[string]string, len(mediaHeaders))
	for key, value := range mediaHeaders {
		checksumMismatchHeaders[key] = value
	}
	checksumMismatchHeaders["X-Cabinet-Media-SHA256"] = strings.Repeat("a", 64)
	checksumMismatch := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/media-submissions", strings.NewReader("png-data"), checksumMismatchHeaders)
	if checksumMismatch.Code != http.StatusBadRequest || !strings.Contains(checksumMismatch.Body.String(), "companion_media_checksum_mismatch") {
		t.Fatalf("media checksum mismatch status=%d body=%s", checksumMismatch.Code, checksumMismatch.Body.String())
	}

	unsupportedHeaders := make(map[string]string, len(mediaHeaders))
	for key, value := range mediaHeaders {
		unsupportedHeaders[key] = value
	}
	unsupportedHeaders["Content-Type"] = "text/html"
	unsupportedMedia := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/media-submissions", strings.NewReader("html"), unsupportedHeaders)
	if unsupportedMedia.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("unsupported media status=%d body=%s", unsupportedMedia.Code, unsupportedMedia.Body.String())
	}

	oversizedMedia := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/media-submissions", strings.NewReader(strings.Repeat("a", companionMediaBodyLimit+1)), mediaHeaders)
	if oversizedMedia.Code != http.StatusRequestEntityTooLarge || !strings.Contains(oversizedMedia.Body.String(), "companion_media_too_large") {
		t.Fatalf("oversized media status=%d body=%s", oversizedMedia.Code, oversizedMedia.Body.String())
	}

	missingManagedSessionID := doCompanionManagementRequest(t, a, http.MethodDelete, "/api/companion/sessions", nil, nil)
	if missingManagedSessionID.Code != http.StatusBadRequest || !strings.Contains(missingManagedSessionID.Body.String(), "companion_session_id_required") {
		t.Fatalf("missing managed session id status=%d body=%s", missingManagedSessionID.Code, missingManagedSessionID.Body.String())
	}
}

type companionPairingReceipt struct {
	RequestID       string `json:"request_id"`
	ExchangeSecret  string `json:"exchange_secret"`
	ProtocolVersion string `json:"protocol_version"`
}

func prepareCompanionAPIProfile(t *testing.T, a *App) string {
	t.Helper()
	profiles := profile.NewRepository(a.db)
	created, err := profiles.Create(context.Background(), "Companion API")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("set active profile: %v", err)
	}
	if _, err := profiles.UpsertIntegrationInstance(context.Background(), created.ID, profile.IntegrationInstancePatch{
		ProviderID: "ebay", Config: map[string]string{"region": "AU", "token": "must-not-project"},
	}); err != nil {
		t.Fatalf("upsert integration: %v", err)
	}
	if _, err := a.authService.CreateUnlockedSession(created.ID); err != nil {
		t.Fatalf("unlock profile: %v", err)
	}
	return created.ID
}

func requestCompanionPairing(t *testing.T, a *App, capabilities []string) companionPairingReceipt {
	t.Helper()
	rawCapabilities, _ := json.Marshal(capabilities)
	body := `{"device_id":"device-a","device_name":"Chrome on Windows","protocol_version":"1","capabilities":` + string(rawCapabilities) + `}`
	response := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/requests", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusCreated {
		t.Fatalf("pairing request status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("pairing request leaked cacheable exchange material: headers=%v", response.Header())
	}
	var receipt companionPairingReceipt
	if err := json.NewDecoder(response.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode pairing request: %v", err)
	}
	return receipt
}

func exchangeCompanionPairing(t *testing.T, a *App, receipt companionPairingReceipt, expectSuccess bool) *httptest.ResponseRecorder {
	t.Helper()
	body := `{"request_id":"` + receipt.RequestID + `","exchange_secret":"` + receipt.ExchangeSecret + `","device_id":"device-a","protocol_version":"` + receipt.ProtocolVersion + `"}`
	response := doCompanionExtensionRequest(t, a, http.MethodPost, "/api/companion/pairing/exchanges", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if expectSuccess && response.Code != http.StatusOK {
		t.Fatalf("pairing exchange status=%d body=%s", response.Code, response.Body.String())
	}
	if expectSuccess && response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("pairing exchange credential was cacheable: headers=%v", response.Header())
	}
	return response
}

func doCompanionExtensionRequest(t *testing.T, a *App, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	all := map[string]string{"Origin": companionAPIOrigin, "X-Cabinet-Companion-Device": "device-a"}
	for key, value := range headers {
		all[key] = value
	}
	return doRawCompanionRequest(t, a, method, path, body, all, "127.0.0.1:17880", "127.0.0.1:1234")
}

func doCompanionManagementRequest(t *testing.T, a *App, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	all := map[string]string{"Origin": "http://127.0.0.1:17880"}
	for key, value := range headers {
		all[key] = value
	}
	if body != nil && all["Content-Type"] == "" {
		all["Content-Type"] = "application/json"
	}
	return doRawCompanionRequest(t, a, method, path, body, all, "127.0.0.1:17880", "127.0.0.1:1234")
}

func doRawCompanionRequest(t *testing.T, a *App, method, path string, body io.Reader, headers map[string]string, host, remoteAddress string) *httptest.ResponseRecorder {
	t.Helper()
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Host = host
	req.RemoteAddr = remoteAddress
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(recorder, req.WithContext(context.Background()))
	return recorder
}
