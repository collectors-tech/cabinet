package forwarding

import "testing"

func TestNormalizeForwarderPackageImportPreservesProvenance(t *testing.T) {
	t.Parallel()

	pkg, err := NormalizePackageImport(PackageImport{
		ProfileID:         " profile-1 ",
		Provider:          " Stackry ",
		Source:            "email",
		ExternalPackageID: " PKG-123 ",
		ShipmentID:        " SHIP-9 ",
		TrackingNumber:    " 1Z999 ",
		Status:            " READY_TO_SHIP ",
		ReceivedAt:        "2026-05-24T08:30:00Z",
		Sender:            "eBay seller",
		WarehouseLocation: " Suite A-42 ",
		WeightGrams:       425,
		RawPayload: map[string]any{
			"stackry_id": "PKG-123",
			"source":     "email",
		},
	})
	if err != nil {
		t.Fatalf("NormalizePackageImport() error = %v", err)
	}

	if pkg.ProfileID != "profile-1" || pkg.Provider != ProviderStackry || pkg.Source != SourceEmail {
		t.Fatalf("expected normalized identity fields, got %+v", pkg)
	}
	if pkg.Status != StatusReadyToShip {
		t.Fatalf("expected ready_to_ship status, got %q", pkg.Status)
	}
	if pkg.ExternalPackageID != "PKG-123" || pkg.ShipmentID != "SHIP-9" || pkg.TrackingNumber != "1Z999" {
		t.Fatalf("expected external ids to be preserved, got %+v", pkg)
	}
	if pkg.ProvenanceKey != "stackry:email:PKG-123" {
		t.Fatalf("expected stable provenance key, got %q", pkg.ProvenanceKey)
	}
	if pkg.RawPayload["stackry_id"] != "PKG-123" {
		t.Fatalf("expected raw payload provenance to be retained, got %+v", pkg.RawPayload)
	}
}

func TestNormalizeForwarderPackageImportRequiresStableIdentity(t *testing.T) {
	t.Parallel()

	_, err := NormalizePackageImport(PackageImport{
		ProfileID: "profile-1",
		Provider:  ProviderStackry,
		Source:    SourceManual,
		Status:    StatusReceived,
	})
	if err == nil {
		t.Fatal("expected missing external package id to fail")
	}
	if got := err.Error(); got != "external_package_id is required" {
		t.Fatalf("expected external_package_id validation error, got %q", got)
	}
}

func TestPackageInboxDeduplicatesByProvenanceKey(t *testing.T) {
	t.Parallel()

	inbox := NewMemoryInbox()
	first, err := inbox.Upsert(PackageImport{
		ProfileID:         "profile-1",
		Provider:          ProviderStackry,
		Source:            SourceAPI,
		ExternalPackageID: "PKG-1",
		Status:            StatusReceived,
		WeightGrams:       100,
	})
	if err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}
	second, err := inbox.Upsert(PackageImport{
		ProfileID:         "profile-1",
		Provider:          ProviderStackry,
		Source:            SourceAPI,
		ExternalPackageID: "PKG-1",
		Status:            StatusReadyToShip,
		WeightGrams:       140,
	})
	if err != nil {
		t.Fatalf("second Upsert() error = %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected duplicate provenance to update existing package, got %q then %q", first.ID, second.ID)
	}
	if second.Status != StatusReadyToShip || second.WeightGrams != 140 {
		t.Fatalf("expected package update to retain latest package fields, got %+v", second)
	}
	packages := inbox.List("profile-1")
	if len(packages) != 1 {
		t.Fatalf("expected one deduplicated package, got %d", len(packages))
	}
}
