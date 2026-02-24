# 09 Component Specs (Strict)

## Scope
Defines strict component contracts per UI section: inputs/outputs, states, accessibility, and test cases.

## Contract Format (mandatory)
- `Inputs`: props/parameters (required vs optional)
- `Outputs`: emitted callbacks/events
- `States`: loading/empty/error/success + component-specific variants
- `Accessibility`: roles, labels, keyboard behavior, focus behavior
- `Telemetry`: optional event names for usage instrumentation
- `Tests`: required test ids and behaviors

## Global Primitives

### Button
- Inputs:
  - `variant`: primary | secondary | ghost | danger
  - `size`: sm | md | lg
  - `disabled`: boolean
  - `loading`: boolean
  - `aria-label`: required when no visible text
- Outputs:
  - `onClick()`
- States:
  - default, hover, active, focus-visible, disabled, loading
- Accessibility:
  - keyboard `Enter/Space` activation
  - focus ring visible
- Tests:
  - `BTN-001` disabled blocks click
  - `BTN-002` loading blocks repeat submits

### Text Input
- Inputs:
  - `value`, `onChange`
  - `placeholder`
  - `aria-label` or associated `<label>`
  - `invalid` and `message`
- Outputs:
  - `onChange(value)`
  - `onBlur()`
- States:
  - default, focused, invalid, disabled
- Accessibility:
  - label association required
- Tests:
  - `INP-001` invalid state shows message

### Select
- Inputs:
  - `value`, `options`, `onChange`
- Outputs:
  - `onChange(value)`
- Accessibility:
  - labeled control
- Tests:
  - `SEL-001` option change updates bound state

### Modal/Dialog
- Inputs:
  - `open`, `onOpenChange`
  - `title`, `description`
- Outputs:
  - close actions (`Esc`, overlay click, close button)
- States:
  - open, closed
- Accessibility:
  - `role="dialog"`, `aria-modal="true"`
  - focus trap and restore
- Tests:
  - `DIA-001` escape closes dialog
  - `DIA-002` focus returns to trigger

### Drawer (mobile nav)
- Inputs:
  - `open`, `onClose`
  - `items[]`
- Outputs:
  - `onNavigate(route)`
- Accessibility:
  - `role="dialog"`
- Tests:
  - `DRW-001` opens/closes via trigger + escape
  - `DRW-002` navigation closes drawer

## Screen Component Inventory

## Home Components
1. AttentionCardList
2. AttentionCard
3. SnapshotKPIGrid
4. RecentActivityList
5. QuickActionBar

### AttentionCard
- Inputs:
  - `title`, `severity`, `count`, `rows[]`, `actions[]`
  - `snoozed_until` optional
- Outputs:
  - `onAction(actionId, payload)`
  - `onSnooze(duration)`
  - `onDismiss()`
- States:
  - normal, empty, error
- Accessibility:
  - action buttons keyboard reachable
- Tests:
  - `HOME-CARD-001` action click emits correct payload
  - `HOME-CARD-002` dismiss updates local card state

## Inventory: Items Components
1. InventoryFilterBar
2. ItemListTable
3. ItemDetailsPanel
4. QuickAddItemForm
5. AdvancedItemForm
6. InstanceList
7. InstanceForm

### QuickAddItemForm
- Inputs:
  - `initialValues`
  - `onSubmit(values)`
  - `isSubmitting`
- Outputs:
  - submit valid payload
- States:
  - pristine, dirty, submitting, success, error
- Accessibility:
  - field labels required
- Tests:
  - `INV-QF-001` required fields enforced
  - `INV-QF-002` submit disabled while submitting

### ItemListTable
- Inputs:
  - `rows[]`, `sort`, `filters`, `selectedRowId`
- Outputs:
  - `onSelectRow(id)`, `onSort(change)`, `onPage(change)`
- States:
  - loading skeleton, empty rows, error banner
