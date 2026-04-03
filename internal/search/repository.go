package search

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/collectors-tech/cabinet/internal/collection"
)

type Query struct {
	Text      string `json:"text"`
	Brand     string `json:"brand"`
	Category  string `json:"category"`
	Condition string `json:"condition"`
	Status    string `json:"status"`
	Tags      string `json:"tags"`
	Scale     string `json:"scale"`
	SortBy    string `json:"sort_by"`
	Limit     int    `json:"limit"`
}

type SavedFilter struct {
	ID        string `json:"id"`
	ProfileID string `json:"profile_id"`
	Name      string `json:"name"`
	Query     Query  `json:"query"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SearchItems(ctx context.Context, q Query) ([]collection.Item, error) {
	return r.SearchItemsByProfile(ctx, "", q)
}

func (r *Repository) SearchItemsByProfile(ctx context.Context, profileID string, q Query) ([]collection.Item, error) {
	limit := q.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	sortExpr := "c.created_at ASC"
	requiresInstanceJoin := false
	switch strings.ToLower(strings.TrimSpace(q.SortBy)) {
	case "part_number":
		sortExpr = "c.part_number ASC"
	case "price":
		sortExpr = "COALESCE(MIN(i.acquisition_price), 0) ASC, c.part_number ASC"
		requiresInstanceJoin = true
	case "date_added":
		sortExpr = "c.created_at ASC"
	}

	args := []any{}
	var where []string
	if profile := strings.TrimSpace(profileID); profile != "" {
		where = append(where, "c.profile_id = ?")
		args = append(args, profile)
	}
	if b := strings.TrimSpace(q.Brand); b != "" {
		where = append(where, "c.brand = ?")
		args = append(args, b)
	}
	if c := strings.TrimSpace(q.Category); c != "" {
		where = append(where, "c.category = ?")
		args = append(args, c)
	}

	if scale := strings.TrimSpace(q.Scale); scale != "" {
		where = append(where, "c.scale = ?")
		args = append(args, scale)
	}
	if tags := strings.TrimSpace(q.Tags); tags != "" {
		where = append(where, "c.tags_json LIKE ?")
		args = append(args, "%"+tags+"%")
	}
	if cond := strings.TrimSpace(q.Condition); cond != "" {
		where = append(where, "EXISTS (SELECT 1 FROM instances ix WHERE ix.item_id = c.id AND ix.condition = ?)")
		args = append(args, cond)
	}
	if status := strings.TrimSpace(q.Status); status != "" {
		where = append(where, "EXISTS (SELECT 1 FROM instances is1 WHERE is1.item_id = c.id AND is1.status = ?)")
		args = append(args, status)
	}
	if t := strings.TrimSpace(q.Text); t != "" {
		where = append(where, "(c.id IN (SELECT item_id FROM canonical_items_fts WHERE canonical_items_fts MATCH ?) OR c.id IN (SELECT it.item_id FROM instances it WHERE it.notes LIKE ? OR it.storage_location LIKE ? OR it.status LIKE ? OR it.condition LIKE ?))")
		like := "%" + t + "%"
		args = append(args, buildFTSMatchQuery(t), like, like, like, like)
	}

	sqlText := `
		SELECT c.id, c.brand, c.category, c.part_number, c.title, c.make, c.model, c.year, c.scale, c.series, c.description, c.tags_json, c.created_at, c.updated_at
		FROM canonical_items c
	`
	if requiresInstanceJoin {
		sqlText += ` LEFT JOIN instances i ON i.item_id = c.id `
	}
	if len(where) > 0 {
		sqlText += " WHERE " + strings.Join(where, " AND ")
	}
	if requiresInstanceJoin {
		sqlText += " GROUP BY c.id"
	}
	sqlText += " ORDER BY " + sortExpr + " LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("search items: %w", err)
	}
	defer rows.Close()

	var out []collection.Item
	for rows.Next() {
		var it collection.Item
		var tagsRaw string
		if err := rows.Scan(&it.ID, &it.Brand, &it.Category, &it.PartNumber, &it.Title, &it.Make, &it.Model, &it.Year, &it.Scale, &it.Series, &it.Description, &tagsRaw, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan search item: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsRaw), &it.Tags); err != nil {
			return nil, fmt.Errorf("decode tags: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search items: %w", err)
	}
	return out, nil
}

func buildFTSMatchQuery(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	if strings.ContainsAny(text, " \t\r\n") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(text, "\"", "\"\""))
	}
	return text
}

func (r *Repository) SaveFilter(ctx context.Context, profileID, name string, q Query) (SavedFilter, error) {
	profileID = strings.TrimSpace(profileID)
	name = strings.TrimSpace(name)
	if profileID == "" || name == "" {
		return SavedFilter{}, fmt.Errorf("profile_id and name are required")
	}
	raw, err := json.Marshal(q)
	if err != nil {
		return SavedFilter{}, fmt.Errorf("encode query: %w", err)
	}
	sf := SavedFilter{
		ID:        uuid.NewString(),
		ProfileID: profileID,
		Name:      name,
		Query:     q,
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO saved_filters (id, profile_id, name, query_json) VALUES (?, ?, ?, ?)`, sf.ID, sf.ProfileID, sf.Name, string(raw)); err != nil {
		return SavedFilter{}, fmt.Errorf("insert saved filter: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT created_at, updated_at FROM saved_filters WHERE id = ?`, sf.ID).Scan(&sf.CreatedAt, &sf.UpdatedAt); err != nil {
		return SavedFilter{}, fmt.Errorf("load saved filter: %w", err)
	}
	return sf, nil
}

