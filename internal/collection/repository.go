package collection

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID                      string   `json:"id"`
	Brand                   string   `json:"brand"`
	Category                string   `json:"category"`
	PartNumber              string   `json:"part_number"`
	Title                   string   `json:"title"`
	Status                  string   `json:"status"`
	Priority                string   `json:"priority"`
	GradingStatus           string   `json:"grading_status"`
	Grader                  string   `json:"grader"`
	GradeNumeric            float64  `json:"grade_numeric"`
	Slabbed                 bool     `json:"slabbed"`
	CollectorClassification string   `json:"collector_classification"`
	CarGradeType            string   `json:"car_grade_type"`
	PackagingGradeType      string   `json:"packaging_grade_type"`
	Make                    string   `json:"make"`
	Model                   string   `json:"model"`
	Year                    string   `json:"year"`
	Scale                   string   `json:"scale"`
	Series                  string   `json:"series"`
	Description             string   `json:"description"`
	Tags                    []string `json:"tags"`
	CreatedAt               string   `json:"created_at"`
	CreatedBy               string   `json:"created_by"`
	UpdatedAt               string   `json:"updated_at"`
	UpdatedBy               string   `json:"updated_by"`
	DeletedAt               string   `json:"deleted_at"`
	DeletedBy               string   `json:"deleted_by"`
}

type AuditEvent struct {
	ID         string         `json:"id"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Action     string         `json:"action"`
	Actor      string         `json:"actor"`
	Source     string         `json:"source"`
	Before     map[string]any `json:"before"`
	After      map[string]any `json:"after"`
	CreatedAt  string         `json:"created_at"`
}

type Instance struct {
	ID               string  `json:"id"`
	ItemID           string  `json:"item_id"`
	Condition        string  `json:"condition"`
	Status           string  `json:"status"`
	Quantity         int     `json:"quantity"`
	StorageLocation  string  `json:"storage_location"`
	AcquisitionPrice float64 `json:"acquisition_price"`
	AcquisitionDate  string  `json:"acquisition_date"`
	Notes            string  `json:"notes"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

const (
	defaultAuditActor  = "system"
	defaultAuditSource = "collection.repository"
)

type BulkEditResult struct {
	UpdatedCount int `json:"updated_count"`
}

var allowedStatuses = map[string]struct{}{
	"sealed":   {},
	"blister":  {},
	"loose":    {},
	"custom":   {},
	"on_track": {},
}

var allowedItemLifecycleStatuses = map[string]struct{}{
	"active":  {},
	"deleted": {},
	"recycle": {},
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateItem(ctx context.Context, in Item) (Item, error) {
	return r.CreateItemForProfile(ctx, "", in)
}

func (r *Repository) CreateItemForProfile(ctx context.Context, profileID string, in Item) (Item, error) {
	in = normalizeItemForCreate(in)
	if in.PartNumber == "" || in.Title == "" {
		return Item{}, fmt.Errorf("part_number and title are required")
	}

	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return Item{}, fmt.Errorf("marshal tags: %w", err)
	}

	in.ID = uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO canonical_items (
			id, profile_id, brand, category, part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, tags_json, created_at, created_by, updated_at, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, strings.TrimSpace(profileID), in.Brand, in.Category, in.PartNumber, in.Title, in.Status, in.Priority, in.GradingStatus, in.Grader, in.GradeNumeric, boolToInt(in.Slabbed), in.CollectorClassification, in.CarGradeType, in.PackagingGradeType, in.Make, in.Model, in.Year, in.Scale, in.Series, in.Description, string(tagsJSON), now, defaultAuditActor, now, defaultAuditActor); err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}

	created, err := r.GetItemByID(ctx, in.ID)
	if err != nil {
		return Item{}, err
	}
	if err := r.appendAuditEvent(ctx, AuditEvent{
		EntityType: "canonical_item",
		EntityID:   created.ID,
		Action:     "create",
		Actor:      defaultAuditActor,
		Source:     defaultAuditSource,
		After:      itemAuditMap(created),
	}); err != nil {
		return Item{}, err
	}
	return created, nil
}

