package forwarding

import (
	"database/sql"
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

func TestForwarderPackageServiceLinksPackageToExpectedArrival(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	seedForwarderPackageLinkTarget(t, conn, "profile-link", "item-link", "life-link", "arrival-link")
	svc := NewService(conn)
	pkg, err := svc.UpsertPackage(t.Context(), PackageImport{ProfileID: "profile-link", Provider: ProviderStackry, Source: SourceEmail, ExternalPackageID: "PKG-LINK", Status: StatusReadyToShip})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}

	link, err := svc.LinkPackage(t.Context(), PackageLinkRequest{
		ProfileID:         "profile-link",
		PackageID:         pkg.ID,
		ItemID:            "item-link",
		LifecycleEntryID:  "life-link",
		ExpectedArrivalID: "arrival-link",
		Source:            "manual-review",
		Notes:             "matched by tracking number",
	})
	if err != nil {
		t.Fatalf("LinkPackage() error = %v", err)
	}
	if link.PackageID != pkg.ID || link.ItemID != "item-link" || link.ExpectedArrivalID != "arrival-link" {
		t.Fatalf("expected package to link to purchase arrival, got %+v", link)
	}
	if link.Source != "manual-review" || link.Notes != "matched by tracking number" {
		t.Fatalf("expected link provenance, got %+v", link)
	}
	links, err := svc.ListPackageLinks(t.Context(), "profile-link", pkg.ID)
	if err != nil {
		t.Fatalf("ListPackageLinks() error = %v", err)
	}
	if len(links) != 1 || links[0].ID != link.ID {
		t.Fatalf("expected one stored link, got %+v", links)
	}
}

func TestForwarderPackageServiceRejectsAmbiguousPackageRelink(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	seedForwarderPackageLinkTarget(t, conn, "profile-link", "item-one", "life-one", "arrival-one")
	seedForwarderPackageLinkTarget(t, conn, "profile-link", "item-two", "life-two", "arrival-two")
	svc := NewService(conn)
	pkg, err := svc.UpsertPackage(t.Context(), PackageImport{ProfileID: "profile-link", Provider: ProviderStackry, Source: SourceManual, ExternalPackageID: "PKG-AMBIG", Status: StatusReceived})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}
	_, err = svc.LinkPackage(t.Context(), PackageLinkRequest{ProfileID: "profile-link", PackageID: pkg.ID, ItemID: "item-one", LifecycleEntryID: "life-one", ExpectedArrivalID: "arrival-one"})
	if err != nil {
		t.Fatalf("first LinkPackage() error = %v", err)
	}

	_, err = svc.LinkPackage(t.Context(), PackageLinkRequest{ProfileID: "profile-link", PackageID: pkg.ID, ItemID: "item-two", LifecycleEntryID: "life-two", ExpectedArrivalID: "arrival-two"})
	if err == nil {
		t.Fatal("expected ambiguous relink to fail")
	}
	if got := err.Error(); got != "forwarder package already linked to a different target" {
		t.Fatalf("expected ambiguous relink error, got %q", got)
	}
}

