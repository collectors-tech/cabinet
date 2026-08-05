package companion

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/profile"
)

const (
	ProtocolVersionV1        = "1"
	CapabilityModulesRead    = "modules:read"
	CapabilityCapturesSubmit = "captures:submit"
	CapabilityMediaSubmit    = "media:submit"
	CapabilitySessionManage  = "session:manage"
	defaultPairingTTL        = 5 * time.Minute
	defaultSessionTTL        = 30 * 24 * time.Hour
)

var chromeExtensionIDPattern = regexp.MustCompile(`^[a-p]{32}$`)

type ProtocolError struct {
	Code string
}

func (e *ProtocolError) Error() string { return e.Code }

func protocolError(code string) error { return &ProtocolError{Code: code} }

func ErrorCode(err error) string {
	if typed, ok := err.(*ProtocolError); ok {
		return typed.Code
	}
	return "companion_internal_error"
}

type RequestMetadata struct {
	Origin         string
	DeviceID       string
	RemoteAddress  string
	IdempotencyKey string
}

type PairingRequestInput struct {
	DeviceID        string   `json:"device_id"`
	DeviceName      string   `json:"device_name"`
	ProtocolVersion string   `json:"protocol_version"`
	Capabilities    []string `json:"capabilities"`
}

type PairingRequestReceipt struct {
	RequestID       string   `json:"request_id"`
	ExchangeSecret  string   `json:"exchange_secret"`
	PairingCode     string   `json:"pairing_code"`
	Status          string   `json:"status"`
	ExpiresAt       string   `json:"expires_at"`
	ProtocolVersion string   `json:"protocol_version"`
	Capabilities    []string `json:"capabilities"`
}

type PairingRequestSummary struct {
	RequestID       string   `json:"request_id"`
	PairingCode     string   `json:"pairing_code"`
	DeviceID        string   `json:"device_id"`
	DeviceName      string   `json:"device_name"`
	Origin          string   `json:"origin"`
	ProtocolVersion string   `json:"protocol_version"`
	Capabilities    []string `json:"capabilities"`
	Status          string   `json:"status"`
	ExpiresAt       string   `json:"expires_at"`
	CreatedAt       string   `json:"created_at"`
}

type PairingApprovalInput struct {
	RequestID    string   `json:"request_id"`
	ProfileID    string   `json:"profile_id"`
	Capabilities []string `json:"capabilities"`
}

type PairingExchangeInput struct {
	RequestID       string `json:"request_id"`
	ExchangeSecret  string `json:"exchange_secret"`
	DeviceID        string `json:"device_id"`
	ProtocolVersion string `json:"protocol_version"`
}

type CredentialResponse struct {
	Credential string  `json:"credential"`
	Session    Session `json:"session"`
}

type Session struct {
	ID                string   `json:"id"`
	CabinetInstanceID string   `json:"cabinet_instance_id"`
	ProfileID         string   `json:"profile_id"`
	DeviceID          string   `json:"device_id"`
	DeviceName        string   `json:"device_name"`
	Origin            string   `json:"origin"`
	ProtocolVersion   string   `json:"protocol_version"`
	Capabilities      []string `json:"capabilities"`
	CreatedAt         string   `json:"created_at"`
	ExpiresAt         string   `json:"expires_at"`
	LastUsedAt        string   `json:"last_used_at,omitempty"`
	RevokedAt         string   `json:"revoked_at,omitempty"`
	RotatedFromID     string   `json:"rotated_from_id,omitempty"`
}

type Options struct {
	Now                      func() time.Time
	Random                   io.Reader
	PairingTTL               time.Duration
	SessionTTL               time.Duration
	PairingRequestsPerMinute int
	SessionRequestsPerMinute int
	MaxConcurrentPerSession  int
}

type rateWindow struct {
	StartedAt time.Time
	Count     int
}

func NewPersistentService(ctx context.Context, conn *sql.DB, profiles *profile.Repository, modules []Module, options Options) (*Service, error) {
	if conn == nil || profiles == nil {
		return nil, fmt.Errorf("companion persistent dependencies are required")
	}
	svc := newConfiguredService(modules, options)
	svc.db = conn
	svc.profiles = profiles
	instanceID, err := svc.ensureCabinetInstanceID(ctx)
	if err != nil {
		return nil, err
	}
	svc.instanceID = instanceID
	if err := svc.ResumePendingCaptures(ctx); err != nil {
		return nil, fmt.Errorf("resume companion captures: %w", err)
	}
	return svc, nil
}

