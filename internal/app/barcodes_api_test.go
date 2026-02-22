package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestBarcodeExternalSearchEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/barcodes/12345/external-search?source=ebay&region=US", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	urlValue, _ := payload["url"].(string)
	if !strings.Contains(urlValue, "ebay.com") || !strings.Contains(urlValue, "_nkw=12345") {
		t.Fatalf("unexpected search url: %q", urlValue)
	}
}
