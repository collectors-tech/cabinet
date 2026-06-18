## Purpose
Define global shell, navigation, context pane, and layout ownership behavior.

## Requirements
### Requirement UI-FOUNDATION-SHELL-NAVIGATION-001: App shell SHALL define fixed navigation and scroll ownership
Cabinet SHALL keep primary navigation fixed and assign page-body scroll to content container.

#### Scenario: Page content scroll
- **GIVEN** desktop viewport is active and page body content exceeds available vertical height
- **WHEN** content exceeds viewport height
- **THEN** only content column SHALL scroll while primary nav remains fixed and header actions remain visible at top of content pane

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-002: Primary navigation SHALL support collapse and edit behavior
Cabinet SHALL support nav collapse and configurable ordering and visibility controls.

#### Scenario: Edit navigation order
- **GIVEN** user enters navigation edit mode from shell footer controls with profile-scoped nav preferences available
- **WHEN** user enters nav edit mode and reorders entries
- **THEN** nav order SHALL persist for active profile context and reload with same order and visibility toggles

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-003: Context navigation SHALL support collection-focused workspace context
Cabinet SHALL provide context pane behavior for collection and folder navigation.

#### Scenario: Context pane selection
- **GIVEN** collection context pane is visible and at least two contexts/folders are available
- **WHEN** user changes active collection context
- **THEN** page content SHALL reflect selected context state and context label in header/body SHALL match selected entry

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-004: Version/build metadata SHALL be visible in shell footer
Cabinet SHALL display app version and build date metadata in sidebar footer region.

#### Scenario: Render shell footer metadata
- **GIVEN** runtime metadata endpoint returns `app_version` and `build_date`
- **WHEN** shell renders
- **THEN** version/build metadata SHALL be visible in footer

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-005: Database switcher SHALL remain DB/profile-only
Top shell switcher SHALL represent database/profile context only and SHALL NOT render collections management widget above primary navigation.

#### Scenario: Sidebar top area renders DB/profile only
- **GIVEN** authenticated shell sidebar is rendered
- **WHEN** Database switcher is visible
- **THEN** top area MUST show DB/profile switcher context only
- **AND** collections list/add widget MUST NOT appear above primary nav items

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-008: Database switcher label SHALL use explicit Database terminology
Shell switcher labels MUST use `Database` terminology instead of ambiguous `Local Workspace` wording.

#### Scenario: Render database switcher labels
- **GIVEN** sidebar top switcher is visible
- **WHEN** switcher renders title/subtitle
- **THEN** UI MUST display Database-oriented labels (for example `Database`, `Primary DB`, `Showcase DB`)

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-006: Collections management SHALL live in Collections section and inline pickers
Collections creation/list management SHALL be provided via dedicated Collections section and inline picker quick-create flows, not via sidebar-top widget.

#### Scenario: Collections management placement
- **GIVEN** user needs to manage collections
- **WHEN** user uses navigation or picker flows
- **THEN** collections list/create MUST be available in Collections section and relevant pickers
- **AND** sidebar top area remains uncluttered

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-009: Database switcher SHALL support functional profile switching
Selecting a different database profile from switcher MUST change active data context across app screens.

#### Scenario: Switch from Primary DB to Showcase DB
- **GIVEN** at least two configured database profiles exist (`Primary DB`, `Showcase DB`)
- **WHEN** user selects `Showcase DB` in switcher
- **THEN** runtime MUST switch active profile context and reload data views from selected DB without cross-profile leakage

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-010: App SHALL provide a seeded Showcase DB profile
Cabinet SHALL support a pre-seeded showcase database profile for demos/testing with sample content and SHALL distinguish it from normal working databases in the shell switcher.

#### Scenario: Open showcase profile
- **GIVEN** showcase profile is provisioned
- **WHEN** user switches to `Showcase DB`
- **THEN** inventory, wishlist, media, and account/demo context MUST be populated with sample seed content suitable for end-to-end demos
- **AND** the switcher MUST label Showcase DB as sample/demo context rather than a generic working database

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-007: Navigation edit dialog SHALL reflect live item order during reordering
When user reorders items in nav edit mode through move buttons or a drag handle, edit dialog list order MUST update immediately to match resulting navigation order.

#### Scenario: Move menu item and verify edit list order
- **GIVEN** navigation edit dialog is open with reorder controls
- **WHEN** user moves an item up or down
- **THEN** edit dialog list MUST re-render in new order immediately
- **AND** resulting saved order MUST match what was shown in edit dialog before save