func defaultOptions(options Options) Options {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.PairingTTL <= 0 {
		options.PairingTTL = defaultPairingTTL
	}
	if options.SessionTTL <= 0 {
		options.SessionTTL = defaultSessionTTL
	}
	if options.PairingRequestsPerMinute <= 0 {
		options.PairingRequestsPerMinute = 10
	}
	if options.SessionRequestsPerMinute <= 0 {
		options.SessionRequestsPerMinute = 120
	}
	if options.MaxConcurrentPerSession <= 0 {
		options.MaxConcurrentPerSession = 4
	}
	return options
}

func (s *Service) CreatePairingRequest(ctx context.Context, input PairingRequestInput, metadata RequestMetadata) (PairingRequestReceipt, error) {
	if s.db == nil {
		return PairingRequestReceipt{}, protocolError("companion_unavailable")
	}
	origin, err := ValidateExtensionOrigin(metadata.Origin)
	if err != nil {
		return PairingRequestReceipt{}, err
	}
	deviceID := boundedText(input.DeviceID, 128)
	deviceName := boundedText(input.DeviceName, 128)
	if deviceID == "" || deviceName == "" {
		return PairingRequestReceipt{}, protocolError("companion_device_required")
	}
	if boundedText(metadata.DeviceID, 128) != deviceID {
		return PairingRequestReceipt{}, protocolError("companion_device_mismatch")
	}
	if input.ProtocolVersion != ProtocolVersionV1 {
		return PairingRequestReceipt{}, protocolError("companion_protocol_unsupported")
	}
	capabilities, err := normaliseCapabilities(input.Capabilities, true)
	if err != nil {
		return PairingRequestReceipt{}, err
	}
	if err := s.checkRate("pairing:"+origin, s.options.PairingRequestsPerMinute); err != nil {
		return PairingRequestReceipt{}, err
	}
	requestID, err := s.randomValue(18)
	if err != nil {
		return PairingRequestReceipt{}, protocolError("companion_random_failed")
	}
	exchangeSecret, err := s.randomValue(32)
	if err != nil {
		return PairingRequestReceipt{}, protocolError("companion_random_failed")
	}
	pairingCode, err := s.randomPairingCode()
	if err != nil {
		return PairingRequestReceipt{}, protocolError("companion_random_failed")
	}
	now := s.options.Now().UTC()
	expiresAt := now.Add(s.options.PairingTTL)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO companion_pairing_requests(
			id, exchange_verifier, pairing_code, device_id, device_name, origin,
			protocol_version, requested_capabilities_json, status, expires_at, created_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?, ?)
	`, requestID, verifier(exchangeSecret), pairingCode, deviceID, deviceName, origin,
		ProtocolVersionV1, encodeStrings(capabilities), formatTime(expiresAt), formatTime(now))
	if err != nil {
		return PairingRequestReceipt{}, protocolError("companion_pairing_store_failed")
	}
	s.recordAudit(ctx, "", "", "pairing.requested", "accepted", metadata)
	return PairingRequestReceipt{
		RequestID:       requestID,
		ExchangeSecret:  exchangeSecret,
		PairingCode:     pairingCode,
		Status:          "pending",
		ExpiresAt:       formatTime(expiresAt),
		ProtocolVersion: ProtocolVersionV1,
		Capabilities:    capabilities,
	}, nil
}

func (s *Service) ListPairingRequests(ctx context.Context, profileID string) ([]PairingRequestSummary, error) {
	if s.db == nil {
		return nil, protocolError("companion_unavailable")
	}
	s.expirePairingRequests(ctx)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, pairing_code, device_id, device_name, origin, protocol_version,
			CASE WHEN status = 'approved' THEN granted_capabilities_json ELSE requested_capabilities_json END,
			status, expires_at, created_at
		FROM companion_pairing_requests
		WHERE status = 'pending' OR (status = 'approved' AND profile_id = ?)
		ORDER BY created_at ASC
	`, strings.TrimSpace(profileID))
	if err != nil {
		return nil, protocolError("companion_pairing_list_failed")
	}
	defer rows.Close()
	var result []PairingRequestSummary
	for rows.Next() {
		var item PairingRequestSummary
		var capabilities string
		if err := rows.Scan(&item.RequestID, &item.PairingCode, &item.DeviceID, &item.DeviceName, &item.Origin,
			&item.ProtocolVersion, &capabilities, &item.Status, &item.ExpiresAt, &item.CreatedAt); err != nil {
			return nil, protocolError("companion_pairing_list_failed")
		}
		item.Capabilities = decodeStrings(capabilities)
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) ApprovePairing(ctx context.Context, input PairingApprovalInput, metadata RequestMetadata) (PairingRequestSummary, error) {
	requestID := strings.TrimSpace(input.RequestID)
	profileID := strings.TrimSpace(input.ProfileID)
	if requestID == "" || profileID == "" {
		return PairingRequestSummary{}, protocolError("companion_pairing_approval_invalid")
	}
	if _, err := s.profiles.GetByID(ctx, profileID); err != nil {
		return PairingRequestSummary{}, protocolError("companion_profile_not_found")
	}
	summary, requested, err := s.loadPairingRequest(ctx, requestID)
	if err != nil {
		return PairingRequestSummary{}, err
	}
	if summary.Status != "pending" || !parseTime(summary.ExpiresAt).After(s.options.Now()) {
		return PairingRequestSummary{}, protocolError("companion_pairing_not_pending")
	}
	granted := input.Capabilities
	if len(granted) == 0 {
		granted = requested
	}
	granted, err = normaliseCapabilities(granted, false)
	if err != nil || !isSubset(granted, requested) {
		return PairingRequestSummary{}, protocolError("companion_capability_not_requested")
	}
	now := formatTime(s.options.Now().UTC())
	result, err := s.db.ExecContext(ctx, `
		UPDATE companion_pairing_requests
		SET profile_id = ?, granted_capabilities_json = ?, status = 'approved', approved_at = ?
		WHERE id = ? AND status = 'pending'
	`, profileID, encodeStrings(granted), now, requestID)
	if err != nil {
		return PairingRequestSummary{}, protocolError("companion_pairing_approval_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return PairingRequestSummary{}, protocolError("companion_pairing_not_pending")
	}
	summary.Status = "approved"
	summary.Capabilities = granted
	s.recordAudit(ctx, profileID, "", "pairing.approved", "accepted", metadata)
	return summary, nil
}

func (s *Service) RejectPairing(ctx context.Context, requestID, profileID string, metadata RequestMetadata) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE companion_pairing_requests SET status = 'rejected'
		WHERE id = ? AND (status = 'pending' OR (status = 'approved' AND profile_id = ?))
	`, strings.TrimSpace(requestID), strings.TrimSpace(profileID))
	if err != nil {
		return protocolError("companion_pairing_reject_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return protocolError("companion_pairing_not_pending")
	}
	s.recordAudit(ctx, "", "", "pairing.rejected", "accepted", metadata)
	return nil
}

func (s *Service) ExchangePairing(ctx context.Context, input PairingExchangeInput, metadata RequestMetadata) (CredentialResponse, error) {
	origin, err := ValidateExtensionOrigin(metadata.Origin)
	if err != nil {
		return CredentialResponse{}, err
	}
	if err := s.checkRate("exchange:"+origin, s.options.PairingRequestsPerMinute); err != nil {
		return CredentialResponse{}, err
	}
	requestID := strings.TrimSpace(input.RequestID)
	deviceID := boundedText(input.DeviceID, 128)
	if requestID == "" || deviceID == "" || input.ProtocolVersion != ProtocolVersionV1 {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_invalid")
	}
	if boundedText(metadata.DeviceID, 128) != deviceID {
		return CredentialResponse{}, protocolError("companion_device_mismatch")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_failed")
	}
	defer tx.Rollback()
	var storedVerifier, storedDeviceID, storedDeviceName, storedOrigin, protocolVersion, capabilitiesRaw, profileID, status, expiresAt string
	err = tx.QueryRowContext(ctx, `
		SELECT exchange_verifier, device_id, device_name, origin, protocol_version,
			granted_capabilities_json, profile_id, status, expires_at
		FROM companion_pairing_requests WHERE id = ?
	`, requestID).Scan(&storedVerifier, &storedDeviceID, &storedDeviceName, &storedOrigin, &protocolVersion,
		&capabilitiesRaw, &profileID, &status, &expiresAt)
	if err != nil {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_invalid")
	}
	if status == "exchanged" {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_replayed")
	}
	if status != "approved" || !parseTime(expiresAt).After(s.options.Now()) ||
		storedOrigin != origin || storedDeviceID != deviceID || protocolVersion != input.ProtocolVersion ||
		!verifierMatches(storedVerifier, input.ExchangeSecret) {
		s.recordAudit(ctx, profileID, "", "pairing.exchange", "rejected", metadata)
		return CredentialResponse{}, protocolError("companion_pairing_exchange_invalid")
	}
	credential, session, err := s.newSessionRecord(profileID, storedDeviceID, storedDeviceName, storedOrigin, protocolVersion, decodeStrings(capabilitiesRaw), "")
	if err != nil {
		return CredentialResponse{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO companion_sessions(
			id, credential_verifier, cabinet_instance_id, profile_id, device_id, device_name,
			origin, protocol_version, capabilities_json, created_at, expires_at, rotated_from_id
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, verifier(credential), session.CabinetInstanceID, session.ProfileID, session.DeviceID,
		session.DeviceName, session.Origin, session.ProtocolVersion, encodeStrings(session.Capabilities),
		session.CreatedAt, session.ExpiresAt, session.RotatedFromID); err != nil {
		return CredentialResponse{}, protocolError("companion_session_store_failed")
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE companion_pairing_requests SET status = 'exchanged', exchanged_at = ?
		WHERE id = ? AND status = 'approved'
	`, formatTime(s.options.Now().UTC()), requestID)
	if err != nil {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_replayed")
	}
	if err := tx.Commit(); err != nil {
		return CredentialResponse{}, protocolError("companion_pairing_exchange_failed")
	}
	s.recordAudit(ctx, profileID, session.ID, "pairing.exchanged", "accepted", metadata)
	return CredentialResponse{Credential: credential, Session: session}, nil
}

