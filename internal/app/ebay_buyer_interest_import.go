package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/google/uuid"
)

type ebayBuyerInterestImportResult struct {
	ebay.BuyerInterestMapping
	PersistedID string `json:"persisted_id"`
	ItemID      string `json:"item_id,omitempty"`
	CandidateID string `json:"candidate_id,omitempty"`
}

func registerEbayBuyerInterestImportRoute(mux *http.ServeMux, conn *sql.DB, profiles *profile.Repository) {
	mux.HandleFunc("/api/providers/ebay/buyer-interest/import", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SourceAccount       string                    `json:"source_account"`
			WriteBackCapability string                    `json:"write_back_capability"`
			Items               []ebay.BuyerInterestInput `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if len(req.Items) == 0 {
			http.Error(w, `{"error":"missing_items"}`, http.StatusBadRequest)
			return
		}

		profileID := ""
		if active, err := profiles.GetActiveProfile(r.Context()); err == nil {
			profileID = strings.TrimSpace(active.ID)
		}
		results := make([]ebayBuyerInterestImportResult, 0, len(req.Items))
		summary := map[string]int{
			ebay.InterestDestinationWishlist:  0,
			ebay.InterestDestinationDiscovery: 0,
			"write_back_allowed":              0,
			"write_back_blocked":              0,
		}
		for _, item := range req.Items {
			if strings.TrimSpace(item.SourceAccount) == "" {
				item.SourceAccount = req.SourceAccount
			}
			if strings.TrimSpace(item.WriteBackCapability) == "" {
				item.WriteBackCapability = req.WriteBackCapability
			}
			mapped := ebay.MapBuyerInterest(item)
			persisted, err := persistEbayBuyerInterest(r.Context(), conn, profileID, mapped)
			if err != nil {
				http.Error(w, `{"error":"failed_to_import_buyer_interest"}`, http.StatusBadRequest)
				return
			}
			results = append(results, persisted)
			summary[mapped.Destination]++
			if mapped.WriteBackAllowed {
				summary["write_back_allowed"]++
			} else {
				summary["write_back_blocked"]++
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "ebay",
			"mode":     "import",
			"items":    results,
			"summary":  summary,
		})
	})
}

func persistEbayBuyerInterest(ctx context.Context, conn *sql.DB, profileID string, mapped ebay.BuyerInterestMapping) (ebayBuyerInterestImportResult, error) {
	if strings.TrimSpace(mapped.ListingID) == "" {
		return ebayBuyerInterestImportResult{}, fmt.Errorf("listing_id is required")
	}
	if mapped.Destination == ebay.InterestDestinationWishlist {
		itemID, entryID, err := persistEbayBuyerInterestWishlist(ctx, conn, profileID, mapped)
		if err != nil {
			return ebayBuyerInterestImportResult{}, err
		}
		return ebayBuyerInterestImportResult{BuyerInterestMapping: mapped, PersistedID: entryID, ItemID: itemID}, nil
	}
	candidateID, err := persistEbayBuyerInterestDiscovery(ctx, conn, profileID, mapped)
	if err != nil {
		return ebayBuyerInterestImportResult{}, err
	}
	return ebayBuyerInterestImportResult{BuyerInterestMapping: mapped, PersistedID: candidateID, CandidateID: candidateID}, nil
}

func persistEbayBuyerInterestWishlist(ctx context.Context, conn *sql.DB, profileID string, mapped ebay.BuyerInterestMapping) (string, string, error) {
	itemID := deterministicEbayBuyerInterestID("item", mapped.ProvenanceKey)
	entryID := deterministicEbayBuyerInterestID("wishlist", mapped.ProvenanceKey)
	partNumber := "EBAY-" + ebayBuyerInterestDigest(mapped.ListingID)[:12]
	notes := ebayBuyerInterestMetadataNote(mapped)
	sourceURLs, _ := json.Marshal([]string{strings.TrimSpace(mapped.URL)})
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority, notes, source_urls_json, updated_by)
		VALUES (?, ?, 'eBay', 'Buyer interest', ?, ?, 'wishlist', 'medium', ?, ?, 'ebay.buyer_interest_import')
		ON CONFLICT(part_number) DO UPDATE SET
			status = 'wishlist',
			priority = 'medium',
			title = excluded.title,
			notes = excluded.notes,
			source_urls_json = excluded.source_urls_json,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = 'ebay.buyer_interest_import'
	`, itemID, profileID, partNumber, strings.TrimSpace(mapped.Title), notes, string(sourceURLs)); err != nil {
		return "", "", fmt.Errorf("upsert buyer interest item: %w", err)
	}
	if err := conn.QueryRowContext(ctx, `SELECT id FROM canonical_items WHERE part_number = ?`, partNumber).Scan(&itemID); err != nil {
		return "", "", fmt.Errorf("load buyer interest item: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit, owned, purchase_url, needed_quantity)
		VALUES (?, ?, ?, 0, 'medium', ?, 1, 0, ?, 1)
		ON CONFLICT(item_id) DO UPDATE SET
			notes = excluded.notes,
			highlight_hit = 1,
			owned = 0,
			purchase_url = excluded.purchase_url,
			updated_at = CURRENT_TIMESTAMP
	`, entryID, profileID, itemID, notes, strings.TrimSpace(mapped.URL)); err != nil {
		return "", "", fmt.Errorf("upsert buyer interest wishlist entry: %w", err)
	}
	if err := conn.QueryRowContext(ctx, `SELECT id FROM wishlist_entries WHERE item_id = ?`, itemID).Scan(&entryID); err != nil {
		return "", "", fmt.Errorf("load buyer interest wishlist entry: %w", err)
	}
	return itemID, entryID, nil
}

func persistEbayBuyerInterestDiscovery(ctx context.Context, conn *sql.DB, profileID string, mapped ebay.BuyerInterestMapping) (string, error) {
	const querySetID = "ebay-buyer-interest"
	candidateID := deterministicEbayBuyerInterestID("candidate", mapped.ProvenanceKey)
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, provider_scope_json, enabled)
		VALUES (?, ?, 'eBay buyer interest', '["ebay buyer interest"]', '["ebay"]', 1)
		ON CONFLICT(id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	`, querySetID, profileID); err != nil {
		return "", fmt.Errorf("upsert buyer interest query set: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, url, source, last_seen, status, stock_state)
		VALUES (?, ?, ?, ?, ?, ?, 'ebay_buyer_interest', COALESCE(NULLIF(?, ''), CURRENT_TIMESTAMP), 'new', 'unknown')
		ON CONFLICT(profile_id, query_set_id, source, listing_id) DO UPDATE SET
			title = excluded.title,
			url = excluded.url,
			last_seen = excluded.last_seen,
			source = excluded.source,
			profile_id = excluded.profile_id
	`, candidateID, profileID, querySetID, mapped.ProvenanceKey, strings.TrimSpace(mapped.Title), strings.TrimSpace(mapped.URL), strings.TrimSpace(mapped.ObservedAt)); err != nil {
		return "", fmt.Errorf("upsert buyer interest discovery candidate: %w", err)
	}
	if err := conn.QueryRowContext(ctx, `SELECT id FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND source = 'ebay_buyer_interest' AND listing_id = ?`, profileID, querySetID, mapped.ProvenanceKey).Scan(&candidateID); err != nil {
		return "", fmt.Errorf("load buyer interest discovery candidate: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number)
		VALUES (?, '', 'not_in_collection', 1, 1, ?)
		ON CONFLICT(candidate_id) DO UPDATE SET
			state = 'not_in_collection',
			confidence = 1,
			needs_review = 1,
			extracted_part_number = excluded.extracted_part_number,
			updated_at = CURRENT_TIMESTAMP
	`, candidateID, mapped.ProvenanceKey); err != nil {
		return "", fmt.Errorf("upsert buyer interest discovery match: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO discovery_actions(id, candidate_id, action_type, payload_json)
		VALUES (?, ?, 'buyer_interest_import', ?)
		ON CONFLICT(id) DO UPDATE SET
			payload_json = excluded.payload_json,
			created_at = CURRENT_TIMESTAMP
	`, deterministicEbayBuyerInterestID("discovery-action", mapped.ProvenanceKey), candidateID, ebayBuyerInterestMetadataJSON(mapped)); err != nil {
		return "", fmt.Errorf("record buyer interest discovery action: %w", err)
	}
	return candidateID, nil
}

func ebayBuyerInterestMetadataNote(mapped ebay.BuyerInterestMapping) string {
	return "[ebay_buyer_interest]" + ebayBuyerInterestMetadataJSON(mapped)
}

func ebayBuyerInterestMetadataJSON(mapped ebay.BuyerInterestMapping) string {
	raw, err := json.Marshal(map[string]any{
		"provenance_key":        mapped.ProvenanceKey,
		"listing_id":            mapped.ListingID,
		"state":                 mapped.State,
		"source_provider":       mapped.SourceProvider,
		"source_account":        mapped.SourceAccount,
		"write_back_capability": mapped.WriteBackCapability,
		"write_back_allowed":    mapped.WriteBackAllowed,
		"write_back_blocker":    mapped.WriteBackBlocker,
	})
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func deterministicEbayBuyerInterestID(prefix, value string) string {
	return prefix + "-" + ebayBuyerInterestDigest(value)[:24]
}

func ebayBuyerInterestDigest(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	encoded := hex.EncodeToString(sum[:])
	if encoded == strings.Repeat("0", len(encoded)) {
		return strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return encoded
}
