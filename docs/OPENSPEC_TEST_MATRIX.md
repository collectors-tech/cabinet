# OpenSpec Test Matrix

Last updated: 2026-02-26  
Tracking issue: `#170`

## Purpose
Map OpenSpec UI use-cases to requirements and automated test coverage state.

## CI-Friendly Coverage Summary
```text
total_uc=50
existing_test_mappings=6
planned_only_mappings=44
unmapped_uc=0
follow_up_issue=#172
```

## Existing UI E2E Files (Current)
- `web/cypress/e2e/smoke.cy.ts`
- `web/cypress/e2e/profile-onboarding.cy.ts`
- `web/cypress/e2e/starter-onboarding.cy.ts`
- `web/cypress/e2e/inventory-non500.cy.ts`
- `web/cypress/e2e/ui-matrix.cy.ts`
- `web/cypress/e2e/shell-layout.cy.ts`

## UC -> Requirement -> Test Mapping
| UC ID | OpenSpec Screen Spec | Requirement Focus | Test Path + ID | Coverage Status |
| --- | --- | --- | --- | --- |
| UC-HOME-01 | `ui-screen-home` | command-center ready render | `web/cypress/e2e/smoke.cy.ts` `smoke-home-load` | Existing |
| UC-HOME-02 | `ui-screen-home` | empty state guidance | `web/cypress/e2e/home.cy.ts` `home-empty-state` | Planned |
| UC-HOME-03 | `ui-screen-home` | error + retry | `web/cypress/e2e/home.cy.ts` `home-error-retry` | Planned |
| UC-HOME-04 | `ui-screen-home` | quick action routing | `web/cypress/e2e/home.cy.ts` `home-quick-action-routing` | Planned |
| UC-ONB-01 | `ui-screen-onboarding-auth` | first-run WebAuthn gate | `web/cypress/e2e/profile-onboarding.cy.ts` `onboarding-webauthn-gate` | Existing |
| UC-ONB-02 | `ui-screen-onboarding-auth` | resume onboarding | `web/cypress/e2e/starter-onboarding.cy.ts` `onboarding-resume` | Existing |
| UC-ONB-03 | `ui-screen-onboarding-auth` | onboarding empty state | `web/cypress/e2e/starter-onboarding.cy.ts` `onboarding-empty-state` | Existing |
| UC-ONB-04 | `ui-screen-onboarding-auth` | auth error handling | `web/cypress/e2e/profile-onboarding.cy.ts` `onboarding-auth-error` | Planned |
| UC-INV-01 | `ui-screen-inventory-items` | inventory ready render | `web/cypress/e2e/inventory.cy.ts` `inventory-ready` | Planned |
| UC-INV-02 | `ui-screen-inventory-items` | empty non-500 state | `web/cypress/e2e/inventory-non500.cy.ts` `inventory-empty-non500` | Existing |
| UC-INV-03 | `ui-screen-inventory-items` | error non-500 state | `web/cypress/e2e/inventory-non500.cy.ts` `inventory-load-error-non500` | Existing |
| UC-INV-04 | `ui-screen-inventory-items` | row opens details | `web/cypress/e2e/inventory.cy.ts` `inventory-row-opens-details` | Planned |
| UC-INV-05 | `ui-screen-inventory-items` | checkbox bulk mode | `web/cypress/e2e/inventory.cy.ts` `inventory-bulk-checkbox-mode` | Planned |
| UC-PHO-01 | `ui-screen-inventory-photos` | photo upload | `web/cypress/e2e/photos.cy.ts` `photo-upload` | Planned |
| UC-PHO-02 | `ui-screen-inventory-photos` | set primary photo | `web/cypress/e2e/photos.cy.ts` `photo-set-primary` | Planned |
| UC-PHO-03 | `ui-screen-inventory-photos` | photos empty state | `web/cypress/e2e/photos.cy.ts` `photo-empty-state` | Planned |
| UC-PHO-04 | `ui-screen-inventory-photos` | photos error state | `web/cypress/e2e/photos.cy.ts` `photo-error-state` | Planned |
| UC-PHO-05 | `ui-screen-inventory-photos` | fullscreen view | `web/cypress/e2e/photos.cy.ts` `photo-fullscreen` | Planned |
| UC-BAR-01 | `ui-screen-inventory-barcodes` | add barcode | `web/cypress/e2e/barcodes.cy.ts` `barcode-add` | Planned |
| UC-BAR-02 | `ui-screen-inventory-barcodes` | local lookup | `web/cypress/e2e/barcodes.cy.ts` `barcode-local-lookup` | Planned |
| UC-BAR-03 | `ui-screen-inventory-barcodes` | external fallback | `web/cypress/e2e/barcodes.cy.ts` `barcode-external-fallback` | Planned |
| UC-BAR-04 | `ui-screen-inventory-barcodes` | error + retry | `web/cypress/e2e/barcodes.cy.ts` `barcode-error-state` | Planned |
| UC-AI-01 | `ui-screen-inventory-ai-assist` | suggest from title | `web/cypress/e2e/ai-assist.cy.ts` `ai-title-suggest` | Planned |
| UC-AI-02 | `ui-screen-inventory-ai-assist` | suggest from photo | `web/cypress/e2e/ai-assist.cy.ts` `ai-photo-suggest` | Planned |
| UC-AI-03 | `ui-screen-inventory-ai-assist` | guarded apply | `web/cypress/e2e/ai-assist.cy.ts` `ai-guarded-apply` | Planned |
| UC-AI-04 | `ui-screen-inventory-ai-assist` | error + retry | `web/cypress/e2e/ai-assist.cy.ts` `ai-error-state` | Planned |
| UC-SCN-01 | `ui-screen-scanner` | queryset CRUD | `web/cypress/e2e/scanner.cy.ts` `scanner-queryset-crud` | Planned |
| UC-SCN-02 | `ui-screen-scanner` | run now | `web/cypress/e2e/scanner.cy.ts` `scanner-run-now` | Planned |
| UC-SCN-03 | `ui-screen-scanner` | retry failures | `web/cypress/e2e/scanner.cy.ts` `scanner-retry-failure` | Planned |
| UC-SCN-04 | `ui-screen-scanner` | scanner empty state | `web/cypress/e2e/scanner.cy.ts` `scanner-empty-state` | Planned |
| UC-SCN-05 | `ui-screen-scanner` | scanner error state | `web/cypress/e2e/scanner.cy.ts` `scanner-error-state` | Planned |
| UC-DIS-01 | `ui-screen-discover` | filtered candidates | `web/cypress/e2e/discover.cy.ts` `discover-filtering` | Planned |
| UC-DIS-02 | `ui-screen-discover` | ignore action | `web/cypress/e2e/discover.cy.ts` `discover-ignore` | Planned |
| UC-DIS-03 | `ui-screen-discover` | wishlist action | `web/cypress/e2e/discover.cy.ts` `discover-wishlist` | Planned |
| UC-DIS-04 | `ui-screen-discover` | track/create action | `web/cypress/e2e/discover.cy.ts` `discover-track-create` | Planned |
| UC-DIS-05 | `ui-screen-discover` | discover error state | `web/cypress/e2e/discover.cy.ts` `discover-error-state` | Planned |
| UC-REP-01 | `ui-screen-reports` | reports ready state | `web/cypress/e2e/reports.cy.ts` `reports-ready` | Planned |
| UC-REP-02 | `ui-screen-reports` | reports export | `web/cypress/e2e/reports.cy.ts` `reports-export` | Planned |
| UC-REP-03 | `ui-screen-reports` | reports empty state | `web/cypress/e2e/reports.cy.ts` `reports-empty-state` | Planned |
| UC-REP-04 | `ui-screen-reports` | reports error state | `web/cypress/e2e/reports.cy.ts` `reports-error-state` | Planned |
| UC-SET-01 | `ui-screen-settings` | settings save/persist | `web/cypress/e2e/settings.cy.ts` `settings-save-persist` | Planned |
| UC-SET-02 | `ui-screen-settings` | backup/list | `web/cypress/e2e/settings.cy.ts` `settings-backup-run-list` | Planned |
| UC-SET-03 | `ui-screen-settings` | restore | `web/cypress/e2e/settings.cy.ts` `settings-restore` | Planned |
| UC-SET-04 | `ui-screen-settings` | diagnostics actions | `web/cypress/e2e/settings.cy.ts` `settings-diagnostics` | Planned |
| UC-SET-05 | `ui-screen-settings` | settings error state | `web/cypress/e2e/settings.cy.ts` `settings-error-state` | Planned |
| UC-CHAT-01 | `ui-screen-chat-copilot` | list threads | `web/cypress/e2e/chat.cy.ts` `chat-thread-list` | Planned |
| UC-CHAT-02 | `ui-screen-chat-copilot` | reopen history | `web/cypress/e2e/chat.cy.ts` `chat-thread-history` | Planned |
| UC-CHAT-03 | `ui-screen-chat-copilot` | attachment add | `web/cypress/e2e/chat.cy.ts` `chat-attachment` | Planned |
| UC-CHAT-04 | `ui-screen-chat-copilot` | guarded apply | `web/cypress/e2e/chat.cy.ts` `chat-guarded-apply` | Planned |
| UC-CHAT-05 | `ui-screen-chat-copilot` | chat error state | `web/cypress/e2e/chat.cy.ts` `chat-error-state` | Planned |

## Unmapped UC Tracker
Definition:
- `Unmapped` means UC has no test mapping at all.
- `Planned` means mapping exists but test file/test id does not yet exist.

Current state:
- `unmapped_uc=0`
- `planned_only_mappings=44`

Follow-up issue for planned-only gaps:
- `#172` `[Testing] Implement planned E2E suites from OpenSpec UC matrix`
