## MODIFIED Requirements

### Requirement: Cabinet UI interactions SHALL be keyboard and pointer safe
Cabinet MUST ensure shared UI interactions behave predictably with mouse,
keyboard, touch, assistive technology, and responsive shell layouts across
reusable components.

#### Scenario: Route titles remain consistent and unobstructed
- **GIVEN** a user navigates to an authenticated Cabinet route
- **WHEN** the page shell renders on desktop or narrow viewports
- **THEN** the visible `HeaderTitle` SHALL use the canonical route title,
  description, and icon unless the route documents an accessibility-only layout
- **AND** the browser document title SHALL be `Cabinet - <canonical page title>`
  for every route that participates in document-title metadata
- **AND** `/purchases` SHALL resolve to `Cabinet - Purchases`
- **AND** `/scanner` SHALL resolve and render as `Market Watch`, not `Scanner`
- **AND** header title text and icon SHALL NOT collide with page actions,
  global utilities, or workspace controls at supported viewport widths.