func (s *Service) Authenticate(ctx context.Context, authorization string, metadata RequestMetadata, capability string) (Session, error) {
	if s.db == nil {
		return Session{}, protocolError("companion_auth_required")
	}
	credential := bearerCredential(authorization)
	if !strings.HasPrefix(credential, "cabcmp_") || strings.HasPrefix(credential, "companion:") {
		return Session{}, protocolError("companion_auth_required")
	}
	origin, err := ValidateExtensionOrigin(metadata.Origin)
	if err != nil {
		return Session{}, protocolError("companion_origin_rejected")
	}
	deviceID := boundedText(metadata.DeviceID, 128)
	if deviceID == "" {
		return Session{}, protocolError("companion_device_required")
	}
	session, err := s.sessionByVerifier(ctx, verifier(credential))
	if err != nil {
		return Session{}, protocolError("companion_auth_required")
	}
	if session.RevokedAt != "" {
		return Session{}, protocolError("companion_session_revoked")
	}
	if !parseTime(session.ExpiresAt).After(s.options.Now()) {
		return Session{}, protocolError("companion_session_expired")
	}
	if session.CabinetInstanceID != s.instanceID {
		return Session{}, protocolError("companion_session_binding_mismatch")
	}
	if session.Origin != origin || session.DeviceID != deviceID {
		return Session{}, protocolError("companion_session_binding_mismatch")
	}
	if capability != "" && !containsString(session.Capabilities, capability) {
		return Session{}, protocolError("companion_capability_denied")
	}
	if err := s.checkRate("session:"+session.ID, s.options.SessionRequestsPerMinute); err != nil {
		return Session{}, err
	}
	now := formatTime(s.options.Now().UTC())
	_, _ = s.db.ExecContext(ctx, `UPDATE companion_sessions SET last_used_at = ? WHERE id = ?`, now, session.ID)
	session.LastUsedAt = now
	return session, nil
}

