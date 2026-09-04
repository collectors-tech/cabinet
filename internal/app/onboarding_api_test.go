package app

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestOnboardingSampleDataEndpointIsIdempotent(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
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

	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	firstSeed := doRequest(t, a, http.MethodPost, "/api/onboarding/sample-data", nil, nil)
	if firstSeed.Code != http.StatusOK {
		t.Fatalf("first sample seed status=%d body=%s", firstSeed.Code, firstSeed.Body.String())
	}
	var firstPayload struct {
		DatasetKind             string `json:"dataset_kind"`
		DatasetLabel            string `json:"dataset_label"`
		SampleDataDisclosure    string `json:"sample_data_disclosure"`
		CreatedItems            int    `json:"created_items"`
		CreatedWishlist         int    `json:"created_wishlist_entries"`
		TotalItems              int    `json:"total_items"`
		TotalWishlist           int    `json:"total_wishlist_entries"`
		AlreadySeededForProfile bool   `json:"already_seeded_for_profile"`
	}
	if err := json.NewDecoder(firstSeed.Body).Decode(&firstPayload); err != nil {
		t.Fatalf("decode first seed payload: %v", err)
	}
	if firstPayload.CreatedItems == 0 {
		t.Fatalf("expected created_items > 0, got %+v", firstPayload)
	}
	if firstPayload.DatasetKind != "sample_showcase" {
		t.Fatalf("expected sample_showcase dataset kind, got %+v", firstPayload)
	}
	if !strings.Contains(strings.ToLower(firstPayload.DatasetLabel), "sample") {
		t.Fatalf("expected sample dataset label, got %+v", firstPayload)
	}
	if !strings.Contains(strings.ToLower(firstPayload.SampleDataDisclosure), "example records") {
		t.Fatalf("expected explicit sample data disclosure, got %+v", firstPayload)
	}
	if firstPayload.AlreadySeededForProfile {
		t.Fatalf("expected already_seeded_for_profile=false on first run, got %+v", firstPayload)
	}

	secondSeed := doRequest(t, a, http.MethodPost, "/api/onboarding/sample-data", nil, nil)
	if secondSeed.Code != http.StatusOK {
		t.Fatalf("second sample seed status=%d body=%s", secondSeed.Code, secondSeed.Body.String())
	}
	var secondPayload struct {
		DatasetKind             string `json:"dataset_kind"`
		DatasetLabel            string `json:"dataset_label"`
		SampleDataDisclosure    string `json:"sample_data_disclosure"`
		CreatedItems            int    `json:"created_items"`
		CreatedWishlist         int    `json:"created_wishlist_entries"`
		TotalItems              int    `json:"total_items"`
		TotalWishlist           int    `json:"total_wishlist_entries"`
		AlreadySeededForProfile bool   `json:"already_seeded_for_profile"`
	}
	if err := json.NewDecoder(secondSeed.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("decode second seed payload: %v", err)
	}
	if secondPayload.CreatedItems != 0 {
		t.Fatalf("expected no new items on second run, got %+v", secondPayload)
	}
	if secondPayload.DatasetKind != firstPayload.DatasetKind ||
		secondPayload.DatasetLabel != firstPayload.DatasetLabel ||
		secondPayload.SampleDataDisclosure != firstPayload.SampleDataDisclosure {
		t.Fatalf("expected stable sample provenance across rerun, first=%+v second=%+v", firstPayload, secondPayload)
	}
	if !secondPayload.AlreadySeededForProfile {
		t.Fatalf("expected already_seeded_for_profile=true on second run, got %+v", secondPayload)
	}
	if secondPayload.TotalItems != firstPayload.TotalItems {
		t.Fatalf("expected stable total_items across reruns, first=%d second=%d", firstPayload.TotalItems, secondPayload.TotalItems)
	}

	categories := make(map[string]struct{})
	for _, status := range []string{"active", "wishlist"} {
		itemsResp := doRequest(t, a, http.MethodGet, "/api/items?status="+status, nil, nil)
		if itemsResp.Code != http.StatusOK {
			t.Fatalf("list items status=%q code=%d body=%s", status, itemsResp.Code, itemsResp.Body.String())
		}
		var itemsPayload struct {
			Items []struct {
				Category string `json:"category"`
			} `json:"items"`
		}
		if err := json.NewDecoder(itemsResp.Body).Decode(&itemsPayload); err != nil {
			t.Fatalf("decode items payload for status=%q: %v", status, err)
		}
		for _, item := range itemsPayload.Items {
			if strings.TrimSpace(item.Category) != "" {
				categories[strings.TrimSpace(item.Category)] = struct{}{}
			}
		}
	}
	if len(categories) < 6 {
		keys := make([]string, 0, len(categories))
		for category := range categories {
			keys = append(keys, category)
		}
		sort.Strings(keys)
		t.Fatalf("expected representative seeded category coverage (>=6 categories), got %d: %v", len(categories), keys)
	}
	for _, required := range []string{"Diecast", "Slot Car", "Trading Card", "Action Figure", "Comic", "Model Kit"} {
		if _, ok := categories[required]; !ok {
			keys := make([]string, 0, len(categories))
			for category := range categories {
				keys = append(keys, category)
			}
			sort.Strings(keys)
			t.Fatalf("expected seeded category %q, got %v", required, keys)
		}
	}
	if firstPayload.TotalItems < 6 {
		t.Fatalf("expected at least 6 seeded items for representative sample content, got %+v", firstPayload)
	}
	if firstPayload.TotalItems < 30 {
		t.Fatalf("expected richer inventory sample content across folders, got %+v", firstPayload)
	}
	if firstPayload.TotalWishlist < 3 {
		t.Fatalf("expected seeded wishlist coverage, got %+v", firstPayload)
	}

	settingsResp := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/settings", nil, nil)
	if settingsResp.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settingsResp.Code, settingsResp.Body.String())
	}
	var settingsPayload struct {
		Settings map[string]string `json:"settings"`
	}
	if err := json.NewDecoder(settingsResp.Body).Decode(&settingsPayload); err != nil {
		t.Fatalf("decode settings payload: %v", err)
	}
	var folderAssignments map[string]string
	if err := json.Unmarshal([]byte(settingsPayload.Settings["inventory.folder-item-assignments.v1"]), &folderAssignments); err != nil {
		t.Fatalf("decode inventory folder assignments: %v", err)
	}
	if len(folderAssignments) < 30 {
		t.Fatalf("expected folder assignments for richer sample inventory, got %d: %+v", len(folderAssignments), folderAssignments)
	}
	if settingsPayload.Settings["onboarding.sample_data.dataset_kind"] != "sample_showcase" {
		t.Fatalf("expected persisted sample dataset kind, got %+v", settingsPayload.Settings)
	}
	if !strings.Contains(strings.ToLower(settingsPayload.Settings["onboarding.sample_data.disclosure"]), "example records") {
		t.Fatalf("expected persisted sample disclosure, got %+v", settingsPayload.Settings)
	}
	seenFolders := make(map[string]struct{})
	for itemID, folderName := range folderAssignments {
		if strings.TrimSpace(itemID) == "" || strings.TrimSpace(folderName) == "" {
			t.Fatalf("expected non-empty item and folder assignments, got item=%q folder=%q", itemID, folderName)
		}
		seenFolders[strings.TrimSpace(folderName)] = struct{}{}
	}
	for _, requiredFolder := range []string{"Store 1", "Store 3", "Warehouse 2", "Archive A"} {
		if _, ok := seenFolders[requiredFolder]; !ok {
			t.Fatalf("expected seeded folder assignment for %q, got folders=%+v", requiredFolder, seenFolders)
		}
	}

	wishlistResp := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if wishlistResp.Code != http.StatusOK {
		t.Fatalf("wishlist sample status=%d body=%s", wishlistResp.Code, wishlistResp.Body.String())
	}
	var wishlistPayload struct {
		Items []struct {
			Priority       string `json:"priority"`
			Notes          string `json:"notes"`
			BelowTargetNow bool   `json:"below_target_now"`
			HighlightHit   bool   `json:"highlight_hit"`
		} `json:"items"`
	}
	if err := json.NewDecoder(wishlistResp.Body).Decode(&wishlistPayload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if len(wishlistPayload.Items) < 5 {
		t.Fatalf("expected at least 5 representative wishlist sample rows, got %+v", wishlistPayload.Items)
	}
	priorities := make(map[string]struct{})
	hasBelowTarget := false
	hasActionableNote := false
	hasHighlightHit := false
	for _, entry := range wishlistPayload.Items {
		priorities[strings.TrimSpace(entry.Priority)] = struct{}{}
		hasBelowTarget = hasBelowTarget || entry.BelowTargetNow
		hasActionableNote = hasActionableNote || strings.Contains(strings.ToLower(entry.Notes), "sample")
		hasHighlightHit = hasHighlightHit || entry.HighlightHit
	}
	for _, required := range []string{"high", "medium", "low"} {
		if _, ok := priorities[required]; !ok {
			t.Fatalf("expected representative wishlist priority %q, got %+v", required, priorities)
		}
	}
	if !hasBelowTarget {
		t.Fatalf("expected at least one below-target wishlist sample row, got %+v", wishlistPayload.Items)
	}
	if !hasActionableNote {
		t.Fatalf("expected wishlist sample rows to include review notes, got %+v", wishlistPayload.Items)
	}
	if !hasHighlightHit {
		t.Fatalf("expected wishlist sample rows to include highlighted hit coverage, got %+v", wishlistPayload.Items)
	}
}

