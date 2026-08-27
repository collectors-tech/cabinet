package wishlist

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Entry struct {
	ID                string  `json:"id"`
	ItemID            string  `json:"item_id"`
	TargetPrice       float64 `json:"target_price"`
	Currency          string  `json:"currency"`
	Priority          string  `json:"priority"`
	Notes             string  `json:"notes"`
	HighlightHit      bool    `json:"highlight_hit"`
	BelowTargetNow    bool    `json:"below_target_now"`
	Owned             bool    `json:"owned"`
	Delivered         bool    `json:"delivered"`
	PricePaid         float64 `json:"price_paid"`
	PurchaseURL       string  `json:"purchase_url"`
	PurchaseDate      string  `json:"purchase_date"`
	PurchaseCondition string  `json:"purchase_condition"`
	Quantity          int     `json:"quantity"`
	NeededQuantity    int     `json:"needed_quantity"`
	Deleted           bool    `json:"deleted"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type Hit struct {
	CandidateID string  `json:"candidate_id"`
	ItemID      string  `json:"item_id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	LastSeen    string  `json:"last_seen"`
}

type queryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Service struct {
	db queryExecutor
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) WithTx(tx *sql.Tx) *Service {
	return &Service{db: tx}
}

func (s *Service) Create(ctx context.Context, in Entry) (Entry, error) {
	return s.CreateForProfile(ctx, "", in)
}

func (s *Service) CreateForProfile(ctx context.Context, profileID string, in Entry) (Entry, error) {
	if strings.TrimSpace(in.ItemID) == "" {
		return Entry{}, fmt.Errorf("item_id is required")
	}
	trimmedProfileID := strings.TrimSpace(profileID)
	in.Priority = normalizeWishlistPriority(in.Priority)
	var err error
	in.Currency, err = normalizeWishlistCurrency(in.Currency)
	if err != nil {
		return Entry{}, err
	}
	in.ID = uuid.NewString()
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	belowTargetNow := 0
	if in.BelowTargetNow {
		belowTargetNow = 1
	}
	owned := 0
	if in.Owned {
		owned = 1
	}
	delivered := 0
	if in.Delivered {
		in.Owned = true
		owned = 1
		delivered = 1
	}
	in.Quantity = normalizeWishlistCount(in.Quantity)
	in.NeededQuantity = normalizeWishlistCount(in.NeededQuantity)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, currency, priority, notes, highlight_hit, below_target_now, owned, delivered, price_paid, purchase_url, purchase_date, purchase_condition, quantity, needed_quantity)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, trimmedProfileID, in.ItemID, in.TargetPrice, in.Currency, in.Priority, in.Notes, highlight, belowTargetNow, owned, delivered, in.PricePaid, strings.TrimSpace(in.PurchaseURL), strings.TrimSpace(in.PurchaseDate), strings.TrimSpace(in.PurchaseCondition), in.Quantity, in.NeededQuantity)
	if err != nil {
		return Entry{}, fmt.Errorf("create wishlist entry: %w", err)
	}
	if err := s.syncItemWishlistState(ctx, trimmedProfileID, in.ItemID, "wishlist", in.Priority); err != nil {
		return Entry{}, err
	}
	if err := s.syncPurchaseDeliveryState(ctx, trimmedProfileID, in.ID, in); err != nil {
		return Entry{}, err
	}
	return s.GetByIDForProfile(ctx, trimmedProfileID, in.ID)
}

func (s *Service) Update(ctx context.Context, in Entry) error {
	return s.UpdateForProfile(ctx, "", in)
}