func (s *Service) BeginBoundedRequest(ctx context.Context, authorization string, metadata RequestMetadata, capability string) (Session, func(), error) {
	session, err := s.Authenticate(ctx, authorization, metadata, capability)
	if err != nil {
		return Session{}, nil, err
	}
	release, err := s.acquireSession(session.ID)
	if err != nil {
		return Session{}, nil, err
	}
	return session, release, nil
}

func (s *Service) RecordMediaTransport(ctx context.Context, session Session, metadata RequestMetadata, result string) {
	s.recordAudit(ctx, session.ProfileID, session.ID, "media.transport.validated", boundedCaptureText(result, 64), metadata)
}

func (s *Service) RotateCredential(ctx context.Context, authorization string, metadata RequestMetadata) (CredentialResponse, error) {
	current, err := s.Authenticate(ctx, authorization, metadata, CapabilitySessionManage)
	if err != nil {
		return CredentialResponse{}, err
	}
	credential, next, err := s.newSessionRecord(current.ProfileID, current.DeviceID, current.DeviceName,
		current.Origin, current.ProtocolVersion, current.Capabilities, current.ID)
	if err != nil {
		return CredentialResponse{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CredentialResponse{}, protocolError("companion_session_rotation_failed")
	}
	defer tx.Rollback()
	now := formatTime(s.options.Now().UTC())
	result, err := tx.ExecContext(ctx, `UPDATE companion_sessions SET revoked_at = ? WHERE id = ? AND revoked_at = ''`, now, current.ID)
	if err != nil {
		return CredentialResponse{}, protocolError("companion_session_rotation_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return CredentialResponse{}, protocolError("companion_session_revoked")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO companion_sessions(
			id, credential_verifier, cabinet_instance_id, profile_id, device_id, device_name,
			origin, protocol_version, capabilities_json, created_at, expires_at, rotated_from_id
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, next.ID, verifier(credential), next.CabinetInstanceID, next.ProfileID, next.DeviceID, next.DeviceName,
		next.Origin, next.ProtocolVersion, encodeStrings(next.Capabilities), next.CreatedAt, next.ExpiresAt, next.RotatedFromID); err != nil {
		return CredentialResponse{}, protocolError("companion_session_rotation_failed")
	}
	if err := tx.Commit(); err != nil {
		return CredentialResponse{}, protocolError("companion_session_rotation_failed")
	}
	s.recordAudit(ctx, current.ProfileID, next.ID, "session.rotated", "accepted", metadata)
	return CredentialResponse{Credential: credential, Session: next}, nil
}

func (s *Service) RevokeCredential(ctx context.Context, authorization string, metadata RequestMetadata) error {
	session, err := s.Authenticate(ctx, authorization, metadata, "")
	if err != nil {
		return err
	}
	return s.revokeSession(ctx, session.ProfileID, session.ID, metadata)
}

func (s *Service) ListSessions(ctx context.Context, profileID string) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, cabinet_instance_id, profile_id, device_id, device_name, origin,
			protocol_version, capabilities_json, created_at, expires_at, last_used_at, revoked_at, rotated_from_id
		FROM companion_sessions WHERE profile_id = ? ORDER BY created_at DESC
	`, strings.TrimSpace(profileID))
	if err != nil {
		return nil, protocolError("companion_session_list_failed")
	}
	defer rows.Close()
	var result []Session
	for rows.Next() {
		var item Session
		var capabilities string
		if err := rows.Scan(&item.ID, &item.CabinetInstanceID, &item.ProfileID, &item.DeviceID, &item.DeviceName,
			&item.Origin, &item.ProtocolVersion, &capabilities, &item.CreatedAt, &item.ExpiresAt,
			&item.LastUsedAt, &item.RevokedAt, &item.RotatedFromID); err != nil {
			return nil, protocolError("companion_session_list_failed")
		}
		item.Capabilities = decodeStrings(capabilities)
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) RevokeManagedSessions(ctx context.Context, profileID, sessionID string, revokeAll bool, metadata RequestMetadata) (int64, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return 0, protocolError("companion_profile_required")
	}
	if !revokeAll && strings.TrimSpace(sessionID) == "" {
		return 0, protocolError("companion_session_id_required")
	}
	now := formatTime(s.options.Now().UTC())
	var result sql.Result
	var err error
	if revokeAll {
		result, err = s.db.ExecContext(ctx, `UPDATE companion_sessions SET revoked_at = ? WHERE profile_id = ? AND revoked_at = ''`, now, profileID)
	} else {
		result, err = s.db.ExecContext(ctx, `UPDATE companion_sessions SET revoked_at = ? WHERE profile_id = ? AND id = ? AND revoked_at = ''`, now, profileID, strings.TrimSpace(sessionID))
	}
	if err != nil {
		return 0, protocolError("companion_session_revoke_failed")
	}
	count, _ := result.RowsAffected()
	s.recordAudit(ctx, profileID, strings.TrimSpace(sessionID), "session.revoked", "accepted", metadata)
	return count, nil
}

func (s *Service) acquireSession(sessionID string) (func(), error) {
	s.concurrencyMu.Lock()
	if s.activeBySession[sessionID] >= s.options.MaxConcurrentPerSession {
		s.concurrencyMu.Unlock()
		return nil, protocolError("companion_concurrency_limited")
	}
	s.activeBySession[sessionID]++
	s.concurrencyMu.Unlock()
	return func() {
		s.concurrencyMu.Lock()
		s.activeBySession[sessionID]--
		s.concurrencyMu.Unlock()
	}, nil
}

func (s *Service) sessionByVerifier(ctx context.Context, credentialVerifier string) (Session, error) {
	var session Session
	var capabilities string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, cabinet_instance_id, profile_id, device_id, device_name, origin,
			protocol_version, capabilities_json, created_at, expires_at, last_used_at, revoked_at, rotated_from_id
		FROM companion_sessions WHERE credential_verifier = ?
	`, credentialVerifier).Scan(&session.ID, &session.CabinetInstanceID, &session.ProfileID, &session.DeviceID,
		&session.DeviceName, &session.Origin, &session.ProtocolVersion, &capabilities, &session.CreatedAt,
		&session.ExpiresAt, &session.LastUsedAt, &session.RevokedAt, &session.RotatedFromID)
	if err != nil {
		return Session{}, err
	}
	session.Capabilities = decodeStrings(capabilities)
	return session, nil
}

