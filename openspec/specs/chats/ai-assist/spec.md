## Purpose
Define AI assist behavior, controls, and mutation safety model.

## Requirements
### Requirement AI-ASSIST-001: AI assist SHALL be OpenAI-backed in v1 with user-provided API key
Cabinet SHALL require profile-level OpenAI credentials for AI operations.

#### Scenario: AI operation without key
- **GIVEN** profile `p1` is active, setting `ai_enabled=true`, and secret `openai_api_key` is missing
- **WHEN** client calls `POST /api/ai/test` with `profile_id=p1`
- **THEN** runtime MUST return `400` with an actionable configuration error (`error: "ai_unavailable"` or equivalent)

### Requirement AI-ASSIST-002: AI suggestions SHALL include confidence and review state
Cabinet SHALL provide suggestion output with confidence signals.

#### Scenario: Suggest from title
- **GIVEN** profile `p1` has `ai_enabled=true`, valid OpenAI base URL, valid API key secret, and license feature `ai_assist`
- **WHEN** client calls `POST /api/ai/suggest/title` with `profile_id` and a raw listing title
- **THEN** runtime MUST return `200` and payload MUST include normalized suggestion fields and `confidence` score

### Requirement AI-ASSIST-003: AI mutations SHALL require explicit confirmation
Cabinet SHALL not auto-create or auto-update collection data from AI output.

#### Scenario: Apply AI suggestion
- **GIVEN** AI suggestion `sug_123` exists for item draft `draft_42` in active profile `p1` and `confirm_token` has not been issued
- **WHEN** user attempts to apply a create/update mutation from AI output
- **THEN** runtime MUST reject mutation with `409` and payload field `error_code="AI_CONFIRM_REQUIRED"` until explicit confirm-before-apply action is provided

### Requirement AI-ASSIST-004: AI control toggle SHALL be profile-scoped
Cabinet SHALL support per-profile enable/disable state for AI features.

#### Scenario: Disable AI in profile
- **GIVEN** profile `p1` has persisted setting `ai_enabled=false`
- **WHEN** client calls `POST /api/ai/test` or `POST /api/ai/suggest/title` for `p1`
- **THEN** runtime MUST reject request with `400` and MUST keep AI unavailable until `ai_enabled=true`