func (s *Service) UpdateForProfile(ctx context.Context, profileID string, in Entry) error {
	if strings.TrimSpace(in.ID) == "" {
		return fmt.Errorf("id is required")
	}
	trimmedProfileID := strings.TrimSpace(profileID)
	in.Priority = normalizeWishlistPriority(in.Priority)
	itemID, err := s.itemIDForEntry(ctx, trimmedProfileID, in.ID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Currency) == "" {
		if err := s.db.QueryRowContext(ctx, `SELECT currency FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)`, in.ID, trimmedProfileID, trimmedProfileID).Scan(&in.Currency); err != nil {
			return err
		}
	}
	in.Currency, err = normalizeWishlistCurrency(in.Currency)
	if err != nil {
		return err
	}
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	belowTargetNow := 0
	if in.BelowTargetNow {
		belowTargetNow = 1
	}
	owned := 0
	if in.Owned {
		owned = 1
	}
	delivered := 0
	if in.Delivered {
		in.Owned = true
		owned = 1
		delivered = 1
	}
	in.Quantity = normalizeWishlistCount(in.Quantity)
	in.NeededQuantity = normalizeWishlistCount(in.NeededQuantity)
	_, err = s.db.ExecContext(ctx, `
		UPDATE wishlist_entries
		SET target_price = ?, currency = ?, priority = ?, notes = ?, highlight_hit = ?, below_target_now = ?, owned = ?, delivered = ?, price_paid = ?, purchase_url = ?, purchase_date = ?, purchase_condition = ?, quantity = ?, needed_quantity = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, in.TargetPrice, in.Currency, in.Priority, in.Notes, highlight, belowTargetNow, owned, delivered, in.PricePaid, strings.TrimSpace(in.PurchaseURL), strings.TrimSpace(in.PurchaseDate), strings.TrimSpace(in.PurchaseCondition), in.Quantity, in.NeededQuantity, in.ID, trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	if err := s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "wishlist", in.Priority); err != nil {
		return err
	}
	return s.syncPurchaseDeliveryState(ctx, trimmedProfileID, in.ID, in)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.DeleteForProfile(ctx, "", id)
}

func (s *Service) ConvertToOwnedForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, err := s.itemIDForEntry(ctx, trimmedProfileID, id)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID); err != nil {
		return err
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "active", "")
}

func (s *Service) DeleteForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, _ := s.itemIDForEntry(ctx, trimmedProfileID, id)
	result, err := s.db.ExecContext(ctx, `
		UPDATE wishlist_entries
		SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("wishlist entry not found")
	}
	if strings.TrimSpace(itemID) == "" {
		return nil
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "active", "")
}

func (s *Service) PermanentDeleteForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, _ := s.itemIDForEntry(ctx, trimmedProfileID, id)
	result, err := s.db.ExecContext(ctx, `DELETE FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("wishlist entry not found")
	}
	if strings.TrimSpace(itemID) == "" {
		return nil
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "active", "")
}

func (s *Service) RestoreForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, err := s.itemIDForEntry(ctx, trimmedProfileID, id)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE wishlist_entries
		SET deleted = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("wishlist entry not found")
	}
	var priority string
	if err := s.db.QueryRowContext(ctx, `
		SELECT priority
		FROM wishlist_entries
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID).Scan(&priority); err != nil {
		return err
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "wishlist", priority)
}

func (s *Service) GetByID(ctx context.Context, id string) (Entry, error) {
	return s.GetByIDForProfile(ctx, "", id)
}

