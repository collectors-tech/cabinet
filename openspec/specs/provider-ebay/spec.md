## Purpose
Define eBay provider contract for scanner/search integration.

## Requirements
### Requirement: eBay provider SHALL support authenticated listing search
Cabinet SHALL execute eBay listing queries using profile-scoped credentials and query-set criteria.

#### Scenario: Search eBay listings
- **GIVEN** active profile has valid eBay credentials
- **WHEN** scanner executes an eBay-backed query set
- **THEN** Cabinet SHALL fetch and normalize eBay listing candidates

### Requirement: eBay provider SHALL expose health state
Cabinet SHALL report eBay provider health and recent failure telemetry via provider health endpoints.

#### Scenario: eBay health check
- **GIVEN** provider health service is available
- **WHEN** user requests eBay health state
- **THEN** Cabinet SHALL return current status, error summary, and retryability state

### Requirement: eBay provider SHALL capture stock observations when available
Cabinet SHALL persist stock/availability observations from eBay listing payloads when present.

#### Scenario: Persist eBay stock observation
- **GIVEN** normalized eBay candidate includes stock/availability data
- **WHEN** candidate is persisted
- **THEN** stock observations SHALL be stored with candidate record