func (s *Service) newSessionRecord(profileID, deviceID, deviceName, origin, protocolVersion string, capabilities []string, rotatedFromID string) (string, Session, error) {
	credentialRaw, err := s.randomValue(32)
	if err != nil {
		return "", Session{}, protocolError("companion_random_failed")
	}
	sessionID, err := s.randomValue(18)
	if err != nil {
		return "", Session{}, protocolError("companion_random_failed")
	}
	now := s.options.Now().UTC()
	return "cabcmp_" + credentialRaw, Session{
		ID:                sessionID,
		CabinetInstanceID: s.instanceID,
		ProfileID:         profileID,
		DeviceID:          deviceID,
		DeviceName:        deviceName,
		Origin:            origin,
		ProtocolVersion:   protocolVersion,
		Capabilities:      append([]string(nil), capabilities...),
		CreatedAt:         formatTime(now),
		ExpiresAt:         formatTime(now.Add(s.options.SessionTTL)),
		RotatedFromID:     rotatedFromID,
	}, nil
}

func (s *Service) revokeSession(ctx context.Context, profileID, sessionID string, metadata RequestMetadata) error {
	now := formatTime(s.options.Now().UTC())
	result, err := s.db.ExecContext(ctx, `UPDATE companion_sessions SET revoked_at = ? WHERE profile_id = ? AND id = ? AND revoked_at = ''`, now, profileID, sessionID)
	if err != nil {
		return protocolError("companion_session_revoke_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return protocolError("companion_session_revoked")
	}
	s.recordAudit(ctx, profileID, sessionID, "session.revoked", "accepted", metadata)
	return nil
}

func (s *Service) loadPairingRequest(ctx context.Context, requestID string) (PairingRequestSummary, []string, error) {
	var item PairingRequestSummary
	var requested string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, pairing_code, device_id, device_name, origin, protocol_version,
			requested_capabilities_json, status, expires_at, created_at
		FROM companion_pairing_requests WHERE id = ?
	`, requestID).Scan(&item.RequestID, &item.PairingCode, &item.DeviceID, &item.DeviceName,
		&item.Origin, &item.ProtocolVersion, &requested, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	if err != nil {
		return PairingRequestSummary{}, nil, protocolError("companion_pairing_not_found")
	}
	capabilities := decodeStrings(requested)
	item.Capabilities = capabilities
	return item, capabilities, nil
}

func (s *Service) expirePairingRequests(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, `UPDATE companion_pairing_requests SET status = 'expired' WHERE status IN ('pending', 'approved') AND expires_at <= ?`, formatTime(s.options.Now().UTC()))
}

func (s *Service) ensureCabinetInstanceID(ctx context.Context) (string, error) {
	var instanceID string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'companion.instance_id'`).Scan(&instanceID)
	if err == nil && strings.TrimSpace(instanceID) != "" {
		return instanceID, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("load companion instance id: %w", err)
	}
	value, err := s.randomValue(18)
	if err != nil {
		return "", err
	}
	instanceID = "cabinet_" + value
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at) VALUES('companion.instance_id', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO NOTHING
	`, instanceID)
	if err != nil {
		return "", fmt.Errorf("store companion instance id: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'companion.instance_id'`).Scan(&instanceID); err != nil {
		return "", fmt.Errorf("reload companion instance id: %w", err)
	}
	return instanceID, nil
}

