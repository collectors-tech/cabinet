package app

import (
	"encoding/base64"
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
		Evidence   map[string]any `json:"evidence"`
		Duplicates []struct {
			ItemID  string   `json:"item_id"`
			Reasons []string `json:"reasons"`
		} `json:"duplicates"`
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
	if len(payload.Duplicates) != 0 {
		t.Fatalf("expected no duplicate candidates, got %+v", payload.Duplicates)
	}
}

func TestProviderProductURLIngestReturnsDuplicateCandidates(t *testing.T) {
	t.Parallel()

	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":19603,
			"name":"BONZA MUG WHITE",
			"slug":"bonza-mug-white",
			"permalink":"https://bonzaslotcars.com.au/product/bonza-mug-white/",
			"prices":{"currency_code":"AUD","price":"995"},
			"is_in_stock":true,
			"attributes":[{"name":"Brand","terms":[{"name":"AFX","slug":"afx"}]}]
		}]`))
	}))
	defer bonza.Close()

	a, profileID := newBonzaIngestTestApp(t)
	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profileID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}
	createExisting := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{
		"part_number":"BONZA-19603",
		"title":"Existing Bonza Mug",
		"brand":"AFX",
		"category":"MERCHANDISE",
		"notes":"provider_product_id=19603",
		"source_urls":["https://bonzaslotcars.com.au/product/bonza-mug-white/"]
	}`), map[string]string{"Content-Type": "application/json"})
	if createExisting.Code != http.StatusCreated {
		t.Fatalf("create existing status=%d body=%s", createExisting.Code, createExisting.Body.String())
	}
	var existing struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createExisting.Body).Decode(&existing); err != nil {
		t.Fatalf("decode existing item: %v", err)
	}

	ingest := doRequest(t, a, http.MethodPost, "/api/providers/product-url/ingest", strings.NewReader(`{"url":"https://bonzaslotcars.com.au/product/bonza-mug-white/"}`), map[string]string{"Content-Type": "application/json"})
	if ingest.Code != http.StatusOK {
		t.Fatalf("ingest status=%d body=%s", ingest.Code, ingest.Body.String())
	}
	var payload struct {
		Duplicates []struct {
			ItemID     string   `json:"item_id"`
			Title      string   `json:"title"`
			SourceURLs []string `json:"source_urls"`
			Reasons    []string `json:"reasons"`
		} `json:"duplicates"`
	}
	if err := json.NewDecoder(ingest.Body).Decode(&payload); err != nil {
		t.Fatalf("decode ingest payload: %v", err)
	}
	if len(payload.Duplicates) != 1 {
		t.Fatalf("expected one duplicate candidate, got %+v", payload.Duplicates)
	}
	duplicate := payload.Duplicates[0]
	if duplicate.ItemID != existing.ID || duplicate.Title != "Existing Bonza Mug" {
		t.Fatalf("unexpected duplicate identity: %+v", duplicate)
	}
	if !containsString(duplicate.Reasons, "source_url") || !containsString(duplicate.Reasons, "provider_product_id") {
		t.Fatalf("expected source URL and provider product id reasons, got %+v", duplicate.Reasons)
	}
}

