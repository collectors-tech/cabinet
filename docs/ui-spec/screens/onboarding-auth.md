# Onboarding and Authentication Screen Spec

## Use Cases
1. First-time user creates profile and identity quickly.
2. User completes guided onboarding without form overload.
3. Returning user authenticates with WebAuthn and enters workspace.

## UI Sections
1. Profile selection/create
2. Identity setup/login actions
3. Starter onboarding wizard (5 steps)
4. Completion/unlock action

## State Behavior
- Loading: profiles and auth requirements loading.
- Empty: no profiles with create CTA.
- Error: auth/onboarding operation errors with retry.
- Success: completed onboarding + workspace unlock.

## Onboarding Steps (strict)
1. Welcome/setup path
2. Identity
3. Starter data choice
4. First item quick add
5. Preferences + finish

## Acceptance Criteria
- [ ] Navigation remains locked until onboarding completion.
- [ ] Identity step blocks progression until success.
- [ ] Starter data flow supports sample and empty paths.
- [ ] Finish action persists onboarding complete state per profile.
- [ ] WebAuthn begin/finish flows show deterministic session/status feedback.

## Test Cases
- `ONB-001` first profile creation flow.
- `ONB-002` identity failure and retry behavior.
- `ONB-003` sample data path and idempotent rerun.
- `ONB-004` finish onboarding unlocks workspace.
- `AUTH-001` begin and finish WebAuthn registration.
- `AUTH-002` begin and finish WebAuthn login.

