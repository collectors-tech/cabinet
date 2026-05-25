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
	ProfileID         string   `json:"profile_id"`
	PackageID         string   `json:"package_id"`
	ItemID            string   `json:"item_id"`
	LifecycleEntryID  string   `json:"lifecycle_entry_id"`
	ExpectedArrivalID string   `json:"expected_arrival_id"`
	Source            string   `json:"source"`
	Decision          string   `json:"decision"`
	Notes             string   `json:"notes"`
	Override          bool     `json:"override"`
	Actor             string   `json:"actor"`
	AuditTrail        []string `json:"audit_trail"`
}

type PackageLink struct {
	ID                string   `json:"id"`
	ProfileID         string   `json:"profile_id"`
	PackageID         string   `json:"package_id"`
	ItemID            string   `json:"item_id"`
	LifecycleEntryID  string   `json:"lifecycle_entry_id,omitempty"`
	ExpectedArrivalID string   `json:"expected_arrival_id,omitempty"`
	Source            string   `json:"source"`
	Decision          string   `json:"decision"`
	Notes             string   `json:"notes"`
	AuditTrail        []string `json:"audit_trail"`
	CreatedAt         string   `json:"created_at,omitempty"`
	UpdatedAt         string   `json:"updated_at,omitempty"`
}

type PackageUnlinkRequest struct {
	ProfileID  string   `json:"profile_id"`
	PackageID  string   `json:"package_id"`
	Source     string   `json:"source"`
	Notes      string   `json:"notes"`
	Actor      string   `json:"actor"`
	AuditTrail []string `json:"audit_trail"`
}

type PackageLinkEvent struct {
	ID                        string   `json:"id"`
	ProfileID                 string   `json:"profile_id"`
	PackageID                 string   `json:"package_id"`
	LinkID                    string   `json:"link_id,omitempty"`
	Action                    string   `json:"action"`
	ItemID                    string   `json:"item_id,omitempty"`
	LifecycleEntryID          string   `json:"lifecycle_entry_id,omitempty"`
	ExpectedArrivalID         string   `json:"expected_arrival_id,omitempty"`
	PreviousItemID            string   `json:"previous_item_id,omitempty"`
	PreviousLifecycleEntryID  string   `json:"previous_lifecycle_entry_id,omitempty"`
	PreviousExpectedArrivalID string   `json:"previous_expected_arrival_id,omitempty"`
	Source                    string   `json:"source"`
	Notes                     string   `json:"notes"`
	AuditTrail                []string `json:"audit_trail"`
	CreatedAt                 string   `json:"created_at,omitempty"`
}

type PackageMatchSignal struct {
	Name     string  `json:"name"`
	Matched  bool    `json:"matched"`
	Weight   float64 `json:"weight"`
	Evidence string  `json:"evidence"`
}

