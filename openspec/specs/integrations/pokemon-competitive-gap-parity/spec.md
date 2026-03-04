## Purpose
Define competitive parity requirements derived from Pokemon-focused collector apps and map them into testable Cabinet feature slices.

## Requirements
### Requirement POKEMON-COMP-001: Scanner workflow MUST support confidence-first batch capture
Cabinet SHALL provide a batch-capable camera/upload capture flow with confidence scoring, alternate suggestions, and explicit manual override before mutation apply.

#### Scenario: Batch scan with confidence and override
- **GIVEN** user captures multiple item photos from mobile or desktop upload queue
- **WHEN** scanner returns candidate matches
- **THEN** each candidate row MUST include confidence score and at least one manual override action
- **AND** no inventory mutation occurs until user confirms apply

### Requirement POKEMON-COMP-002: Collection model MUST support set progress dimensions
Cabinet SHALL model set completion across card variants, language, and graded state for progress reporting.

#### Scenario: Progress model computation
- **GIVEN** active profile inventory contains card items tagged with `set:<id>`, `variant:<name>`, and `language:<code>`
- **AND** grading state is represented by each item `grading_status` (`graded` or `ungraded`)
- **WHEN** client requests `GET /api/integrations/pokemon/set-progress?set_id=<id>&total_count=<n>`
- **THEN** API MUST return `200` with fields `set_id`, `owned_count`, `total_count`, `completion_percent`, and `breakdown`
- **AND** `breakdown` MUST include nested maps `variant`, `language`, and `graded`
- **AND** `completion_percent` MUST be deterministic as `(owned_count / total_count) * 100` rounded to two decimals when `total_count > 0`

#### Scenario: Missing set identifier is rejected deterministically
- **GIVEN** request omits `set_id`
- **WHEN** client requests `GET /api/integrations/pokemon/set-progress`
- **THEN** API MUST return `400` with error envelope field `error` set to `missing_set_id`

### Requirement POKEMON-COMP-003: Pricing MUST support multi-source historical trends and alerts
Cabinet SHALL aggregate pricing snapshots across configured providers and expose trend deltas with threshold alert rules.

#### Scenario: Trend and alert evaluation
- **GIVEN** tracked Pokemon items in the same set have daily snapshots across at least two sources
- **WHEN** client requests `GET /api/integrations/pokemon/price-alerts?set_id=<id>&drop_threshold_pct=<n>`
- **THEN** API MUST return `200` with fields `set_id`, `sources`, `items`, and `alerts`
- **AND** each `sources` entry MUST include `source`, `min_price`, `median_price`, and `latest_price`
- **AND** each alert in `alerts` MUST include `item_id`, `source`, `change_pct`, and `threshold_pct`
- **AND** only crossed-threshold records (negative change at or beyond threshold) MUST appear in `alerts`

#### Scenario: Missing set identifier is rejected deterministically
- **GIVEN** request omits `set_id`
- **WHEN** client requests `GET /api/integrations/pokemon/price-alerts`
- **THEN** API MUST return `400` with error envelope field `error` set to `missing_set_id`

### Requirement POKEMON-COMP-004: Discovery handoff MUST preserve marketplace decision metadata
Cabinet SHALL include seller reputation, stock signal, and direct buy-link context when handing candidates to wishlist/discovery actions.

#### Scenario: Discovery to wishlist handoff
- **GIVEN** scanner candidate `candidate_id=<id>` exists with listing metadata fields `url`, `seller`, `stock_state`, and `price`
- **AND** candidate has a resolved match with non-empty `item_id`
- **WHEN** client calls `POST /api/discovery/action` with body `{"candidate_id":"<id>","type":"add_to_wishlist"}`
- **THEN** API MUST return `200` with body field `ok=true`
- **AND** subsequent `GET /api/wishlist` response MUST include a wishlist record for `item_id`
- **AND** that wishlist record MUST retain listing decision metadata in persisted notes payload with fields `listing_url`, `seller`, `stock_signal`, and `observed_price`

#### Scenario: Missing candidate identifier is rejected deterministically
- **GIVEN** request omits `candidate_id`
- **WHEN** client calls `POST /api/discovery/action` with body `{"type":"add_to_wishlist"}`
- **THEN** API MUST return `400` with error envelope field `error` set to `failed_to_apply_discovery_action`

### Requirement POKEMON-COMP-005: Share controls MUST expose visibility policy per list/profile
Cabinet SHALL support deterministic visibility controls (`private`, `shared_link`, `team`) for collections and dynamic lists.

#### Scenario: Visibility policy enforcement
- **GIVEN** visibility policy is configured via request parameter `visibility` with allowed values `private`, `shared_link`, `team`
- **WHEN** client requests `GET /api/integrations/pokemon/visibility-access?visibility=<policy>&actor=<anonymous|authenticated|team_member>&share_token=<token>`
- **THEN** API MUST return deterministic envelopes:
  - `private` + `anonymous` => `403` with `{"error":"visibility_forbidden","visibility":"private","required":"authenticated"}`
  - `shared_link` + missing `share_token` => `403` with `{"error":"missing_share_token","visibility":"shared_link","required":"share_token"}`
  - `team` + actor not `team_member` => `403` with `{"error":"visibility_forbidden","visibility":"team","required":"team_member"}`
  - allowed access => `200` with `{"ok":true,"visibility":"<policy>","actor":"<actor>"}` and optional `share_token_present`

#### Scenario: Invalid visibility values are rejected deterministically
- **GIVEN** `visibility` is omitted or outside `private|shared_link|team`
- **WHEN** client requests `GET /api/integrations/pokemon/visibility-access`
- **THEN** API MUST return `400` with `{"error":"invalid_visibility"}`

### Requirement POKEMON-COMP-006: Dynamic list templates MUST be first-class and reusable
Cabinet SHALL support reusable list templates for wishlist, trade binder, and watchlist views with saved filters and sort order.

#### Scenario: Template-based list creation
- **GIVEN** user selects `trade_binder` template
- **WHEN** list is created
- **THEN** list MUST preload template fields, default filters, and saved sort order

### Requirement POKEMON-COMP-007: Graded workflow MUST capture slab metadata with valuation overrides
Cabinet SHALL support graded-card metadata fields (grader, grade, cert number, slab state) and optional valuation override with source attribution.

#### Scenario: Graded card valuation override
- **GIVEN** graded instance has grader + numeric grade + certificate data
- **WHEN** user saves valuation override
- **THEN** record MUST persist override amount, currency, timestamp, and override source note

### Requirement POKEMON-COMP-008: Social sharing hooks SHOULD support progress snapshots
Cabinet SHOULD provide share-ready progress snapshots for selected lists/sets.

#### Scenario: Generate progress snapshot
- **GIVEN** user requests share snapshot for a set progress report
- **WHEN** snapshot API executes
- **THEN** response SHOULD include summary metrics and canonical share payload

### Requirement POKEMON-COMP-009: Gamified milestones SHOULD support deterministic badge triggers
Cabinet SHOULD expose milestone trigger rules for onboarding and collection progression badges.

#### Scenario: Milestone badge trigger
- **GIVEN** user crosses configured milestone threshold
- **WHEN** milestone evaluation runs
- **THEN** badge event SHOULD be emitted with milestone id and timestamp

### Requirement POKEMON-COMP-010: Goal bundle presets SHOULD support collector objective workflows
Cabinet SHOULD provide objective presets that create grouped filters/actions for common collector goals.

#### Scenario: Goal bundle activation
- **GIVEN** user activates objective preset
- **WHEN** preset is saved
- **THEN** system SHOULD create linked filter set and dashboard action shortcuts
