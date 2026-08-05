package companion

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
)

const companionTestOrigin = "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestCompanionRegistryNormalizesPassiveModules(t *testing.T) {
	t.Parallel()

	svc := NewService([]Module{{
		ID: " ebay-purchase-capture ", ModuleVersion: "1.0.0", Site: " ebay ",
		Actions: []string{"capture_tracking", "capture_item"}, PassiveOnly: true, Display: ModuleDisplay{Name: "eBay"},
		CaptureSchemas: []CaptureSchema{{PayloadType: "purchase_order", Fields: []string{"title"}}}, Workflows: []string{"manual_capture"},
		RedactionRules: []string{"no_tokens"}, FixtureVersion: "1",
		Browser: BrowserContract{StartURL: "https://www.ebay.com/mye/myebay/purchase", Origins: []string{"https://www.ebay.com/*"}, URLPatterns: []string{"https://www.ebay.com/mye/myebay/purchase*"},
			Readiness: BrowserReadiness{Ready: []string{"#ready"}, LoggedOut: []string{"#logged-out"}, Challenge: []string{"#challenge"}}},
		Configuration: ModuleConfiguration{CaptureMode: "manual_user_present", ItemFields: []string{"title"}, MediaPolicy: "review", ReviewDestination: "purchases", RateLimitPerMinute: 6, HelpURL: "/help"},
	}})
	registry := svc.Registry()
	if len(registry.Modules) != 1 || registry.ProtocolVersion != ProtocolVersionV1 {
		t.Fatalf("expected one v1 module, got %+v", registry)
	}
	module := registry.Modules[0]
	if module.ID != "ebay-purchase-capture" || module.Site != "ebay" || module.ProviderID != "ebay" || !module.PassiveOnly {
		t.Fatalf("unexpected module %+v", module)
	}
	if got := module.Actions; len(got) != 2 || got[0] != "capture_item" || got[1] != "capture_tracking" {
		t.Fatalf("actions were not sorted and preserved: %+v", got)
	}
}

func TestCompanionRegistryPublishesDataDrivenBrowserHostContract(t *testing.T) {
	t.Parallel()

	registry := NewService(DefaultModules()).Registry()
	if len(registry.Modules) != 1 {
		t.Fatalf("expected one default browser module, got %+v", registry.Modules)
	}
	module := registry.Modules[0]
	if module.ModuleVersion != "1.0.0" || module.Display.Name != "eBay purchases" {
		t.Fatalf("missing versioned display contract: %+v", module)
	}
	if module.Browser.StartURL != "https://www.ebay.com/mye/myebay/purchase" ||
		len(module.Browser.Origins) != 1 || module.Browser.Origins[0] != "https://www.ebay.com/*" {
		t.Fatalf("unsafe or incomplete browser origin contract: %+v", module.Browser)
	}
	if len(module.Browser.Readiness.Ready) == 0 || len(module.Browser.Readiness.LoggedOut) == 0 ||
		len(module.Browser.Readiness.Challenge) == 0 {
		t.Fatalf("readiness evidence contract is incomplete: %+v", module.Browser.Readiness)
	}
	if len(module.Browser.URLPatterns) == 0 || len(module.CaptureSchemas) == 0 || len(module.Workflows) == 0 ||
		len(module.RedactionRules) == 0 || module.FixtureVersion != "1" {
		t.Fatalf("versioned capture module contract is incomplete: %+v", module)
	}
	if module.Configuration.CaptureMode != "manual_user_present" || module.Configuration.SyncAvailable ||
		module.Configuration.ReviewDestination != "purchases" || module.Configuration.RateLimitPerMinute != 6 {
		t.Fatalf("Cabinet module configuration is not authoritative or truthful: %+v", module.Configuration)
	}
}

func TestCompanionRegistryAdvertisesSyncOnlyForPackagedCaptureModules(t *testing.T) {
	t.Parallel()

	modules := DefaultModules()
	modules[0].Browser.CaptureScript = "modules/fixture.js"
	modules[0].Configuration.SyncAvailable = true
	registry := NewService(modules).Registry()
	if len(registry.Modules) != 1 || !registry.Modules[0].Configuration.SyncAvailable {
		t.Fatalf("durable persistence did not enable the packaged capture module: %+v", registry.Modules)
	}
}