func TestProviderProductURLIngestSolvesBonzaSucuriChallenge(t *testing.T) {
	t.Parallel()

	const challengeCookie = "sucuri_cloudproxy_uuid_d7cbc918a=6b8cc2469d5d14ef58e8d55a94069178"
	challengeScript := `r=String.fromCharCode(54)+'b'+"8"+"c"+String.fromCharCode(99)+"2"+"4"+"6"+"9"+"d"+"5"+"d"+"1"+"4"+"e"+"f"+"5"+"8"+"e"+"8"+"d"+"5"+"5"+"a"+"9"+"4"+"0"+"6"+"9"+"1"+"7"+"8"+'';document.cookie='s'+'u'+'c'+'u'+'r'+'i'+'_'+'c'+'l'+'o'+'u'+'d'+'p'+'r'+'o'+'x'+'y'+'_'+'u'+'u'+'i'+'d'+'_'+'d'+'7'+'c'+'b'+'c'+'9'+'1'+'8'+'a'+"=" + r + ';path=/;max-age=86400'; location.reload();`
	challengeBody := "Javascript is required.<script>S='" + base64.StdEncoding.EncodeToString([]byte(challengeScript)) + "';sucuri_cloudproxy_js='';</script>"
	cookie, err := bonzaSucuriChallengeCookie(challengeBody)
	if err != nil {
		t.Fatalf("challenge cookie failed: %v", err)
	}
	if cookie != challengeCookie {
		t.Fatalf("challenge cookie=%q want %q", cookie, challengeCookie)
	}
	requests := 0
	var cookies []string
	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		cookies = append(cookies, r.Header.Get("Cookie"))
		if r.Header.Get("Cookie") != challengeCookie {
			w.WriteHeader(http.StatusTemporaryRedirect)
			_, _ = w.Write([]byte(challengeBody))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":19603,
			"name":"BONZA MUG WHITE",
			"slug":"bonza-mug-white",
			"permalink":"https://bonzaslotcars.com.au/product/bonza-mug-white/",
			"prices":{"currency_code":"AUD","price":"995"},
			"is_in_stock":true,
			"attributes":[{"name":"Brand","terms":[{"name":"AFX","slug":"afx"}]}]
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
		t.Fatalf("ingest status=%d requests=%d cookies=%q body=%s", ingest.Code, requests, cookies, ingest.Body.String())
	}
	if requests != 2 {
		t.Fatalf("expected challenge request plus cookie retry, got %d requests", requests)
	}
	if !strings.Contains(ingest.Body.String(), `"provider_product_id":"19603"`) {
		t.Fatalf("expected Bonza product payload after challenge retry, got %s", ingest.Body.String())
	}
}

func TestProviderProductURLIngestSolvesChainedBonzaSucuriChallenges(t *testing.T) {
	t.Parallel()

	firstCookie := "sucuri_cloudproxy_uuid_first=first-cookie-value"
	secondCookie := "sucuri_cloudproxy_uuid_second=second-cookie-value"
	firstChallenge := bonzaSucuriChallengeBody(t, "sucuri_cloudproxy_uuid_first", "first-cookie-value")
	secondChallenge := bonzaSucuriChallengeBody(t, "sucuri_cloudproxy_uuid_second", "second-cookie-value")
	requests := 0
	var cookies []string
	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		cookie := r.Header.Get("Cookie")
		cookies = append(cookies, cookie)
		switch {
		case !strings.Contains(cookie, firstCookie):
			w.WriteHeader(http.StatusTemporaryRedirect)
			_, _ = w.Write([]byte(firstChallenge))
			return
		case !strings.Contains(cookie, secondCookie):
			w.WriteHeader(http.StatusTemporaryRedirect)
			_, _ = w.Write([]byte(secondChallenge))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"id":19603,
			"name":"BONZA MUG WHITE",
			"slug":"bonza-mug-white",
			"permalink":"https://bonzaslotcars.com.au/product/bonza-mug-white/",
			"prices":{"currency_code":"AUD","price":"995"},
			"is_in_stock":true,
			"attributes":[{"name":"Brand","terms":[{"name":"AFX","slug":"afx"}]}]
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
		t.Fatalf("ingest status=%d requests=%d cookies=%q body=%s", ingest.Code, requests, cookies, ingest.Body.String())
	}
	if requests != 3 {
		t.Fatalf("expected first challenge, second challenge, and cookie retry, got %d requests", requests)
	}
	if !strings.Contains(cookies[2], firstCookie) || !strings.Contains(cookies[2], secondCookie) {
		t.Fatalf("expected final retry to send both challenge cookies, got %q", cookies)
	}
	if !strings.Contains(ingest.Body.String(), `"provider_product_id":"19603"`) {
		t.Fatalf("expected Bonza product payload after chained challenge retries, got %s", ingest.Body.String())
	}
	if !strings.Contains(ingest.Body.String(), `"Brand":"AFX"`) {
		t.Fatalf("expected object-shaped Bonza attribute terms to normalize, got %s", ingest.Body.String())
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
	if !strings.Contains(ingest.Body.String(), `"fallback_state":"manual_url_capture"`) || !strings.Contains(ingest.Body.String(), `"static_extraction_attempted":false`) {
		t.Fatalf("expected manual URL capture guidance for known non-product page, got %s", ingest.Body.String())
	}
}

func TestProviderProductURLIngestReturnsHeadlessRequiredGuidanceAfterStaticFailure(t *testing.T) {
	t.Parallel()

	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/wp-json/wc/store/v1/products") {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`temporarily blocked by storefront challenge`))
	}))
	defer bonza.Close()

	a, profileID := newBonzaIngestTestApp(t)
	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profileID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	ingest := doRequest(t, a, http.MethodPost, "/api/providers/product-url/ingest", strings.NewReader(`{"url":"https://bonzaslotcars.com.au/product/bonza-mug-white/"}`), map[string]string{"Content-Type": "application/json"})
	if ingest.Code != http.StatusBadRequest {
		t.Fatalf("ingest status=%d body=%s", ingest.Code, ingest.Body.String())
	}
	body := ingest.Body.String()
	for _, want := range []string{
		`"error":"failed_to_ingest_bonza_product_url"`,
		`"fallback_state":"headless_required"`,
		`"static_extraction_attempted":true`,
		`"next_action":"capture_url_for_manual_review"`,
		`"guidance":"Static product extraction was attempted first but the storefront did not return usable public product data. Keep the URL as a manual review item; do not run headless browsing unless this provider is explicitly opted in."`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %s in failure guidance, got %s", want, body)
		}
	}
}

