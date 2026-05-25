package forwarding

import (
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

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

func TestForwarderPackageServicePersistsAndListsImports(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	svc := NewService(conn)

	created, err := svc.UpsertPackage(t.Context(), PackageImport{
		ProfileID:         "profile-1",
		Provider:          ProviderStackry,
		Source:            SourceCSV,
		ExternalPackageID: "PKG-77",
		ShipmentID:        "SHIP-77",
		TrackingNumber:    "TRACK-77",
		Status:            StatusReceived,
		WeightGrams:       320,
		RawPayload:        map[string]any{"row": "77"},
	})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}
	updated, err := svc.UpsertPackage(t.Context(), PackageImport{
		ProfileID:         "profile-1",
		Provider:          ProviderStackry,
		Source:            SourceCSV,
		ExternalPackageID: "PKG-77",
		ShipmentID:        "SHIP-77",
		TrackingNumber:    "TRACK-77",
		Status:            StatusReadyToShip,
		WeightGrams:       420,
		RawPayload:        map[string]any{"row": "77", "status": "ready"},
	})
	if err != nil {
		t.Fatalf("second UpsertPackage() error = %v", err)
	}
	if created.ID != updated.ID {
		t.Fatalf("expected persistent upsert to reuse id, got %q then %q", created.ID, updated.ID)
	}
	packages, err := svc.ListPackages(t.Context(), "profile-1", StatusReadyToShip)
	if err != nil {
		t.Fatalf("ListPackages() error = %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected one ready package, got %d", len(packages))
	}
	if packages[0].Status != StatusReadyToShip || packages[0].WeightGrams != 420 || packages[0].RawPayload["status"] != "ready" {
		t.Fatalf("expected latest package fields and raw payload, got %+v", packages[0])
	}
}

func TestParsePackageCSVBuildsImportsAndPreservesRawRows(t *testing.T) {
	t.Parallel()

	csv := strings.NewReader(`Stackry Package ID,Status,Shipment ID,Tracking Number,Received At,Sender,Warehouse Location,Weight Grams
PKG-CSV-1,received,SHIP-1,TRACK-1,2026-05-25T03:51:00Z,eBay seller,A-12,425
PKG-CSV-2,ready_to_ship,SHIP-2,TRACK-2,2026-05-25T04:10:00Z,Card shop,B-7,510
`)
	imports, rowErrors, err := ParsePackageCSV("profile-csv", ProviderStackry, csv)
	if err != nil {
		t.Fatalf("ParsePackageCSV() error = %v", err)
	}
	if len(rowErrors) != 0 {
		t.Fatalf("expected no row errors, got %+v", rowErrors)
	}
	if len(imports) != 2 {
		t.Fatalf("expected two imports, got %d", len(imports))
	}
	first := imports[0]
	if first.ProfileID != "profile-csv" || first.Provider != ProviderStackry || first.Source != SourceCSV {
		t.Fatalf("expected CSV import identity, got %+v", first)
	}
	if first.ExternalPackageID != "PKG-CSV-1" || first.ShipmentID != "SHIP-1" || first.TrackingNumber != "TRACK-1" || first.WeightGrams != 425 {
		t.Fatalf("expected first CSV row fields to map into package import, got %+v", first)
	}
	if first.RawPayload["row"] != 2 || first.RawPayload["stackry_package_id"] != "PKG-CSV-1" {
		t.Fatalf("expected raw CSV row provenance, got %+v", first.RawPayload)
	}
	pkg, err := NormalizePackageImport(first)
	if err != nil {
		t.Fatalf("NormalizePackageImport(first CSV import) error = %v", err)
	}
	if pkg.ProvenanceKey != "stackry:csv:PKG-CSV-1" {
		t.Fatalf("expected CSV provenance key, got %q", pkg.ProvenanceKey)
	}
}

func TestParsePackageCSVReportsRowErrorsWithoutDroppingValidRows(t *testing.T) {
	t.Parallel()

	csv := strings.NewReader(`package_id,status,weight_grams
PKG-OK,received,100
,received,200
PKG-BAD-WEIGHT,received,abc
PKG-READY,ready_to_ship,250
`)
	imports, rowErrors, err := ParsePackageCSV("profile-csv", ProviderStackry, csv)
	if err != nil {
		t.Fatalf("ParsePackageCSV() error = %v", err)
	}
	if len(imports) != 2 {
		t.Fatalf("expected two valid imports, got %d: %+v", len(imports), imports)
	}
	if imports[0].ExternalPackageID != "PKG-OK" || imports[1].ExternalPackageID != "PKG-READY" {
		t.Fatalf("expected valid rows to be preserved in order, got %+v", imports)
	}
	if len(rowErrors) != 2 {
		t.Fatalf("expected two row errors, got %+v", rowErrors)
	}
	if rowErrors[0].Row != 3 || rowErrors[0].Error != "external_package_id is required" {
		t.Fatalf("expected row 3 identity error, got %+v", rowErrors[0])
	}
	if rowErrors[1].Row != 4 || rowErrors[1].Error != "weight_grams must be an integer" {
		t.Fatalf("expected row 4 weight error, got %+v", rowErrors[1])
	}
}

func TestParsePackageEmailBuildsImportAndPreservesMessageProvenance(t *testing.T) {
	t.Parallel()

	email := strings.NewReader(`From: notifications@stackry.example
Subject: Package PKG-EMAIL-1 received at your locker

Package ID: PKG-EMAIL-1
Status: Ready to Ship
Shipment ID: SHIP-EMAIL-1
Tracking Number: 1ZEMAIL
Received At: 2026-05-25T04:30:00Z
Sender: eBay seller
Warehouse Location: Suite A-42
Weight: 425 g
`)
	imported, err := ParsePackageEmail("profile-email", ProviderStackry, "msg-123", email)
	if err != nil {
		t.Fatalf("ParsePackageEmail() error = %v", err)
	}

	if imported.ProfileID != "profile-email" || imported.Provider != ProviderStackry || imported.Source != SourceEmail {
		t.Fatalf("expected email import identity, got %+v", imported)
	}
	if imported.ExternalPackageID != "PKG-EMAIL-1" || imported.Status != StatusReadyToShip || imported.WeightGrams != 425 {
		t.Fatalf("expected email fields to map into package import, got %+v", imported)
	}
	if imported.ShipmentID != "SHIP-EMAIL-1" || imported.TrackingNumber != "1ZEMAIL" {
		t.Fatalf("expected shipment and tracking fields, got %+v", imported)
	}
	if imported.RawPayload["source"] != "email" || imported.RawPayload["message_id"] != "msg-123" {
		t.Fatalf("expected message provenance to be retained, got %+v", imported.RawPayload)
	}
	pkg, err := NormalizePackageImport(imported)
	if err != nil {
		t.Fatalf("NormalizePackageImport(email import) error = %v", err)
	}
	if pkg.ProvenanceKey != "stackry:email:PKG-EMAIL-1" {
		t.Fatalf("expected email provenance key, got %q", pkg.ProvenanceKey)
	}
}

func TestParsePackageEmailRejectsMissingPackageIdentity(t *testing.T) {
	t.Parallel()

	_, err := ParsePackageEmail("profile-email", ProviderStackry, "msg-missing", strings.NewReader(`Subject: package update

Status: received
Weight: 100 g
`))
	if err == nil {
		t.Fatal("expected missing external package id to fail")
	}
	if got := err.Error(); got != "external_package_id is required" {
		t.Fatalf("expected identity validation error, got %q", got)
	}
}
