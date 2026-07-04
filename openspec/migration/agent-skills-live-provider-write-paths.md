# Agent Skills Live Provider Write-Path Evidence

Issue: #1780

## Eligible Surfaces

- `cabinet.integrations.test_connection`: provider-health-backed readiness check. It is non-mutating and may run with live or stubbed provider health evidence.
- `cabinet.integrations.configure_provider`: confirmed provider setup metadata persistence. It may persist non-secret provider settings and write-only provider secrets, then instruct the operator to run provider health validation.
- `cabinet.integrations.repair_provider`: confirmed setup review/repair metadata persistence. It does not claim an external provider write.
- `cabinet.integrations.disable_provider`: confirmed profile-scoped disable metadata persistence. It does not claim an external provider write.
- Settings/Data companion skills remain covered by the shared apply path in `internal/app/agent_skills_api_test.go`; provider-live proof for this slice is limited to Integrations provider setup/readiness flows.

## Preconditions

- Live-provider validation requires provider-specific credentials, a selected profile, and a provider health record from the matching provider runtime.
- Stubbed validation is acceptable when live credentials are unavailable, provided the stub records the same readiness taxonomy fields and uses non-secret fixture values.
- Provider secrets must be accepted only as write-only inputs. They must not appear in API responses, logs, screenshots, chat text, or Action Timeline/user-facing result text.

## Evidence Added

- `TestAgentSkillApplyAPICapturesStubbedProviderWritePathEvidence` seeds provider health for `openai`, exercises test/configure/repair/disable through `/api/agent/skills/apply`, verifies provider-not-ready remediation copy, verifies profile settings and secret persistence, and asserts response bodies omit the fixture secret.

## Traceability

- Requirement: `AGENT-SKILLS-REGISTRY-003`
- Issue binding: #1780
- Test target: `internal/app/agent_skills_api_test.go` (`TestAgentSkillApplyAPICapturesStubbedProviderWritePathEvidence`)