func normalizeItemForCreate(in Item) Item {
	in.Brand = strings.TrimSpace(in.Brand)
	if in.Brand == "" {
		in.Brand = "Unknown"
	}
	in.Category = strings.TrimSpace(in.Category)
	if in.Category == "" {
		in.Category = "General"
	}
	in.PartNumber = strings.TrimSpace(in.PartNumber)
	in.Title = strings.TrimSpace(in.Title)
	in.Status = normalizeItemLifecycleStatus(in.Status)
	in.Priority = normalizeItemPriority(in.Priority)
	in.GradingStatus = normalizeItemGradingStatus(in.GradingStatus)
	in.Grader = strings.TrimSpace(in.Grader)
	in.CollectorClassification = strings.TrimSpace(in.CollectorClassification)
	in.CarGradeType = strings.TrimSpace(in.CarGradeType)
	in.PackagingGradeType = strings.TrimSpace(in.PackagingGradeType)
	in.Make = strings.TrimSpace(in.Make)
	in.Model = strings.TrimSpace(in.Model)
	in.Year = strings.TrimSpace(in.Year)
	in.Scale = strings.TrimSpace(in.Scale)
	in.Series = strings.TrimSpace(in.Series)
	in.Description = strings.TrimSpace(in.Description)
	return in
}

func normalizeItemLifecycleStatus(raw string) string {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status == "" {
		return "active"
	}
	if _, ok := allowedItemLifecycleStatuses[status]; !ok {
		return "active"
	}
	return status
}

func normalizeItemPriority(raw string) string {
	priority := strings.ToLower(strings.TrimSpace(raw))
	if priority == "" {
		return "medium"
	}
	return priority
}

func normalizeItemGradingStatus(raw string) string {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status == "" {
		return "ungraded"
	}
	return status
}

