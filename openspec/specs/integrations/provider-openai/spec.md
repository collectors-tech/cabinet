## Purpose
Define OpenAI provider adapter behavior for ChatGPT-compatible tasks through the AI gateway.

## Requirements
### Requirement PROVIDER-OPENAI-001: OpenAI adapter SHALL implement gateway canonical contract
Cabinet SHALL map OpenAI request/response semantics to gateway canonical envelope without leaking provider-specific structure to callers.

#### Scenario: Structured extraction response
- **GIVEN** gateway routes request to OpenAI adapter with structured extraction task
- **WHEN** adapter receives provider response
- **THEN** adapter MUST normalize output into `result` object with deterministic keys and include `provider="openai"` metadata

### Requirement PROVIDER-OPENAI-002: OpenAI adapter SHALL support multimodal image-assisted suggestions
Cabinet SHALL support image-assisted suggestion flows for inventory media analysis through OpenAI-capable models when enabled.

#### Scenario: Photo suggestion task
- **GIVEN** request includes `media_refs` and task type `inventory_photo_suggest`
- **WHEN** adapter calls configured OpenAI model
- **THEN** normalized result MUST include suggested `title`, `brand`, `part_number`, and `confidence`

### Requirement PROVIDER-OPENAI-003: OpenAI adapter SHALL classify and map provider errors deterministically
Cabinet SHALL normalize OpenAI upstream failures into gateway error taxonomy.

#### Scenario: Upstream rate limit
- **GIVEN** OpenAI returns rate limit error
- **WHEN** adapter maps provider error
- **THEN** gateway error response MUST use stable internal code (for example `AI_PROVIDER_RATE_LIMIT`) with retry guidance and no secret leakage
