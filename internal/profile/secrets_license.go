package profile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/collectors-tech/cabinet/internal/securestore"
)

func (r *Repository) PutSecret(ctx context.Context, profileID, key, value string) error {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("secret key is required")
	}
	if isAPISecretKey(key) {
		if err := securestore.Set(profileID, key, value); err == nil {
			return nil
		} else if os.Getenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK") != "1" {
			return fmt.Errorf("secure store set failed: %w", err)
		}
		encrypted, encErr := encryptFallbackSecret(profileID, key, value)
		if encErr != nil {
			return fmt.Errorf("fallback secret encryption failed: %w", encErr)
		}
		value = encrypted
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
	if isAPISecretKey(key) {
		if value, err := securestore.Get(profileID, key); err == nil {
			return value, nil
		} else if os.Getenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK") != "1" {
			return "", fmt.Errorf("secure store get failed: %w", err)
		}
	}
	var value string
	if err := r.db.QueryRowContext(ctx, `SELECT value FROM profile_secrets WHERE profile_id = ? AND key = ?`, profileID, key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("secret not found")
		}
		return "", fmt.Errorf("get secret: %w", err)
	}
	if isAPISecretKey(key) {
		decoded, err := decryptFallbackSecret(profileID, key, value)
		if err != nil {
			return "", fmt.Errorf("decode fallback secret: %w", err)
		}
		return decoded, nil
	}
	return value, nil
}

func isAPISecretKey(key string) bool {
	lk := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(lk, "api_key") || strings.Contains(lk, "token")
}

const fallbackSecretPrefix = "enc:v1:"

func encryptFallbackSecret(profileID, key, plaintext string) (string, error) {
	block, err := aes.NewCipher(fallbackSecretKey(profileID, key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	raw := append(nonce, ciphertext...)
	return fallbackSecretPrefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func decryptFallbackSecret(profileID, key, stored string) (string, error) {
	if !strings.HasPrefix(stored, fallbackSecretPrefix) {
		return "", fmt.Errorf("secret is not encrypted fallback format")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(stored, fallbackSecretPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(fallbackSecretKey(profileID, key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted payload too short")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func fallbackSecretKey(profileID, key string) []byte {
	material := strings.Join([]string{
		"cabinet-fallback-secret-v1",
		strings.TrimSpace(profileID),
		strings.TrimSpace(strings.ToLower(key)),
		strings.TrimSpace(os.Getenv("CABINET_FALLBACK_SECRET_PEPPER")),
	}, "|")
	sum := sha256.Sum256([]byte(material))
	return sum[:]
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
