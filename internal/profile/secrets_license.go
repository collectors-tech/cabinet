package profile

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (r *Repository) PutSecret(ctx context.Context, profileID, key, value string) error {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("secret key is required")
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO profile_secrets(profile_id, key, value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id, key) DO UPDATE SET
			value=excluded.value,
			updated_at=CURRENT_TIMESTAMP
	`, profileID, key, value); err != nil {
		return fmt.Errorf("upsert secret: %w", err)
	}
	return nil
}

func (r *Repository) GetSecret(ctx context.Context, profileID, key string) (string, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return "", err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("secret key is required")
	}
	var value string
	if err := r.db.QueryRowContext(ctx, `SELECT value FROM profile_secrets WHERE profile_id = ? AND key = ?`, profileID, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("secret not found")
		}
		return "", fmt.Errorf("get secret: %w", err)
	}
	return value, nil
}

func (r *Repository) PutLicense(ctx context.Context, profileID, licenseJSON string) error {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return err
	}
	if strings.TrimSpace(licenseJSON) == "" {
		return fmt.Errorf("license payload is required")
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO profile_licenses(profile_id, license_json, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET
			license_json=excluded.license_json,
			updated_at=CURRENT_TIMESTAMP
	`, profileID, licenseJSON); err != nil {
		return fmt.Errorf("upsert license: %w", err)
	}
	return nil
}

func (r *Repository) GetLicense(ctx context.Context, profileID string) (string, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return "", err
	}
	var value string
	if err := r.db.QueryRowContext(ctx, `SELECT license_json FROM profile_licenses WHERE profile_id = ?`, profileID).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("license not found")
		}
		return "", fmt.Errorf("get license: %w", err)
	}
	return value, nil
}