func (s *Service) randomValue(byteLength int) (string, error) {
	buffer := make([]byte, byteLength)
	if _, err := io.ReadFull(s.options.Random, buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (s *Service) randomPairingCode() (string, error) {
	buffer := make([]byte, 4)
	if _, err := io.ReadFull(s.options.Random, buffer); err != nil {
		return "", err
	}
	value := (uint32(buffer[0])<<24 | uint32(buffer[1])<<16 | uint32(buffer[2])<<8 | uint32(buffer[3])) % 1000000
	return fmt.Sprintf("%06d", value), nil
}

func (s *Service) checkRate(key string, limit int) error {
	now := s.options.Now()
	s.rateMu.Lock()
	defer s.rateMu.Unlock()
	window := s.rateWindows[key]
	if window.StartedAt.IsZero() || now.Sub(window.StartedAt) >= time.Minute {
		window = rateWindow{StartedAt: now}
	}
	if window.Count >= limit {
		return protocolError("companion_rate_limited")
	}
	window.Count++
	s.rateWindows[key] = window
	return nil
}

func (s *Service) recordAudit(ctx context.Context, profileID, sessionID, action, result string, metadata RequestMetadata) {
	if s.db == nil {
		return
	}
	id, err := s.randomValue(12)
	if err != nil {
		return
	}
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO companion_audit_events(id, profile_id, session_id, action, result_code, origin, remote_address, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	`, id, boundedText(profileID, 128), boundedText(sessionID, 128), boundedText(action, 128),
		boundedText(result, 128), boundedText(metadata.Origin, 256), boundedText(metadata.RemoteAddress, 128), formatTime(s.options.Now().UTC()))
}

func ValidateExtensionOrigin(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", protocolError("companion_origin_rejected")
	}
	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Host)
	switch scheme {
	case "chrome-extension":
		if !chromeExtensionIDPattern.MatchString(host) {
			return "", protocolError("companion_origin_rejected")
		}
	case "moz-extension":
		if len(host) < 8 || len(host) > 128 || strings.ContainsAny(host, " /\\") {
			return "", protocolError("companion_origin_rejected")
		}
	default:
		return "", protocolError("companion_origin_rejected")
	}
	return scheme + "://" + host, nil
}

func normaliseCapabilities(input []string, defaultWhenEmpty bool) ([]string, error) {
	if len(input) == 0 && defaultWhenEmpty {
		input = []string{CapabilityModulesRead, CapabilityCapturesSubmit, CapabilitySessionManage}
	}
	allowed := map[string]bool{
		CapabilityModulesRead: true, CapabilityCapturesSubmit: true,
		CapabilityMediaSubmit: true, CapabilitySessionManage: true,
	}
	seen := map[string]bool{}
	for _, item := range input {
		item = strings.TrimSpace(item)
		if !allowed[item] {
			return nil, protocolError("companion_capability_unknown")
		}
		seen[item] = true
	}
	result := make([]string, 0, len(seen))
	for item := range seen {
		result = append(result, item)
	}
	sort.Strings(result)
	return result, nil
}

func verifier(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

func verifierMatches(stored, candidate string) bool {
	return subtle.ConstantTimeCompare([]byte(stored), []byte(verifier(candidate))) == 1
}

func bearerCredential(authorization string) string {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(authorization)), "bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimSpace(authorization)[7:])
}

func encodeStrings(values []string) string {
	encoded, _ := json.Marshal(values)
	return string(encoded)
}

func decodeStrings(raw string) []string {
	var result []string
	_ = json.Unmarshal([]byte(raw), &result)
	if result == nil {
		result = []string{}
	}
	return result
}

func isSubset(subset, superset []string) bool {
	allowed := map[string]bool{}
	for _, item := range superset {
		allowed[item] = true
	}
	for _, item := range subset {
		if !allowed[item] {
			return false
		}
	}
	return true
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func boundedText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) > limit {
		return ""
	}
	return value
}

func formatTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func parseTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed
}
