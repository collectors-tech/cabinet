package ebaypurchasecapture

import (
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PurchaseCard is a normalized, passive capture of one eBay purchase-history item card.
// It intentionally separates broad listing metadata from the actual purchased item
// variant/aspect metadata shown inside the purchase card.
type PurchaseCard struct {
	ListingID         string
	VariationID       string
	TransactionID     string
	OrderID           string
	ListingTitle      string
	PurchasedIdentity string
	Aspects           map[string]string
	Quantity          int
	ItemPrice         string
	ImageURL          string
	ItemURL           string
	SellerUsername    string
	SellerProfileURL  string
	OrderTotal        string
	Currency          string
	Shipping          string
	Tax               string
	ImportCharges     string
	DestinationMarker string
	OrderStatus       string
	OrderDetailURL    string
	ItemStatus        string
	TrackingStatus    string
	NoteCapability    NoteCapability
	Actions           []Action
}

type NoteCapability struct {
	TextareaID string
	MaxLength  int
	Actions    []Action
}

type Action struct {
	Name     string
	Label    string
	URL      string
	Enabled  bool
	Metadata map[string]string
}

var (
	inputListingIDRe = regexp.MustCompile(`(?i)input-listing-id=["']([^"']+)["']`)
	hrefRe           = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	imgSrcRe         = regexp.MustCompile(`(?is)<img\b[^>]*\bsrc=["']([^"']+)["'][^>]*>`)
	titleRe          = regexp.MustCompile(`(?is)<h3\b[^>]*class=["'][^"']*title-heading[^"']*["'][^>]*>\s*<a\b[^>]*>(.*?)</a>\s*</h3>`)
	aspectBlockRe    = regexp.MustCompile(`(?is)<div\b[^>]*class=["'][^"']*container-item-col__info-item-info-aspectValuesList[^"']*["'][^>]*>(.*?)</div>\s*</div>`)
	aspectDivRe      = regexp.MustCompile(`(?is)<div\b[^>]*>(.*?)</div>`)
	aspectPairRe     = regexp.MustCompile(`(?is)<div\b[^>]*>\s*([^<:]+):\s*([^<]+)</div>`)
	priceRe          = regexp.MustCompile(`(?is)container-item-col__info-item-info-additionalPrice[^>]*>\s*<div>\s*<span>(.*?)</span>`)
	sellerLinkRe     = regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*/usr/[^"']+)["'][^>]*>\s*<span[^>]*>(.*?)</span>`)
	textareaRe       = regexp.MustCompile(`(?is)<textarea\b([^>]*)>`)
	idAttrRe         = regexp.MustCompile(`(?i)\bid=["']([^"']+)["']`)
	maxLengthRe      = regexp.MustCompile(`(?i)\bmaxlength=["']?(\d+)`)
	actionTagRe      = regexp.MustCompile(`(?is)<(a|button)\b([^>]*(?:data-action|data-action-name)[^>]*)>(.*?)</(?:a|button)>`)
	attrActionNameRe = regexp.MustCompile(`(?i)\bdata-action-name=["']([^"']+)["']`)
	attrActionRe     = regexp.MustCompile(`(?i)\bdata-action=["']([^"']+)["']`)
	attrHrefRe       = regexp.MustCompile(`(?i)\bhref=["']([^"']+)["']`)
	attrDataHrefRe   = regexp.MustCompile(`(?i)\bdata-href=["']([^"']*)["']`)
	ariaDisabledRe   = regexp.MustCompile(`(?i)\baria-disabled=["']true["']`)
	dataActionURLRe  = regexp.MustCompile(`(?i)&quot;URL&quot;\s*:\s*&quot;([^&]+(?:&[^q][^u][^o][^t][^;][^&]*)*)`) // best-effort only
	jsonURLRe        = regexp.MustCompile(`(?i)"URL"\s*:\s*"([^"]+)"`)
)

func ParsePurchaseCardHTML(cardHTML string) PurchaseCard {
	card := PurchaseCard{Aspects: map[string]string{}}
	card.ListingID = firstMatch(inputListingIDRe, cardHTML)
	card.ItemURL = firstItemURL(cardHTML)
	if card.ListingID == "" {
		card.ListingID = listingIDFromURL(card.ItemURL)
	}
	card.VariationID = queryValue(card.ItemURL, "var")
	card.ImageURL = html.UnescapeString(firstMatch(imgSrcRe, cardHTML))
	card.ListingTitle = cleanText(firstMatch(titleRe, cardHTML))
	card.Aspects = parseAspects(cardHTML)
	card.PurchasedIdentity = firstNonEmpty(card.Aspects["Card"], card.Aspects["Item"], card.ListingTitle)
	card.Quantity = parseQuantity(card.Aspects["Quantity"])
	if card.Quantity == 0 {
		card.Quantity = 1
	}
	card.ItemPrice = cleanText(firstMatch(priceRe, cardHTML))
	card.SellerProfileURL, card.SellerUsername = parseSeller(cardHTML)
	card.NoteCapability = parseNoteCapability(cardHTML)
	card.Actions = parseActions(cardHTML)
	for _, action := range card.Actions {
		if card.TransactionID == "" {
			card.TransactionID = firstNonEmpty(action.Metadata["transaction_id"], action.Metadata["transactionId"], action.Metadata["transId"])
		}
		if card.OrderID == "" {
			card.OrderID = action.Metadata["orderId"]
		}
		if card.VariationID == "" {
			card.VariationID = action.Metadata["var"]
		}
	}
	return card
}

func parseAspects(s string) map[string]string {
	out := map[string]string{}
	for _, m := range aspectPairRe.FindAllStringSubmatch(s, -1) {
		key := strings.TrimSpace(cleanText(m[1]))
		value := strings.TrimSpace(cleanText(m[2]))
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}

func parseSeller(s string) (string, string) {
	m := sellerLinkRe.FindStringSubmatch(s)
	if len(m) < 3 {
		return "", ""
	}
	profile := html.UnescapeString(m[1])
	username := cleanText(m[2])
	username = strings.TrimSuffix(username, " username")
	return profile, strings.TrimSpace(username)
}

func parseNoteCapability(s string) NoteCapability {
	cap := NoteCapability{}
	if m := textareaRe.FindStringSubmatch(s); len(m) > 1 {
		attrs := m[1]
		cap.TextareaID = html.UnescapeString(firstMatch(idAttrRe, attrs))
		if n, err := strconv.Atoi(firstMatch(maxLengthRe, attrs)); err == nil {
			cap.MaxLength = n
		}
	}
	for _, action := range parseActions(s) {
		switch action.Name {
		case "EDIT_NOTE", "SAVE_NOTE", "DELETE_NOTE", "CANCEL_EDIT_NOTE":
			cap.Actions = append(cap.Actions, action)
		}
	}
	return cap
}

func parseActions(s string) []Action {
	matches := actionTagRe.FindAllStringSubmatch(s, -1)
	actions := make([]Action, 0, len(matches))
	for _, m := range matches {
		attrs := m[2]
		label := cleanText(m[3])
		name := firstNonEmpty(firstMatch(attrActionNameRe, attrs), actionNameFromDataAction(firstMatch(attrActionRe, attrs)))
		if name == "" {
			continue
		}
		actionURL := firstNonEmpty(firstMatch(attrHrefRe, attrs), firstMatch(attrDataHrefRe, attrs), urlFromDataAction(attrs))
		actionURL = html.UnescapeString(actionURL)
		action := Action{
			Name:     html.UnescapeString(name),
			Label:    label,
			URL:      actionURL,
			Enabled:  !ariaDisabledRe.MatchString(attrs),
			Metadata: parseURLMetadata(actionURL),
		}
		actions = append(actions, action)
	}
	return actions
}

func actionNameFromDataAction(v string) string {
	v = html.UnescapeString(v)
	if strings.HasPrefix(strings.TrimSpace(v), "{") {
		if m := regexp.MustCompile(`(?i)"name"\s*:\s*"([^"]+)"`).FindStringSubmatch(v); len(m) > 1 {
			return m[1]
		}
		return ""
	}
	return v
}

func urlFromDataAction(attrs string) string {
	v := firstMatch(attrActionRe, attrs)
	decoded := html.UnescapeString(v)
	if m := jsonURLRe.FindStringSubmatch(decoded); len(m) > 1 {
		return m[1]
	}
	if m := dataActionURLRe.FindStringSubmatch(v); len(m) > 1 {
		return html.UnescapeString(m[1])
	}
	return ""
}

func parseURLMetadata(raw string) map[string]string {
	out := map[string]string{}
	if raw == "" {
		return out
	}
	u, err := url.Parse(html.UnescapeString(raw))
	if err != nil {
		return out
	}
	for key, values := range u.Query() {
		if len(values) > 0 {
			out[key] = values[0]
		}
	}
	if id := listingIDFromURL(raw); id != "" {
		out["listing_id"] = id
	}
	return out
}

func firstItemURL(s string) string {
	for _, m := range hrefRe.FindAllStringSubmatch(s, -1) {
		if strings.Contains(m[1], "/itm/") {
			return html.UnescapeString(m[1])
		}
	}
	return ""
}

func listingIDFromURL(raw string) string {
	u, err := url.Parse(html.UnescapeString(raw))
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, part := range parts {
		if part == "itm" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func queryValue(raw, key string) string {
	u, err := url.Parse(html.UnescapeString(raw))
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

func parseQuantity(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func firstMatch(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func cleanText(s string) string {
	s = regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return strings.Join(strings.Fields(s), " ")
}
