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
- **GIVEN** template catalog includes `wishlist`, `trade_binder`, and `watchlist`
- **WHEN** client requests `GET /api/integrations/pokemon/list-templates`
- **THEN** API MUST return `200` with `templates` array
- **AND** each template MUST expose `id`, `label`, `default_fields`, `default_filters`, and `sort_order`

#### Scenario: Apply template to create dynamic list definition
- **GIVEN** user selects `trade_binder` template
- **WHEN** client calls `POST /api/integrations/pokemon/list-templates/apply` with `{"template_id":"trade_binder","list_name":"Trade Binder"}`
- **THEN** API MUST return `201` with payload fields `list_id`, `list_name`, `template_id`, `default_fields`, `default_filters`, and `sort_order`
- **AND** `default_fields` MUST include grading-centric fields (`grader`, `grade_numeric`, `collector_classification`)
- **AND** `default_filters` MUST include at least one deterministic status filter

#### Scenario: Unknown template is rejected deterministically
- **GIVEN** `template_id` is not a known template
- **WHEN** client calls `POST /api/integrations/pokemon/list-templates/apply`
- **THEN** API MUST return `400` with `{"error":"invalid_template_id"}`

### Requirement POKEMON-COMP-007: Graded workflow MUST capture slab metadata with valuation overrides
Cabinet SHALL support graded-card metadata fields (grader, grade, cert number, slab state) and optional valuation override with source attribution.

#### Scenario: Graded card valuation override
- **GIVEN** canonical item exists and belongs to active profile
- **WHEN** client calls `POST /api/integrations/pokemon/graded-overrides` with `item_id`, `grader`, `grade_numeric`, `cert_number`, `slab_state`, `valuation_override_amount`, `currency`, and `source_note`
- **THEN** API MUST return `200` and persist override metadata with deterministic `overridden_at` timestamp
- **AND** subsequent `GET /api/integrations/pokemon/graded-overrides?item_id=<id>` MUST return fields `item_id`, `grader`, `grade_numeric`, `cert_number`, `slab_state`, `valuation_override_amount`, `currency`, `source_note`, and `overridden_at`
- **AND** canonical item grading fields MUST be updated to reflect graded slab metadata (`grading_status`, `grader`, `grade_numeric`, `slabbed`)

#### Scenario: Missing item id is rejected deterministically
- **GIVEN** request omits `item_id`
- **WHEN** client calls `POST /api/integrations/pokemon/graded-overrides`
- **THEN** API MUST return `400` with error envelope `{"error":"missing_item_id"}`

### Requirement POKEMON-COMP-008: Social sharing hooks MUST support progress snapshots
Cabinet SHALL provide deterministic share-ready progress snapshots for selected lists/sets.

#### Scenario: Generate progress snapshot with canonical share payload
- **GIVEN** active profile has canonical inventory tagged with `set:<id>` and optional `language:<code>` tags
- **AND** request includes `set_id` and `total_count`
- **WHEN** client requests `GET /api/integrations/pokemon/progress-snapshot?set_id=<id>&total_count=<n>`
- **THEN** API MUST return `200` with fields `set_id`, `owned_count`, `total_count`, `completion_percent`, `share_payload`, and `generated_at`
- **AND** `share_payload` MUST include fields `headline`, `summary`, `visibility`, and `share_link`
- **AND** `completion_percent` MUST be deterministic as `(owned_count / total_count) * 100` rounded to two decimals when `total_count > 0`
- **AND** `share_payload.visibility` MUST default to `private`

#### Scenario: Missing set identifier is rejected deterministically
- **GIVEN** request omits `set_id`
- **WHEN** client requests `GET /api/integrations/pokemon/progress-snapshot`
- **THEN** API MUST return `400` with error envelope `{"error":"missing_set_id"}`

### Requirement POKEMON-COMP-009: Gamified milestones MUST support deterministic badge triggers
Cabinet SHALL expose deterministic milestone trigger evaluation for collector progression badges.

#### Scenario: Milestone badge trigger evaluation
- **GIVEN** active profile has canonical inventory tagged with `set:<id>`
- **AND** request includes `set_id` and `total_count`
- **WHEN** client calls `POST /api/integrations/pokemon/milestone-evaluate` with `{"set_id":"<id>","total_count":<n>}`
- **THEN** API MUST return `200` with fields `set_id`, `owned_count`, `completion_percent`, and `events`
- **AND** each entry in `events` MUST include `milestone_id`, `threshold_pct`, and `triggered_at`
- **AND** milestone ids MUST be deterministic using thresholds `25`, `50`, `75`, and `100` (`milestone-25`, `milestone-50`, `milestone-75`, `milestone-100`)
- **AND** only thresholds less than or equal to `completion_percent` MUST be returned

#### Scenario: Missing set identifier is rejected deterministically
- **GIVEN** request omits `set_id`
- **WHEN** client calls `POST /api/integrations/pokemon/milestone-evaluate`
- **THEN** API MUST return `400` with error envelope `{"error":"missing_set_id"}`

### Requirement POKEMON-COMP-010: Goal bundle presets MUST support collector objective workflows
Cabinet SHALL provide objective presets that create grouped filters/actions for common collector goals.

#### Scenario: Goal bundle catalog is deterministic
- **GIVEN** client requests collector objective presets
- **WHEN** client requests `GET /api/integrations/pokemon/goal-bundles`
- **THEN** API MUST return `200` with field `bundles`
- **AND** each bundle MUST include `id`, `label`, `filters`, `actions`, and `shortcut`
- **AND** catalog MUST include ids `finish-master-set`, `optimize-trade-binder`, and `price-drop-watch`

#### Scenario: Goal bundle apply creates deterministic workspace definition
- **GIVEN** user selects `finish-master-set` goal bundle
- **WHEN** client calls `POST /api/integrations/pokemon/goal-bundles/apply` with `{"bundle_id":"finish-master-set","workspace_name":"Master Set Focus"}`
- **THEN** API MUST return `201` with fields `workspace_id`, `workspace_name`, `bundle_id`, `filters`, `actions`, and `shortcut`
- **AND** `bundle_id` in response MUST match the requested bundle
- **AND** `filters` and `actions` MUST be copied from the selected bundle definition

#### Scenario: Unknown goal bundle id is rejected deterministically
- **GIVEN** `bundle_id` is not in the catalog
- **WHEN** client calls `POST /api/integrations/pokemon/goal-bundles/apply`
- **THEN** API MUST return `400` with error envelope `{"error":"invalid_bundle_id"}`
