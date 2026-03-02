## Purpose
Define a provider-agnostic AI gateway contract for Cabinet inference and assistant workflows.

## Requirements
### Requirement AI-GATEWAY-001: Gateway SHALL provide a canonical request/response envelope
Cabinet SHALL route AI requests through a single gateway contract so UI and services do not depend on provider-specific payload shapes.

#### Scenario: Canonical completion request
- **GIVEN** caller submits `POST /api/ai/gateway/complete` with `task_type`, `input`, optional `media_refs`, and optional `provider_hint`
- **WHEN** gateway processes request
- **THEN** response MUST return `200` with canonical envelope containing `status`, `provider`, `model`, `result`, `confidence`, `requires_confirmation`, and `meta`

### Requirement AI-GATEWAY-002: Gateway SHALL support deterministic provider routing and fallback
Cabinet SHALL support policy-driven provider selection and fallback without changing caller contract.

#### Scenario: Fallback route on provider failure
- **GIVEN** routing policy maps task to preferred provider and fallback order
- **WHEN** preferred provider fails with timeout or retriable upstream error
- **THEN** gateway MUST attempt next provider in policy chain and return final provider metadata in response envelope

### Requirement AI-GATEWAY-003: Gateway SHALL enforce auditable inference logging
Cabinet SHALL capture request/response metadata for diagnostics and governance.

#### Scenario: Audit log persistence
- **GIVEN** gateway request completes (success or failure)
- **WHEN** audit logging executes
- **THEN** runtime MUST persist `provider`, `model`, latency, token/usage metrics, decision flags, and error taxonomy without storing plaintext secrets

### Requirement AI-GATEWAY-004: Gateway SHALL enforce confirmation for low-confidence actions
Cabinet SHALL require explicit user confirmation before applying low-confidence or high-risk AI actions.

#### Scenario: Low-confidence classification
- **GIVEN** gateway result has `confidence < configured_threshold` or `requires_confirmation=true`
- **WHEN** caller requests auto-apply action
- **THEN** runtime MUST reject auto-apply with actionable confirmation response contract and SHALL NOT mutate inventory/wishlist state
