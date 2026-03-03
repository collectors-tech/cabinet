package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/google/uuid"
)

const usersStateKeyPrefix = "users.profile."

type runtimeUser struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Status      string `json:"status"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func registerUsersRoutes(mux *http.ServeMux, conn *sql.DB, profiles *profile.Repository) {
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			profileID, err := resolveUsersProfileID(r.Context(), profiles)
			if err != nil {
				http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusNotFound)
				return
			}
			users, err := listRuntimeUsers(r.Context(), conn, profileID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_users"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"users": users})
		case http.MethodPost:
			profileID, err := resolveUsersProfileID(r.Context(), profiles)
			if err != nil {
				http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusNotFound)
				return
			}
			var req struct {
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				Username    string `json:"username"`
				Email       string `json:"email"`
				PhoneNumber string `json:"phoneNumber"`
				Role        string `json:"role"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := createRuntimeUser(r.Context(), conn, profileID, req.FirstName, req.LastName, req.Username, req.Email, req.PhoneNumber, req.Role)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/invite", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		profileID, err := resolveUsersProfileID(r.Context(), profiles)
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusNotFound)
			return
		}
		var req struct {
			Email string `json:"email"`
			Role  string `json:"role"`
			Desc  string `json:"desc"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		invited, err := inviteRuntimeUser(r.Context(), conn, profileID, req.Email, req.Role)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(invited)
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/users/"))
		if id == "" || strings.ContainsRune(id, '/') {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		profileID, err := resolveUsersProfileID(r.Context(), profiles)
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodPut:
			var req struct {
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				Username    string `json:"username"`
				Email       string `json:"email"`
				PhoneNumber string `json:"phoneNumber"`
				Status      string `json:"status"`
				Role        string `json:"role"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			updated, err := updateRuntimeUser(r.Context(), conn, profileID, id, req.FirstName, req.LastName, req.Username, req.Email, req.PhoneNumber, req.Status, req.Role)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(updated)
		case http.MethodDelete:
			if err := deleteRuntimeUser(r.Context(), conn, profileID, id); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}

func resolveUsersProfileID(ctx context.Context, profiles *profile.Repository) (string, error) {
	active, err := profiles.GetActiveProfile(ctx)
	if err != nil {
		return "local-default", nil
	}
	id := strings.TrimSpace(active.ID)
	if id == "" {
		return "local-default", nil
	}
	return id, nil
}

func listRuntimeUsers(ctx context.Context, conn *sql.DB, profileID string) ([]runtimeUser, error) {
	key := usersStateKey(profileID)
	var raw string
	err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, key).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows || strings.TrimSpace(raw) == "" {
		seed := []runtimeUser{defaultOwnerUser(profileID)}
		if err := putRuntimeUsers(ctx, conn, profileID, seed); err != nil {
			return nil, err
		}
		return seed, nil
	}
	var users []runtimeUser
	if decodeErr := json.Unmarshal([]byte(raw), &users); decodeErr != nil {
		return nil, decodeErr
	}
	if users == nil {
		users = make([]runtimeUser, 0)
	}
	return users, nil
}