type PackageMatchSuggestion struct {
	PackageID         string               `json:"package_id"`
	ItemID            string               `json:"item_id"`
	LifecycleEntryID  string               `json:"lifecycle_entry_id,omitempty"`
	ExpectedArrivalID string               `json:"expected_arrival_id,omitempty"`
	Confidence        float64              `json:"confidence"`
	ConfidenceLabel   string               `json:"confidence_label"`
	Explanation       []string             `json:"explanation"`
	Signals           []PackageMatchSignal `json:"signals"`
	AuditTrail        []string             `json:"audit_trail"`
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
	req.Decision = normalizeLinkDecision(req.Decision)
	req.Notes = strings.TrimSpace(req.Notes)
	req.Actor = strings.TrimSpace(req.Actor)
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
	auditTrail := normalizeAuditTrail(req.AuditTrail)
	if len(auditTrail) == 0 {
		auditTrail = append(auditTrail, linkAuditLine(req.Decision, req.Source, req.Actor, req.Notes))
	}
	auditJSON, err := encodeStringList(auditTrail)
	if err != nil {
		return PackageLink{}, fmt.Errorf("encode forwarder package link audit trail: %w", err)
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
			if !req.Override {
				return PackageLink{}, fmt.Errorf("forwarder package already linked to a different target")
			}
			req.Decision = "override"
			auditTrail = append([]string{fmt.Sprintf("override accepted previous target item=%s lifecycle=%s expected_arrival=%s", existing.ItemID, existing.LifecycleEntryID, existing.ExpectedArrivalID)}, auditTrail...)
			auditJSON, err = encodeStringList(auditTrail)
			if err != nil {
				return PackageLink{}, fmt.Errorf("encode forwarder package override audit trail: %w", err)
			}
			_, err = s.db.ExecContext(ctx, `
				UPDATE forwarder_package_links
				SET item_id = ?, lifecycle_entry_id = ?, expected_arrival_id = ?, source = ?, decision = ?, notes = ?, audit_trail_json = ?, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, req.ItemID, req.LifecycleEntryID, req.ExpectedArrivalID, req.Source, req.Decision, req.Notes, auditJSON, existing.ID)
			if err != nil {
				return PackageLink{}, fmt.Errorf("override forwarder package link: %w", err)
			}
			link, err := s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
			if err != nil {
				return PackageLink{}, err
			}
			if _, err := s.recordPackageLinkEvent(ctx, "override", link, existing, req.Source, req.Notes, auditTrail); err != nil {
				return PackageLink{}, err
			}
			return link, nil
		}
		_, err = s.db.ExecContext(ctx, `
			UPDATE forwarder_package_links
			SET source = ?, decision = ?, notes = ?, audit_trail_json = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Source, req.Decision, req.Notes, auditJSON, existing.ID)
		if err != nil {
			return PackageLink{}, fmt.Errorf("update forwarder package link: %w", err)
		}
		link, err := s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
		if err != nil {
			return PackageLink{}, err
		}
		if _, err := s.recordPackageLinkEvent(ctx, req.Decision, link, PackageLink{}, req.Source, req.Notes, auditTrail); err != nil {
			return PackageLink{}, err
		}
		return link, nil
	}
	if !strings.Contains(err.Error(), "not found") {
		return PackageLink{}, err
	}
	id := packageLinkID(req.ProfileID, req.PackageID, req.ItemID, req.LifecycleEntryID, req.ExpectedArrivalID)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO forwarder_package_links(id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, decision, notes, audit_trail_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.ProfileID, req.PackageID, req.ItemID, req.LifecycleEntryID, req.ExpectedArrivalID, req.Source, req.Decision, req.Notes, auditJSON)
	if err != nil {
		return PackageLink{}, fmt.Errorf("create forwarder package link: %w", err)
	}
	link, err := s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
	if err != nil {
		return PackageLink{}, err
	}
	if _, err := s.recordPackageLinkEvent(ctx, req.Decision, link, PackageLink{}, req.Source, req.Notes, auditTrail); err != nil {
		return PackageLink{}, err
	}
	return link, nil
}

