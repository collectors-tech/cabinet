## Purpose
Define Gemini provider adapter behavior through the AI gateway for text and multimodal analysis.

## Requirements
### Requirement PROVIDER-GEMINI-001: Gemini adapter SHALL implement gateway canonical contract
Cabinet SHALL normalize Gemini responses into the same canonical gateway envelope used by all providers.

#### Scenario: Canonical Gemini response mapping
- **GIVEN** request is routed to Gemini adapter
- **WHEN** provider response is received
- **THEN** adapter MUST return normalized `result`, `confidence`, and gateway `meta` with `provider="gemini"`

### Requirement PROVIDER-GEMINI-002: Gemini adapter SHALL support media analysis tasks for inventory workflows
Cabinet SHALL allow Gemini-backed media analysis when routing policy selects Gemini for multimodal tasks.

#### Scenario: Inventory asset classification
- **GIVEN** task type `inventory_asset_classify` with image reference set
- **WHEN** Gemini adapter processes the request
- **THEN** normalized output MUST include candidate labels/tags, confidence, and confirmation flag for low-confidence outcomes

### Requirement PROVIDER-GEMINI-003: Gemini adapter SHALL comply with gateway retry/fallback policy
Cabinet SHALL surface retriable error classes so gateway fallback policy can execute predictably.

#### Scenario: Retriable transient failure
- **GIVEN** Gemini returns transient upstream failure
- **WHEN** adapter maps error for gateway
- **THEN** response metadata MUST mark failure retriable and preserve fallback eligibility for next provider
