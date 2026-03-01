# Auto Wave 21 Summary

- Issue: #152
- Scope: remove template/sample placeholder copy from user-visible UI and add banned-phrase regression contract.
- Date: 2026-03-02

## Changes delivered
- Replaced template-origin branding/content in visible UI:
  - app title/subtitle
  - auth layout and sign-in variant headings/alt text
  - Clerk auth panel quote/help copy
  - profile username placeholder
  - HTML head/meta title and descriptions
- Replaced social preview image reference to `cabinet-ui.png`.
- Added automated contract test for banned placeholder strings:
  - `tests/ui_placeholder_copy_contract_test.go`
- Updated backend root-shell test to assert Cabinet title and no template metadata phrase.

## Manual sweep completed (updated files)
- `ui.web/index.html`
- `ui.web/src/components/layout/app-title.tsx`
- `ui.web/src/features/auth/auth-layout.tsx`
- `ui.web/src/features/auth/sign-in/sign-in-2.tsx`
- `ui.web/src/routes/clerk/(auth)/route.tsx`
- `ui.web/src/assets/logo.tsx`
- `ui.web/src/features/settings/profile/profile-form.tsx`
- `internal/app/ui_root_test.go`
- `tests/ui_placeholder_copy_contract_test.go`
- `ui.web/public/images/cabinet-ui.png`
- regenerated embedded UI assets under `internal/ui/static/**`

## Commands run
1. `npm run build` (workdir: `ui.web`)
2. `npm run e2e:run-smoke` (workdir: `ui.web`)
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`

## Results
- `npm run e2e:run-smoke`: **pass** (`3 passing, 0 failing`)
- `go test ./internal/app -count=1`: **pass**
- `go test ./tests -count=1`: **pass**
- `openspec validate --all`: **pass**
