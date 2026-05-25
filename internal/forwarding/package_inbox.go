package forwarding

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	ProviderStackry = "stackry"

	SourceAPI     = "api"
	SourceCapture = "capture"
	SourceCSV     = "csv"
	SourceEmail   = "email"
	SourceManual  = "manual"

	StatusReceived    = "received"
	StatusReadyToShip = "ready_to_ship"
	StatusShipped     = "shipped"
	StatusDelivered   = "delivered"
	StatusException   = "exception"
)

type PackageImport struct {
	ProfileID         string
	Provider          string
	Source            string
	ExternalPackageID string
	ShipmentID        string
	TrackingNumber    string
	Status            string
	ReceivedAt        string
	Sender            string
	WarehouseLocation string
	WeightGrams       int
	RawPayload        map[string]any
}

type Package struct {
	ID                string
	ProfileID         string
	Provider          string
	Source            string
	ExternalPackageID string
	ShipmentID        string
	TrackingNumber    string
	Status            string
	ReceivedAt        string
	Sender            string
	WarehouseLocation string
	WeightGrams       int
	ProvenanceKey     string
	RawPayload        map[string]any
}

func NormalizePackageImport(in PackageImport) (Package, error) {
	profileID := strings.TrimSpace(in.ProfileID)
	if profileID == "" {
		return Package{}, fmt.Errorf("profile_id is required")
	}
	provider := normalizeProvider(in.Provider)
	if provider == "" {
		return Package{}, fmt.Errorf("provider is required")
	}
	source := normalizeSource(in.Source)
	if source == "" {
		return Package{}, fmt.Errorf("source is required")
	}
	externalPackageID := strings.TrimSpace(in.ExternalPackageID)
	if externalPackageID == "" {
		return Package{}, fmt.Errorf("external_package_id is required")
	}
	status := normalizeStatus(in.Status)
	if status == "" {
		return Package{}, fmt.Errorf("status is required")
	}
	if in.WeightGrams < 0 {
		return Package{}, fmt.Errorf("weight_grams must be non-negative")
	}

	provenanceKey := provider + ":" + source + ":" + externalPackageID
	return Package{
		ID:                packageID(profileID, provenanceKey),
		ProfileID:         profileID,
		Provider:          provider,
		Source:            source,
		ExternalPackageID: externalPackageID,
		ShipmentID:        strings.TrimSpace(in.ShipmentID),
		TrackingNumber:    strings.TrimSpace(in.TrackingNumber),
		Status:            status,
		ReceivedAt:        strings.TrimSpace(in.ReceivedAt),
		Sender:            strings.TrimSpace(in.Sender),
		WarehouseLocation: strings.TrimSpace(in.WarehouseLocation),
		WeightGrams:       in.WeightGrams,
		ProvenanceKey:     provenanceKey,
		RawPayload:        clonePayload(in.RawPayload),
	}, nil
}

type MemoryInbox struct {
	mu       sync.Mutex
	packages map[string]Package
}

func NewMemoryInbox() *MemoryInbox {
	return &MemoryInbox{packages: map[string]Package{}}
}

func (m *MemoryInbox) Upsert(in PackageImport) (Package, error) {
	pkg, err := NormalizePackageImport(in)
	if err != nil {
		return Package{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.packages[pkg.ProfileID+"|"+pkg.ProvenanceKey] = pkg
	return pkg, nil
}

func (m *MemoryInbox) List(profileID string) []Package {
	profileID = strings.TrimSpace(profileID)
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Package, 0)
	for _, pkg := range m.packages {
		if pkg.ProfileID == profileID {
			out = append(out, pkg)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ProvenanceKey < out[j].ProvenanceKey
	})
	return out
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderStackry:
		return ProviderStackry
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func normalizeSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case SourceAPI, SourceCapture, SourceCSV, SourceEmail, SourceManual:
		return strings.ToLower(strings.TrimSpace(source))
	default:
		return ""
	}
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusReceived, StatusReadyToShip, StatusShipped, StatusDelivered, StatusException:
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return ""
	}
}

func packageID(profileID, provenanceKey string) string {
	sum := sha1.Sum([]byte(profileID + "|" + provenanceKey))
	return "fwdpkg_" + hex.EncodeToString(sum[:])[:16]
}

func clonePayload(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
