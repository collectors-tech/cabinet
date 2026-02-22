package profile

import (
	"context"
	"fmt"
)

func (r *Repository) GetSettings(ctx context.Context, profileID string) (map[string]string, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM profile_settings WHERE profile_id = ? ORDER BY key ASC`, profileID)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan settings: %w", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}
	return out, nil
}

func (r *Repository) PutSettings(ctx context.Context, profileID string, values map[string]string) error {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin settings tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for k, v := range values {
		if k == "" {
			return fmt.Errorf("setting key is required")
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO profile_settings(profile_id, key, value, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(profile_id, key) DO UPDATE SET
				value=excluded.value,
				updated_at=CURRENT_TIMESTAMP
		`, profileID, k, v); err != nil {
			return fmt.Errorf("upsert setting %q: %w", k, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings tx: %w", err)
	}
	return nil
}