func TestForwarderPackageServiceOverridesAndUnlinksWithAuditEvents(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	seedForwarderPackageLinkTarget(t, conn, "profile-decision", "item-one", "life-one", "arrival-one")
	seedForwarderPackageLinkTarget(t, conn, "profile-decision", "item-two", "life-two", "arrival-two")
	svc := NewService(conn)
	pkg, err := svc.UpsertPackage(t.Context(), PackageImport{ProfileID: "profile-decision", Provider: ProviderStackry, Source: SourceManual, ExternalPackageID: "PKG-DECISION", Status: StatusReceived})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}
	confirmed, err := svc.LinkPackage(t.Context(), PackageLinkRequest{
		ProfileID: "profile-decision", PackageID: pkg.ID, ItemID: "item-one", LifecycleEntryID: "life-one", ExpectedArrivalID: "arrival-one",
		Source: "suggested-match", Decision: "confirmed", Actor: "reviewer", Notes: "high confidence suggestion accepted",
	})
	if err != nil {
		t.Fatalf("confirm LinkPackage() error = %v", err)
	}
	if confirmed.Decision != "confirmed" || len(confirmed.AuditTrail) == 0 || !strings.Contains(confirmed.AuditTrail[0], "decision=confirmed") {
		t.Fatalf("expected confirmed audit trail, got %+v", confirmed)
	}
	overridden, err := svc.LinkPackage(t.Context(), PackageLinkRequest{
		ProfileID: "profile-decision", PackageID: pkg.ID, ItemID: "item-two", LifecycleEntryID: "life-two", ExpectedArrivalID: "arrival-two",
		Source: "manual-override", Decision: "override", Override: true, Actor: "reviewer", Notes: "seller note contradicted package text",
	})
	if err != nil {
		t.Fatalf("override LinkPackage() error = %v", err)
	}
	if overridden.ItemID != "item-two" || overridden.Decision != "override" || !strings.Contains(strings.Join(overridden.AuditTrail, " "), "previous target item=item-one") {
		t.Fatalf("expected override target and audit trail, got %+v", overridden)
	}
	unlinked, err := svc.UnlinkPackage(t.Context(), PackageUnlinkRequest{ProfileID: "profile-decision", PackageID: pkg.ID, Source: "manual-unlink", Actor: "reviewer", Notes: "waiting for better evidence"})
	if err != nil {
		t.Fatalf("UnlinkPackage() error = %v", err)
	}
	if unlinked.Action != "unlinked" || unlinked.PreviousItemID != "item-two" || !strings.Contains(strings.Join(unlinked.AuditTrail, " "), "decision=unlinked") {
		t.Fatalf("expected unlink event with previous target, got %+v", unlinked)
	}
	if _, err := svc.GetPackageLink(t.Context(), "profile-decision", pkg.ID); err == nil {
		t.Fatal("expected package link to be removed after unlink")
	}
	events, err := svc.ListPackageLinkEvents(t.Context(), "profile-decision", pkg.ID)
	if err != nil {
		t.Fatalf("ListPackageLinkEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected confirm, override, and unlink events, got %+v", events)
	}
	seen := map[string]bool{}
	for _, event := range events {
		seen[event.Action] = true
	}
	if !seen["confirmed"] || !seen["override"] || !seen["unlinked"] {
		t.Fatalf("expected decision events, got %+v", events)
	}
}

func TestForwarderPackageServiceSuggestsDeterministicMatchesWithAuditSignals(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	seedForwarderPackageLinkTarget(t, conn, "profile-match", "item-match", "life-match", "arrival-match")
	if _, err := conn.Exec(`UPDATE canonical_items SET title = 'AFX Turbo slot car', brand = 'AFX', part_number = 'AFX-900' WHERE id = 'item-match'`); err != nil {
		t.Fatalf("update item: %v", err)
	}
	if _, err := conn.Exec(`UPDATE commerce_lifecycle_entries SET source = 'ebay', external_ref = 'ORDER-900', quantity = 2, notes = 'Seller: slot shop; tracking TRACK-900; package PKG-900' WHERE id = 'life-match'`); err != nil {
		t.Fatalf("update lifecycle: %v", err)
	}
	if _, err := conn.Exec(`UPDATE expected_arrivals SET external_ref = 'ORDER-900', quantity = 2, expected_on = '2026-05-27', notes = 'TRACK-900 PKG-900 AFX Turbo' WHERE id = 'arrival-match'`); err != nil {
		t.Fatalf("update arrival: %v", err)
	}
	svc := NewService(conn)
	pkg, err := svc.UpsertPackage(t.Context(), PackageImport{
		ProfileID:         "profile-match",
		Provider:          ProviderStackry,
		Source:            SourceEmail,
		ExternalPackageID: "PKG-900",
		Status:            StatusReceived,
		TrackingNumber:    "TRACK-900",
		ReceivedAt:        "2026-05-27T09:15:00Z",
		Sender:            "slot shop",
		RawPayload: map[string]any{
			"title":    "AFX Turbo slot car",
			"quantity": 2,
		},
	})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}

	suggestions, err := svc.SuggestPackageMatches(t.Context(), "profile-match", pkg.ID)
	if err != nil {
		t.Fatalf("SuggestPackageMatches() error = %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected one match suggestion, got %+v", suggestions)
	}
	got := suggestions[0]
	if got.PackageID != pkg.ID || got.ItemID != "item-match" || got.ExpectedArrivalID != "arrival-match" {
		t.Fatalf("unexpected match target: %+v", got)
	}
	if got.Confidence != 1 || got.ConfidenceLabel != "high" {
		t.Fatalf("expected high full-confidence suggestion, got %+v", got)
	}
	if len(got.Signals) != 6 || len(got.Explanation) != 6 {
		t.Fatalf("expected six scored signals and explanations, got %+v", got)
	}
	if len(got.AuditTrail) == 0 || !strings.Contains(got.AuditTrail[0], "without mutating") {
		t.Fatalf("expected non-mutating audit trail, got %+v", got.AuditTrail)
	}
}

