package forwarding

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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
	ProfileID         string         `json:"profile_id"`
	Provider          string         `json:"provider"`
	Source            string         `json:"source"`
	ExternalPackageID string         `json:"external_package_id"`
	ShipmentID        string         `json:"shipment_id"`
	TrackingNumber    string         `json:"tracking_number"`
	Status            string         `json:"status"`
	ReceivedAt        string         `json:"received_at"`
	Sender            string         `json:"sender"`
	WarehouseLocation string         `json:"warehouse_location"`
	WeightGrams       int            `json:"weight_grams"`
	RawPayload        map[string]any `json:"raw_payload"`
}

type Package struct {
	ID                string         `json:"id"`
	ProfileID         string         `json:"profile_id"`
	Provider          string         `json:"provider"`
	Source            string         `json:"source"`
	ExternalPackageID string         `json:"external_package_id"`
	ShipmentID        string         `json:"shipment_id"`
	TrackingNumber    string         `json:"tracking_number"`
	Status            string         `json:"status"`
	ReceivedAt        string         `json:"received_at"`
	Sender            string         `json:"sender"`
	WarehouseLocation string         `json:"warehouse_location"`
	WeightGrams       int            `json:"weight_grams"`
	ProvenanceKey     string         `json:"provenance_key"`
	RawPayload        map[string]any `json:"raw_payload,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
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

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) UpsertPackage(ctx context.Context, in PackageImport) (Package, error) {
	pkg, err := NormalizePackageImport(in)
	if err != nil {
		return Package{}, err
	}
	rawJSON, err := encodePayload(pkg.RawPayload)
	if err != nil {
		return Package{}, fmt.Errorf("encode raw payload: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO forwarder_packages(
			id, profile_id, provider, source, external_package_id, shipment_id, tracking_number,
			status, received_at, sender, warehouse_location, weight_grams, provenance_key, raw_payload_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, provider, source, external_package_id) DO UPDATE SET
			shipment_id = excluded.shipment_id,
			tracking_number = excluded.tracking_number,
			status = excluded.status,
			received_at = excluded.received_at,
			sender = excluded.sender,
			warehouse_location = excluded.warehouse_location,
			weight_grams = excluded.weight_grams,
			provenance_key = excluded.provenance_key,
			raw_payload_json = excluded.raw_payload_json,
			updated_at = CURRENT_TIMESTAMP
	`, pkg.ID, pkg.ProfileID, pkg.Provider, pkg.Source, pkg.ExternalPackageID, pkg.ShipmentID, pkg.TrackingNumber, pkg.Status, pkg.ReceivedAt, pkg.Sender, pkg.WarehouseLocation, pkg.WeightGrams, pkg.ProvenanceKey, rawJSON)
	if err != nil {
		return Package{}, fmt.Errorf("upsert forwarder package: %w", err)
	}
	return s.GetPackage(ctx, pkg.ProfileID, pkg.Provider, pkg.Source, pkg.ExternalPackageID)
}

func (s *Service) GetPackage(ctx context.Context, profileID, provider, source, externalPackageID string) (Package, error) {
	var pkg Package
	var rawJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, provider, source, external_package_id, shipment_id, tracking_number,
			status, received_at, sender, warehouse_location, weight_grams, provenance_key, raw_payload_json,
			created_at, updated_at
		FROM forwarder_packages
		WHERE profile_id = ? AND provider = ? AND source = ? AND external_package_id = ?
	`, strings.TrimSpace(profileID), normalizeProvider(provider), normalizeSource(source), strings.TrimSpace(externalPackageID)).Scan(
		&pkg.ID, &pkg.ProfileID, &pkg.Provider, &pkg.Source, &pkg.ExternalPackageID, &pkg.ShipmentID, &pkg.TrackingNumber,
		&pkg.Status, &pkg.ReceivedAt, &pkg.Sender, &pkg.WarehouseLocation, &pkg.WeightGrams, &pkg.ProvenanceKey, &rawJSON,
		&pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Package{}, fmt.Errorf("forwarder package not found")
		}
		return Package{}, err
	}
	payload, err := decodePayload(rawJSON)
	if err != nil {
		return Package{}, err
	}
	pkg.RawPayload = payload
	return pkg, nil
}

func (s *Service) ListPackages(ctx context.Context, profileID, status string) ([]Package, error) {
	profileID = strings.TrimSpace(profileID)
	status = normalizeStatus(status)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, provider, source, external_package_id, shipment_id, tracking_number,
			status, received_at, sender, warehouse_location, weight_grams, provenance_key, raw_payload_json,
			created_at, updated_at
		FROM forwarder_packages
		WHERE (? = '' OR profile_id = ?) AND (? = '' OR status = ?)
		ORDER BY updated_at DESC, provenance_key ASC
	`, profileID, profileID, status, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Package
	for rows.Next() {
		var pkg Package
		var rawJSON string
		if err := rows.Scan(
			&pkg.ID, &pkg.ProfileID, &pkg.Provider, &pkg.Source, &pkg.ExternalPackageID, &pkg.ShipmentID, &pkg.TrackingNumber,
			&pkg.Status, &pkg.ReceivedAt, &pkg.Sender, &pkg.WarehouseLocation, &pkg.WeightGrams, &pkg.ProvenanceKey, &rawJSON,
			&pkg.CreatedAt, &pkg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payload, err := decodePayload(rawJSON)
		if err != nil {
			return nil, err
		}
		pkg.RawPayload = payload
		out = append(out, pkg)
	}
	return out, rows.Err()
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

func encodePayload(in map[string]any) (string, error) {
	if len(in) == 0 {
		return "", nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodePayload(in string) (map[string]any, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		return nil, fmt.Errorf("decode raw payload: %w", err)
	}
	return out, nil
}
