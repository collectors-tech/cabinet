package profile

import (
	"context"
	"database/sql"
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
	if tx, ok := r.db.(*sql.Tx); ok {
		return putSettingsValues(ctx, tx, profileID, values)
	}

	db, ok := r.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("profile settings database does not support transactions")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin settings tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := putSettingsValues(ctx, tx, profileID, values); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings tx: %w", err)
	}
	return nil
}

func putSettingsValues(ctx context.Context, exec queryExecutor, profileID string, values map[string]string) error {
	for k, v := range values {
		if k == "" {
			return fmt.Errorf("setting key is required")
		}
		if _, err := exec.ExecContext(ctx, `
			INSERT INTO profile_settings(profile_id, key, value, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(profile_id, key) DO UPDATE SET
				value=excluded.value,
				updated_at=CURRENT_TIMESTAMP
		`, profileID, k, v); err != nil {
			return fmt.Errorf("upsert setting %q: %w", k, err)
		}
	}
	return nil
}