func (s *Service) GetByIDForProfile(ctx context.Context, profileID, id string) (Entry, error) {
	var e Entry
	var highlight int
	var belowTargetNow int
	var owned int
	var delivered int
	var deleted int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, item_id, target_price, currency, priority, notes, highlight_hit, below_target_now, owned, delivered, price_paid, purchase_url, purchase_date, purchase_condition, quantity, needed_quantity, deleted, created_at, updated_at
		FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)
	`, id, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&e.ID, &e.ItemID, &e.TargetPrice, &e.Currency, &e.Priority, &e.Notes, &highlight, &belowTargetNow, &owned, &delivered, &e.PricePaid, &e.PurchaseURL, &e.PurchaseDate, &e.PurchaseCondition, &e.Quantity, &e.NeededQuantity, &deleted, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Entry{}, fmt.Errorf("wishlist entry not found")
		}
		return Entry{}, err
	}
	e.HighlightHit = highlight == 1
	e.BelowTargetNow = belowTargetNow == 1 || s.isBelowTarget(ctx, e.ItemID, e.TargetPrice)
	e.Owned = owned == 1
	e.Delivered = delivered == 1
	e.Deleted = deleted == 1
	return e, nil
}

func (s *Service) List(ctx context.Context) ([]Entry, error) {
	return s.ListByProfile(ctx, "")
}

func (s *Service) ListByProfile(ctx context.Context, profileID string) ([]Entry, error) {
	return s.ListByProfileDeleted(ctx, profileID, false)
}

func (s *Service) ListByProfileDeleted(ctx context.Context, profileID string, deleted bool) ([]Entry, error) {
	deletedValue := 0
	if deleted {
		deletedValue = 1
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM wishlist_entries
		WHERE (? = '' OR profile_id = ?) AND deleted = ?
		ORDER BY created_at ASC
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), deletedValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Entry, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		e, err := s.GetByIDForProfile(ctx, strings.TrimSpace(profileID), id)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Service) Hits(ctx context.Context, itemID string) ([]Hit, error) {
	return s.HitsByProfile(ctx, "", itemID)
}

func (s *Service) HitsByProfile(ctx context.Context, profileID, itemID string) ([]Hit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, w.item_id, c.title, c.price, c.url, c.last_seen
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_candidates c ON LOWER(c.title) LIKE '%' || LOWER(i.part_number) || '%'
		WHERE w.highlight_hit = 1 AND (? = '' OR w.profile_id = ?) AND (? = '' OR w.item_id = ?)
		ORDER BY c.last_seen DESC
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), itemID, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.CandidateID, &h.ItemID, &h.Title, &h.Price, &h.URL, &h.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Service) isBelowTarget(ctx context.Context, itemID string, target float64) bool {
	if target <= 0 {
		return false
	}
	var part string
	if err := s.db.QueryRowContext(ctx, `SELECT part_number FROM canonical_items WHERE id = ?`, itemID).Scan(&part); err != nil {
		return false
	}
	var minPrice float64
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MIN(price), 0)
		FROM scanner_candidates
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
	`, part).Scan(&minPrice)
	if err != nil {
		return false
	}
	return minPrice > 0 && minPrice <= target
}

func normalizeWishlistPriority(raw string) string {
	priority := strings.ToLower(strings.TrimSpace(raw))
	if priority == "" || priority == "normal" {
		return "medium"
	}
	return priority
}

func normalizeWishlistCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeWishlistCurrency(raw string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(raw))
	if currency == "" {
		return "USD", nil
	}
	if len(currency) != 3 {
		return "", fmt.Errorf("currency must be a three-letter ISO 4217 code")
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return "", fmt.Errorf("currency must be a three-letter ISO 4217 code")
		}
	}
	return currency, nil
}