func (s *Service) GetPackageLink(ctx context.Context, profileID, packageID string) (PackageLink, error) {
	var out PackageLink
	var auditJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, decision, notes, audit_trail_json, created_at, updated_at
		FROM forwarder_package_links
		WHERE profile_id = ? AND package_id = ?
	`, strings.TrimSpace(profileID), strings.TrimSpace(packageID)).Scan(
		&out.ID, &out.ProfileID, &out.PackageID, &out.ItemID, &out.LifecycleEntryID, &out.ExpectedArrivalID, &out.Source, &out.Decision, &out.Notes, &auditJSON, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return PackageLink{}, fmt.Errorf("forwarder package link not found")
		}
		return PackageLink{}, err
	}
	out.AuditTrail, err = decodeStringList(auditJSON)
	if err != nil {
		return PackageLink{}, err
	}
	return out, nil
}

func (s *Service) ListPackageLinks(ctx context.Context, profileID, packageID string) ([]PackageLink, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, package_id, item_id, lifecycle_entry_id, expected_arrival_id, source, decision, notes, audit_trail_json, created_at, updated_at
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
		var auditJSON string
		if err := rows.Scan(&link.ID, &link.ProfileID, &link.PackageID, &link.ItemID, &link.LifecycleEntryID, &link.ExpectedArrivalID, &link.Source, &link.Decision, &link.Notes, &auditJSON, &link.CreatedAt, &link.UpdatedAt); err != nil {
			return nil, err
		}
		link.AuditTrail, err = decodeStringList(auditJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, rows.Err()
}

func (s *Service) UnlinkPackage(ctx context.Context, req PackageUnlinkRequest) (PackageLinkEvent, error) {
	req.ProfileID = strings.TrimSpace(req.ProfileID)
	req.PackageID = strings.TrimSpace(req.PackageID)
	req.Source = strings.TrimSpace(req.Source)
	req.Notes = strings.TrimSpace(req.Notes)
	req.Actor = strings.TrimSpace(req.Actor)
	if req.ProfileID == "" {
		return PackageLinkEvent{}, fmt.Errorf("profile_id is required")
	}
	if req.PackageID == "" {
		return PackageLinkEvent{}, fmt.Errorf("package_id is required")
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	if err := s.requirePackageProfile(ctx, req.ProfileID, req.PackageID); err != nil {
		return PackageLinkEvent{}, err
	}
	existing, err := s.GetPackageLink(ctx, req.ProfileID, req.PackageID)
	if err != nil {
		return PackageLinkEvent{}, err
	}
	auditTrail := normalizeAuditTrail(req.AuditTrail)
	if len(auditTrail) == 0 {
		auditTrail = []string{linkAuditLine("unlinked", req.Source, req.Actor, req.Notes)}
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM forwarder_package_links WHERE id = ?`, existing.ID); err != nil {
		return PackageLinkEvent{}, fmt.Errorf("unlink forwarder package: %w", err)
	}
	return s.recordPackageLinkEvent(ctx, "unlinked", PackageLink{ProfileID: req.ProfileID, PackageID: req.PackageID}, existing, req.Source, req.Notes, auditTrail)
}

