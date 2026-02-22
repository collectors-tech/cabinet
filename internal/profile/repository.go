package profile

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Profile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM profiles ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var out []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate profiles: %w", err)
	}

	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Profile, error) {
	var p Profile
	err := r.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM profiles WHERE id = ?`, id).Scan(&p.ID, &p.Name, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Profile{}, fmt.Errorf("profile not found")
		}
		return Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}

func (r *Repository) Create(ctx context.Context, name string) (Profile, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Profile{}, fmt.Errorf("name is required")
	}

	p := Profile{
		ID:   uuid.NewString(),
		Name: name,
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO profiles (id, name) VALUES (?, ?)`, p.ID, p.Name); err != nil {
		return Profile{}, fmt.Errorf("create profile: %w", err)
	}

	if err := r.db.QueryRowContext(ctx, `SELECT created_at FROM profiles WHERE id = ?`, p.ID).Scan(&p.CreatedAt); err != nil {
		return Profile{}, fmt.Errorf("load profile: %w", err)
	}

	return p, nil
}

func (r *Repository) SetActiveProfile(ctx context.Context, id string) error {
	if _, err := r.GetByID(ctx, id); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES('active_profile_id', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value=excluded.value,
			updated_at=CURRENT_TIMESTAMP
	`, id)
	if err != nil {
		return fmt.Errorf("set active profile: %w", err)
	}
	return nil
}

func (r *Repository) GetActiveProfile(ctx context.Context) (Profile, error) {
	var id string
	if err := r.db.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'active_profile_id'`).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return Profile{}, fmt.Errorf("active profile not set")
		}
		return Profile{}, fmt.Errorf("get active profile id: %w", err)
	}
	return r.GetByID(ctx, id)
}
