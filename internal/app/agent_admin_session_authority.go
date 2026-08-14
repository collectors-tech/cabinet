package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/auth"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/profile"
)

type serverSessionPrincipalContextKey struct{}

type serverSessionPrincipal struct {
	IdentityMode string
	ProfileID    string
	Subject      string
	Email        string
	Roles        []string
}

type agentAdminAuthority struct {
	ProfileID    string
	UserID       string
	Role         string
	IdentityMode string
}

type agentAdminAuthorityError struct {
	code   string
	status int
}

func (err *agentAdminAuthorityError) Error() string { return err.code }

func withServerSessionPrincipal(ctx context.Context, principal serverSessionPrincipal) context.Context {
	return context.WithValue(ctx, serverSessionPrincipalContextKey{}, principal)
}

func serverSessionPrincipalFromContext(ctx context.Context) (serverSessionPrincipal, bool) {
	principal, ok := ctx.Value(serverSessionPrincipalContextKey{}).(serverSessionPrincipal)
	return principal, ok
}

func attachLocalServerSessionPrincipal(r *http.Request, authService *auth.Service) *http.Request {
	if r == nil || authService == nil {
		return r
	}
	token := sessionTokenFromRequest(r)
	if token == "" {
		return r
	}
	profileID, err := authService.ValidateUnlockedSessionProfile(token)
	if err != nil || profileID == "" {
		return r
	}
	ctx := withServerSessionPrincipal(r.Context(), serverSessionPrincipal{
		IdentityMode: "local",
		ProfileID:    profileID,
		Roles:        []string{"local-owner"},
	})
	return r.WithContext(ctx)
}

func attachZitadelServerSessionPrincipal(r *http.Request, session zitadelSession) *http.Request {
	if r == nil {
		return r
	}
	ctx := withServerSessionPrincipal(r.Context(), serverSessionPrincipal{
		IdentityMode: "zitadel",
		Subject:      strings.TrimSpace(session.Identity.Subject),
		Email:        strings.ToLower(strings.TrimSpace(session.Identity.Email)),
		Roles:        append([]string(nil), session.Identity.Roles...),
	})
	return r.WithContext(ctx)
}

func resolveAgentAdminAuthority(ctx context.Context, conn *sql.DB, profileID string) (agentAdminAuthority, error) {
	principal, ok := serverSessionPrincipalFromContext(ctx)
	if !ok {
		return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_authentication_required", status: http.StatusUnauthorized}
	}
	profileID = strings.TrimSpace(profileID)
	if profileID == "" || conn == nil {
		return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_profile_forbidden", status: http.StatusForbidden}
	}
	active, err := profile.NewRepository(conn).GetActiveProfile(ctx)
	if err != nil || strings.TrimSpace(active.ID) != profileID {
		return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_profile_forbidden", status: http.StatusForbidden}
	}
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_membership_unavailable", status: http.StatusForbidden}
	}

	switch principal.IdentityMode {
	case "local":
		if strings.TrimSpace(principal.ProfileID) != profileID {
			return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_profile_forbidden", status: http.StatusForbidden}
		}
		for _, user := range users {
			if isProtectedLocalOwner(user) && strings.EqualFold(user.Status, "active") {
				return agentAdminAuthority{ProfileID: profileID, UserID: user.ID, Role: "admin", IdentityMode: "local"}, nil
			}
		}
	case "zitadel":
		if !serverPrincipalHasRole(principal, "cabinet.admin") {
			return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_role_required", status: http.StatusForbidden}
		}
		for _, user := range users {
			if strings.EqualFold(strings.TrimSpace(user.Email), principal.Email) &&
				strings.EqualFold(user.Status, "active") && strings.EqualFold(user.Role, "admin") {
				return agentAdminAuthority{ProfileID: profileID, UserID: user.ID, Role: "admin", IdentityMode: "zitadel"}, nil
			}
		}
	}
	return agentAdminAuthority{}, &agentAdminAuthorityError{code: "agent_admin_membership_required", status: http.StatusForbidden}
}

func serverPrincipalHasRole(principal serverSessionPrincipal, required string) bool {
	for _, role := range principal.Roles {
		if strings.EqualFold(strings.TrimSpace(role), strings.TrimSpace(required)) {
			return true
		}
	}
	return false
}

func authorizeAgentUsersRequest(ctx context.Context, conn *sql.DB, req agentskills.PreviewRequest) (agentskills.PreviewRequest, agentAdminAuthority, error) {
	if !strings.HasPrefix(strings.TrimSpace(req.SkillID), "cabinet.users.") {
		return req, agentAdminAuthority{}, nil
	}
	authority, err := resolveAgentAdminAuthority(ctx, conn, req.ProfileID)
	if err != nil {
		return req, agentAdminAuthority{}, err
	}
	req = stripClientAssertedAdminAuthority(req)
	req, err = hydrateAgentUsersTargetFromServer(ctx, conn, req)
	return req, authority, err
}