func (s *Service) ListPackageLinkEvents(ctx context.Context, profileID, packageID string) ([]PackageLinkEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, package_id, link_id, action, item_id, lifecycle_entry_id, expected_arrival_id,
			previous_item_id, previous_lifecycle_entry_id, previous_expected_arrival_id, source, notes, audit_trail_json, created_at
		FROM forwarder_package_link_events
		WHERE profile_id = ? AND (? = '' OR package_id = ?)
		ORDER BY created_at DESC, id DESC
	`, strings.TrimSpace(profileID), strings.TrimSpace(packageID), strings.TrimSpace(packageID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PackageLinkEvent
	for rows.Next() {
		var event PackageLinkEvent
		var auditJSON string
		if err := rows.Scan(&event.ID, &event.ProfileID, &event.PackageID, &event.LinkID, &event.Action, &event.ItemID, &event.LifecycleEntryID, &event.ExpectedArrivalID, &event.PreviousItemID, &event.PreviousLifecycleEntryID, &event.PreviousExpectedArrivalID, &event.Source, &event.Notes, &auditJSON, &event.CreatedAt); err != nil {
			return nil, err
		}
		event.AuditTrail, err = decodeStringList(auditJSON)
		if err != nil {
			return nil, err
		}
		out = append(out, event)
	}
	return out, rows.Err()
}

func (s *Service) recordPackageLinkEvent(ctx context.Context, action string, link PackageLink, previous PackageLink, source, notes string, auditTrail []string) (PackageLinkEvent, error) {
	auditJSON, err := encodeStringList(auditTrail)
	if err != nil {
		return PackageLinkEvent{}, fmt.Errorf("encode forwarder package link event audit trail: %w", err)
	}
	profileID := strings.TrimSpace(link.ProfileID)
	if profileID == "" {
		profileID = previous.ProfileID
	}
	packageID := strings.TrimSpace(link.PackageID)
	if packageID == "" {
		packageID = previous.PackageID
	}
	event := PackageLinkEvent{
		ID:                        packageLinkEventID(profileID, packageID, action, link.ItemID, link.LifecycleEntryID, link.ExpectedArrivalID, previous.ItemID, previous.LifecycleEntryID, previous.ExpectedArrivalID, notes),
		ProfileID:                 profileID,
		PackageID:                 packageID,
		LinkID:                    link.ID,
		Action:                    action,
		ItemID:                    link.ItemID,
		LifecycleEntryID:          link.LifecycleEntryID,
		ExpectedArrivalID:         link.ExpectedArrivalID,
		PreviousItemID:            previous.ItemID,
		PreviousLifecycleEntryID:  previous.LifecycleEntryID,
		PreviousExpectedArrivalID: previous.ExpectedArrivalID,
		Source:                    source,
		Notes:                     notes,
		AuditTrail:                append([]string(nil), auditTrail...),
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO forwarder_package_link_events(
			id, profile_id, package_id, link_id, action, item_id, lifecycle_entry_id, expected_arrival_id,
			previous_item_id, previous_lifecycle_entry_id, previous_expected_arrival_id, source, notes, audit_trail_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.ProfileID, event.PackageID, event.LinkID, event.Action, event.ItemID, event.LifecycleEntryID, event.ExpectedArrivalID, event.PreviousItemID, event.PreviousLifecycleEntryID, event.PreviousExpectedArrivalID, event.Source, event.Notes, auditJSON)
	if err != nil {
		return PackageLinkEvent{}, fmt.Errorf("record forwarder package link event: %w", err)
	}
	return event, nil
}

func (s *Service) SuggestPackageMatches(ctx context.Context, profileID, packageID string) ([]PackageMatchSuggestion, error) {
	profileID = strings.TrimSpace(profileID)
	packageID = strings.TrimSpace(packageID)
	if profileID == "" {
		return nil, fmt.Errorf("profile_id is required")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			p.id, p.external_package_id, p.shipment_id, p.tracking_number, p.received_at, p.sender, p.weight_grams, p.raw_payload_json,
			i.id, i.title, i.brand, i.part_number,
			COALESCE(l.id, ''), COALESCE(l.source, ''), COALESCE(l.external_ref, ''), COALESCE(l.quantity, 0), COALESCE(l.notes, ''),
			a.id, a.external_ref, a.quantity, a.expected_on, a.notes
		FROM forwarder_packages p
		JOIN expected_arrivals a ON a.profile_id = p.profile_id
		JOIN canonical_items i ON i.id = a.item_id AND i.profile_id = p.profile_id
		LEFT JOIN commerce_lifecycle_entries l ON l.id = a.lifecycle_entry_id AND l.profile_id = p.profile_id AND l.item_id = i.id
		LEFT JOIN forwarder_package_links existing ON existing.package_id = p.id
		WHERE p.profile_id = ? AND (? = '' OR p.id = ?) AND existing.id IS NULL
		ORDER BY p.updated_at DESC, a.updated_at DESC, i.id ASC
	`, profileID, packageID, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PackageMatchSuggestion
	for rows.Next() {
		var pkg matchPackageRow
		var item matchItemRow
		var lifecycle matchLifecycleRow
		var arrival matchArrivalRow
		var rawJSON string
		if err := rows.Scan(
			&pkg.ID, &pkg.ExternalPackageID, &pkg.ShipmentID, &pkg.TrackingNumber, &pkg.ReceivedAt, &pkg.Sender, &pkg.WeightGrams, &rawJSON,
			&item.ID, &item.Title, &item.Brand, &item.PartNumber,
			&lifecycle.ID, &lifecycle.Source, &lifecycle.ExternalRef, &lifecycle.Quantity, &lifecycle.Notes,
			&arrival.ID, &arrival.ExternalRef, &arrival.Quantity, &arrival.ExpectedOn, &arrival.Notes,
		); err != nil {
			return nil, err
		}
		raw, err := decodePayload(rawJSON)
		if err != nil {
			return nil, err
		}
		pkg.RawPayload = raw
		suggestion := scorePackageMatch(pkg, item, lifecycle, arrival)
		if suggestion.Confidence > 0 {
			out = append(out, suggestion)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence == out[j].Confidence {
			if out[i].PackageID == out[j].PackageID {
				return out[i].ExpectedArrivalID < out[j].ExpectedArrivalID
			}
			return out[i].PackageID < out[j].PackageID
		}
		return out[i].Confidence > out[j].Confidence
	})
	return out, nil
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

type matchPackageRow struct {
	ID                string
	ExternalPackageID string
	ShipmentID        string
	TrackingNumber    string
	ReceivedAt        string
	Sender            string
	WeightGrams       int
	RawPayload        map[string]any
}

type matchItemRow struct {
	ID         string
	Title      string
	Brand      string
	PartNumber string
}

type matchLifecycleRow struct {
	ID          string
	Source      string
	ExternalRef string
	Quantity    int
	Notes       string
}

type matchArrivalRow struct {
	ID          string
	ExternalRef string
	Quantity    int
	ExpectedOn  string
	Notes       string
}

func scorePackageMatch(pkg matchPackageRow, item matchItemRow, lifecycle matchLifecycleRow, arrival matchArrivalRow) PackageMatchSuggestion {
	suggestion := PackageMatchSuggestion{
		PackageID:         pkg.ID,
		ItemID:            item.ID,
		LifecycleEntryID:  lifecycle.ID,
		ExpectedArrivalID: arrival.ID,
		AuditTrail: []string{
			"match suggestion generated without mutating package reconciliation links",
			"signals scored deterministically from package, item, lifecycle, and expected-arrival fields",
		},
	}
	addSignal := func(name string, weight float64, matched bool, evidence string) {
		signal := PackageMatchSignal{Name: name, Matched: matched, Weight: weight, Evidence: evidence}
		suggestion.Signals = append(suggestion.Signals, signal)
		if matched {
			suggestion.Confidence += weight
			suggestion.Explanation = append(suggestion.Explanation, evidence)
		}
	}

	trackingHaystack := strings.Join([]string{arrival.ExternalRef, arrival.Notes, lifecycle.ExternalRef, lifecycle.Notes}, " ")
	addSignal("tracking", 0.35, containsToken(trackingHaystack, pkg.TrackingNumber), "package tracking number appears in purchase or expected-arrival evidence")

	externalHaystack := strings.Join([]string{arrival.ExternalRef, arrival.Notes, lifecycle.ExternalRef, lifecycle.Notes}, " ")
	addSignal("package_or_order_reference", 0.25, containsToken(externalHaystack, pkg.ExternalPackageID) || containsToken(externalHaystack, pkg.ShipmentID), "package or shipment id appears in purchase/order evidence")

	titleNeedle := strings.Join([]string{rawString(pkg.RawPayload, "title"), rawString(pkg.RawPayload, "item_title"), rawString(pkg.RawPayload, "description")}, " ")
	titleHaystack := strings.Join([]string{item.Title, item.Brand, item.PartNumber, lifecycle.Notes, arrival.Notes}, " ")
	addSignal("title_or_part", 0.20, sharedMeaningfulToken(titleNeedle, titleHaystack), "package item text overlaps the purchase item title, brand, or part number")

	sellerNeedle := strings.Join([]string{pkg.Sender, rawString(pkg.RawPayload, "seller"), rawString(pkg.RawPayload, "merchant")}, " ")
	sellerHaystack := strings.Join([]string{lifecycle.Source, lifecycle.Notes, arrival.Notes}, " ")
	addSignal("seller_or_source", 0.10, sharedMeaningfulToken(sellerNeedle, sellerHaystack), "package sender/source overlaps purchase source evidence")

	quantity := rawInt(pkg.RawPayload, "quantity")
	addSignal("quantity", 0.05, quantity > 0 && (quantity == lifecycle.Quantity || quantity == arrival.Quantity), "package quantity matches lifecycle or expected-arrival quantity")

	addSignal("date", 0.05, sameDate(pkg.ReceivedAt, arrival.ExpectedOn), "package received date matches expected-arrival date")

	suggestion.Confidence = roundConfidence(suggestion.Confidence)
	suggestion.ConfidenceLabel = confidenceLabel(suggestion.Confidence)
	if len(suggestion.Explanation) == 0 {
		suggestion.Explanation = []string{"no positive matching signals found"}
	}
	return suggestion
}

func containsToken(haystack, needle string) bool {
	needle = normalizeMatchText(needle)
	if len(needle) < 3 {
		return false
	}
	return strings.Contains(normalizeMatchText(haystack), needle)
}

func sharedMeaningfulToken(left, right string) bool {
	right = normalizeMatchText(right)
	for _, token := range strings.Fields(normalizeMatchText(left)) {
		if len(token) < 3 {
			continue
		}
		if strings.Contains(right, token) {
			return true
		}
	}
	return false
}

func normalizeMatchText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", "\\", " ", ".", " ", ",", " ", ":", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func rawString(payload map[string]any, key string) string {
	if len(payload) == 0 {
		return ""
	}
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	switch value := value.(type) {
	case string:
		return value
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprint(value)
	}
}

