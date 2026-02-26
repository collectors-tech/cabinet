# OpenSpec Migration TODO (Strict, Testable)

Last updated: 2026-02-26  
Tracking issue: `#171`

## Objective
Complete migration of spec-like docs into OpenSpec with testable use-cases for every UI feature.

## Non-Negotiable Rule
Each UI feature use-case in OpenSpec MUST be testable:
- Unique use-case ID (`UC-<screen>-<nn>`)
- Requirement reference
- Expected state/result
- Planned automated test mapping (`Cypress` spec path and test ID)

---

## Phase 0: Governance and Enforcement
- [x] Add CI gate: `openspec validate --all --strict --no-interactive` required on PR.
- [x] Add pre-commit hook to validate changed OpenSpec artifacts.
- [x] Add pre-push hook to run OpenSpec validation + relevant tests.
- [x] Add commit-msg hook to enforce `#<issue>` prefix.
- [x] Add policy note in `AGENTS.md` linking OpenSpec migration catalog and TODO.

## Phase 1: Catalog-to-Spec Gap Closure
- [x] Create `openspec/specs/ui-foundation-components/spec.md` from `docs/ui-spec/09-COMPONENT-SPECS-STRICT.md`.
- [x] Create `openspec/specs/ui-data-contract-parity/spec.md` from `docs/ui-spec/04-DATA-CONTRACTS-UI.md` + `docs/UI_ENDPOINT_PARITY.md`.
- [x] Create `openspec/specs/ui-scale-and-performance/spec.md` from `docs/ui-spec/07-SCALABILITY-DATA-PLAN.md` and `12-PERF-VALIDATION-S2-S3.*`.
- [x] Create `openspec/specs/ui-governance-gates/spec.md` from `docs/ui-spec/13-UI-UX-STRATEGY-GATE.md`.
- [x] Create `openspec/specs/ui-semantic-component-layer/spec.md` from `docs/ui-spec/14-SEMANTIC-COMPONENT-LAYER.md`.
- [x] Create `openspec/specs/cloud-auth-billing/spec.md` from `docs/auth/CLERK_BILLING_SETUP.md` if billing remains in product scope.

## Phase 2: Per-Screen Use-Case Testability Hardening
- [x] Update `openspec/specs/ui-screen-home/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-onboarding-auth/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-inventory-items/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-inventory-photos/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-inventory-barcodes/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-inventory-ai-assist/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-scanner/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-discover/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-reports/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-settings/spec.md` with UC IDs and Cypress mapping table.
- [x] Update `openspec/specs/ui-screen-chat-copilot/spec.md` with UC IDs and Cypress mapping table.

## Phase 3: Cross-Feature Use-Case Normalization
- [x] Add canonical UC ID namespace and naming standard to `openspec/specs/README.md`.
- [x] Map existing `docs/USE_CASES_AND_SCENARIOS.md` UCs to OpenSpec capability/screen specs.
- [x] Ensure each major capability spec has at least one integration/API test mapping.
- [x] Ensure each UI screen spec has at least one E2E test mapping per critical flow.

## Phase 4: Test Matrix Synchronization
- [x] Create `docs/OPENSPEC_TEST_MATRIX.md` generated from OpenSpec UC mappings.
- [x] Link every UC to current or planned Cypress spec path.
- [x] Identify all unmapped UCs and open issues for missing tests.
- [x] Add required status report in CI output: mapped UCs vs unmapped UCs.

## Phase 5: Archival and Source-of-Truth Cleanup
- [x] Mark legacy docs sections as "migrated to OpenSpec" where fully covered.
- [x] Keep non-spec docs as reference-only with explicit labels.
- [x] Archive completed OpenSpec changes and roll specs into baseline.
- [x] Update `docs/OPENSPEC_WORKFLOW.md` with final migration workflow.

---

## UI Use-Case Testability Checklist (Must Hold True)
- [x] Every UI use-case has UC ID.
- [x] Every UC has a deterministic expected result.
- [x] Every UC maps to at least one automated test (current or planned).
- [x] Every critical UC has negative/error-path scenario.
- [x] Every screen has loading/empty/error/ready scenario coverage.

## Execution Order (Do Not Reorder)
1. Phase 0 (enforcement first)
2. Phase 1 (remaining baseline specs)
3. Phase 2 (screen-level UC hardening)
4. Phase 3 (cross-feature normalization)
5. Phase 4 (test matrix sync)
6. Phase 5 (cleanup and lock-in)
