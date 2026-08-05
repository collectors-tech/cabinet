package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/collectors-tech/cabinet/internal/auth"
	"github.com/collectors-tech/cabinet/internal/companion"
	"github.com/collectors-tech/cabinet/internal/profile"
)

const (
	companionJSONBodyLimit  = 1 << 20
	companionSmallBodyLimit = 16 << 10
	companionMediaBodyLimit = 8 << 20
)

func registerCompanionRoutes(mux *http.ServeMux, svc *companion.Service, profiles *profile.Repository, authService *auth.Service) {
	mux.HandleFunc("/api/companion/pairing/requests", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, r.Method == http.MethodPost || r.Method == http.MethodOptions)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input companion.PairingRequestInput
			if !decodeBoundedJSON(w, r, companionSmallBodyLimit, &input) {
				return
			}
			receipt, err := svc.CreatePairingRequest(r.Context(), input, metadata)
			if err != nil {
				writeCompanionError(w, err)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(receipt)
		case http.MethodGet:
			profileID, authorised := companionManagementProfile(w, r, profiles, authService)
			if !authorised || profileID == "" {
				return
			}
			requests, err := svc.ListPairingRequests(r.Context(), profileID)
			if err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"requests": requests})
		case http.MethodDelete:
			profileID, authorised := companionManagementProfile(w, r, profiles, authService)
			if !authorised {
				return
			}
			if err := svc.RejectPairing(r.Context(), r.URL.Query().Get("id"), profileID, metadata); err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"rejected": true})
		default:
			writeCompanionMethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/companion/pairing/approvals", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, false)
		if !ok {
			return
		}
		if r.Method != http.MethodPost {
			writeCompanionMethodNotAllowed(w)
			return
		}
		profileID, authorised := companionManagementProfile(w, r, profiles, authService)
		if !authorised {
			return
		}
		var input companion.PairingApprovalInput
		if !decodeBoundedJSON(w, r, companionSmallBodyLimit, &input) {
			return
		}
		if strings.TrimSpace(input.ProfileID) == "" {
			input.ProfileID = profileID
		}
		if input.ProfileID != profileID {
			writeCompanionCode(w, http.StatusForbidden, "companion_profile_mismatch")
			return
		}
		approved, err := svc.ApprovePairing(r.Context(), input, metadata)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(approved)
	})

	mux.HandleFunc("/api/companion/pairing/exchanges", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			writeCompanionMethodNotAllowed(w)
			return
		}
		var input companion.PairingExchangeInput
		if !decodeBoundedJSON(w, r, companionSmallBodyLimit, &input) {
			return
		}
		credential, err := svc.ExchangePairing(r.Context(), input, metadata)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(credential)
	})

	mux.HandleFunc("/api/companion/session", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			session, err := svc.Authenticate(r.Context(), r.Header.Get("Authorization"), metadata, "")
			if err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(session)
		case http.MethodDelete:
			if err := svc.RevokeCredential(r.Context(), r.Header.Get("Authorization"), metadata); err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"revoked": true})
		default:
			writeCompanionMethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/companion/session/rotate", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			writeCompanionMethodNotAllowed(w)
			return
		}
		credential, err := svc.RotateCredential(r.Context(), r.Header.Get("Authorization"), metadata)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(credential)
	})

	mux.HandleFunc("/api/companion/sessions", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, false)
		if !ok {
			return
		}
		profileID, authorised := companionManagementProfile(w, r, profiles, authService)
		if !authorised {
			return
		}
		switch r.Method {
		case http.MethodGet:
			sessions, err := svc.ListSessions(r.Context(), profileID)
			if err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"sessions": sessions})
		case http.MethodDelete:
			count, err := svc.RevokeManagedSessions(r.Context(), profileID, r.URL.Query().Get("id"), r.URL.Query().Get("all") == "true", metadata)
			if err != nil {
				writeCompanionError(w, err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"revoked_count": count})
		default:
			writeCompanionMethodNotAllowed(w)
		}
	})

	mux.HandleFunc("/api/companion/modules", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		if r.Method != http.MethodGet {
			writeCompanionMethodNotAllowed(w)
			return
		}
		registry, err := svc.RegistryForSession(r.Context(), r.Header.Get("Authorization"), metadata)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(registry)
	})

	mux.HandleFunc("/api/companion/payloads", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			writeCompanionMethodNotAllowed(w)
			return
		}
		var input companion.PayloadSubmission
		if !decodeBoundedJSON(w, r, companionJSONBodyLimit, &input) {
			return
		}
		accepted, err := svc.AcceptPayload(r.Context(), input, r.Header.Get("Authorization"), metadata)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(accepted)
	})

	mux.HandleFunc("/api/companion/media-submissions", func(w http.ResponseWriter, r *http.Request) {
		metadata, ok := prepareCompanionRequest(w, r, true)
		if !ok {
			return
		}
		if handleCompanionPreflight(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			writeCompanionMethodNotAllowed(w)
			return
		}
		session, release, err := svc.BeginBoundedRequest(r.Context(), r.Header.Get("Authorization"), metadata, companion.CapabilityMediaSubmit)
		if err != nil {
			writeCompanionError(w, err)
			return
		}
		defer release()
		if strings.TrimSpace(r.Header.Get("X-Cabinet-Profile")) != session.ProfileID {
			writeCompanionCode(w, http.StatusForbidden, "companion_profile_mismatch")
			return
		}
		if !validSHA256Header(r.Header.Get("X-Cabinet-Media-SHA256")) || boundedHeader(r.Header.Get("X-Cabinet-Idempotency-Key"), 128) == "" {
			writeCompanionCode(w, http.StatusBadRequest, "companion_media_metadata_invalid")
			return
		}
		contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
		if !strings.HasPrefix(contentType, "image/") && contentType != "application/octet-stream" {
			writeCompanionCode(w, http.StatusUnsupportedMediaType, "companion_media_content_type_rejected")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, companionMediaBodyLimit)
		digest := sha256.New()
		written, err := io.Copy(digest, r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeCompanionCode(w, http.StatusRequestEntityTooLarge, "companion_media_too_large")
				return
			}
			writeCompanionCode(w, http.StatusBadRequest, "companion_media_read_failed")
			return
		}
		if written == 0 {
			writeCompanionCode(w, http.StatusBadRequest, "companion_media_empty")
			return
		}
		if !strings.EqualFold(hex.EncodeToString(digest.Sum(nil)), r.Header.Get("X-Cabinet-Media-SHA256")) {
			writeCompanionCode(w, http.StatusBadRequest, "companion_media_checksum_mismatch")
			return
		}
		svc.RecordMediaTransport(r.Context(), session, metadata)
		writeCompanionCode(w, http.StatusNotImplemented, "companion_media_persistence_not_implemented")
	})
}

