package commerce

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
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

type PurchaseOrderLineItem struct {
	ItemID            string  `json:"item_id"`
	Title             string  `json:"title"`
	Quantity          int     `json:"quantity"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
	LifecycleEntryID  string  `json:"lifecycle_entry_id"`
	ExpectedArrivalID string  `json:"expected_arrival_id"`
}

type PurchaseOrder struct {
	OrderID         string                  `json:"order_id"`
	Source          string                  `json:"source"`
	Seller          string                  `json:"seller"`
	Tracking        string                  `json:"tracking"`
	Status          string                  `json:"status"`
	TotalAmount     float64                 `json:"total_amount"`
	Currency        string                  `json:"currency"`
	LineItemCount   int                     `json:"line_item_count"`
	ReceivedCount   int                     `json:"received_count"`
	UnreceivedCount int                     `json:"unreceived_count"`
	LineItems       []PurchaseOrderLineItem `json:"line_items"`
	CreatedAt       string                  `json:"created_at"`
}

type PurchaseOrderList struct {
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
	Orders     []PurchaseOrder `json:"orders"`
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

func normalizePurchaseOrderStatusFilter(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "all":
		return "all"
	case "active", "reviews", "shipped", "received":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "all"
	}
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

func (s *Service) ListPurchaseOrdersByProfile(ctx context.Context, profileID, status, search string, page, pageSize int) (PurchaseOrderList, error) {
	profileID = strings.TrimSpace(profileID)
	filter := normalizePurchaseOrderStatusFilter(status)
	search = strings.ToLower(strings.TrimSpace(search))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT e.id, e.source, e.external_ref, e.quantity, e.amount, e.currency, e.notes, e.created_at,
			i.id, i.title,
			COALESCE(a.id, ''), COALESCE(a.status, ''), COALESCE(a.notes, ''), COALESCE(a.created_at, '')
		FROM commerce_lifecycle_entries e
		JOIN canonical_items i ON i.id = e.item_id AND i.profile_id = e.profile_id
		LEFT JOIN expected_arrivals a ON a.lifecycle_entry_id = e.id AND a.profile_id = e.profile_id
		WHERE e.profile_id = ? AND e.state = 'purchase'
		ORDER BY e.created_at ASC, e.id ASC
	`, profileID)
	if err != nil {
		return PurchaseOrderList{}, err
	}
	defer rows.Close()

	ordersByKey := map[string]*PurchaseOrder{}
	var orderKeys []string
	for rows.Next() {
		var entryID, source, externalRef, currency, notes, createdAt string
		var itemID, title, arrivalID, arrivalStatus, arrivalNotes, arrivalCreatedAt string
		var quantity int
		var amount float64
		if err := rows.Scan(&entryID, &source, &externalRef, &quantity, &amount, &currency, &notes, &createdAt, &itemID, &title, &arrivalID, &arrivalStatus, &arrivalNotes, &arrivalCreatedAt); err != nil {
			return PurchaseOrderList{}, err
		}
		orderID := strings.TrimSpace(externalRef)
		if orderID == "" {
			orderID = entryID
		}
		key := strings.ToLower(strings.TrimSpace(source)) + ":" + orderID
		order := ordersByKey[key]
		if order == nil {
			order = &PurchaseOrder{
				OrderID:   orderID,
				Source:    strings.TrimSpace(source),
				Seller:    valueFromPurchaseNotes(notes, "seller"),
				Tracking:  valueFromPurchaseNotes(notes, "tracking"),
				Currency:  strings.ToUpper(strings.TrimSpace(currency)),
				CreatedAt: createdAt,
			}
			ordersByKey[key] = order
			orderKeys = append(orderKeys, key)
		}
		if order.Seller == "" {
			order.Seller = valueFromPurchaseNotes(notes, "seller")
		}
		if order.Tracking == "" {
			order.Tracking = valueFromPurchaseNotes(notes, "tracking")
		}
		lineStatus := normalizeArrivalStatus(arrivalStatus)
		if lineStatus == "" {
			lineStatus = "expected"
		}
		order.LineItems = append(order.LineItems, PurchaseOrderLineItem{
			ItemID:            itemID,
			Title:             title,
			Quantity:          quantity,
			Amount:            amount,
			Status:            lineStatus,
			LifecycleEntryID:  entryID,
			ExpectedArrivalID: arrivalID,
		})
		order.TotalAmount += amount
		order.LineItemCount++
		if lineStatus == "delivered" || lineStatus == "reconciled" {
			order.ReceivedCount++
		} else {
			order.UnreceivedCount++
		}
		if order.CreatedAt == "" || (arrivalCreatedAt != "" && arrivalCreatedAt < order.CreatedAt) {
			order.CreatedAt = arrivalCreatedAt
		}
		_ = arrivalNotes
	}
	if err := rows.Err(); err != nil {
		return PurchaseOrderList{}, err
	}

	allOrders := make([]PurchaseOrder, 0, len(orderKeys))
	for _, key := range orderKeys {
		order := *ordersByKey[key]
		if order.Currency == "" {
			order.Currency = "AUD"
		}
		order.TotalAmount = math.Round(order.TotalAmount*100) / 100
		if order.UnreceivedCount > 0 {
			order.Status = "active"
		} else {
			order.Status = "received"
		}
		if !purchaseOrderMatchesStatus(order, filter) || !purchaseOrderMatchesSearch(order, search) {
			continue
		}
		allOrders = append(allOrders, order)
	}
	sort.SliceStable(allOrders, func(i, j int) bool {
		if allOrders[i].Status != allOrders[j].Status {
			return allOrders[i].Status == "active"
		}
		if allOrders[i].CreatedAt == allOrders[j].CreatedAt {
			return allOrders[i].OrderID < allOrders[j].OrderID
		}
		return allOrders[i].CreatedAt < allOrders[j].CreatedAt
	})

	total := len(allOrders)
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return PurchaseOrderList{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		Orders:     allOrders[start:end],
	}, nil
}

func valueFromPurchaseNotes(notes, key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, field := range strings.Fields(notes) {
		name, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		if strings.ToLower(strings.TrimSpace(name)) == key {
			return strings.Trim(strings.TrimSpace(value), ",;")
		}
	}
	return ""
}

func purchaseOrderMatchesStatus(order PurchaseOrder, filter string) bool {
	switch filter {
	case "all":
		return true
	case "active", "shipped":
		return order.UnreceivedCount > 0
	case "reviews":
		return order.ReceivedCount > 0
	case "received":
		return order.ReceivedCount > 0 && order.UnreceivedCount == 0
	default:
		return true
	}
}

func purchaseOrderMatchesSearch(order PurchaseOrder, search string) bool {
	if search == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		order.OrderID,
		order.Source,
		order.Seller,
		order.Tracking,
		order.Status,
	}, " "))
	if strings.Contains(haystack, search) {
		return true
	}
	for _, item := range order.LineItems {
		line := strings.ToLower(strings.Join([]string{
			item.ItemID,
			item.Title,
			item.Status,
			item.LifecycleEntryID,
			item.ExpectedArrivalID,
		}, " "))
		if strings.Contains(line, search) {
			return true
		}
	}
	return false
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
