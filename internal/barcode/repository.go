package barcode

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Record struct {
	ID        string `json:"id"`
	ItemID    string `json:"item_id"`
	Barcode   string `json:"barcode"`
	CreatedAt string `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, itemID, barcode string) (Record, error) {
	itemID = strings.TrimSpace(itemID)
	barcode = strings.TrimSpace(barcode)
	if itemID == "" || barcode == "" {
		return Record{}, fmt.Errorf("item_id and barcode are required")
	}
	rec := Record{
		ID:      uuid.NewString(),
		ItemID:  itemID,
		Barcode: barcode,
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO item_barcodes (id, item_id, barcode) VALUES (?, ?, ?)`, rec.ID, rec.ItemID, rec.Barcode); err != nil {
		return Record{}, fmt.Errorf("insert barcode: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT created_at FROM item_barcodes WHERE id = ?`, rec.ID).Scan(&rec.CreatedAt); err != nil {
		return Record{}, fmt.Errorf("load barcode: %w", err)
	}
	return rec, nil
}

func (r *Repository) ListByItem(ctx context.Context, itemID string) ([]Record, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, item_id, barcode, created_at FROM item_barcodes WHERE item_id = ? ORDER BY created_at ASC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list barcodes: %w", err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.ItemID, &rec.Barcode, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan barcode: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate barcodes: %w", err)
	}
	return out, nil
}

func (r *Repository) Lookup(ctx context.Context, barcode string) ([]Record, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, item_id, barcode, created_at FROM item_barcodes WHERE barcode = ? ORDER BY created_at ASC`, barcode)
	if err != nil {
		return nil, fmt.Errorf("lookup barcode: %w", err)
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.ItemID, &rec.Barcode, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan lookup barcode: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lookup barcode: %w", err)
	}
	return out, nil
}
