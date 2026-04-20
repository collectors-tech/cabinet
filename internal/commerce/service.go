package commerce

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var validLifecycleStates = map[string]struct{}{
	"wishlist":  {},
	"watchlist": {},
	"cart":      {},
	"offer":     {},
	"purchase":  {},
}

var validArrivalStatuses = map[string]struct{}{
	"expected":   {},
	"delivered":  {},
	"reconciled": {},
	"cancelled":  {},
}

type LifecycleEntry struct {
	ID                string  `json:"id"`
	ProfileID         string  `json:"profile_id,omitempty"`
	ItemID            string  `json:"item_id"`
	State             string  `json:"state"`
	Source            string  `json:"source"`
	ExternalRef       string  `json:"external_ref"`
	Quantity          int     `json:"quantity"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Notes             string  `json:"notes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ExpectedArrivalID string  `json:"expected_arrival_id,omitempty"`
}

type ExpectedArrival struct {
	ID                   string  `json:"id"`
	ProfileID            string  `json:"profile_id,omitempty"`
	ItemID               string  `json:"item_id"`
	LifecycleEntryID     string  `json:"lifecycle_entry_id"`
	Source               string  `json:"source"`
	ExternalRef          string  `json:"external_ref"`
	Quantity             int     `json:"quantity"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Status               string  `json:"status"`
	ExpectedOn           string  `json:"expected_on"`
	DeliveredOn          string  `json:"delivered_on"`
	ReconciledInstanceID string  `json:"reconciled_instance_id"`
	Notes                string  `json:"notes"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func normalizeLifecycleState(state string) string {
	return strings.ToLower(strings.TrimSpace(state))
}

func normalizeArrivalStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func (s *Service) CreateLifecycleForProfile(ctx context.Context, profileID string, in LifecycleEntry) (LifecycleEntry, *ExpectedArrival, error) {
	profileID = strings.TrimSpace(profileID)
	in.ItemID = strings.TrimSpace(in.ItemID)
	in.State = normalizeLifecycleState(in.State)
	in.Source = strings.TrimSpace(in.Source)
	in.ExternalRef = strings.TrimSpace(in.ExternalRef)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.Notes = strings.TrimSpace(in.Notes)
	if in.ItemID == "" {
		return LifecycleEntry{}, nil, fmt.Errorf("item_id is required")
	}
	if _, ok := validLifecycleStates[in.State]; !ok {
		return LifecycleEntry{}, nil, fmt.Errorf("invalid state")
	}
	if in.Quantity <= 0 {
		in.Quantity = 1
	}
	if in.Currency == "" {
		in.Currency = "AUD"
	}
	in.ID = uuid.NewString()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO commerce_lifecycle_entries(
			id, profile_id, item_id, state, source, external_ref, quantity, amount, currency, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, profileID, in.ItemID, in.State, in.Source, in.ExternalRef, in.Quantity, in.Amount, in.Currency, in.Notes)
	if err != nil {
		return LifecycleEntry{}, nil, fmt.Errorf("create lifecycle entry: %w", err)
	}
	created, err := s.GetLifecycleByIDForProfile(ctx, profileID, in.ID)
	if err != nil {
		return LifecycleEntry{}, nil, err
	}
	var arrival *ExpectedArrival
	if created.State == "purchase" {
		made, err := s.CreateArrivalForProfile(ctx, profileID, ExpectedArrival{
			ItemID:           created.ItemID,
			LifecycleEntryID: created.ID,
			Source:           created.Source,
			ExternalRef:      created.ExternalRef,
			Quantity:         created.Quantity,
			Amount:           created.Amount,
			Currency:         created.Currency,
			Status:           "expected",
			Notes:            created.Notes,
		})
		if err != nil {
			return LifecycleEntry{}, nil, fmt.Errorf("create expected arrival for purchase: %w", err)
		}
		created.ExpectedArrivalID = made.ID
		arrival = &made
	}
	return created, arrival, nil
}

func (s *Service) GetLifecycleByIDForProfile(ctx context.Context, profileID, id string) (LifecycleEntry, error) {
	var out LifecycleEntry
	var expectedArrivalID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT e.id, e.profile_id, e.item_id, e.state, e.source, e.external_ref, e.quantity, e.amount, e.currency, e.notes, e.created_at, e.updated_at,
			(SELECT a.id FROM expected_arrivals a WHERE a.lifecycle_entry_id = e.id LIMIT 1)
		FROM commerce_lifecycle_entries e
		WHERE e.id = ? AND (? = '' OR e.profile_id = ?)
	`, strings.TrimSpace(id), strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(
		&out.ID, &out.ProfileID, &out.ItemID, &out.State, &out.Source, &out.ExternalRef, &out.Quantity, &out.Amount, &out.Currency, &out.Notes, &out.CreatedAt, &out.UpdatedAt, &expectedArrivalID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return LifecycleEntry{}, fmt.Errorf("lifecycle entry not found")
		}
		return LifecycleEntry{}, err
	}
	if expectedArrivalID.Valid {
		out.ExpectedArrivalID = expectedArrivalID.String
	}
	return out, nil
}

func (s *Service) ListLifecycleByProfile(ctx context.Context, profileID, itemID string) ([]LifecycleEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM commerce_lifecycle_entries
		WHERE (? = '' OR profile_id = ?) AND (? = '' OR item_id = ?)
		ORDER BY created_at ASC, id ASC
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), strings.TrimSpace(itemID), strings.TrimSpace(itemID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LifecycleEntry
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		entry, err := s.GetLifecycleByIDForProfile(ctx, profileID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (s *Service) CreateArrivalForProfile(ctx context.Context, profileID string, in ExpectedArrival) (ExpectedArrival, error) {
	profileID = strings.TrimSpace(profileID)
	in.ItemID = strings.TrimSpace(in.ItemID)
	in.LifecycleEntryID = strings.TrimSpace(in.LifecycleEntryID)
	in.Source = strings.TrimSpace(in.Source)
	in.ExternalRef = strings.TrimSpace(in.ExternalRef)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.Status = normalizeArrivalStatus(in.Status)
	in.ExpectedOn = strings.TrimSpace(in.ExpectedOn)
	in.DeliveredOn = strings.TrimSpace(in.DeliveredOn)
	in.ReconciledInstanceID = strings.TrimSpace(in.ReconciledInstanceID)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.ItemID == "" {
		return ExpectedArrival{}, fmt.Errorf("item_id is required")
	}
	if in.Quantity <= 0 {
		in.Quantity = 1
	}
	if in.Currency == "" {
		in.Currency = "AUD"
	}
	if in.Status == "" {
		in.Status = "expected"
	}
	if _, ok := validArrivalStatuses[in.Status]; !ok {
		return ExpectedArrival{}, fmt.Errorf("invalid status")
	}
	in.ID = uuid.NewString()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO expected_arrivals(
			id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, expected_on, delivered_on, reconciled_instance_id, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, profileID, in.ItemID, in.LifecycleEntryID, in.Source, in.ExternalRef, in.Quantity, in.Amount, in.Currency, in.Status, in.ExpectedOn, in.DeliveredOn, in.ReconciledInstanceID, in.Notes)
	if err != nil {
		return ExpectedArrival{}, fmt.Errorf("create expected arrival: %w", err)
	}
	return s.GetArrivalByIDForProfile(ctx, profileID, in.ID)
}

func (s *Service) GetArrivalByIDForProfile(ctx context.Context, profileID, id string) (ExpectedArrival, error) {
	var out ExpectedArrival
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, expected_on, delivered_on, reconciled_instance_id, notes, created_at, updated_at
		FROM expected_arrivals
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(id), strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(
		&out.ID, &out.ProfileID, &out.ItemID, &out.LifecycleEntryID, &out.Source, &out.ExternalRef, &out.Quantity, &out.Amount, &out.Currency, &out.Status, &out.ExpectedOn, &out.DeliveredOn, &out.ReconciledInstanceID, &out.Notes, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ExpectedArrival{}, fmt.Errorf("expected arrival not found")
		}
		return ExpectedArrival{}, err
	}
	return out, nil
}

func (s *Service) ListArrivalsByProfile(ctx context.Context, profileID, itemID, status string) ([]ExpectedArrival, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM expected_arrivals
		WHERE (? = '' OR profile_id = ?) AND (? = '' OR item_id = ?) AND (? = '' OR status = ?)
		ORDER BY created_at ASC, id ASC
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), strings.TrimSpace(itemID), strings.TrimSpace(itemID), normalizeArrivalStatus(status), normalizeArrivalStatus(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ExpectedArrival
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		entry, err := s.GetArrivalByIDForProfile(ctx, profileID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (s *Service) UpdateArrivalForProfile(ctx context.Context, profileID string, in ExpectedArrival) error {
	profileID = strings.TrimSpace(profileID)
	in.ID = strings.TrimSpace(in.ID)
	in.Status = normalizeArrivalStatus(in.Status)
	in.ExpectedOn = strings.TrimSpace(in.ExpectedOn)
	in.DeliveredOn = strings.TrimSpace(in.DeliveredOn)
	in.ReconciledInstanceID = strings.TrimSpace(in.ReconciledInstanceID)
	in.Notes = strings.TrimSpace(in.Notes)
	if in.ID == "" {
		return fmt.Errorf("id is required")
	}
	if _, ok := validArrivalStatuses[in.Status]; !ok {
		return fmt.Errorf("invalid status")
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE expected_arrivals
		SET status = ?, expected_on = ?, delivered_on = ?, reconciled_instance_id = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, in.Status, in.ExpectedOn, in.DeliveredOn, in.ReconciledInstanceID, in.Notes, in.ID, profileID, profileID)
	if err != nil {
		return fmt.Errorf("update expected arrival: %w", err)
	}
	return nil
}