func TestCompanionPredictableBearerCannotImpersonateProfile(t *testing.T) {
	t.Parallel()

	svc, _, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	_, err := svc.AcceptPayload(context.Background(), PayloadSubmission{
		ProfileID:       profileID,
		ModuleID:        "ebay-purchase-capture",
		URL:             "https://www.ebay.com/itm/123",
		PayloadType:     "purchase_order",
		Passive:         true,
		ConfidenceScore: 0.9,
	}, "Bearer companion:"+profileID, metadata)
	if ErrorCode(err) != "companion_auth_required" {
		t.Fatalf("predictable companion:<profile_id> bearer was not rejected: %v", err)
	}
}

func TestCompanionAcceptedPayloadIsPersistedBeforeAcknowledgement(t *testing.T) {
	t.Parallel()

	svc, _, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	accepted, err := svc.AcceptPayload(context.Background(), validPurchasePayload(t, svc, profileID, "capture-persisted", "123"), authorization, metadata)
	if err != nil || !accepted.Accepted {
		t.Fatalf("AcceptPayload() = %+v, %v", accepted, err)
	}

	var count int
	if err := svc.db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM companion_captures
		WHERE profile_id = ? AND module_id = ? AND payload_type = ?
	`, profileID, "ebay-purchase-capture", "purchase_order").Scan(&count); err != nil {
		t.Fatalf("accepted companion payload disappeared instead of being persisted: %v", err)
	}
	if count != 1 {
		t.Fatalf("persisted companion capture count = %d, want 1", count)
	}
}

func TestCompanionPairApproveExchangeRotateRevokeAndRestart(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	receipt, err := svc.CreatePairingRequest(context.Background(), PairingRequestInput{
		DeviceID: "device-a", DeviceName: "Chrome on Windows", ProtocolVersion: ProtocolVersionV1,
		Capabilities: []string{CapabilityModulesRead, CapabilityCapturesSubmit, CapabilitySessionManage},
	}, metadata)
	if err != nil {
		t.Fatalf("CreatePairingRequest() error = %v", err)
	}
	if receipt.ExchangeSecret == "" || len(receipt.ExchangeSecret) < 40 || receipt.Status != "pending" || len(receipt.PairingCode) != 6 {
		t.Fatalf("unexpected pairing receipt %+v", receipt)
	}
	_, err = svc.ExchangePairing(context.Background(), PairingExchangeInput{
		RequestID: receipt.RequestID, ExchangeSecret: receipt.ExchangeSecret,
		DeviceID: "device-a", ProtocolVersion: ProtocolVersionV1,
	}, metadata)
	if ErrorCode(err) != "companion_pairing_exchange_invalid" {
		t.Fatalf("exchange before approval error = %v", err)
	}
	if _, err := svc.ApprovePairing(context.Background(), PairingApprovalInput{
		RequestID: receipt.RequestID, ProfileID: profileID,
	}, RequestMetadata{RemoteAddress: "127.0.0.1:1000"}); err != nil {
		t.Fatalf("ApprovePairing() error = %v", err)
	}
	approvedRequests, err := svc.ListPairingRequests(context.Background(), profileID)
	if err != nil || len(approvedRequests) != 1 || approvedRequests[0].Status != "approved" {
		t.Fatalf("approved profile pairing requests = %+v, %v", approvedRequests, err)
	}
	otherProfile, err := profiles.Create(context.Background(), "Other companion profile")
	if err != nil {
		t.Fatalf("create other profile: %v", err)
	}
	otherRequests, err := svc.ListPairingRequests(context.Background(), otherProfile.ID)
	if err != nil || len(otherRequests) != 0 {
		t.Fatalf("approved pairing leaked across profiles: %+v, %v", otherRequests, err)
	}
	exchanged, err := svc.ExchangePairing(context.Background(), PairingExchangeInput{
		RequestID: receipt.RequestID, ExchangeSecret: receipt.ExchangeSecret,
		DeviceID: "device-a", ProtocolVersion: ProtocolVersionV1,
	}, metadata)
	if err != nil {
		t.Fatalf("ExchangePairing() error = %v", err)
	}
	if !strings.HasPrefix(exchanged.Credential, "cabcmp_") || exchanged.Session.ProfileID != profileID || exchanged.Session.Origin != companionTestOrigin {
		t.Fatalf("unexpected exchanged session %+v", exchanged)
	}
	var storedCredentialVerifier, storedExchangeVerifier string
	if err := svc.db.QueryRowContext(context.Background(), `SELECT credential_verifier FROM companion_sessions WHERE id = ?`, exchanged.Session.ID).Scan(&storedCredentialVerifier); err != nil {
		t.Fatalf("load stored credential verifier: %v", err)
	}
	if err := svc.db.QueryRowContext(context.Background(), `SELECT exchange_verifier FROM companion_pairing_requests WHERE id = ?`, receipt.RequestID).Scan(&storedExchangeVerifier); err != nil {
		t.Fatalf("load stored exchange verifier: %v", err)
	}
	if storedCredentialVerifier == exchanged.Credential || storedCredentialVerifier != verifier(exchanged.Credential) ||
		storedExchangeVerifier == receipt.ExchangeSecret || storedExchangeVerifier != verifier(receipt.ExchangeSecret) {
		t.Fatalf("raw pairing/session material was persisted: credential=%q exchange=%q", storedCredentialVerifier, storedExchangeVerifier)
	}
	if _, err := svc.ExchangePairing(context.Background(), PairingExchangeInput{
		RequestID: receipt.RequestID, ExchangeSecret: receipt.ExchangeSecret,
		DeviceID: "device-a", ProtocolVersion: ProtocolVersionV1,
	}, metadata); ErrorCode(err) != "companion_pairing_exchange_replayed" {
		t.Fatalf("replayed exchange error = %v", err)
	}

	authorization := "Bearer " + exchanged.Credential
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, CapabilityModulesRead); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	restarted, err := NewPersistentService(context.Background(), svc.db, profiles, DefaultModules(), Options{})
	if err != nil {
		t.Fatalf("restart service error = %v", err)
	}
	if _, err := restarted.Authenticate(context.Background(), authorization, metadata, CapabilityModulesRead); err != nil {
		t.Fatalf("restarted Authenticate() error = %v", err)
	}
	rotated, err := restarted.RotateCredential(context.Background(), authorization, metadata)
	if err != nil {
		t.Fatalf("RotateCredential() error = %v", err)
	}
	if rotated.Credential == exchanged.Credential || rotated.Session.RotatedFromID != exchanged.Session.ID {
		t.Fatalf("rotation did not replace credential: %+v", rotated)
	}
	if _, err := restarted.Authenticate(context.Background(), authorization, metadata, ""); ErrorCode(err) != "companion_session_revoked" {
		t.Fatalf("old credential after rotation error = %v", err)
	}
	rotatedAuthorization := "Bearer " + rotated.Credential
	if err := restarted.RevokeCredential(context.Background(), rotatedAuthorization, metadata); err != nil {
		t.Fatalf("RevokeCredential() error = %v", err)
	}
	if _, err := restarted.Authenticate(context.Background(), rotatedAuthorization, metadata, ""); ErrorCode(err) != "companion_session_revoked" {
		t.Fatalf("revoked credential error = %v", err)
	}
	sessions, err := restarted.ListSessions(context.Background(), profileID)
	if err != nil || len(sessions) != 2 {
		t.Fatalf("ListSessions() = %+v, %v", sessions, err)
	}
	encoded := encodeStrings([]string{sessions[0].ID, sessions[1].ID})
	if strings.Contains(encoded, exchanged.Credential) || strings.Contains(encoded, rotated.Credential) {
		t.Fatal("session management response leaked credential material")
	}
}

func TestCompanionSessionBindingProfileModulesAndPayloadSafety(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityModulesRead, CapabilityCapturesSubmit})

	registry, err := svc.RegistryForSession(context.Background(), authorization, metadata)
	if err != nil {
		t.Fatalf("RegistryForSession() error = %v", err)
	}
	if registry.ProfileID != profileID || len(registry.Modules) != 1 || registry.Modules[0].IntegrationInstanceID == "" {
		t.Fatalf("unexpected profile module registry %+v", registry)
	}
	if registry.Modules[0].SafeConfig["region"] != "AU" {
		t.Fatalf("safe config missing: %+v", registry.Modules[0].SafeConfig)
	}
	for key := range registry.Modules[0].SafeConfig {
		if strings.Contains(strings.ToLower(key), "token") {
			t.Fatalf("registry leaked secret-like key %q", key)
		}
	}

	other, err := profiles.Create(context.Background(), "Other profile")
	if err != nil {
		t.Fatalf("create other profile: %v", err)
	}
	_, err = svc.AcceptPayload(context.Background(), PayloadSubmission{
		ProfileID: other.ID, ModuleID: "ebay-purchase-capture", URL: "https://www.ebay.com/itm/123",
		PayloadType: "purchase_order", Passive: true, ConfidenceScore: 0.9,
	}, authorization, metadata)
	if ErrorCode(err) != "companion_profile_mismatch" {
		t.Fatalf("cross-profile payload error = %v", err)
	}

	_, err = svc.Authenticate(context.Background(), authorization, RequestMetadata{
		Origin: companionTestOrigin, DeviceID: "other-device", RemoteAddress: metadata.RemoteAddress,
	}, CapabilityModulesRead)
	if ErrorCode(err) != "companion_session_binding_mismatch" {
		t.Fatalf("device binding error = %v", err)
	}
	_, err = svc.Authenticate(context.Background(), authorization, RequestMetadata{
		Origin: "chrome-extension://bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", DeviceID: metadata.DeviceID, RemoteAddress: metadata.RemoteAddress,
	}, CapabilityModulesRead)
	if ErrorCode(err) != "companion_session_binding_mismatch" {
		t.Fatalf("origin binding error = %v", err)
	}

	accepted, err := svc.AcceptPayload(context.Background(), validPurchasePayload(t, svc, profileID, "capture-authorised", "123"), authorization, metadata)
	if err != nil || !accepted.Accepted || accepted.RemoteWrite {
		t.Fatalf("AcceptPayload() = %+v, %v", accepted, err)
	}
	captureOnlyAuthorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})
	if _, err := svc.AcceptPayload(context.Background(), validPurchasePayload(t, svc, profileID, "capture-only", "456"), captureOnlyAuthorization, metadata); err != nil {
		t.Fatalf("capture-only capability was incorrectly coupled to module discovery: %v", err)
	}
	if _, err := svc.db.ExecContext(context.Background(), `UPDATE companion_sessions SET cabinet_instance_id = 'another-instance'`); err != nil {
		t.Fatalf("alter companion instance binding: %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, ""); ErrorCode(err) != "companion_session_binding_mismatch" {
		t.Fatalf("instance binding error = %v", err)
	}
}

func TestCompanionLimitsExpiryAndCapabilitiesFailClosed(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	svc, _, profileID, metadata := newPersistentCompanionTestService(t, Options{
		Now: func() time.Time { return now }, PairingTTL: time.Minute, SessionTTL: 2 * time.Minute,
		SessionRequestsPerMinute: 2, MaxConcurrentPerSession: 1,
	})
	staleReceipt, err := svc.CreatePairingRequest(context.Background(), PairingRequestInput{
		DeviceID: metadata.DeviceID, DeviceName: "Stale browser", ProtocolVersion: ProtocolVersionV1,
		Capabilities: []string{CapabilityModulesRead},
	}, metadata)
	if err != nil {
		t.Fatalf("create stale pairing request: %v", err)
	}
	if _, err := svc.ApprovePairing(context.Background(), PairingApprovalInput{RequestID: staleReceipt.RequestID, ProfileID: profileID}, metadata); err != nil {
		t.Fatalf("approve stale pairing request: %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := svc.ExchangePairing(context.Background(), PairingExchangeInput{
		RequestID: staleReceipt.RequestID, ExchangeSecret: staleReceipt.ExchangeSecret,
		DeviceID: metadata.DeviceID, ProtocolVersion: ProtocolVersionV1,
	}, metadata); ErrorCode(err) != "companion_pairing_exchange_invalid" {
		t.Fatalf("stale pairing exchange error = %v", err)
	}
	now = now.Add(-2 * time.Minute)
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityModulesRead})
	session, err := svc.Authenticate(context.Background(), authorization, metadata, CapabilityModulesRead)
	if err != nil {
		t.Fatalf("first Authenticate() error = %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, CapabilityMediaSubmit); ErrorCode(err) != "companion_capability_denied" {
		t.Fatalf("capability error = %v", err)
	}
	if _, err := svc.RotateCredential(context.Background(), authorization, metadata); ErrorCode(err) != "companion_capability_denied" {
		t.Fatalf("rotation capability error = %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, CapabilityModulesRead); err != nil {
		t.Fatalf("second rate-limited request unexpectedly failed: %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, CapabilityModulesRead); ErrorCode(err) != "companion_rate_limited" {
		t.Fatalf("rate limit error = %v", err)
	}
	release, err := svc.acquireSession(session.ID)
	if err != nil {
		t.Fatalf("first acquireSession() error = %v", err)
	}
	defer release()
	if _, err := svc.acquireSession(session.ID); ErrorCode(err) != "companion_concurrency_limited" {
		t.Fatalf("concurrency error = %v", err)
	}
	now = now.Add(3 * time.Minute)
	if _, err := svc.Authenticate(context.Background(), authorization, metadata, ""); ErrorCode(err) != "companion_session_expired" {
		t.Fatalf("expiry error = %v", err)
	}
}

func newPersistentCompanionTestService(t *testing.T, options Options) (*Service, *profile.Repository, string, RequestMetadata) {
	t.Helper()
	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	profiles := profile.NewRepository(conn)
	created, err := profiles.Create(context.Background(), "Companion test")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	_, err = profiles.UpsertIntegrationInstance(context.Background(), created.ID, profile.IntegrationInstancePatch{
		ProviderID: "ebay", Config: map[string]string{"region": "AU", "access_token": "must-not-project"},
	})
	if err != nil {
		t.Fatalf("create integration instance: %v", err)
	}
	svc, err := NewPersistentService(context.Background(), conn, profiles, DefaultModules(), options)
	if err != nil {
		t.Fatalf("NewPersistentService() error = %v", err)
	}
	metadata := RequestMetadata{Origin: companionTestOrigin, DeviceID: "device-a", RemoteAddress: "127.0.0.1:1000"}
	return svc, profiles, created.ID, metadata
}

func pairCompanionTestSession(t *testing.T, svc *Service, profileID string, metadata RequestMetadata, capabilities []string) string {
	t.Helper()
	receipt, err := svc.CreatePairingRequest(context.Background(), PairingRequestInput{
		DeviceID: metadata.DeviceID, DeviceName: "Test browser", ProtocolVersion: ProtocolVersionV1, Capabilities: capabilities,
	}, metadata)
	if err != nil {
		t.Fatalf("CreatePairingRequest() error = %v", err)
	}
	if _, err := svc.ApprovePairing(context.Background(), PairingApprovalInput{
		RequestID: receipt.RequestID, ProfileID: profileID,
	}, RequestMetadata{RemoteAddress: metadata.RemoteAddress}); err != nil {
		t.Fatalf("ApprovePairing() error = %v", err)
	}
	exchanged, err := svc.ExchangePairing(context.Background(), PairingExchangeInput{
		RequestID: receipt.RequestID, ExchangeSecret: receipt.ExchangeSecret,
		DeviceID: metadata.DeviceID, ProtocolVersion: ProtocolVersionV1,
	}, metadata)
	if err != nil {
		t.Fatalf("ExchangePairing() error = %v", err)
	}
	return "Bearer " + exchanged.Credential
}

func validPurchasePayload(t *testing.T, svc *Service, profileID, idempotencyKey, listingID string) PayloadSubmission {
	t.Helper()
	registry, err := svc.registryForProfile(context.Background(), profileID)
	if err != nil || len(registry.Modules) == 0 {
		t.Fatalf("load companion module for payload: %+v, %v", registry, err)
	}
	module := registry.Modules[0]
	data := map[string]any{"cards": []map[string]any{{
		"order_id": "ORDER-" + listingID, "listing_id": listingID, "title": "Captured item " + listingID,
		"quantity": 1, "item_price": "AU $2.40", "currency": "AUD",
		"item_url": "https://www.ebay.com/itm/" + listingID,
	}}}
	return PayloadSubmission{
		ProtocolVersion: ProtocolVersionV1, ProfileID: profileID, ModuleID: module.ID,
		ModuleVersion: module.ModuleVersion, SchemaVersion: module.FixtureVersion,
		IntegrationInstanceID: module.IntegrationInstanceID, ProviderID: module.ProviderID,
		URL: "https://www.ebay.com/mye/myebay/purchase", PayloadType: "purchase_order",
		CapturedAt: svc.options.Now().UTC().Format(time.RFC3339), PageComplete: true, Passive: true,
		ConfidenceScore: 0.9, RedactionSummary: append([]string(nil), module.RedactionRules...),
		PayloadHash: PayloadDigest(data), IdempotencyKey: idempotencyKey, Data: data,
	}
}
