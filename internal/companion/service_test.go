package companion

import "testing"

func TestCompanionRegistryNormalizesPassiveModules(t *testing.T) {
	t.Parallel()

	svc := NewService([]Module{{ID: " ebay-purchase-capture ", Site: " ebay ", Actions: []string{"capture_tracking", "capture_item"}}})
	registry := svc.Registry()
	if len(registry.Modules) != 1 {
		t.Fatalf("expected one module, got %+v", registry.Modules)
	}
	module := registry.Modules[0]
	if module.ID != "ebay-purchase-capture" || module.Site != "ebay" || !module.PassiveOnly {
		t.Fatalf("unexpected module %+v", module)
	}
	if got := module.Actions; len(got) != 2 || got[0] != "capture_item" || got[1] != "capture_tracking" {
		t.Fatalf("actions were not sorted and preserved: %+v", got)
	}
}

func TestCompanionAcceptPayloadRequiresProfileScopedBearerToken(t *testing.T) {
	t.Parallel()

	svc := DefaultService()
	_, err := svc.AcceptPayload(PayloadSubmission{
		ProfileID:       "profile-1",
		ModuleID:        "ebay-purchase-capture",
		URL:             "https://www.ebay.com/itm/123",
		PayloadType:     "purchase_order",
		Passive:         true,
		ConfidenceScore: 0.9,
	}, "Bearer companion:other-profile")
	if err == nil || err.Error() != "companion_auth_required" {
		t.Fatalf("expected profile-scoped auth error, got %v", err)
	}
}

func TestCompanionAcceptPayloadBlocksWritesAndAcceptsPassiveCapture(t *testing.T) {
	t.Parallel()

	svc := DefaultService()
	_, err := svc.AcceptPayload(PayloadSubmission{
		ProfileID:       "profile-1",
		ModuleID:        "ebay-purchase-capture",
		URL:             "https://www.ebay.com/itm/123",
		PayloadType:     "purchase_order",
		Passive:         false,
		AttemptedWrite:  true,
		ConfidenceScore: 0.9,
	}, "Bearer companion:profile-1")
	if err == nil || err.Error() != "companion_payload_must_be_passive" {
		t.Fatalf("expected passive capture error, got %v", err)
	}

	accepted, err := svc.AcceptPayload(PayloadSubmission{
		ProfileID:       "profile-1",
		ModuleID:        "ebay-purchase-capture",
		URL:             "https://www.ebay.com/itm/123",
		PayloadType:     "purchase_order",
		Passive:         true,
		ConfidenceScore: 0.9,
		Data:            map[string]any{"order_id": "123"},
	}, "Bearer companion:profile-1")
	if err != nil {
		t.Fatalf("AcceptPayload error = %v", err)
	}
	if !accepted.Accepted || accepted.SyncMode != SyncModePassiveCapture || accepted.RemoteWrite || accepted.ConfidenceLabel != "high" {
		t.Fatalf("unexpected accepted payload %+v", accepted)
	}
}
