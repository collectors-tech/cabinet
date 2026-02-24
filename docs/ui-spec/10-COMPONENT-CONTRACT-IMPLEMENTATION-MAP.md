# 10 Component Contract Implementation Map

## Purpose
Provide a concrete map from `09-COMPONENT-SPECS-STRICT.md` contracts to implementation files and executable tests.

## Global Primitives

| Contract | Implementation | Tests |
|---|---|---|
| Button (`variant`, `size`, `disabled`, `loading`) | `web/src/components/ui-primitives.tsx` (`Button`) | `web/src/components/ui-primitives-contracts.test.tsx` (`BTN-001`, `BTN-002`) |
| Text Input (`invalid`, `message`) | `web/src/components/ui-primitives.tsx` (`TextField`) | `web/src/components/ui-primitives-contracts.test.tsx` (`INP-001`) |
| Select | `web/src/components/ui-primitives.tsx` (`SelectField`) | `web/src/components/ui-primitives-contracts.test.tsx` (`SEL-001`) |
| Modal/Dialog | `web/src/components/ui-primitives.tsx` (`Dialog`) | `web/src/components/ui-primitives-contracts.test.tsx` (`DIA-001`, `DIA-002`) |
| Drawer | `web/src/components/ui-primitives.tsx` (`Drawer`) | `web/src/components/ui-primitives-contracts.test.tsx` (`DRW-001`, `DRW-002`) |

## Per-Screen Contracts (strict coverage set)

| Contract | Implementation | Tests |
|---|---|---|
| QuickAddItemForm | `web/src/components/screen-components.tsx` (`QuickAddItemForm`) | `web/src/components/screen-components-contracts.test.tsx` (`INV-QF-001`, `INV-QF-002`) |
| ItemListTable | `web/src/components/screen-components.tsx` (`ItemListTable`) | `web/src/components/screen-components-contracts.test.tsx` (`INV-TBL-001`) |
| PhotoUploadPanel | `web/src/components/screen-components.tsx` (`PhotoUploadPanel`) | `web/src/components/screen-components-contracts.test.tsx` (`PHO-UP-001`, `PHO-UP-002`) |
| BarcodeLookupResult | `web/src/components/screen-components.tsx` (`BarcodeLookupResult`) | `web/src/components/screen-components-contracts.test.tsx` (`BAR-RES-001`, `BAR-RES-002`) |
| AISuggestionPreview | `web/src/components/screen-components.tsx` (`AISuggestionPreview`) | `web/src/components/screen-components-contracts.test.tsx` (`AI-PRV-001`, `AI-PRV-002`) |
| DiscoveryCandidateRowActions | `web/src/components/screen-components.tsx` (`DiscoveryCandidateRowActions`) | `web/src/components/screen-components-contracts.test.tsx` (`DIS-ACT-001`) |
| ScannerFailureList | `web/src/components/screen-components.tsx` (`ScannerFailureList`) | `web/src/components/screen-components-contracts.test.tsx` (`SCN-F-001`) |
| ExportPanel | `web/src/components/screen-components.tsx` (`ExportPanel`) | `web/src/components/screen-components-contracts.test.tsx` (`REP-EXP-001`) |
| BackupRestorePanel | `web/src/components/screen-components.tsx` (`BackupRestorePanel`) | `web/src/components/screen-components-contracts.test.tsx` (`SET-BK-001`) |

## Accessibility Non-Negotiables Status

- Inputs with explicit labels: covered via `aria-label` and label associations in primitives and existing forms.
- Dialog/drawer semantics: `role="dialog"` + `aria-modal="true"` in primitives and app shell mobile drawer.
- Keyboard core workflows: existing app tests for nav/editor/workspace plus primitive contract tests.
- Status not color-only: explicit text labels are rendered for state chips/statuses across screens.

## Notes

- Existing feature-level flows remain covered in `web/src/App.test.tsx`.
- Strict primitive and component contract tests are additive and intended as regression guards for issue `#102`.
