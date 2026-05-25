package forwarding

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
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

type PackageLinkRequest struct {
	ProfileID         string `json:"profile_id"`
	PackageID         string `json:"package_id"`
	ItemID            string `json:"item_id"`
	LifecycleEntryID  string `json:"lifecycle_entry_id"`
	ExpectedArrivalID string `json:"expected_arrival_id"`
	Source            string `json:"source"`
	Notes             string `json:"notes"`
}

type PackageLink struct {
	ID                string `json:"id"`
	ProfileID         string `json:"profile_id"`
	PackageID         string `json:"package_id"`
	ItemID            string `json:"item_id"`
	LifecycleEntryID  string `json:"lifecycle_entry_id,omitempty"`
	ExpectedArrivalID string `json:"expected_arrival_id,omitempty"`
	Source            string `json:"source"`
	Notes             string `json:"notes"`
	CreatedAt         string `json:"created_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

type PackageCSVRowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
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

func ParsePackageCSV(profileID, provider string, r io.Reader) ([]PackageImport, []PackageCSVRowError, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read forwarder package csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("forwarder package csv is empty")
	}
	headers := mapHeaders(rows[0])
	var imports []PackageImport
	var rowErrors []PackageCSVRowError
	for i, row := range rows[1:] {
		rowNumber := i + 2
		if rowIsEmpty(row) {
			continue
		}
		raw := rawCSVPayload(rows[0], row, rowNumber)
		weightGrams, err := parseCSVWeight(valueForCSV(row, headers, "weight_grams"))
		if err != nil {
			rowErrors = append(rowErrors, PackageCSVRowError{Row: rowNumber, Error: err.Error()})
			continue
		}
		candidate := PackageImport{
			ProfileID:         profileID,
			Provider:          provider,
			Source:            SourceCSV,
			ExternalPackageID: firstCSVValue(row, headers, "external_package_id", "package_id", "stackry_package_id"),
			ShipmentID:        firstCSVValue(row, headers, "shipment_id", "shipment", "shipment_number"),
			TrackingNumber:    firstCSVValue(row, headers, "tracking_number", "tracking", "tracking_no"),
			Status:            firstCSVValue(row, headers, "status", "package_status"),
			ReceivedAt:        firstCSVValue(row, headers, "received_at", "received", "date_received"),
			Sender:            firstCSVValue(row, headers, "sender", "merchant", "from"),
			WarehouseLocation: firstCSVValue(row, headers, "warehouse_location", "warehouse", "suite"),
			WeightGrams:       weightGrams,
			RawPayload:        raw,
		}
		if _, err := NormalizePackageImport(candidate); err != nil {
			rowErrors = append(rowErrors, PackageCSVRowError{Row: rowNumber, Error: err.Error()})
			continue
		}
		imports = append(imports, candidate)
	}
	return imports, rowErrors, nil
}

func ParsePackageEmail(profileID, provider, messageID string, r io.Reader) (PackageImport, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return PackageImport{}, fmt.Errorf("read forwarder package email: %w", err)
	}
	body := string(data)
	fields := parseEmailFields(body)
	weightGrams, err := parseEmailWeight(firstEmailValue(fields, "weight_grams", "weight"))
	if err != nil {
		return PackageImport{}, err
	}
	imported := PackageImport{
		ProfileID:         profileID,
		Provider:          provider,
		Source:            SourceEmail,
		ExternalPackageID: firstEmailValue(fields, "external_package_id", "package_id", "stackry_package_id"),
		ShipmentID:        firstEmailValue(fields, "shipment_id", "shipment", "shipment_number"),
		TrackingNumber:    firstEmailValue(fields, "tracking_number", "tracking", "tracking_no"),
		Status:            normalizeEmailStatus(firstEmailValue(fields, "status", "package_status")),
		ReceivedAt:        firstEmailValue(fields, "received_at", "received", "date_received"),
		Sender:            firstEmailValue(fields, "sender", "merchant", "from"),
		WarehouseLocation: firstEmailValue(fields, "warehouse_location", "warehouse", "suite"),
		WeightGrams:       weightGrams,
		RawPayload: map[string]any{
			"source":     SourceEmail,
			"message_id": strings.TrimSpace(messageID),
			"body":       body,
		},
	}
	if _, err := NormalizePackageImport(imported); err != nil {
		return PackageImport{}, err
	}
	return imported, nil
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

func (s *Service) LinkPackage(ctx context.Context, req PackageLinkRequest) (PackageLink, error) {
	req.ProfileID = strings.TrimSpace(req.ProfileID)
	req.PackageID = strings.TrimSpace(req.PackageID)
	req.ItemID = strings.TrimSpace(req.ItemID)
	req.LifecycleEntryID = strings.TrimSpace(req.LifecycleEntryID)
	req.ExpectedArrivalID = strings.TrimSpace(req.ExpectedArrivalID)
	req.Source = strings.TrimSpace(req.Source)
	req.Notes = strings.TrimSpace(req.Notes)
	if req.ProfileID == "" {
		return PackageLink{}, fmt.Errorf("profile_id is required")
	}
	if req.PackageID == "" {
		return PackageLink{}, fmt.Errorf("package_id is required")
	}
	if req.ItemID == "" {
		return PackageLink{}, fmt.Errorf("item_id is required")
	}
	if req.ExpectedArrivalID == "" && req.LifecycleEntryID == "" {
		return PackageLink{}, fmt.Errorf("expected_arrival_id or lifecycle_entry_id is required")
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	if err := s.requirePackageProfile(ctx, req.ProfileID, req.PackageID); err != nil {
		return PackageLink{}, err
	}
	if err := s.requireItemProfile(ctx, req.ProfileID, req.ItemID); err != nil {
		return PackageLink{}, err
	}
	if req.LifecycleEntryID != "" {
		if err := s.requireLifecycleProfile(ctx, req.ProfileID, req.LifecycleEntryID, req.ItemID); err != nil {
			return PackageLink{}, err
		}
	}
	if req.ExpectedArrivalID != "" {
		if err := s.requireArrivalProfile(ctx, req.ProfileID, req.ExpectedArrivalID, req.ItemID, req.LifecycleEntryID); err != nil {
			return PackageLink{}, err
		}
	}
	existing, err := s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
	if err == nil {
		if existing.ItemID != req.ItemID || existing.LifecycleEntryID != req.LifecycleEntryID || existing.ExpectedArrivalID != req.ExpectedArrivalID {
			return PackageLink{}, fmt.Errorf("forwarder package already linked to a different target")
		}
		_, err = s.db.ExecContext(ctx, `
			UPDATE forwarder_package_links
			SET source = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Source, req.Notes, existing.ID)
		if err != nil {
			return PackageLink{}, fmt.Errorf("update forwarder package link: %w", err)
		}
		return s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
	}
	if !strings.Contains(err.Error(), "not found") {
		return PackageLink{}, err
	}
	id := packageLinkID(req.ProfileID, req.PackageID, req.ItemID, req.LifecycleEntryID, req.ExpectedArrivalID)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO forwarder_package_links(id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.ProfileID, req.PackageID, req.ItemID, req.LifecycleEntryID, req.ExpectedArrivalID, req.Source, req.Notes)
	if err != nil {
		return PackageLink{}, fmt.Errorf("create forwarder package link: %w", err)
	}
	return s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
}