func TestProviderRegistryPublishesManualURLFallbackStates(t *testing.T) {
	t.Parallel()

	a, _ := newBonzaIngestTestApp(t)
	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	var bonza map[string]any
	for _, provider := range payload.Providers {
		if fmt.Sprintf("%v", provider["provider_id"]) == "au-webshop-bonzaslotcars-com-au" {
			bonza = provider
			break
		}
	}
	if bonza == nil {
		t.Fatal("expected Bonza provider registry entry")
	}
	if got := fmt.Sprintf("%v", bonza["fallback_state"]); got != "manual_url_capture" {
		t.Fatalf("Bonza fallback_state got %q want manual_url_capture: %+v", got, bonza)
	}
	if got := fmt.Sprintf("%v", bonza["headless_state"]); got != "opt_in_required" {
		t.Fatalf("Bonza headless_state got %q want opt_in_required: %+v", got, bonza)
	}
	if got := fmt.Sprintf("%v", bonza["manual_capture_action"]); got != "provider_product_url_ingest" {
		t.Fatalf("Bonza manual_capture_action got %q want provider_product_url_ingest: %+v", got, bonza)
	}
	capabilities, ok := bonza["capabilities"].(map[string]any)
	if !ok || capabilities["manual_url_capture"] != true || capabilities["headless_default"] != false {
		t.Fatalf("expected manual URL capture capability and no default headless crawling, got %+v", capabilities)
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

func bonzaSucuriChallengeBody(t *testing.T, cookieName, cookieValue string) string {
	t.Helper()

	nameExpr := sucuriConcatExpression(cookieName)
	valueExpr := sucuriConcatExpression(cookieValue)
	challengeScript := "r=" + valueExpr + ";document.cookie=" + nameExpr + "+\"=\"+r+';path=/;max-age=86400'; location.reload();"
	return "Javascript is required.<script>S='" + base64.StdEncoding.EncodeToString([]byte(challengeScript)) + "';sucuri_cloudproxy_js='';</script>"
}

func sucuriConcatExpression(value string) string {
	parts := make([]string, 0, len(value))
	for _, char := range value {
		parts = append(parts, fmt.Sprintf("%q", string(char)))
	}
	return strings.Join(parts, "+")
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