func companionManagementProfile(w http.ResponseWriter, r *http.Request, profiles *profile.Repository, authService *auth.Service) (string, bool) {
	active, err := profiles.GetActiveProfile(r.Context())
	if err != nil || strings.TrimSpace(active.ID) == "" {
		writeCompanionCode(w, http.StatusUnauthorized, "companion_cabinet_auth_required")
		return "", false
	}
	token := sessionTokenFromRequest(r)
	if token != "" {
		if err := authService.ValidateUnlockedSession(token); err != nil {
			writeCompanionCode(w, http.StatusUnauthorized, "companion_cabinet_auth_required")
			return "", false
		}
	} else if !authService.HasUnlockedSession(active.ID) {
		writeCompanionCode(w, http.StatusUnauthorized, "companion_cabinet_auth_required")
		return "", false
	}
	return active.ID, true
}

func prepareCompanionRequest(w http.ResponseWriter, r *http.Request, requireExtensionOrigin bool) (companion.RequestMetadata, bool) {
	w.Header().Set("Content-Type", "application/json")
	if !isLoopbackHost(r.Host) || !isLoopbackRemote(r.RemoteAddr) {
		writeCompanionCode(w, http.StatusForbidden, "companion_loopback_required")
		return companion.RequestMetadata{}, false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if requireExtensionOrigin && origin == "" {
		writeCompanionCode(w, http.StatusForbidden, "companion_origin_rejected")
		return companion.RequestMetadata{}, false
	}
	if requireExtensionOrigin {
		validatedOrigin, err := companion.ValidateExtensionOrigin(origin)
		if err != nil {
			writeCompanionCode(w, http.StatusForbidden, "companion_origin_rejected")
			return companion.RequestMetadata{}, false
		}
		origin = validatedOrigin
	}
	if !requireExtensionOrigin && origin != "" && !sameLoopbackOrigin(origin, r.Host) {
		writeCompanionCode(w, http.StatusForbidden, "companion_origin_rejected")
		return companion.RequestMetadata{}, false
	}
	if origin != "" && (strings.HasPrefix(strings.ToLower(origin), "chrome-extension://") || strings.HasPrefix(strings.ToLower(origin), "moz-extension://")) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	return companion.RequestMetadata{
		Origin: origin, DeviceID: r.Header.Get("X-Cabinet-Companion-Device"), RemoteAddress: r.RemoteAddr,
	}, true
}

func handleCompanionPreflight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodOptions {
		return false
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Cabinet-Companion-Device, X-Cabinet-Idempotency-Key, X-Cabinet-Media-SHA256, X-Cabinet-Profile")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
	return true
}

func decodeBoundedJSON(w http.ResponseWriter, r *http.Request, limit int64, target any) bool {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "application/json") {
		writeCompanionCode(w, http.StatusUnsupportedMediaType, "companion_content_type_required")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeCompanionCode(w, http.StatusRequestEntityTooLarge, "companion_payload_too_large")
		} else {
			writeCompanionCode(w, http.StatusBadRequest, "companion_invalid_json")
		}
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeCompanionCode(w, http.StatusBadRequest, "companion_invalid_json")
		return false
	}
	return true
}

