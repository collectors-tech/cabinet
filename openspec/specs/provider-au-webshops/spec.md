## Purpose
Define AU webshop provider family contract for slot-car collector discovery and stock-aware candidate ingestion.

## Requirements
### Requirement: AU webshop provider family SHALL maintain domain catalog
Cabinet SHALL maintain provider entries for:
- bonzaslotcars.com.au
- frontlinehobbies.com.au
- hobbytechtoys.com.au
- andrewshobbies.com.au
- voglers.com.au
- acercmodels.com
- mrtoys.com.au

#### Scenario: AU provider catalog list
- **GIVEN** AU webshop provider family is enabled
- **WHEN** integrations registry is requested
- **THEN** all configured AU webshop domains SHALL be listed

### Requirement: AU webshop ingestion SHALL extract stock observations
Cabinet SHALL parse stock/availability from webshop listing pages where available and persist normalized stock observations.

#### Scenario: Webshop stock extraction
- **GIVEN** a webshop listing page includes stock/availability signal
- **WHEN** candidate ingestion normalizes the listing
- **THEN** candidate record SHALL include stock observation fields with source attribution

### Requirement: AU webshop providers SHALL declare scraping policy and throttling
Cabinet SHALL store per-domain crawling policy metadata including crawl delay/rate limit and failure backoff behavior.

#### Scenario: Domain throttle applied
- **GIVEN** domain policy declares rate limit constraints
- **WHEN** scanner executes multiple AU webshop requests
- **THEN** requests SHALL respect per-domain throttling policy

