package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderProductURLIngestRoutesBonzaProductURL(t *testing.T) {
	t.Parallel()

	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/wp-json/wc/store/v1/products") {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("search"); got != "bonza mug white" {
			t.Fatalf("expected slug-derived Store API search, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":19603,
			"name":"BONZA MUG WHITE",
			"slug":"bonza-mug-white",
			"permalink":"https://bonzaslotcars.com.au/product/bonza-mug-white/",
			"description":"<p>White ceramic Bonza mug.</p>",
			"prices":{"currency_code":"AUD","price":"995"},
			"is_in_stock":true,
			"low_stock_remaining":3,
			"categories":[{"name":"AFX ACCESSORIES HO"},{"name":"MERCHANDISE"}],
			"attributes":[{"name":"Brand","terms":["AFX"]},{"name":"Scale","terms":["1:64"]},{"name":"Type","terms":["Tracks"]}],
			"images":[{"src":"https://bonzaslotcars.com.au/wp-content/uploads/BONZA-MUG.jpg"}]
		}]`))
	}))
	defer bonza.Close()

	a, profileID := newBonzaIngestTestApp(t)
	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profileID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	ingest := doRequest(t, a, http.MethodPost, "/api/providers/product-url/ingest", strings.NewReader(`{"url":"https://bonzaslotcars.com.au/product/bonza-mug-white/"}`), map[string]string{"Content-Type": "application/json"})
	if ingest.Code != http.StatusOK {
		t.Fatalf("ingest status=%d body=%s", ingest.Code, ingest.Body.String())
	}
	var payload struct {
		Mode     string         `json:"mode"`
		Provider string         `json:"provider"`
		Family   string         `json:"family"`
		Route    map[string]any `json:"route"`
		Draft    struct {
			ProviderProductID string            `json:"provider_product_id"`
			Title             string            `json:"title"`
			SourceURL         string            `json:"source_url"`
			Price             float64           `json:"price"`
			Currency          string            `json:"currency"`
			StockState        string            `json:"stock_state"`
			StockCount        int               `json:"stock_count"`
			Description       string            `json:"description"`
			Categories        []string          `json:"categories"`
			Attributes        map[string]string `json:"attributes"`
			ImageURLs         []string          `json:"image_urls"`
		} `json:"draft"`
		Evidence map[string]any `json:"evidence"`
	}
	if err := json.NewDecoder(ingest.Body).Decode(&payload); err != nil {
		t.Fatalf("decode ingest payload: %v", err)
	}
	if payload.Mode != "provider_product_url_ingest" || payload.Provider != "bonzaslotcars" || payload.Family != "woocommerce" {
		t.Fatalf("unexpected route envelope: %+v", payload)
	}
	if payload.Route["slug"] != "bonza-mug-white" || payload.Route["action"] != "ingest_product_url" {
		t.Fatalf("unexpected route detail: %+v", payload.Route)
	}
	if payload.Draft.ProviderProductID != "19603" || payload.Draft.Title != "BONZA MUG WHITE" {
		t.Fatalf("unexpected Bonza draft identity: %+v", payload.Draft)
	}
	if payload.Draft.SourceURL != "https://bonzaslotcars.com.au/product/bonza-mug-white/" {
		t.Fatalf("unexpected source URL: %s", payload.Draft.SourceURL)
	}
	if payload.Draft.Price != 9.95 || payload.Draft.Currency != "AUD" {
		t.Fatalf("unexpected price/currency: %+v", payload.Draft)
	}
	if payload.Draft.StockState != "in_stock" || payload.Draft.StockCount != 3 {
		t.Fatalf("unexpected stock: %+v", payload.Draft)
	}
	if len(payload.Draft.Categories) != 2 || payload.Draft.Attributes["Brand"] != "AFX" || payload.Draft.Attributes["Scale"] != "1:64" || payload.Draft.Attributes["Type"] != "Tracks" {
		t.Fatalf("unexpected category/attribute mapping: %+v", payload.Draft)
	}
	if len(payload.Draft.ImageURLs) != 1 || !strings.Contains(payload.Draft.Description, "White ceramic Bonza mug") {
		t.Fatalf("unexpected image/description mapping: %+v", payload.Draft)
	}
	if payload.Evidence["provider"] != "bonzaslotcars" || payload.Evidence["family"] != "woocommerce" || payload.Evidence["extraction_method"] != "store_api" {
		t.Fatalf("unexpected evidence: %+v", payload.Evidence)
	}
}

func TestProviderProductURLIngestRejectsKnownProviderNonProductURL(t *testing.T) {
	t.Parallel()

	a, _ := newBonzaIngestTestApp(t)
	ingest := doRequest(t, a, http.MethodPost, "/api/providers/product-url/ingest", strings.NewReader(`{"url":"https://bonzaslotcars.com.au/shop/"}`), map[string]string{"Content-Type": "application/json"})
	if ingest.Code != http.StatusBadRequest {
		t.Fatalf("ingest status=%d body=%s", ingest.Code, ingest.Body.String())
	}
	if !strings.Contains(ingest.Body.String(), "supported_provider_unsupported_page") {
		t.Fatalf("expected unsupported-page envelope, got %s", ingest.Body.String())
	}
}

func newBonzaIngestTestApp(t *testing.T) (*App, string) {
	t.Helper()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BonzaIngestProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	return a, profile.ID
}