- Accessibility:
  - header cells readable and sortable with keyboard
- Tests:
  - `INV-TBL-001` row select opens details

## Inventory: Photos Components
1. PhotoItemPicker
2. PhotoUploadPanel
3. CameraCapturePanel
4. PhotoList
5. FullscreenPhotoDialog

### PhotoUploadPanel
- Inputs:
  - `selectedItemId`
  - `onUpload(file)`
- Outputs:
  - upload request
- States:
  - idle, file-staged, uploading, success, error
- Tests:
  - `PHO-UP-001` upload without item id blocked
  - `PHO-UP-002` drag/drop stages file

## Inventory: Barcodes Components
1. BarcodeInputBar
2. BarcodeList
3. BarcodeLookupResult
4. BarcodeExternalSearchLink

### BarcodeLookupResult
- Inputs:
  - `matches[]`
  - `status`: matched | none | error
- Outputs:
  - none
- Tests:
  - `BAR-RES-001` matched state shows count
  - `BAR-RES-002` no-match state suggests external search

## Inventory: AI Components
1. AITogglePanel
2. TitleNormalizeForm
3. PhotoIdentifyForm
4. AISuggestionPreview
5. AIConfirmApplyAction

### AISuggestionPreview
- Inputs:
  - `suggestion`, `confidence`, `error`
- Outputs:
  - `onApply()`
  - `onRetry()`
- Tests:
  - `AI-PRV-001` apply hidden when no suggestion
  - `AI-PRV-002` confidence displayed when provided

## Discover Components
1. DiscoverFilterBar
2. DiscoveryCandidateList
3. DiscoveryCandidateRowActions

### DiscoveryCandidateRowActions
- Inputs:
  - `candidateId`
  - capability flags
- Outputs:
  - `onIgnore`, `onWishlist`, `onTrack`, `onCreateItem`
- States:
  - idle, action-pending, action-error
- Tests:
  - `DIS-ACT-001` each action sends correct payload

## Scanner Components
1. QuerySetForm
2. QuerySetList
3. ScanRunControls
4. ScannerFailureList
5. ProviderHealthPanel
6. MatchingSummaryPanel

### ScannerFailureList
- Inputs:
  - `failures[]`
- Outputs:
  - `onRetry(querySetId)`
- States:
  - empty, populated, retrying-row
- Tests:
  - `SCN-F-001` retry only enabled with query_set_id

## Reports Components
1. WishlistSummaryCard
2. PricingTrendPanel
3. SourceBreakdownPanel
4. ExportPanel

### ExportPanel
- Inputs:
  - selected scope (`item`, `date range`, `source`)
- Outputs:
  - `onExport(type, filters)`
- Tests:
  - `REP-EXP-001` export disabled with invalid scope

## Settings Components
1. DiagnosticsPanel
2. MaintenancePanel
3. LicensePanel
4. BackupRestorePanel
5. ProfileSettingsForm
6. SecretsForm

### BackupRestorePanel
- Inputs:
  - `backups[]`, `selectedBackupPath`, `confirmRestore`
- Outputs:
  - `onRestore(path)`
- States:
  - list-empty, list-populated, restore-pending, restore-error
- Tests:
  - `SET-BK-001` restore blocked until confirmation checked

## Cross-Component Interaction Contracts
1. Navigation -> Screen mount
- Navigating updates active route and unmounts irrelevant heavy panels.

2. Action feedback
- Any write action must return one of:
  - inline success message
  - inline error with retry

3. Busy-state policy
- Double-submit protection required on mutating actions.

## Accessibility Non-Negotiables
1. All inputs have explicit labels.
2. Dialog/drawer must trap and restore focus.
3. Keyboard-only completion for core workflows.
4. Color is never sole status indicator.

## Component Coverage Targets
- Critical components: 95% behavior coverage.
- Standard components: 85% behavior coverage.
- Every component has at least one error-state test.