func (r *Repository) UpdateItem(ctx context.Context, id string, changes Item) (Item, error) {
	current, err := r.GetItemByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return Item{}, err
	}
	next := current
	if v := strings.TrimSpace(changes.Brand); v != "" {
		next.Brand = v
	}
	if v := strings.TrimSpace(changes.Category); v != "" {
		next.Category = v
	}
	if v := strings.TrimSpace(changes.PartNumber); v != "" {
		next.PartNumber = v
	}
	if v := strings.TrimSpace(changes.Title); v != "" {
		next.Title = v
	}
	if v := strings.TrimSpace(changes.Status); v != "" {
		status := normalizeItemLifecycleStatus(v)
		next.Status = status
	}
	if v := strings.TrimSpace(changes.Priority); v != "" {
		next.Priority = normalizeItemPriority(v)
	}
	if v := strings.TrimSpace(changes.GradingStatus); v != "" {
		next.GradingStatus = normalizeItemGradingStatus(v)
	}
	if v := strings.TrimSpace(changes.Grader); v != "" {
		next.Grader = v
	}
	if changes.GradeNumeric > 0 {
		next.GradeNumeric = changes.GradeNumeric
	}
	if changes.Slabbed {
		next.Slabbed = true
	}
	if v := strings.TrimSpace(changes.CollectorClassification); v != "" {
		next.CollectorClassification = v
	}
	if v := strings.TrimSpace(changes.CarGradeType); v != "" {
		next.CarGradeType = v
	}
	if v := strings.TrimSpace(changes.PackagingGradeType); v != "" {
		next.PackagingGradeType = v
	}
	if v := strings.TrimSpace(changes.Make); v != "" {
		next.Make = v
	}
	if v := strings.TrimSpace(changes.Model); v != "" {
		next.Model = v
	}
	if v := strings.TrimSpace(changes.Year); v != "" {
		next.Year = v
	}
	if v := strings.TrimSpace(changes.Scale); v != "" {
		next.Scale = v
	}
	if v := strings.TrimSpace(changes.Series); v != "" {
		next.Series = v
	}
	if v := strings.TrimSpace(changes.Description); v != "" {
		next.Description = v
	}
	if len(changes.Tags) > 0 {
		next.Tags = changes.Tags
	}
	if next.Brand == "" || next.Category == "" || next.PartNumber == "" || next.Title == "" {
		return Item{}, fmt.Errorf("brand, category, part_number, and title are required")
	}

	tagsJSON, err := json.Marshal(next.Tags)
	if err != nil {
		return Item{}, fmt.Errorf("marshal tags: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET brand = ?, category = ?, part_number = ?, title = ?, status = ?, priority = ?, grading_status = ?, grader = ?, grade_numeric = ?, slabbed = ?, collector_classification = ?, car_grade_type = ?, packaging_grade_type = ?, make = ?, model = ?, year = ?, scale = ?, series = ?, description = ?, tags_json = ?, updated_at = ?, updated_by = ?
		WHERE id = ?
	`, next.Brand, next.Category, next.PartNumber, next.Title, next.Status, next.Priority, next.GradingStatus, next.Grader, next.GradeNumeric, boolToInt(next.Slabbed), next.CollectorClassification, next.CarGradeType, next.PackagingGradeType, next.Make, next.Model, next.Year, next.Scale, next.Series, next.Description, string(tagsJSON), time.Now().UTC().Format(time.RFC3339), defaultAuditActor, id); err != nil {
		return Item{}, fmt.Errorf("update item: %w", err)
	}
	updated, err := r.GetItemByID(ctx, id)
	if err != nil {
		return Item{}, err
	}
	if err := r.appendAuditEvent(ctx, AuditEvent{
		EntityType: "canonical_item",
		EntityID:   updated.ID,
		Action:     "update",
		Actor:      defaultAuditActor,
		Source:     defaultAuditSource,
		Before:     itemAuditMap(current),
		After:      itemAuditMap(updated),
	}); err != nil {
		return Item{}, err
	}
	return updated, nil
}

func (r *Repository) BulkEditItems(ctx context.Context, ids []string, changes Item) (BulkEditResult, error) {
	if len(ids) == 0 {
		return BulkEditResult{}, fmt.Errorf("item_ids are required")
	}
	updated := 0
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, err := r.UpdateItem(ctx, trimmed, changes); err != nil {
			return BulkEditResult{}, err
		}
		updated++
	}
	return BulkEditResult{UpdatedCount: updated}, nil
}

func (r *Repository) GetItemByID(ctx context.Context, id string) (Item, error) {
	var item Item
	var tagsRaw string
	var slabbed int
	err := r.db.QueryRowContext(ctx, `
		SELECT id, brand, category, part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, tags_json, created_at, created_by, updated_at, updated_by, COALESCE(deleted_at,''), COALESCE(deleted_by,'')
		FROM canonical_items WHERE id = ?
	`, id).Scan(
		&item.ID, &item.Brand, &item.Category, &item.PartNumber, &item.Title, &item.Status, &item.Priority, &item.GradingStatus, &item.Grader, &item.GradeNumeric, &slabbed, &item.CollectorClassification, &item.CarGradeType, &item.PackagingGradeType, &item.Make, &item.Model, &item.Year,
		&item.Scale, &item.Series, &item.Description, &tagsRaw, &item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy, &item.DeletedAt, &item.DeletedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, fmt.Errorf("item not found")
		}
		return Item{}, fmt.Errorf("get item: %w", err)
	}
	if err := json.Unmarshal([]byte(tagsRaw), &item.Tags); err != nil {
		return Item{}, fmt.Errorf("unmarshal tags: %w", err)
	}
	item.Slabbed = slabbed == 1
	return item, nil
}

func (r *Repository) ListItems(ctx context.Context) ([]Item, error) {
	return r.listItemsByQuery(ctx, `
		SELECT id, brand, category, part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, tags_json, created_at, created_by, updated_at, updated_by, COALESCE(deleted_at,''), COALESCE(deleted_by,'')
		FROM canonical_items ORDER BY created_at ASC
	`)
}

func (r *Repository) ListItemsByProfile(ctx context.Context, profileID string) ([]Item, error) {
	return r.listItemsByQuery(ctx, `
		SELECT id, brand, category, part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, tags_json, created_at, created_by, updated_at, updated_by, COALESCE(deleted_at,''), COALESCE(deleted_by,'')
		FROM canonical_items WHERE profile_id = ? ORDER BY created_at ASC
	`, strings.TrimSpace(profileID))
}

func (r *Repository) listItemsByQuery(ctx context.Context, query string, args ...any) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `
	`+query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var out []Item
	for rows.Next() {
		var item Item
		var tagsRaw string
		var slabbed int
		if err := rows.Scan(
			&item.ID, &item.Brand, &item.Category, &item.PartNumber, &item.Title, &item.Status, &item.Priority, &item.GradingStatus, &item.Grader, &item.GradeNumeric, &slabbed, &item.CollectorClassification, &item.CarGradeType, &item.PackagingGradeType, &item.Make, &item.Model, &item.Year,
			&item.Scale, &item.Series, &item.Description, &tagsRaw, &item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy, &item.DeletedAt, &item.DeletedBy,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsRaw), &item.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal item tags: %w", err)
		}
		item.Slabbed = slabbed == 1
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return out, nil
}

func (r *Repository) SetItemLifecycleStatus(ctx context.Context, id, status string) (Item, error) {
	next := normalizeItemLifecycleStatus(status)
	if _, ok := allowedItemLifecycleStatuses[next]; !ok {
		return Item{}, fmt.Errorf("invalid lifecycle status %q", status)
	}
	current, err := r.GetItemByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return Item{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	deletedAt := ""
	deletedBy := ""
	if next == "deleted" || next == "recycle" {
		deletedAt = now
		deletedBy = defaultAuditActor
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET status = ?, updated_at = ?, updated_by = ?, deleted_at = ?, deleted_by = ?
		WHERE id = ?
	`, next, now, defaultAuditActor, deletedAt, deletedBy, strings.TrimSpace(id)); err != nil {
		return Item{}, fmt.Errorf("set item lifecycle status: %w", err)
	}
	updated, err := r.GetItemByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return Item{}, err
	}
	action := "status_change"
	if next == "deleted" || next == "recycle" {
		action = "soft_delete"
	}
	if next == "active" && strings.TrimSpace(current.DeletedAt) != "" {
		action = "restore"
	}
	if err := r.appendAuditEvent(ctx, AuditEvent{
		EntityType: "canonical_item",
		EntityID:   updated.ID,
		Action:     action,
		Actor:      defaultAuditActor,
		Source:     defaultAuditSource,
		Before:     itemAuditMap(current),
		After:      itemAuditMap(updated),
	}); err != nil {
		return Item{}, err
	}
	return updated, nil
}

