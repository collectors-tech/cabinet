# Docs History Migration Tracker

| Source (`commit:path`) | OpenSpec targets | Status | Verified By/Date | Evidence (Requirement IDs / PR / Commit) | Notes |
| --- | --- | --- | --- | --- | --- |
| `82294546bf0b715fe49394e1c5a885d3045294d2:docs/SHOP_PROVIDERS.md` | `openspec/specs/provider-registry/spec.md`; `openspec/specs/provider-ebay/spec.md`; `openspec/specs/provider-amazon/spec.md`; `openspec/specs/provider-au-webshops/spec.md`; `openspec/specs/integrations/spec.md`; `openspec/specs/scanner/spec.md` | migrated-verified | codex / 2026-02-26 | IDs: `INTEGRATION-001..015`, `OPS-001`; Closure Commit: `e81cca8`; Issue: `#182` | Normative requirements extracted from provider list, API feasibility notes (eBay/Amazon), and integration policy bullets. |

## Traceability Details

### Source: `82294546bf0b715fe49394e1c5a885d3045294d2:docs/SHOP_PROVIDERS.md`
- Legacy section intent: provider catalog list and integration mode per provider
  - New IDs: `INTEGRATION-001`, `INTEGRATION-002`, `INTEGRATION-003`, `INTEGRATION-004`, `INTEGRATION-011`
  - Validating tests:
    - `TestProviderSpecsExistAndRegistryLinksThem` (`internal/app/openspec_provider_specs_test.go`)
    - `TestOpenSpecScenariosRequireGivenWhenThen` (`internal/app/openspec_scenario_contract_test.go`)
- Legacy section intent: eBay official API feasibility and policy
  - New IDs: `INTEGRATION-005`, `INTEGRATION-006`, `INTEGRATION-007`
  - Validating tests:
    - `TestProviderSpecsExistAndRegistryLinksThem` (`internal/app/openspec_provider_specs_test.go`)
    - `TestEbayProviderResponseContract` (`tests/shop_providers_contract_test.go`)
  - Verification:
    - status: done
    - commit: `e81cca8` (`#182`)
- Legacy section intent: Amazon constrained API/program eligibility
  - New IDs: `INTEGRATION-008`, `INTEGRATION-009`, `INTEGRATION-010`
  - Validating tests:
    - `TestProviderSpecsExistAndRegistryLinksThem` (`internal/app/openspec_provider_specs_test.go`)
    - `TestAmazonDisabledModeReturns409ContractEnvelope` (`tests/shop_providers_contract_test.go`)
  - Verification:
    - status: done
    - commit: `e81cca8` (`#182`)
- Legacy section intent: integration policy for web ingestion
  - New IDs: `OPS-001`, `INTEGRATION-012`, `INTEGRATION-013`, `INTEGRATION-014`, `INTEGRATION-015`
  - Validating tests:
    - `TestProviderSpecsExistAndRegistryLinksThem` (`internal/app/openspec_provider_specs_test.go`)
    - `TestAUWebshopThrottlingConformanceOPS001` (`tests/shop_providers_contract_test.go`)
  - Verification:
    - status: done
    - commit: `e81cca8` (`#182`)

## PR Gate Notes
- Requirement IDs touched:
  - `INTEGRATION-001..015`
  - `OPS-001`
- Tests proving touched IDs:
  - `TestProviderSpecsExistAndRegistryLinksThem`
  - `TestOpenSpecScenariosRequireGivenWhenThen`
  - `openspec validate --all`
- OpenAPI changed: `no`
- Changed paths/schemas if yes: `n/a`
