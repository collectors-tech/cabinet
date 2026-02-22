package barcode

import (
	"fmt"
	"net/url"
	"strings"
)

var ebayHostByRegion = map[string]string{
	"":   "www.ebay.com",
	"US": "www.ebay.com",
	"UK": "www.ebay.co.uk",
	"AU": "www.ebay.com.au",
	"DE": "www.ebay.de",
}

func BuildExternalSearchURL(source, region, barcode string) (string, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	if source == "" {
		source = "ebay"
	}
	if source != "ebay" {
		return "", fmt.Errorf("unsupported source")
	}

	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return "", fmt.Errorf("barcode is required")
	}

	host := ebayHostByRegion[strings.ToUpper(strings.TrimSpace(region))]
	if host == "" {
		host = ebayHostByRegion["US"]
	}

	q := url.Values{}
	q.Set("_nkw", barcode)
	q.Set("LH_TitleDesc", "0")
	return "https://" + host + "/sch/i.html?" + q.Encode(), nil
}