func TestForwarderPackageServiceSuggestionsExcludeAlreadyLinkedPackages(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(t.Context(), t.TempDir()+"/cabinet.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	seedForwarderPackageLinkTarget(t, conn, "profile-linked", "item-linked", "life-linked", "arrival-linked")
	svc := NewService(conn)
	pkg, err := svc.UpsertPackage(t.Context(), PackageImport{
		ProfileID:         "profile-linked",
		Provider:          ProviderStackry,
		Source:            SourceManual,
		ExternalPackageID: "PKG-LINKED",
		Status:            StatusReceived,
	})
	if err != nil {
		t.Fatalf("UpsertPackage() error = %v", err)
	}
	if _, err := svc.LinkPackage(t.Context(), PackageLinkRequest{ProfileID: "profile-linked", PackageID: pkg.ID, ItemID: "item-linked", LifecycleEntryID: "life-linked", ExpectedArrivalID: "arrival-linked"}); err != nil {
		t.Fatalf("LinkPackage() error = %v", err)
	}

	suggestions, err := svc.SuggestPackageMatches(t.Context(), "profile-linked", pkg.ID)
	if err != nil {
		t.Fatalf("SuggestPackageMatches() error = %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected linked package to be excluded from suggestions, got %+v", suggestions)
	}
}

func TestForwarderPackageMatchSuggestionSummaryCountsConfidenceBuckets(t *testing.T) {
	t.Parallel()

	summary := SummarizePackageMatchSuggestions("pkg-scoped", []PackageMatchSuggestion{
		{PackageID: "pkg-scoped", ConfidenceLabel: "high"},
		{PackageID: "pkg-scoped", ConfidenceLabel: "medium"},
		{PackageID: "pkg-other", ConfidenceLabel: "low"},
		{PackageID: "pkg-other", ConfidenceLabel: ""},
	})

	if summary.Count != 4 || summary.ScopedPackages != 2 {
		t.Fatalf("expected scoped four-suggestion summary, got %+v", summary)
	}
	if summary.HighConfidence != 1 || summary.MediumConfidence != 1 || summary.LowConfidence != 2 {
		t.Fatalf("expected confidence bucket counts, got %+v", summary)
	}

	emptyScoped := SummarizePackageMatchSuggestions("pkg-scoped", nil)
	if emptyScoped.Count != 0 || emptyScoped.ScopedPackages != 0 {
		t.Fatalf("expected empty scoped summary to derive zero scoped packages, got %+v", emptyScoped)
	}
}

func TestPackageMatchSuggestionConfidenceFilter(t *testing.T) {
	t.Parallel()

	suggestions := []PackageMatchSuggestion{
		{PackageID: "pkg-high", ConfidenceLabel: "high"},
		{PackageID: "pkg-medium", ConfidenceLabel: "medium"},
		{PackageID: "pkg-low", ConfidenceLabel: "low"},
	}

	label, err := NormalizePackageMatchConfidenceFilter(" Medium ")
	if err != nil {
		t.Fatalf("NormalizePackageMatchConfidenceFilter() error = %v", err)
	}
	filtered := FilterPackageMatchSuggestionsByConfidence(suggestions, label)
	if len(filtered) != 1 || filtered[0].PackageID != "pkg-medium" {
		t.Fatalf("expected only medium confidence suggestion, got %+v", filtered)
	}

	unfiltered := FilterPackageMatchSuggestionsByConfidence(suggestions, "")
	if len(unfiltered) != len(suggestions) {
		t.Fatalf("expected empty filter to preserve suggestions, got %+v", unfiltered)
	}

	if _, err := NormalizePackageMatchConfidenceFilter("certain"); err == nil {
		t.Fatalf("expected invalid confidence label error")
	}
}

func seedForwarderPackageLinkTarget(t *testing.T, conn interface {
	Exec(string, ...any) (sql.Result, error)
}, profileID, itemID, lifecycleID, arrivalID string) {
	t.Helper()
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES (?, ?, 'AFX', 'Slot', ?, 'Forwarder Link Target')`, itemID, profileID, itemID); err != nil {
		t.Fatalf("seed item %s: %v", itemID, err)
	}
	if _, err := conn.Exec(`INSERT INTO commerce_lifecycle_entries(id, profile_id, item_id, state, source, external_ref, quantity, amount, currency, notes) VALUES (?, ?, ?, 'purchase', 'ebay', ?, 1, 10, 'AUD', '')`, lifecycleID, profileID, itemID, lifecycleID); err != nil {
		t.Fatalf("seed lifecycle %s: %v", lifecycleID, err)
	}
	if _, err := conn.Exec(`INSERT INTO expected_arrivals(id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, notes) VALUES (?, ?, ?, ?, 'ebay', ?, 1, 10, 'AUD', 'expected', '')`, arrivalID, profileID, itemID, lifecycleID, arrivalID); err != nil {
		t.Fatalf("seed arrival %s: %v", arrivalID, err)
	}
}
