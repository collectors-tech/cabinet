## MODIFIED Requirements

### Requirement: Agent skill coverage SHALL stay mapped by Cabinet surface
Cabinet SHALL maintain a durable per-surface Agent skill coverage matrix so broad "Agent can do the work" claims remain traceable to exact surfaces, skill ids, safety levels, dependencies, statuses, issue links, and validation evidence.

#### Scenario: Maintain per-surface coverage matrix
- **GIVEN** Agent skill implementation is spread across inventory, wishlist, collections, media, discoveries, market watch, purchases, integrations, inbox, users, settings, storage, data-management, chats, and external channels
- **WHEN** Cabinet reports Agent skill coverage
- **THEN** the coverage matrix SHALL list each major Cabinet surface with the expected skill ids, user-facing request examples, safety level, required context or selection, required integration/provider setup, bound capability ids, bound guided workflow ids, implementation status, missing issue or PR links, validation evidence, and blocked/deferred reason where applicable
- **AND** the matrix SHALL distinguish in-app main Chat, side-panel Chat, Inbox review, and Telegram/external channels instead of treating all Agent entry points as equivalent
- **AND** planned, blocked, and deferred entries SHALL stay explicit until their linked issues provide implementation, validation, and closure evidence

#### Scenario: Dashboard activity summary skill is read-only and profile-scoped
- **GIVEN** the active profile has Dashboard data available
- **WHEN** Cabinet lists, resolves, previews, or executes `cabinet.dashboard.summarise_activity`
- **THEN** the skill SHALL be registered as built-in, read-only, profile-scoped, and available without external provider setup
- **AND** direct execution SHALL return structured current totals, attention signals, recent items, destination links, record identifiers, source context, and non-secret evidence
- **AND** the skill SHALL NOT create mutation previews, confirmation tokens, external writes, database writes outside non-secret audit/workflow evidence, or applied mutations
- **AND** profile mismatch or unavailable Dashboard dependencies SHALL return deterministic guidance without leaking another profile's data
