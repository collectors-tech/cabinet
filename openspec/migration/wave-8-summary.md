# Wave 8 Summary

- wave number: 8
- scope: high-value UI/E2E partial closure (onboarding, settings, chat-copilot, shell/nav, photos fullscreen)
- issue: #189
- status: completed

## IDs moved to implemented
- `UI-SCREEN-ONBOARDING-AUTH-001`
- `UI-SCREEN-ONBOARDING-AUTH-002`
- `UI-SCREEN-ONBOARDING-AUTH-003`
- `UI-SCREEN-SETTINGS-001`
- `UI-SCREEN-SETTINGS-002`
- `UI-SCREEN-SETTINGS-003`
- `UI-SCREEN-CHAT-COPILOT-001`
- `UI-SCREEN-CHAT-COPILOT-002`
- `UI-SCREEN-CHAT-COPILOT-003`
- `UI-FOUNDATION-SHELL-NAVIGATION-001`
- `UI-FOUNDATION-SHELL-NAVIGATION-002`
- `UI-FOUNDATION-SHELL-NAVIGATION-003`
- `UI-FOUNDATION-SHELL-NAVIGATION-004`
- `PHOTOS-MEDIA-004`
- `CHAT-COPILOT-001`

## IDs still partial + blockers
- `AI-ASSIST-003`: Runtime mutation-apply endpoint for AI suggestions with deterministic `409` + `error_code="AI_CONFIRM_REQUIRED"` is not implemented.

## Runtime/UI behavior implemented
- Added onboarding auth-requirements failure handling with explicit retry CTA in starter wizard.
- Added profile-scoped chat history persistence and deterministic chat state handling (`loading`, `error`, `ready`).
- Added fullscreen photo navigation controls (previous/next) plus ArrowLeft/ArrowRight keyboard support.

## Test commands and results
- `npm run e2e:playwright -- playwright/e2e/ui-backlog.spec.ts` -> pass (`20 passed`)
- `openspec validate --all` -> pass (`57 passed, 0 failed`)

## Net counts
- partial before -> after: `90 -> 75`
- implemented before -> after: `76 -> 91`
- reduction: `15`

## Top 5 remaining risk IDs
- `AI-ASSIST-003`
- `COLLECTION-DOMAIN-004`
- `LOGGING-001`
- `LOGGING-002`
- `UI-FOUNDATION-THEME-RTL-I18N-001`

## Commit
- commit: `<pending>`