func putRuntimeUsers(ctx context.Context, conn *sql.DB, profileID string, users []runtimeUser) error {
	payload, err := json.Marshal(users)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
	`, usersStateKey(profileID), string(payload))
	return err
}

func createRuntimeUser(ctx context.Context, conn *sql.DB, profileID, firstName, lastName, username, email, phoneNumber, role string) (runtimeUser, error) {
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return runtimeUser{}, err
	}
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	phoneNumber = strings.TrimSpace(phoneNumber)
	role = normalizeUserRole(role)
	if firstName == "" || lastName == "" || username == "" || email == "" || phoneNumber == "" {
		return runtimeUser{}, fmt.Errorf("invalid_user")
	}
	for _, existing := range users {
		if strings.EqualFold(existing.Username, username) || strings.EqualFold(existing.Email, email) {
			return runtimeUser{}, fmt.Errorf("user_already_exists")
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	created := runtimeUser{
		ID:          uuid.NewString(),
		FirstName:   firstName,
		LastName:    lastName,
		Username:    username,
		Email:       email,
		PhoneNumber: phoneNumber,
		Status:      "active",
		Role:        role,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	users = append(users, created)
	if err := putRuntimeUsers(ctx, conn, profileID, users); err != nil {
		return runtimeUser{}, err
	}
	return created, nil
}

func inviteRuntimeUser(ctx context.Context, conn *sql.DB, profileID, email, role string) (runtimeUser, error) {
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return runtimeUser{}, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	role = normalizeUserRole(role)
	if email == "" || !strings.Contains(email, "@") {
		return runtimeUser{}, fmt.Errorf("invalid_invite")
	}
	username := strings.TrimSpace(strings.SplitN(email, "@", 2)[0])
	if username == "" {
		return runtimeUser{}, fmt.Errorf("invalid_invite")
	}
	for _, existing := range users {
		if strings.EqualFold(existing.Email, email) {
			return runtimeUser{}, fmt.Errorf("user_already_exists")
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	invite := runtimeUser{
		ID:          uuid.NewString(),
		FirstName:   "Invited",
		LastName:    "User",
		Username:    username,
		Email:       email,
		PhoneNumber: "",
		Status:      "invited",
		Role:        role,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	users = append(users, invite)
	if err := putRuntimeUsers(ctx, conn, profileID, users); err != nil {
		return runtimeUser{}, err
	}
	return invite, nil
}

func updateRuntimeUser(ctx context.Context, conn *sql.DB, profileID, id, firstName, lastName, username, email, phoneNumber, status, role string) (runtimeUser, error) {
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return runtimeUser{}, err
	}
	for i := range users {
		if users[i].ID != id {
			continue
		}
		if trimmed := strings.TrimSpace(firstName); trimmed != "" {
			users[i].FirstName = trimmed
		}
		if trimmed := strings.TrimSpace(lastName); trimmed != "" {
			users[i].LastName = trimmed
		}
		if trimmed := strings.TrimSpace(username); trimmed != "" {
			users[i].Username = trimmed
		}
		if trimmed := strings.ToLower(strings.TrimSpace(email)); trimmed != "" {
			users[i].Email = trimmed
		}
		if strings.TrimSpace(phoneNumber) != "" {
			users[i].PhoneNumber = strings.TrimSpace(phoneNumber)
		}
		if normalizedStatus := normalizeUserStatus(status); normalizedStatus != "" {
			users[i].Status = normalizedStatus
		}
		users[i].Role = normalizeUserRole(role)
		users[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		if err := putRuntimeUsers(ctx, conn, profileID, users); err != nil {
			return runtimeUser{}, err
		}
		return users[i], nil
	}
	return runtimeUser{}, fmt.Errorf("user_not_found")
}

func deleteRuntimeUser(ctx context.Context, conn *sql.DB, profileID, id string) error {
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return err
	}
	next := make([]runtimeUser, 0, len(users))
	removed := false
	for _, user := range users {
		if user.ID == id {
			removed = true
			continue
		}
		next = append(next, user)
	}
	if !removed {
		return fmt.Errorf("user_not_found")
	}
	return putRuntimeUsers(ctx, conn, profileID, next)
}

func usersStateKey(profileID string) string {
	return usersStateKeyPrefix + strings.TrimSpace(profileID)
}

func normalizeUserRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "view":
		return "view"
	default:
		return "view"
	}
}

func normalizeUserStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "inactive", "invited", "suspended":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return ""
	}
}

func defaultOwnerUser(profileID string) runtimeUser {
	now := time.Now().UTC().Format(time.RFC3339)
	suffix := strings.ReplaceAll(strings.TrimSpace(profileID), "-", "")
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	if suffix == "" {
		suffix = "local"
	}
	return runtimeUser{
		ID:          uuid.NewString(),
		FirstName:   "Local",
		LastName:    "Admin",
		Username:    "owner_" + strings.ToLower(suffix),
		Email:       "owner+" + strings.ToLower(suffix) + "@cabinet.local",
		PhoneNumber: "",
		Status:      "active",
		Role:        "admin",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