func (s *Service) GetPackageLink(ctx context.Context, profileID, packageID string) (PackageLink, error) {
	var out PackageLink
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, notes, created_at, updated_at
		FROM forwarder_package_links
		WHERE profile_id = ? AND package_id = ?
	`, strings.TrimSpace(profileID), strings.TrimSpace(packageID)).Scan(
		&out.ID, &out.ProfileID, &out.PackageID, &out.ItemID, &out.LifecycleEntryID, &out.ExpectedArrivalID, &out.Source, &out.Notes, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return PackageLink{}, fmt.Errorf("forwarder package link not found")
		}
		return PackageLink{}, err
	}
	return out, nil
}

func (s *Service) ListPackageLinks(ctx context.Context, profileID, packageID string) ([]PackageLink, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, notes, created_at, updated_at
		FROM forwarder_package_links
		WHERE profile_id = ? AND (? = '' OR package_id = ?)
		ORDER BY updated_at DESC, id ASC
	`, strings.TrimSpace(profileID), strings.TrimSpace(packageID), strings.TrimSpace(packageID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PackageLink
	for rows.Next() {
		var link PackageLink
		if err := rows.Scan(&link.ID, &link.ProfileID, &link.PackageID, &link.ItemID, &link.LifecycleEntryID, &link.ExpectedArrivalID, &link.Source, &link.Notes, &link.CreatedAt, &link.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, rows.Err()
}

func (s *Service) requirePackageProfile(ctx context.Context, profileID, packageID string) error {
	var found string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM forwarder_packages WHERE id = ? AND profile_id = ?`, packageID, profileID).Scan(&found)
	if err == sql.ErrNoRows {
		return fmt.Errorf("forwarder package not found for profile")
	}
	return err
}

func (s *Service) requireItemProfile(ctx context.Context, profileID, itemID string) error {
	var found string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM canonical_items WHERE id = ? AND profile_id = ?`, itemID, profileID).Scan(&found)
	if err == sql.ErrNoRows {
		return fmt.Errorf("item not found for profile")
	}
	return err
}

func (s *Service) requireLifecycleProfile(ctx context.Context, profileID, lifecycleEntryID, itemID string) error {
	var found string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM commerce_lifecycle_entries WHERE id = ? AND profile_id = ? AND item_id = ?`, lifecycleEntryID, profileID, itemID).Scan(&found)
	if err == sql.ErrNoRows {
		return fmt.Errorf("lifecycle entry not found for profile item")
	}
	return err
}

func (s *Service) requireArrivalProfile(ctx context.Context, profileID, arrivalID, itemID, lifecycleEntryID string) error {
	var found string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM expected_arrivals
		WHERE id = ? AND profile_id = ? AND item_id = ? AND (? = '' OR lifecycle_entry_id = ?)
	`, arrivalID, profileID, itemID, strings.TrimSpace(lifecycleEntryID), strings.TrimSpace(lifecycleEntryID)).Scan(&found)
	if err == sql.ErrNoRows {
		return fmt.Errorf("expected arrival not found for profile item")
	}
	return err
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

func mapHeaders(row []string) map[string]int {
	headers := make(map[string]int, len(row))
	for i, header := range row {
		headers[normalizeCSVHeader(header)] = i
	}
	return headers
}

func normalizeCSVHeader(header string) string {
	header = strings.ToLower(strings.TrimSpace(header))
	header = strings.ReplaceAll(header, " ", "_")
	header = strings.ReplaceAll(header, "-", "_")
	return header
}

func rowIsEmpty(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func rawCSVPayload(headers, row []string, rowNumber int) map[string]any {
	payload := map[string]any{
		"source": "csv",
		"row":    rowNumber,
	}
	for i, header := range headers {
		key := normalizeCSVHeader(header)
		if key == "" || i >= len(row) {
			continue
		}
		payload[key] = strings.TrimSpace(row[i])
	}
	return payload
}

func firstCSVValue(row []string, headers map[string]int, names ...string) string {
	for _, name := range names {
		if value := valueForCSV(row, headers, name); value != "" {
			return value
		}
	}
	return ""
}

func valueForCSV(row []string, headers map[string]int, name string) string {
	idx, ok := headers[normalizeCSVHeader(name)]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseCSVWeight(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	weight, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("weight_grams must be an integer")
	}
	if weight < 0 {
		return 0, fmt.Errorf("weight_grams must be non-negative")
	}
	return weight, nil
}

func parseEmailFields(body string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = normalizeCSVHeader(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		fields[key] = value
	}
	return fields
}

func firstEmailValue(fields map[string]string, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(fields[normalizeCSVHeader(name)]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeEmailStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	return status
}

func parseEmailWeight(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parts := strings.Fields(value)
	if len(parts) > 0 {
		value = parts[0]
	}
	weight, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("weight_grams must be an integer")
	}
	if weight < 0 {
		return 0, fmt.Errorf("weight_grams must be non-negative")
	}
	return weight, nil
}

func packageID(profileID, provenanceKey string) string {
	sum := sha1.Sum([]byte(profileID + "|" + provenanceKey))
	return "fwdpkg_" + hex.EncodeToString(sum[:])[:16]
}

func packageLinkID(profileID, packageID, itemID, lifecycleEntryID, expectedArrivalID string) string {
	sum := sha1.Sum([]byte(profileID + "|" + packageID + "|" + itemID + "|" + lifecycleEntryID + "|" + expectedArrivalID))
	return "fwdlink_" + hex.EncodeToString(sum[:])[:16]
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
