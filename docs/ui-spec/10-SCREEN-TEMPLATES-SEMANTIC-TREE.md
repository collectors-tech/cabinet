# 10 - Screen Templates and Semantic Tree

This document defines the canonical semantic page template and component composition tree for the current Cabinet UI.

## 1. App Shell Template (Global)

```text
main.cabinet-shell [data-testid=app-shell]
  aside.cabinet-sidebar
    h1 (App Name)
    nav (Primary Navigation)
      button (Dashboard)
      button (Collection)
      button (Scanner)
      button (Discoveries)
      button (AI Assist)
      button (Barcodes)
      button (Photos)
      button (Pricing)
      button (Settings)
    div.cabinet-sidebar-meta (App Build Metadata)
      p > strong "Version"
      p > strong "Build Date"
  section.cabinet-content
    header.cabinet-topbar (Sticky)
      button.cabinet-nav-toggle (mobile)
      strong (Runtime status text)
      button#theme-toggle
    div.cabinet-content-scroll (Primary scroll container)
      section.cabinet-card[role=alert]? (Recovery alert)
      section.cabinet-card (Main body container)
        starter mode OR advanced mode content
```

## 2. Starter Mode Page Template

```text
section.cabinet-card
  h2 (Welcome to Cabinet)
  p (Starter description)
  div.cabinet-empty-state OR div.cabinet-profile-list
  p.cabinet-active-profile?
  details.cabinet-tech-details (Profile and Identity Diagnostics)
    summary
    profile storage details
    authentication diagnostics and tools
  section.cabinet-onboarding (if active profile and workspace not unlocked)
    div.cabinet-onboarding-main
      h3 (Starter Onboarding Wizard)
      h4 (Getting Started)
      p (step summary)
      ol.cabinet-onboarding-progress[aria-label="Onboarding progress"]
        li.cabinet-onboarding-step-* x5
      step-specific panel/actions
      nav row (Back Step / Next Step)
      optional step-specific forms
    aside.cabinet-onboarding-side
      h4 (Quick Status)
      status p elements
  ul (utility links)
    a /healthz
    a /api/runtime
    a /api/runtime/recovery
    a /apidocs
```

## 3. Advanced Mode Page Template

```text
section.cabinet-card
  h2 (Welcome / runtime intro)
  active screen section(s) based on selected nav state
  utility links list (health/runtime/recovery/apidocs)
```

`activeScreen` controls which top-level screen section is visible:
- `dashboard`
- `collection`
- `scanner`
- `discoveries`
- `ai`
- `barcodes`
- `photos`
- `pricing`
- `settings`

## 4. Top-Level Screen Semantic Trees

### 4.1 Dashboard

```text
section[aria-label=Dashboard][data-testid=screen-dashboard]
  h3
  section.cabinet-home-hero
    content cluster (eyebrow/title/description/next action)
    div.cabinet-home-hero-actions (workspace/deep-link CTAs)
  section.cabinet-home-attention
    h4
    div.cabinet-attention-grid
      article.cabinet-attention-card x3
  section.cabinet-home-queue
    h4
    ul
      li.cabinet-home-queue-item xN
    p.cabinet-home-calm? (zero-attention state)
  div.cabinet-home-kpi-grid OR empty-state paragraph
```

### 4.2 Collection

```text
section[aria-label=Collection][data-testid=screen-collection]
  h3
  section.collection-toolbar
    workspace mode toggle action
    search/filter controls
    add item/folder actions
    grid/table view toggle
  section.collection-summary-strip
    folders/items/quantity/value summaries
  section.collection-browser
    aside.collection-tree
      folder search
      folder hierarchy list
      utility links (history/trash)
    section.collection-results
      saved filter controls
      saved filter list
      selection bar (visible when rows/cards selected)
      loading/error/empty states
      table OR card grid (items/folders)
  section.collection-details
    collection item form
    instances block (load + form + list)
    column-view block (brands/categories/series/items lists)
```

### 4.3 Scanner

```text
section[aria-label=Scanner][data-testid=screen-scanner]
  h3
  query set form
  query set controls (load/run/schedule/candidates)
  scanner diagnostics controls (failures/provider/matching)
  status messages
  failures list
  query sets list
  candidates list with actions
  matching summary block
```

### 4.4 Discoveries

```text
section[aria-label=Discoveries][data-testid=screen-discoveries]
  h3
  filter controls (query/max price/date)
  load action
  results list
    each item row with Ignore / Wishlist / Track / Create Item actions
```

### 4.5 AI Assist

```text
section[aria-label=AI Assist][data-testid=screen-ai]
  h3
  AIAssistForms
    toggle/test/suggest/apply/retry actions
    confidence + error/status messaging
```

### 4.6 Barcodes

```text
section[aria-label=Barcodes][data-testid=screen-barcodes]
  h3
  barcode input + actions (add/load/lookup/external search)
  helper text (detect from image flow)
  error/status text
  barcode list
  local matches summary
  external URL output
```

### 4.7 Photos

```text
section[aria-label=Photos][data-testid=screen-photos]
  h3
  item selector + load
  file upload controls
  drop zone
  camera controls
  error text
  photos list (set primary/delete/fullscreen)
  fullscreen preview dialog
```

### 4.8 Pricing

```text
section[aria-label=Pricing][data-testid=screen-pricing]
  h3
  pricing/wishlist action toolbar
  wishlist draft inputs
  wishlist list
  wishlist hits list
  status summaries (track/snapshot/points/source groups/stats/trend)
  source breakdown list
  export bytes summary
```

### 4.9 Settings

```text
section[aria-label=Settings][data-testid=screen-settings]
  h3
  diagnostics block
  maintenance block
  license profile JSON editor
  license import controls
  backup restore controls
  forms block
    ProfileSettingsForm
    SecretsForm
    DataImportExportWizard
  license/log/status outputs
  errors and guarded failure hints
```

## 5. Shared Component Primitives

```text
Form primitives
  .cabinet-form-item
    label.cabinet-form-label
    input/select/textarea
    p.cabinet-form-message?

Screen wrapper primitives
  top-level screen components from ./screens/top-level-screens.tsx
  section + heading + body composition

Interactive primitives
  button.cabinet-nav-link (+ active/disabled variants)
  card surfaces (.cabinet-card, .cabinet-onboarding-*, .cabinet-attention-card)
  list/table surfaces
```

## 6. Composition Rules

1. Keep one dominant heading per visible screen region.
2. Keep primary action row near the top of each screen.
3. Keep technical diagnostics behind `details` by default in starter mode.
4. Keep scroll ownership in `.cabinet-content-scroll` for desktop layout.
5. Preserve explicit labels/roles (`aria-label`, `data-testid`) used by tests.

## 7. Current Gaps Noted from Template

1. Starter mode still carries too many mixed concerns inside one card.
2. Dashboard visual hierarchy is improved but still not at final polish.
3. Utility links should move to a dedicated footer utility strip in a later pass.
4. Several advanced screens still rely on dense action bars and need cardized grouping.