func hydrateAgentUsersTargetFromServer(ctx context.Context, conn *sql.DB, req agentskills.PreviewRequest) (agentskills.PreviewRequest, error) {
	switch strings.TrimSpace(req.SkillID) {
	case "cabinet.users.remove_user", "cabinet.users.update_role", "cabinet.users.activate_or_deactivate":
	default:
		return req, nil
	}
	targetID, blocker, err := resolveAgentSkillUserTarget(ctx, conn, req.ProfileID, req.Parameters)
	if err != nil {
		return req, &agentSkillPreviewLifecycleError{Code: blocker, Recoverable: true, NextAction: "Select a current user from this profile and create a fresh preview."}
	}
	users, err := listRuntimeUsers(ctx, conn, req.ProfileID)
	if err != nil {
		return req, err
	}
	for _, user := range users {
		if user.ID != targetID {
			continue
		}
		parameters := make(map[string]any, len(req.Parameters)+7)
		for key, value := range req.Parameters {
			parameters[key] = value
		}
		parameters["target_user"] = user.ID
		parameters["target_email"] = user.Email
		parameters["target_display_name"] = strings.TrimSpace(user.FirstName + " " + user.LastName)
		parameters["target_role_current"] = user.Role
		parameters["target_status_current"] = user.Status
		parameters["target_updated_at"] = user.UpdatedAt
		parameters["protected"] = isProtectedLocalOwner(user) || protectedUserChangeLeavesNoAdmin(users, user.ID, "", "")
		req.Parameters = parameters
		return req, nil
	}
	return req, &agentSkillPreviewLifecycleError{Code: "users_admin_target_not_found", Recoverable: true, NextAction: "Select a current user from this profile and create a fresh preview."}
}

func authorizeDurableAgentUsersPreview(ctx context.Context, conn *sql.DB, profileID, previewID string) (agentAdminAuthority, error) {
	record, err := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
	if err != nil || !strings.HasPrefix(strings.TrimSpace(record.SkillID), "cabinet.users.") {
		return agentAdminAuthority{}, err
	}
	return resolveAgentAdminAuthority(ctx, conn, profileID)
}

func stripClientAssertedAdminAuthority(req agentskills.PreviewRequest) agentskills.PreviewRequest {
	req.Parameters = withoutClientAssertedAdminAuthority(req.Parameters)
	req.AgentContext = withoutClientAssertedAdminAuthority(req.AgentContext)
	return req
}

func withoutClientAssertedAdminAuthority(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		if isClientAssertedAdminAuthorityKey(key) {
			continue
		}
		out[key] = withoutClientAssertedAdminAuthorityValue(value)
	}
	return out
}

func isClientAssertedAdminAuthorityKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "admin_session", "admin", "role", "roles", "permission", "permissions", "permission_state", "authority", "owner":
		return true
	default:
		return false
	}
}

func withoutClientAssertedAdminAuthorityValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return withoutClientAssertedAdminAuthority(typed)
	case []map[string]any:
		out := make([]map[string]any, 0, len(typed))
		for _, nested := range typed {
			out = append(out, withoutClientAssertedAdminAuthority(nested))
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, nested := range typed {
			out = append(out, withoutClientAssertedAdminAuthorityValue(nested))
		}
		return out
	default:
		return value
	}
}