func TestOnboardingSampleDataEndpointRollsBackDatabaseAndMediaOnFailure(t *testing.T) {
	a := newTestApp(t)
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
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	if _, err := a.db.Exec(`
		CREATE TRIGGER fail_onboarding_sample_price
		BEFORE INSERT ON price_snapshots
		WHEN NEW.id LIKE 'sample-price-%'
		BEGIN
			SELECT RAISE(ABORT, 'forced onboarding seed failure');
		END;
	`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}

	failedSeed := doRequest(t, a, http.MethodPost, "/api/onboarding/sample-data", nil, nil)
	if failedSeed.Code != http.StatusBadRequest {
		t.Fatalf("failed seed status=%d body=%s", failedSeed.Code, failedSeed.Body.String())
	}

	for table, query := range map[string]string{
		"items":  `SELECT COUNT(*) FROM canonical_items WHERE profile_id = ?`,
		"photos": `SELECT COUNT(*) FROM item_photos ip JOIN canonical_items ci ON ci.id = ip.item_id WHERE ci.profile_id = ?`,
	} {
		var count int
		if err := a.db.QueryRow(query, p.ID).Scan(&count); err != nil {
			t.Fatalf("count %s after rollback: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected no %s after rollback, got %d", table, count)
		}
	}

	assetRoot := filepath.Join(a.cfg.DataDir, "media", "assets")
	entries, err := os.ReadDir(assetRoot)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read media assets after rollback: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no canonical media assets after rollback, got %d", len(entries))
	}
}
