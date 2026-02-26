# OpenSpec Traceability Matrix

## SHOP_PROVIDERS Migration

- Source migration commit: `da7c354`
- Closure/verification commit: `e81cca8`
- Tracking issue: `#182`

| Requirement IDs | Contract Focus | Route/Component Scope | Validating Tests | Status |
| --- | --- | --- | --- | --- |
| `INTEGRATION-005` | eBay authenticated listing search contract fields | `internal/ebay.Provider.Search` | `TestEbayProviderResponseContract` (`tests/shop_providers_contract_test.go`) | implemented |
| `INTEGRATION-006` | eBay provider health response stability | `GET /api/provider/health?provider=ebay` + provider contract | `TestEbayProviderResponseContract` (`tests/shop_providers_contract_test.go`), `TestScannerQuerySetsAndProviderHealthEndpoints` (`internal/app/scanner_api_test.go`) | implemented |
| `INTEGRATION-007` | eBay normalized candidate field persistence baseline | eBay candidate normalization output | `TestEbayProviderResponseContract` (`tests/shop_providers_contract_test.go`) | implemented |
| `INTEGRATION-010` | Amazon disabled mode explicit `409 PROVIDER_DISABLED` envelope | provider-disabled API envelope contract harness + app-router feasibility path `GET /api/provider/health?provider=amazon` | `TestAmazonDisabledModeReturns409ContractEnvelope`, `TestAmazonProviderHealthAppRouterPathFeasibility` (`tests/shop_providers_contract_test.go`) | implemented |
| `OPS-001` | AU webshop throttling/backoff conformance and degraded state observability | scanner retry/backoff + provider health/failure logs for AU query-set region | `TestAUWebshopThrottlingConformanceOPS001_RegionAU` (`tests/shop_providers_contract_test.go`) | implemented |

## Remaining Planned Coverage

- `INTEGRATION-008`, `INTEGRATION-009`, `INTEGRATION-011`, `INTEGRATION-012`, `INTEGRATION-013`, `INTEGRATION-014`, `INTEGRATION-015`: covered by migrated spec assertions and existing OpenSpec structure checks; runtime/API implementation tests remain planned as provider registry endpoints are expanded.
