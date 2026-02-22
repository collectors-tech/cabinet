package barcode

import (
	"strings"
	"testing"
)

func TestBuildExternalSearchURL(t *testing.T) {
	t.Parallel()

	u, err := BuildExternalSearchURL("ebay", "US", "12345")
	if err != nil {
		t.Fatalf("BuildExternalSearchURL() error = %v", err)
	}
	if !strings.HasPrefix(u, "https://www.ebay.com/") {
		t.Fatalf("expected ebay US url, got %q", u)
	}
	if !strings.Contains(u, "_nkw=12345") || !strings.Contains(u, "LH_TitleDesc=0") {
		t.Fatalf("unexpected query string in %q", u)
	}
}

func TestBuildExternalSearchURL_Validation(t *testing.T) {
	t.Parallel()

	if _, err := BuildExternalSearchURL("foo", "US", "123"); err == nil {
		t.Fatal("expected unsupported source error")
	}
	if _, err := BuildExternalSearchURL("ebay", "US", ""); err == nil {
		t.Fatal("expected barcode validation error")
	}
}
