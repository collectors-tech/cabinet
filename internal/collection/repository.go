package collection

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Item struct {
	ID          string   `json:"id"`
	Brand       string   `json:"brand"`
	Category    string   `json:"category"`
	PartNumber  string   `json:"part_number"`
	Title       string   `json:"title"`
	Make        string   `json:"make"`
	Model       string   `json:"model"`
	Year        string   `json:"year"`
	Scale       string   `json:"scale"`
	Series      string   `json:"series"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
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

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateItem(ctx context.Context, in Item) (Item, error) {
	in.Brand = strings.TrimSpace(in.Brand)
	in.Category = strings.TrimSpace(in.Category)
	in.PartNumber = strings.TrimSpace(in.PartNumber)
	in.Title = strings.TrimSpace(in.Title)
	if in.Brand == "" || in.Category == "" || in.PartNumber == "" || in.Title == "" {
		return Item{}, fmt.Errorf("brand, category, part_number, and title are required")
	}

	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return Item{}, fmt.Errorf("marshal tags: %w", err)
	}

	in.ID = uuid.NewString()
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO canonical_items (
			id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, in.Brand, in.Category, in.PartNumber, in.Title, in.Make, in.Model, in.Year, in.Scale, in.Series, in.Description, string(tagsJSON)); err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}

	return r.GetItemByID(ctx, in.ID)
}

func (r *Repository) GetItemByID(ctx context.Context, id string) (Item, error) {
	var item Item
	var tagsRaw string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json, created_at, updated_at
		FROM canonical_items WHERE id = ?
	`, id).Scan(
		&item.ID, &item.Brand, &item.Category, &item.PartNumber, &item.Title, &item.Make, &item.Model, &item.Year,
		&item.Scale, &item.Series, &item.Description, &tagsRaw, &item.CreatedAt, &item.UpdatedAt,
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
	return item, nil
}

func (r *Repository) ListItems(ctx context.Context) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json, created_at, updated_at
		FROM canonical_items ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var out []Item
	for rows.Next() {
		var item Item
		var tagsRaw string
		if err := rows.Scan(
			&item.ID, &item.Brand, &item.Category, &item.PartNumber, &item.Title, &item.Make, &item.Model, &item.Year,
			&item.Scale, &item.Series, &item.Description, &tagsRaw, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsRaw), &item.Tags); err != nil {
			return nil, fmt.Errorf("unmarshal item tags: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return out, nil
}

func (r *Repository) CreateInstance(ctx context.Context, in Instance) (Instance, error) {
	in.ItemID = strings.TrimSpace(in.ItemID)
	in.Status = strings.TrimSpace(in.Status)
	if in.ItemID == "" || in.Status == "" {
		return Instance{}, fmt.Errorf("item_id and status are required")
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
