package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/profile"
)

type Service struct {
	wa       *webauthn.WebAuthn
	db       *sql.DB
	profiles *profile.Repository

	mu       sync.Mutex
	sessions map[string]pendingSession
}

type pendingSession struct {
	ProfileID string
	Kind      string
	Session   webauthn.SessionData
	ExpiresAt time.Time
}

type WebAuthnStartResponse struct {
	SessionID string `json:"session_id"`
	Options   any    `json:"options"`
}

func NewService(cfg config.Config, db *sql.DB, profiles *profile.Repository) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.WebAuthnName,
		RPID:          cfg.WebAuthnRPID,
		RPOrigins:     []string{cfg.WebAuthnOrigin},
	})
	if err != nil {
		return nil, fmt.Errorf("create webauthn config: %w", err)
	}

	return &Service{
		wa:       wa,
		db:       db,
		profiles: profiles,
		sessions: map[string]pendingSession{},
	}, nil
}

func (s *Service) BeginRegistration(ctx context.Context, profileID string) (WebAuthnStartResponse, error) {
	user, err := s.loadUser(ctx, profileID)
	if err != nil {
		return WebAuthnStartResponse{}, err
	}

	options, session, err := s.wa.BeginRegistration(user)
	if err != nil {
		return WebAuthnStartResponse{}, fmt.Errorf("begin registration: %w", err)
	}

	token, err := randomToken()
	if err != nil {
		return WebAuthnStartResponse{}, err
	}

	s.storeSession(token, pendingSession{
		ProfileID: profileID,
		Kind:      "registration",
		Session:   *session,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})

	return WebAuthnStartResponse{
		SessionID: token,
		Options:   options,
	}, nil
}

func (s *Service) FinishRegistration(ctx context.Context, sessionID string, credentialPayload any) error {
	sess, err := s.popSession(sessionID, "registration")
	if err != nil {
		return err
	}

	user, err := s.loadUser(ctx, sess.ProfileID)
	if err != nil {
		return err
	}

	req, err := payloadToRequest(credentialPayload)
	if err != nil {
		return err
	}
	cred, err := s.wa.FinishRegistration(user, sess.Session, req)
	if err != nil {
		return fmt.Errorf("finish registration: %w", err)
	}
	if err := s.upsertCredential(ctx, sess.ProfileID, *cred); err != nil {
		return err
	}
	return nil
}

func (s *Service) BeginLogin(ctx context.Context, profileID string) (WebAuthnStartResponse, error) {
	user, err := s.loadUser(ctx, profileID)
	if err != nil {
		return WebAuthnStartResponse{}, err
	}

	options, session, err := s.wa.BeginLogin(user)
	if err != nil {
		return WebAuthnStartResponse{}, fmt.Errorf("begin login: %w", err)
	}
	token, err := randomToken()
	if err != nil {
		return WebAuthnStartResponse{}, err
	}
	s.storeSession(token, pendingSession{
		ProfileID: profileID,
		Kind:      "login",
		Session:   *session,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})

	return WebAuthnStartResponse{
		SessionID: token,
		Options:   options,
	}, nil
}

func (s *Service) FinishLogin(ctx context.Context, sessionID string, credentialPayload any) error {
	sess, err := s.popSession(sessionID, "login")
	if err != nil {
		return err
	}
	user, err := s.loadUser(ctx, sess.ProfileID)
	if err != nil {
		return err
	}

	req, err := payloadToRequest(credentialPayload)
	if err != nil {
		return err
	}
	cred, err := s.wa.FinishLogin(user, sess.Session, req)
	if err != nil {
		return fmt.Errorf("finish login: %w", err)
	}
	if err := s.upsertCredential(ctx, sess.ProfileID, *cred); err != nil {
		return err
	}
	return nil
}

func (s *Service) storeSession(sessionID string, p pendingSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = p
}

func (s *Service) popSession(sessionID, kind string) (pendingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, ok := s.sessions[sessionID]
	if !ok {
		return pendingSession{}, fmt.Errorf("invalid session")
	}
	delete(s.sessions, sessionID)
	if ps.Kind != kind {
		return pendingSession{}, fmt.Errorf("unexpected session kind")
	}
	if time.Now().After(ps.ExpiresAt) {
		return pendingSession{}, fmt.Errorf("session expired")
	}
	return ps, nil
}

func (s *Service) loadUser(ctx context.Context, profileID string) (*user, error) {
	p, err := s.profiles.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	creds, err := s.listCredentials(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return &user{
		id:          []byte(p.ID),
		name:        p.Name,
		displayName: p.Name,
		creds:       creds,
	}, nil
}

func payloadToRequest(payload any) (*http.Request, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal credential payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, "http://localhost/internal", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) listCredentials(ctx context.Context, profileID string) ([]webauthn.Credential, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT credential_json FROM webauthn_credentials WHERE profile_id = ?`, profileID)
	if err != nil {
		return nil, fmt.Errorf("query credentials: %w", err)
	}
	defer rows.Close()

	var out []webauthn.Credential
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan credential: %w", err)
		}
		var c webauthn.Credential
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			return nil, fmt.Errorf("decode credential: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}
	return out, nil
}

func (s *Service) upsertCredential(ctx context.Context, profileID string, cred webauthn.Credential) error {
	raw, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("encode credential: %w", err)
	}
	key := base64.RawURLEncoding.EncodeToString(cred.ID)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO webauthn_credentials(id, profile_id, credential_json, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			credential_json=excluded.credential_json,
			updated_at=CURRENT_TIMESTAMP
	`, key, profileID, string(raw))
	if err != nil {
		return fmt.Errorf("upsert credential: %w", err)
	}
	return nil
}

type user struct {
	id          []byte
	name        string
	displayName string
	creds       []webauthn.Credential
}

func (u *user) WebAuthnID() []byte {
	return u.id
}

func (u *user) WebAuthnName() string {
	return u.name
}

func (u *user) WebAuthnDisplayName() string {
	return u.displayName
}

func (u *user) WebAuthnCredentials() []webauthn.Credential {
	return u.creds
}