func rawInt(payload map[string]any, key string) int {
	if len(payload) == 0 {
		return 0
	}
	switch value := payload[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

func sameDate(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if len(left) >= 10 {
		left = left[:10]
	}
	if len(right) >= 10 {
		right = right[:10]
	}
	return left != "" && left == right
}

func roundConfidence(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func confidenceLabel(confidence float64) string {
	switch {
	case confidence >= 0.80:
		return "high"
	case confidence >= 0.50:
		return "medium"
	default:
		return "low"
	}
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

func packageLinkEventID(profileID, packageID, action, itemID, lifecycleEntryID, expectedArrivalID, previousItemID, previousLifecycleEntryID, previousExpectedArrivalID, notes string) string {
	sum := sha1.Sum([]byte(strings.Join([]string{profileID, packageID, action, itemID, lifecycleEntryID, expectedArrivalID, previousItemID, previousLifecycleEntryID, previousExpectedArrivalID, notes}, "|")))
	return "fwdlinkevent_" + hex.EncodeToString(sum[:])[:16]
}

func normalizeLinkDecision(decision string) string {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "", "confirm", "confirmed":
		return "confirmed"
	case "override", "overridden":
		return "override"
	case "unlinked", "unlink":
		return "unlinked"
	default:
		return strings.ToLower(strings.TrimSpace(decision))
	}
}

func normalizeAuditTrail(in []string) []string {
	out := make([]string, 0, len(in))
	for _, line := range in {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func linkAuditLine(decision, source, actor, notes string) string {
	parts := []string{"decision=" + normalizeLinkDecision(decision)}
	if source != "" {
		parts = append(parts, "source="+source)
	}
	if actor != "" {
		parts = append(parts, "actor="+actor)
	}
	if notes != "" {
		parts = append(parts, "notes="+notes)
	}
	return strings.Join(parts, "; ")
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

func encodeStringList(in []string) (string, error) {
	if len(in) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeStringList(in string) ([]string, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		return nil, fmt.Errorf("decode audit trail: %w", err)
	}
	return out, nil
}