func (r *Repository) ListSavedFilters(ctx context.Context, profileID string) ([]SavedFilter, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, profile_id, name, query_json, created_at, updated_at FROM saved_filters WHERE profile_id = ? ORDER BY created_at ASC`, profileID)
	if err != nil {
		return nil, fmt.Errorf("list saved filters: %w", err)
	}
	defer rows.Close()

	var out []SavedFilter
	for rows.Next() {
		var sf SavedFilter
		var qRaw string
		if err := rows.Scan(&sf.ID, &sf.ProfileID, &sf.Name, &qRaw, &sf.CreatedAt, &sf.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan saved filter: %w", err)
		}
		if err := json.Unmarshal([]byte(qRaw), &sf.Query); err != nil {
			return nil, fmt.Errorf("decode saved filter query: %w", err)
		}
		out = append(out, sf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate saved filters: %w", err)
	}
	return out, nil
}

func (r *Repository) UpdateFilter(ctx context.Context, id, profileID, name string, q Query) (SavedFilter, error) {
	id = strings.TrimSpace(id)
	profileID = strings.TrimSpace(profileID)
	name = strings.TrimSpace(name)
	if id == "" || profileID == "" || name == "" {
		return SavedFilter{}, fmt.Errorf("id, profile_id and name are required")
	}
	raw, err := json.Marshal(q)
	if err != nil {
		return SavedFilter{}, fmt.Errorf("encode query: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE saved_filters
		SET name = ?, query_json = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ?
	`, name, string(raw), id, profileID)
	if err != nil {
		return SavedFilter{}, fmt.Errorf("update saved filter: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return SavedFilter{}, fmt.Errorf("saved filter not found")
	}
	var sf SavedFilter
	var qRaw string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, profile_id, name, query_json, created_at, updated_at
		FROM saved_filters WHERE id = ?
	`, id).Scan(&sf.ID, &sf.ProfileID, &sf.Name, &qRaw, &sf.CreatedAt, &sf.UpdatedAt); err != nil {
		return SavedFilter{}, fmt.Errorf("load saved filter: %w", err)
	}
	if err := json.Unmarshal([]byte(qRaw), &sf.Query); err != nil {
		return SavedFilter{}, fmt.Errorf("decode query: %w", err)
	}
	return sf, nil
}

func (r *Repository) DeleteFilter(ctx context.Context, id, profileID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM saved_filters WHERE id = ? AND profile_id = ?`, id, profileID)
	if err != nil {
		return fmt.Errorf("delete saved filter: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("saved filter not found")
	}
	return nil
}
