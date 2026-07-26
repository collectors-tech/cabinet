package agentcontext

import "strings"

type SelectedRecord struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type Envelope struct {
	ProfileID       string         `json:"profile_id,omitempty"`
	WorkspaceID     string         `json:"workspace_id,omitempty"`
	RouteID         string         `json:"route_id,omitempty"`
	SurfaceID       string         `json:"surface_id,omitempty"`
	SelectedRecord  SelectedRecord `json:"selected_record,omitempty"`
	ThreadID        string         `json:"thread_id,omitempty"`
	IntentText      string         `json:"intent_text,omitempty"`
	AttachmentIDs   []string       `json:"attachment_ids,omitempty"`
	MediaIDs        []string       `json:"media_ids,omitempty"`
	SourceChannel   string         `json:"source_channel,omitempty"`
	PermissionState string         `json:"permission_state,omitempty"`
	SetupState      string         `json:"setup_state,omitempty"`
	WorkflowRunID   string         `json:"workflow_run_id,omitempty"`
	AuditID         string         `json:"audit_id,omitempty"`
}

type NormalizeInput struct {
	ProfileID     string
	ThreadID      string
	IntentText    string
	Context       map[string]any
	AttachmentIDs []string
}

func Normalize(input NormalizeInput) Envelope {
	ctx := input.Context
	envelope := Envelope{
		ProfileID:     firstNonBlank(input.ProfileID, nestedString(ctx, "profile", "id"), stringValue(ctx["profile_id"])),
		WorkspaceID:   firstNonBlank(nestedString(ctx, "workspace", "id"), stringValue(ctx["workspace_id"])),
		ThreadID:      firstNonBlank(input.ThreadID, stringValue(ctx["thread_id"])),
		IntentText:    strings.TrimSpace(input.IntentText),
		AttachmentIDs: cleanStrings(input.AttachmentIDs),
	}
	if envelope.AttachmentIDs == nil {
		envelope.AttachmentIDs = cleanAnyStrings(ctx["attachment_ids"])
	}

	routePath := nestedString(ctx, "route", "pathname")
	routeID := firstNonBlank(stringValue(ctx["route_id"]), routePath)
	envelope.RouteID = routeID
	envelope.SurfaceID = firstNonBlank(stringValue(ctx["surface_id"]), nestedString(ctx, "assistant", "surface_id"), surfaceFromRoute(routePath), stringValue(ctx["source_surface"]))
	envelope.SourceChannel = firstNonBlank(stringValue(ctx["source_channel"]), nestedString(ctx, "assistant", "source_channel"), "in-app")
	envelope.PermissionState = firstNonBlank(stringValue(ctx["permission_state"]), nestedString(ctx, "permissions", "state"), nestedString(ctx, "assistant", "permission_state"))
	envelope.SetupState = firstNonBlank(stringValue(ctx["setup_state"]), nestedString(ctx, "setup", "state"), nestedString(ctx, "assistant", "setup_state"))
	envelope.WorkflowRunID = firstNonBlank(stringValue(ctx["workflow_run_id"]), nestedString(ctx, "workflow", "run_id"))
	envelope.AuditID = firstNonBlank(stringValue(ctx["audit_id"]), nestedString(ctx, "audit", "id"))
	envelope.MediaIDs = cleanAnyStrings(ctx["media_ids"])

	selection := mapValue(ctx["selection"])
	envelope.SelectedRecord = SelectedRecord{
		Type: firstNonBlank(stringValue(selection["record_type"]), stringValue(selection["type"]), stringValue(selection["selected_record_type"])),
		ID:   firstNonBlank(stringValue(selection["record_id"]), stringValue(selection["id"]), stringValue(selection["selected_record_id"])),
	}
	return envelope
}

func (e Envelope) Map() map[string]any {
	out := map[string]any{}
	putString(out, "profile_id", e.ProfileID)
	putString(out, "workspace_id", e.WorkspaceID)
	putString(out, "route_id", e.RouteID)
	putString(out, "surface_id", e.SurfaceID)
	if e.SelectedRecord.Type != "" || e.SelectedRecord.ID != "" {
		out["selected_record"] = map[string]any{}
		selected := out["selected_record"].(map[string]any)
		putString(selected, "type", e.SelectedRecord.Type)
		putString(selected, "id", e.SelectedRecord.ID)
	}
	putString(out, "thread_id", e.ThreadID)
	putString(out, "intent_text", e.IntentText)
	putStrings(out, "attachment_ids", e.AttachmentIDs)
	putStrings(out, "media_ids", e.MediaIDs)
	putString(out, "source_channel", e.SourceChannel)
	putString(out, "permission_state", e.PermissionState)
	putString(out, "setup_state", e.SetupState)
	putString(out, "workflow_run_id", e.WorkflowRunID)
	putString(out, "audit_id", e.AuditID)
	return out
}

func WithEnvelope(messageContext map[string]any, input NormalizeInput) map[string]any {
	out := map[string]any{}
	for key, value := range messageContext {
		out[key] = value
	}
	out["agent_context"] = Normalize(input).Map()
	return out
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func mapValue(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func nestedString(root map[string]any, path ...string) string {
	current := root
	for i, key := range path {
		value, ok := current[key]
		if !ok {
			return ""
		}
		if i == len(path)-1 {
			return stringValue(value)
		}
		current = mapValue(value)
	}
	return ""
}

func stringValue(value any) string {
	if typed, ok := value.(string); ok {
		return strings.TrimSpace(typed)
	}
	return ""
}

func surfaceFromRoute(routePath string) string {
	routePath = strings.TrimSpace(routePath)
	if routePath == "" {
		return ""
	}
	if strings.HasPrefix(routePath, "/chats") {
		return "chats.main"
	}
	return strings.Trim(routePath, "/")
}

func cleanStrings(values []string) []string {
	var out []string
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func cleanAnyStrings(value any) []string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range values {
		if trimmed := stringValue(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func putString(out map[string]any, key, value string) {
	if value != "" {
		out[key] = value
	}
}

func putStrings(out map[string]any, key string, values []string) {
	if len(values) > 0 {
		out[key] = append([]string(nil), values...)
	}
}
