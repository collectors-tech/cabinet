## MODIFIED Requirements

### Requirement: Cabinet UI SHALL provide shared table and dialog primitives
Cabinet MUST expose reusable UI primitives for table interaction, row-level
record actions, CRUD dialogs, and route-level shell metadata so feature
surfaces do not implement incompatible copies of the same behavior.

#### Scenario: Authenticated routes use canonical metadata
- **GIVEN** an authenticated Cabinet route is visible in the shell
- **WHEN** the application resolves its page metadata
- **THEN** the route SHALL resolve to exactly one canonical title,
  description, icon, navigation group, browser-title eligibility, and stable
  header/icon test identifier
- **AND** page headers, sidebar navigation, command/search navigation, and
  browser document titles SHALL use the same canonical title when they describe
  the same route
- **AND** settings child routes SHALL have specific page titles/icons while
  remaining grouped under Settings navigation
- **AND** unknown, unsupported, and dynamic routes SHALL use an explicit
  fallback that does not hide missing metadata for known authenticated routes
- **AND** existing packaged tests that depend on stable page/icon test IDs
  SHALL keep a compatible identifier or receive an intentional documented
  migration.