#### Scenario: Drag nav row from left-side handle and reorder with visible insertion feedback
- **GIVEN** navigation edit dialog is open with drag handles on the left side of each row label
- **WHEN** user drags a row by its handle and reorders it within the list
- **THEN** the edit dialog MUST show visible insertion feedback during drag
- **AND** the list MUST re-render immediately in the dropped order
- **AND** move up, move down, and hide controls MUST remain available on the same row

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-011: Authenticated shell content SHALL default to full-width workspace fill
Authenticated shell pages SHALL let the main content region expand to the full available workspace width beside the sidebar unless a screen explicitly opts into a constrained document-style layout.

#### Scenario: Inventory workspace fills available shell width on wide viewport
- **GIVEN** an authenticated user opens the inventory workspace on a wide desktop viewport with the shell sidebar visible
- **WHEN** the main content region renders beside the sidebar
- **THEN** the primary shell content container MUST expand to nearly the full available width beside the sidebar
- **AND** the shell MUST NOT apply an implicit centered `max-width` cap to the workspace container by default
- **AND** inventory panels MUST retain their grid structure while using the wider workspace area

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-012: Sidebar count pills SHALL render as trailing nav badges
Sidebar navigation rows that expose notification or count pills SHALL render them as right-aligned trailing badges using the shell nav badge pattern instead of inline label chips.

#### Scenario: Chats row renders count pill as trailing sidebar badge
- **GIVEN** the authenticated shell sidebar includes a navigation row with a count pill such as `Chats`
- **WHEN** the row renders in desktop expanded-sidebar mode
- **THEN** the label text MUST remain left-aligned in the row body
- **AND** the count pill MUST render as a trailing right-aligned badge near the row end
- **AND** the label text MUST NOT collapse into or overlap the badge area

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-013: Authenticated routes SHALL set product-first browser titles
Authenticated shell routes SHALL keep `document.title` in the format `Cabinet - <Page Title>`.

#### Scenario: Browser title updates across representative routes
- **GIVEN** an authenticated user navigates between primary shell routes
- **WHEN** the active route changes
- **THEN** the browser title MUST update to `Cabinet - <Page Title>` for that route
- **AND** the title MUST NOT be blank or leak raw route ids/translation keys

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-014: Database switcher SHALL recover from active profile load failure
The database/profile switcher SHALL make active profile load failures visible and SHALL provide a retry action that restores the active database label when the profile endpoint recovers.

#### Scenario: Retry profile loading from shell switcher
- **GIVEN** the authenticated shell cannot load the active profile/database context
- **WHEN** the user opens the database switcher and retries profile loading
- **THEN** the shell SHALL show the profile load failure before retry
- **AND** the recovered active database label SHALL replace the failure guidance after retry succeeds

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-015: Database switcher SHALL create and activate profiles
The database/profile switcher add action SHALL create a new database profile, activate it, and refresh the shell against that profile context.

#### Scenario: Add database profile from switcher
- **GIVEN** the authenticated shell database switcher is visible
- **WHEN** the user chooses Add Database and supplies a profile name
- **THEN** the switcher SHALL call the profile create API with that name
- **AND** the switcher SHALL activate the created profile through the active profile API
- **AND** the shell SHALL reload with the created profile label as active database context

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-016: Existing database selection SHALL carry across core app sections
The shell database/profile switcher SHALL keep an existing selected database profile visible as the active context while the user moves through core authenticated sections.

#### Scenario: Existing database profile selection across sections
- **GIVEN** the authenticated shell has loaded multiple database profiles
- **WHEN** the user selects an existing non-active database profile
- **THEN** the active profile API SHALL report the selected profile
- **AND** the selected database label SHALL remain visible in Inventory, Wishlist, Collections, Settings, Chats, and Integrations

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-017: Browser and install metadata SHALL use approved Cabinet icon assets
Cabinet SHALL publish approved Cabinet app icon assets through browser favicon and install metadata without retaining stale template or prior app icon references.

#### Scenario: Browser and install metadata expose current Cabinet branding
- **GIVEN** the app shell HTML and static image assets are served
- **WHEN** a browser resolves favicon, touch icon, and web app manifest metadata
- **THEN** SVG favicons MUST use the approved Cabinet light and dark icon variants
- **AND** PNG, ICO, touch icon, and web app manifest assets MUST resolve from the approved Cabinet favicon package
- **AND** app install metadata MUST use `Cabinet` naming and Cabinet theme colors

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-018: Primary navigation links SHALL render as icon-only controls
Cabinet SHALL render authenticated primary app navigation links as icon-only controls while preserving stable accessible names, tooltips, active-route affordances, and keyboard focus affordances.

#### Scenario: Primary app navigation is icon-only and accessible
- **GIVEN** the authenticated shell sidebar is visible
- **WHEN** primary app navigation links render in the Navigation workspace
- **THEN** each link MUST show its icon without visible text label content
- **AND** each link MUST expose a stable accessible name matching the destination label
- **AND** the active route state MUST remain visually detectable through the shell nav active affordance
- **AND** keyboard focus MUST remain visible on each icon-only navigation control