func writeAgentAdminAuthorityError(w http.ResponseWriter, err error) {
	var authorityErr *agentAdminAuthorityError
	if !errors.As(err, &authorityErr) {
		authorityErr = &agentAdminAuthorityError{code: "agent_admin_authority_unavailable", status: http.StatusForbidden}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(authorityErr.status)
	_, _ = w.Write([]byte(`{"error":"` + authorityErr.code + `"}`))
}

func agentAdminAuthorityEvidence(ctx context.Context) map[string]any {
	authority, ok := ctx.Value(agentAdminAuthorityContextKey{}).(agentAdminAuthority)
	if !ok || authority.UserID == "" {
		return nil
	}
	return map[string]any{
		"source":        "server_session",
		"identity_mode": authority.IdentityMode,
		"actor_user_id": authority.UserID,
		"actor_role":    authority.Role,
	}
}

type agentAdminAuthorityContextKey struct{}

func withAgentAdminAuthority(ctx context.Context, authority agentAdminAuthority) context.Context {
	return context.WithValue(ctx, agentAdminAuthorityContextKey{}, authority)
}

func filterUnauthorizedAgentAdminWorkflowRuns(ctx context.Context, conn *sql.DB, profileID string, runs []chat.WorkflowRun) []chat.WorkflowRun {
	if _, err := resolveAgentAdminAuthority(ctx, conn, profileID); err == nil {
		return runs
	}
	filtered := make([]chat.WorkflowRun, 0, len(runs))
	for _, run := range runs {
		if strings.TrimSpace(run.WorkflowID) == "chat.agent_capability_explanation" {
			filtered = append(filtered, run)
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(run.CapabilityID), "cabinet.users.") ||
			containsAgentUsersEvidence(run.Input) || containsAgentUsersEvidence(run.Result) || containsAgentUsersEvidence(run.BulkItems) {
			continue
		}
		filtered = append(filtered, run)
	}
	return filtered
}

func filterUnauthorizedAgentAdminMessages(ctx context.Context, conn *sql.DB, profileID string, messages []chat.Message) []chat.Message {
	if _, err := resolveAgentAdminAuthority(ctx, conn, profileID); err == nil {
		return messages
	}
	filtered := make([]chat.Message, 0, len(messages))
	for _, message := range messages {
		if messageContainsAgentUsersEvidence(message.Context) {
			continue
		}
		filtered = append(filtered, message)
	}
	return filtered
}

func workflowRunRequiresAgentAdminAuthority(run chat.WorkflowRun) bool {
	return strings.HasPrefix(strings.TrimSpace(run.CapabilityID), "cabinet.users.") ||
		containsAgentUsersEvidence(run.Input) || containsAgentUsersEvidence(run.ProviderTrace) ||
		containsAgentUsersEvidence(run.Result) || containsAgentUsersEvidence(run.Error) || containsAgentUsersEvidence(run.BulkItems)
}

func workflowRunInputRequiresAgentAdminAuthority(input chat.CreateWorkflowRunInput) bool {
	return strings.HasPrefix(strings.TrimSpace(input.CapabilityID), "cabinet.users.") ||
		containsAgentUsersEvidence(input.Input) || containsAgentUsersEvidence(input.BulkItems)
}

func sanitizeAgentAdminWorkflowInput(input chat.CreateWorkflowRunInput, authority agentAdminAuthority) chat.CreateWorkflowRunInput {
	input.Input = withoutClientAssertedAdminAuthority(input.Input)
	input.ProviderTrace = withoutClientAssertedAdminAuthority(input.ProviderTrace)
	for index, item := range input.BulkItems {
		input.BulkItems[index] = withoutClientAssertedAdminAuthority(item)
	}
	if input.Input == nil {
		input.Input = make(map[string]any)
	}
	input.Input["admin_authority"] = map[string]any{
		"source":        "server_session",
		"identity_mode": authority.IdentityMode,
		"actor_user_id": authority.UserID,
		"actor_role":    authority.Role,
	}
	return input
}

func workflowRunUpdateRequiresAgentAdminAuthority(input chat.UpdateWorkflowRunInput) bool {
	return containsAgentUsersEvidence(input.ProviderTrace) || containsAgentUsersEvidence(input.Result) ||
		containsAgentUsersEvidence(input.Error) || containsAgentUsersEvidence(input.BulkItems)
}

func sanitizeAgentAdminWorkflowUpdate(input chat.UpdateWorkflowRunInput, authority agentAdminAuthority) chat.UpdateWorkflowRunInput {
	input.ProviderTrace = withoutClientAssertedAdminAuthority(input.ProviderTrace)
	input.Result = withoutClientAssertedAdminAuthority(input.Result)
	input.Error = withoutClientAssertedAdminAuthority(input.Error)
	for index, item := range input.BulkItems {
		input.BulkItems[index] = withoutClientAssertedAdminAuthority(item)
	}
	if input.Result == nil {
		input.Result = make(map[string]any)
	}
	input.Result["admin_authority"] = map[string]any{
		"source":        "server_session",
		"identity_mode": authority.IdentityMode,
		"actor_user_id": authority.UserID,
		"actor_role":    authority.Role,
	}
	return input
}

func messageContainsAgentUsersEvidence(messageContext map[string]any) bool {
	for key, nested := range messageContext {
		if strings.EqualFold(strings.TrimSpace(key), "agent_capabilities") {
			continue
		}
		if containsAgentUsersEvidence(nested) {
			return true
		}
	}
	return false
}

func containsAgentUsersEvidence(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			normalizedKey := strings.ToLower(strings.TrimSpace(key))
			if normalizedKey == "skill_id" || normalizedKey == "selected_skill_id" || normalizedKey == "capability_id" {
				if skillID, ok := nested.(string); ok && strings.HasPrefix(strings.TrimSpace(skillID), "cabinet.users.") {
					return true
				}
			}
			if containsAgentUsersEvidence(nested) {
				return true
			}
		}
	case []map[string]any:
		for _, nested := range typed {
			if containsAgentUsersEvidence(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsAgentUsersEvidence(nested) {
				return true
			}
		}
	}
	return false
}
