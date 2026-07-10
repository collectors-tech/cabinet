package licensing

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/update"
)

type File struct {
	PayloadBase64   string `json:"payload_base64"`
	SignatureBase64 string `json:"signature_base64"`
}

type Payload struct {
	ProfileID string   `json:"profile_id"`
	Tier      string   `json:"tier"`
	ExpiresAt string   `json:"expires_at"`
	Features  []string `json:"features"`
}

type Status struct {
	State     string   `json:"state"`
	Tier      string   `json:"tier"`
	Features  []string `json:"features"`
	ExpiresAt string   `json:"expires_at"`
}

type Service struct {
	db        *sql.DB
	profiles  *profile.Repository
	publicKey string
}

func NewService(db *sql.DB, profiles *profile.Repository, publicKey string) *Service {
	return &Service{db: db, profiles: profiles, publicKey: strings.TrimSpace(publicKey)}
}

func (s *Service) Import(ctx context.Context, profileID string, file File) error {
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return err
	}
	raw, err := base64.StdEncoding.DecodeString(file.PayloadBase64)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	if err := update.VerifySignature(s.publicKey, raw, file.SignatureBase64); err != nil {
		return fmt.Errorf("verify license signature: %w", err)
	}
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("decode payload json: %w", err)
	}
	if p.ProfileID != profileID {
		return fmt.Errorf("license profile mismatch")
	}
	blob, _ := json.Marshal(file)
	return s.profiles.PutLicense(ctx, profileID, string(blob))
}

func (s *Service) Status(ctx context.Context, profileID string) (Status, error) {
	licJSON, err := s.profiles.GetLicense(ctx, profileID)
	if err != nil {
		return Status{State: "free", Tier: "free", Features: []string{}}, nil
	}
	var file File
	if err := json.Unmarshal([]byte(licJSON), &file); err != nil {
		return Status{State: "invalid", Tier: "free", Features: []string{}}, nil
	}
	raw, err := base64.StdEncoding.DecodeString(file.PayloadBase64)
	if err != nil {
		return Status{State: "invalid", Tier: "free", Features: []string{}}, nil
	}
	if err := update.VerifySignature(s.publicKey, raw, file.SignatureBase64); err != nil {
		return Status{State: "invalid", Tier: "free", Features: []string{}}, nil
	}
	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return Status{State: "invalid", Tier: "free", Features: []string{}}, nil
	}
	if p.ExpiresAt != "" {
		exp, err := time.Parse(time.RFC3339, p.ExpiresAt)
		if err != nil || time.Now().After(exp) {
			return Status{State: "expired", Tier: "free", Features: []string{}, ExpiresAt: p.ExpiresAt}, nil
		}
	}
	return Status{
		State:     "valid",
		Tier:      NormalizeTier(p.Tier),
		Features:  p.Features,
		ExpiresAt: p.ExpiresAt,
	}, nil
}

func (s *Service) Allow(ctx context.Context, profileID, feature string) (bool, error) {
	status, err := s.Status(ctx, profileID)
	if err != nil {
		return false, err
	}
	if status.State != "valid" || (status.Tier != "plus" && status.Tier != "pro") {
		return false, nil
	}
	for _, f := range status.Features {
		if f == feature {
			return true, nil
		}
	}
	return false, nil
}

func NormalizeTier(tier string) string {
	switch strings.TrimSpace(strings.ToLower(tier)) {
	case "", "free", "mvp":
		return "free"
	case "plus", "creator":
		return "plus"
	case "pro", "paid", "teams":
		return "pro"
	default:
		return strings.TrimSpace(strings.ToLower(tier))
	}
}