func writeCompanionError(w http.ResponseWriter, err error) {
	code := companion.ErrorCode(err)
	status := http.StatusBadRequest
	switch code {
	case "companion_auth_required", "companion_session_revoked", "companion_session_expired":
		status = http.StatusUnauthorized
	case "companion_origin_rejected", "companion_session_binding_mismatch", "companion_capability_denied", "companion_profile_mismatch":
		status = http.StatusForbidden
	case "companion_pairing_not_found":
		status = http.StatusNotFound
	case "companion_pairing_not_pending", "companion_pairing_exchange_replayed":
		status = http.StatusConflict
	case "companion_rate_limited", "companion_concurrency_limited":
		status = http.StatusTooManyRequests
		w.Header().Set("Retry-After", "60")
	case "companion_payload_too_large", "companion_media_too_large":
		status = http.StatusRequestEntityTooLarge
	case "companion_unavailable":
		status = http.StatusServiceUnavailable
	}
	writeCompanionCode(w, status, code)
}

func writeCompanionCode(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": code})
}

func writeCompanionMethodNotAllowed(w http.ResponseWriter) {
	writeCompanionCode(w, http.StatusMethodNotAllowed, "method_not_allowed")
}

func isLoopbackHost(raw string) bool {
	host := strings.TrimSpace(raw)
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	host = strings.Trim(strings.ToLower(host), "[]")
	return host == "localhost" || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())
}

func isLoopbackRemote(raw string) bool {
	host := strings.TrimSpace(raw)
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func sameLoopbackOrigin(rawOrigin, requestHost string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawOrigin))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || !isLoopbackHost(parsed.Host) {
		return false
	}
	return strings.EqualFold(parsed.Host, requestHost)
}

func validSHA256Header(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func boundedHeader(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxLength || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	return value
}
