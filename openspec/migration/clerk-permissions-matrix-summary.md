# Clerk Permissions Matrix Summary (#238)

## Issue
- #238 `[Execution] Clerk integration + multi-account permissions test matrix`

## Spec IDs
- AUTH-PERM-001
- AUTH-PERM-002
- AUTH-PERM-003
- AUTH-PERM-004
- AUTH-PERM-005

## Implemented
- Added explicit plan->capability matrix behavior in runtime entitlement resolution:
  - `mvp` -> `collection_core`
  - `creator` -> `collection_core, ai_assist, scanner_automation`
  - `teams` -> `collection_core, ai_assist, price_tracking, scanner_automation`
- Updated cloud-feature gate path to evaluate feature capability membership (`cloudPlanHasFeature`) instead of `plan == pro` check.
- Added diagnostics endpoint `/api/auth/cloud/session/effective` returning deterministic effective permissions payload:
  - `provider`, `user_id`, `email`, `role`, `plan`, `features`
- Persisted cloud session context (`cloud.email`, `cloud.role`) during bootstrap.

## Test evidence
- `TestAuthPermissionsPlanCapabilityMatrix`
- `TestAuthPermissionsFeatureGateMatrixFromCloudPlan`
- `TestAuthPermissionsEffectiveDiagnosticsContract`
- Existing Clerk identity mode coverage retained in `auth_provider_options_api_test.go`.

## Commands run
1. `go test ./internal/app -run TestAuthPermissions -count=1` (fail-first then pass)
2. `go test ./internal/app -count=1`
3. `go test ./tests -count=1`
4. `openspec validate --all`
5. `pwsh -File .\\scripts\\build-cabinet.ps1`

## Result
- All gates passed.