func (r *Repository) DeleteItemPermanent(ctx context.Context, id string) error {
	itemID := strings.TrimSpace(id)
	before, _ := r.GetItemByID(ctx, itemID)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM canonical_items WHERE id = ?`, itemID); err != nil {
		return fmt.Errorf("delete item permanently: %w", err)
	}
	if before.ID != "" {
		if err := r.appendAuditEvent(ctx, AuditEvent{
			EntityType: "canonical_item",
			EntityID:   before.ID,
			Action:     "permanent_delete",
			Actor:      defaultAuditActor,
			Source:     defaultAuditSource,
			Before:     itemAuditMap(before),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListAuditEventsByEntity(ctx context.Context, entityType, entityID string) ([]AuditEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, entity_type, entity_id, action, actor, source, before_json, after_json, created_at
		FROM audit_events
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY datetime(created_at) ASC, id ASC
	`, strings.TrimSpace(entityType), strings.TrimSpace(entityID))
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	events := make([]AuditEvent, 0)
	for rows.Next() {
		var event AuditEvent
		var beforeJSON string
		var afterJSON string
		if err := rows.Scan(&event.ID, &event.EntityType, &event.EntityID, &event.Action, &event.Actor, &event.Source, &beforeJSON, &afterJSON, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		if err := json.Unmarshal([]byte(beforeJSON), &event.Before); err != nil {
			return nil, fmt.Errorf("unmarshal audit before_json: %w", err)
		}
		if err := json.Unmarshal([]byte(afterJSON), &event.After); err != nil {
			return nil, fmt.Errorf("unmarshal audit after_json: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}
	return events, nil
}

func (r *Repository) appendAuditEvent(ctx context.Context, event AuditEvent) error {
	beforeJSON, err := json.Marshal(event.Before)
	if err != nil {
		return fmt.Errorf("marshal audit before payload: %w", err)
	}
	afterJSON, err := json.Marshal(event.After)
	if err != nil {
		return fmt.Errorf("marshal audit after payload: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(event.Actor) == "" {
		event.Actor = defaultAuditActor
	}
	if strings.TrimSpace(event.Source) == "" {
		event.Source = defaultAuditSource
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_events(id, entity_type, entity_id, action, actor, source, before_json, after_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.NewString(), strings.TrimSpace(event.EntityType), strings.TrimSpace(event.EntityID), strings.TrimSpace(event.Action), strings.TrimSpace(event.Actor), strings.TrimSpace(event.Source), string(beforeJSON), string(afterJSON), now); err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func itemAuditMap(item Item) map[string]any {
	if item.ID == "" {
		return map[string]any{}
	}
	return map[string]any{
		"id":             item.ID,
		"brand":          item.Brand,
		"category":       item.Category,
		"part_number":    item.PartNumber,
		"title":          item.Title,
		"status":         item.Status,
		"priority":       item.Priority,
		"grading_status": item.GradingStatus,
		"created_at":     item.CreatedAt,
		"created_by":     item.CreatedBy,
		"updated_at":     item.UpdatedAt,
		"updated_by":     item.UpdatedBy,
		"deleted_at":     item.DeletedAt,
		"deleted_by":     item.DeletedBy,
	}
}

func (r *Repository) ListItemDependencyCounts(ctx context.Context, id string) (map[string]int, error) {
	itemID := strings.TrimSpace(id)
	counts := map[string]int{}
	queries := map[string]string{
		"barcodes":   `SELECT COUNT(1) FROM item_barcodes WHERE item_id = ?`,
		"instances":  `SELECT COUNT(1) FROM instances WHERE item_id = ?`,
		"photos":     `SELECT COUNT(1) FROM item_photos WHERE item_id = ?`,
		"wishlist":   `SELECT COUNT(1) FROM wishlist_entries WHERE item_id = ?`,
		"tracked":    `SELECT COUNT(1) FROM tracked_items WHERE item_id = ?`,
		"price_data": `SELECT COUNT(1) FROM price_snapshots WHERE item_id = ?`,
	}
	for dependencyType, query := range queries {
		var c int
		if err := r.db.QueryRowContext(ctx, query, itemID).Scan(&c); err != nil {
			return nil, fmt.Errorf("count %s dependencies: %w", dependencyType, err)
		}
		if c > 0 {
			counts[dependencyType] = c
		}
	}
	return counts, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (r *Repository) CreateInstance(ctx context.Context, in Instance) (Instance, error) {
	in.ItemID = strings.TrimSpace(in.ItemID)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	if in.ItemID == "" || in.Status == "" {
		return Instance{}, fmt.Errorf("item_id and status are required")
	}
	if _, ok := allowedStatuses[in.Status]; !ok {
		return Instance{}, fmt.Errorf("invalid status %q", in.Status)
	}
	if in.Quantity <= 0 {
		in.Quantity = 1
	}

	in.ID = uuid.NewString()
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO instances (
			id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, in.ItemID, in.Condition, in.Status, in.Quantity, in.StorageLocation, in.AcquisitionPrice, in.AcquisitionDate, in.Notes); err != nil {
		return Instance{}, fmt.Errorf("create instance: %w", err)
	}

	return r.GetInstanceByID(ctx, in.ID)
}

func (r *Repository) GetInstanceByID(ctx context.Context, id string) (Instance, error) {
	var in Instance
	err := r.db.QueryRowContext(ctx, `
		SELECT id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes, created_at, updated_at
		FROM instances WHERE id = ?
	`, id).Scan(
		&in.ID, &in.ItemID, &in.Condition, &in.Status, &in.Quantity, &in.StorageLocation, &in.AcquisitionPrice, &in.AcquisitionDate, &in.Notes, &in.CreatedAt, &in.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Instance{}, fmt.Errorf("instance not found")
		}
		return Instance{}, fmt.Errorf("get instance: %w", err)
	}
	return in, nil
}

func (r *Repository) ListInstancesByItemID(ctx context.Context, itemID string) ([]Instance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes, created_at, updated_at
		FROM instances WHERE item_id = ? ORDER BY created_at ASC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	defer rows.Close()

	var out []Instance
	for rows.Next() {
		var in Instance
		if err := rows.Scan(
			&in.ID, &in.ItemID, &in.Condition, &in.Status, &in.Quantity, &in.StorageLocation, &in.AcquisitionPrice, &in.AcquisitionDate, &in.Notes, &in.CreatedAt, &in.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}
		out = append(out, in)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}
	return out, nil
}

func (r *Repository) UpdateInstance(ctx context.Context, id string, in Instance) (Instance, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Instance{}, fmt.Errorf("instance id is required")
	}
	current, err := r.GetInstanceByID(ctx, id)
	if err != nil {
		return Instance{}, err
	}
	next := current
	if v := strings.TrimSpace(in.Condition); v != "" {
		next.Condition = v
	}
	if v := strings.ToLower(strings.TrimSpace(in.Status)); v != "" {
		if _, ok := allowedStatuses[v]; !ok {
			return Instance{}, fmt.Errorf("invalid status %q", v)
		}
		next.Status = v
	}
	if in.Quantity > 0 {
		next.Quantity = in.Quantity
	}
	if v := strings.TrimSpace(in.StorageLocation); v != "" {
		next.StorageLocation = v
	}
	if in.AcquisitionPrice > 0 {
		next.AcquisitionPrice = in.AcquisitionPrice
	}
	if v := strings.TrimSpace(in.AcquisitionDate); v != "" {
		next.AcquisitionDate = v
	}
	if v := strings.TrimSpace(in.Notes); v != "" {
		next.Notes = v
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE instances
		SET condition = ?, status = ?, quantity = ?, storage_location = ?, acquisition_price = ?, acquisition_date = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, next.Condition, next.Status, next.Quantity, next.StorageLocation, next.AcquisitionPrice, next.AcquisitionDate, next.Notes, id); err != nil {
		return Instance{}, fmt.Errorf("update instance: %w", err)
	}
	return r.GetInstanceByID(ctx, id)
}