func (s *Service) itemIDForEntry(ctx context.Context, profileID, entryID string) (string, error) {
	var itemID string
	err := s.db.QueryRowContext(ctx, `
		SELECT item_id
		FROM wishlist_entries
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, entryID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("wishlist entry not found")
		}
		return "", err
	}
	return strings.TrimSpace(itemID), nil
}

func (s *Service) syncItemWishlistState(ctx context.Context, profileID, itemID, status, priority string) error {
	if strings.TrimSpace(itemID) == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET status = ?,
			priority = CASE WHEN ? = '' THEN priority ELSE ? END,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = 'wishlist.service'
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(status), strings.TrimSpace(priority), strings.TrimSpace(priority), strings.TrimSpace(itemID), strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return fmt.Errorf("sync wishlist item status: %w", err)
	}
	return nil
}

func (s *Service) syncPurchaseDeliveryState(ctx context.Context, profileID, entryID string, in Entry) error {
	if !in.Owned && !in.Delivered {
		return nil
	}
	lifecycleID, err := s.ensurePurchaseLifecycle(ctx, profileID, entryID, in)
	if err != nil {
		return err
	}
	instanceID, err := s.ensurePurchasedInventoryInstance(ctx, in, !in.Delivered)
	if err != nil {
		return err
	}
	if !in.Delivered {
		return nil
	}
	if err := s.markPurchaseArrivalDelivered(ctx, profileID, lifecycleID, instanceID, in); err != nil {
		return err
	}
	if err := s.syncItemWishlistState(ctx, profileID, in.ItemID, "active", in.Priority); err != nil {
		return err
	}
	return nil
}

func (s *Service) ensurePurchaseLifecycle(ctx context.Context, profileID, entryID string, in Entry) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM commerce_lifecycle_entries
		WHERE (? = '' OR profile_id = ?) AND item_id = ? AND state = 'purchase' AND source = 'wishlist' AND external_ref = ?
		LIMIT 1
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), strings.TrimSpace(in.ItemID), strings.TrimSpace(entryID)).Scan(&id)
	if err == nil {
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE commerce_lifecycle_entries
			SET quantity = ?, amount = ?, currency = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, purchaseQuantity(in), in.PricePaid, in.Currency, purchaseNotes(in), id)
		if updateErr != nil {
			return "", fmt.Errorf("update wishlist purchase lifecycle: %w", updateErr)
		}
		if err := s.ensureExpectedArrival(ctx, profileID, id, in); err != nil {
			return "", err
		}
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("load wishlist purchase lifecycle: %w", err)
	}
	id = uuid.NewString()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO commerce_lifecycle_entries(id, profile_id, item_id, state, source, external_ref, quantity, amount, currency, notes)
		VALUES (?, ?, ?, 'purchase', 'wishlist', ?, ?, ?, ?, ?)
	`, id, strings.TrimSpace(profileID), strings.TrimSpace(in.ItemID), strings.TrimSpace(entryID), purchaseQuantity(in), in.PricePaid, in.Currency, purchaseNotes(in))
	if err != nil {
		return "", fmt.Errorf("create wishlist purchase lifecycle: %w", err)
	}
	if err := s.ensureExpectedArrival(ctx, profileID, id, in); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) ensureExpectedArrival(ctx context.Context, profileID, lifecycleID string, in Entry) error {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM expected_arrivals
		WHERE (? = '' OR profile_id = ?) AND lifecycle_entry_id = ?
		LIMIT 1
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), strings.TrimSpace(lifecycleID)).Scan(&id)
	if err == nil {
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE expected_arrivals
			SET quantity = ?, amount = ?, currency = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, purchaseQuantity(in), in.PricePaid, in.Currency, purchaseNotes(in), id)
		if updateErr != nil {
			return fmt.Errorf("update wishlist expected arrival: %w", updateErr)
		}
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("load wishlist expected arrival: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO expected_arrivals(id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, expected_on, delivered_on, reconciled_instance_id, notes)
		VALUES (?, ?, ?, ?, 'wishlist', ?, ?, ?, ?, 'expected', '', '', '', ?)
	`, uuid.NewString(), strings.TrimSpace(profileID), strings.TrimSpace(in.ItemID), strings.TrimSpace(lifecycleID), strings.TrimSpace(in.ID), purchaseQuantity(in), in.PricePaid, in.Currency, purchaseNotes(in))
	if err != nil {
		return fmt.Errorf("create wishlist expected arrival: %w", err)
	}
	return nil
}

func (s *Service) ensurePurchasedInventoryInstance(ctx context.Context, in Entry, inTransit bool) (string, error) {
	wishlistNote := wishlistInstanceNote(in.ID)
	var id string
	var quantity int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, quantity
		FROM instances
		WHERE item_id = ? AND notes LIKE ?
		LIMIT 1
	`, strings.TrimSpace(in.ItemID), "%"+wishlistNote+"%").Scan(&id, &quantity)
	if err == nil {
		nextQuantity := purchaseQuantity(in)
		if inTransit && quantity > nextQuantity {
			nextQuantity = quantity
		}
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE instances
			SET condition = ?, status = ?, quantity = ?, acquisition_price = ?, acquisition_date = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, purchaseCondition(in, inTransit), purchaseStatus(in, inTransit), nextQuantity, in.PricePaid, strings.TrimSpace(in.PurchaseDate), id)
		if updateErr != nil {
			return "", fmt.Errorf("update wishlist purchase inventory instance: %w", updateErr)
		}
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("load wishlist purchase inventory instance: %w", err)
	}
	err = s.db.QueryRowContext(ctx, `
		SELECT id, quantity
		FROM instances
		WHERE item_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, strings.TrimSpace(in.ItemID)).Scan(&id, &quantity)
	if err == nil {
		_, updateErr := s.db.ExecContext(ctx, `
			UPDATE instances
			SET condition = ?, status = ?, quantity = ?, acquisition_price = ?, acquisition_date = ?, notes = CASE WHEN TRIM(notes) = '' THEN ? ELSE notes || ' ' || ? END, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, purchaseCondition(in, inTransit), purchaseStatus(in, inTransit), quantity+purchaseQuantity(in), in.PricePaid, strings.TrimSpace(in.PurchaseDate), wishlistNote, wishlistNote, id)
		if updateErr != nil {
			return "", fmt.Errorf("link wishlist purchase inventory instance: %w", updateErr)
		}
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("load existing wishlist purchase inventory instance: %w", err)
	}
	id = uuid.NewString()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES (?, ?, ?, ?, ?, '', ?, ?, ?)
	`, id, strings.TrimSpace(in.ItemID), purchaseCondition(in, inTransit), purchaseStatus(in, inTransit), purchaseQuantity(in), in.PricePaid, strings.TrimSpace(in.PurchaseDate), wishlistNote)
	if err != nil {
		return "", fmt.Errorf("create wishlist purchase inventory instance: %w", err)
	}
	return id, nil
}

func (s *Service) markPurchaseArrivalDelivered(ctx context.Context, profileID, lifecycleID, instanceID string, in Entry) error {
	deliveredOn := strings.TrimSpace(in.PurchaseDate)
	_, err := s.db.ExecContext(ctx, `
		UPDATE expected_arrivals
		SET status = 'delivered', delivered_on = ?, reconciled_instance_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE (? = '' OR profile_id = ?) AND lifecycle_entry_id = ?
	`, deliveredOn, strings.TrimSpace(instanceID), strings.TrimSpace(profileID), strings.TrimSpace(profileID), strings.TrimSpace(lifecycleID))
	if err != nil {
		return fmt.Errorf("mark wishlist expected arrival delivered: %w", err)
	}
	return nil
}

func purchaseQuantity(in Entry) int {
	if in.Quantity > 0 {
		return in.Quantity
	}
	if in.NeededQuantity > 0 {
		return in.NeededQuantity
	}
	return 1
}

func purchaseNotes(in Entry) string {
	notes := strings.TrimSpace(in.Notes)
	if notes == "" {
		return "Wishlist purchase"
	}
	return notes
}

func deliveredCondition(in Entry) string {
	condition := strings.ToLower(strings.TrimSpace(in.PurchaseCondition))
	if condition == "" {
		return "custom"
	}
	switch condition {
	case "sealed", "blister", "loose", "custom", "on_track":
		return condition
	default:
		return "custom"
	}
}

func purchaseCondition(in Entry, inTransit bool) string {
	if inTransit && strings.TrimSpace(in.PurchaseCondition) == "" {
		return "on_track"
	}
	return deliveredCondition(in)
}

func purchaseStatus(in Entry, inTransit bool) string {
	if inTransit {
		return "on_track"
	}
	return deliveredCondition(in)
}

func wishlistInstanceNote(entryID string) string {
	return "wishlist:" + strings.TrimSpace(entryID)
}
