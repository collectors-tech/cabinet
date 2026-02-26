## Purpose
Define barcode lookup and external resolution behavior.

## Requirements
### Requirement: Barcode lookup SHALL support local and external resolution
Cabinet SHALL support local match lookup and external search integrations.

#### Scenario: Local barcode lookup
- **GIVEN** a barcode lookup request is issued
- **WHEN** local match exists
- **THEN** Cabinet SHALL return local matches

#### Scenario: External barcode search
- **GIVEN** local match is absent
- **WHEN** external search is invoked
- **THEN** Cabinet SHALL return provider search results or explicit failure state

