# Chat Copilot Screen Spec

## Use Cases
1. User opens chat from any screen and asks a collection question.
2. User asks assistant to find item gaps or suggest buys.
3. User previews assistant-proposed action and confirms apply.
4. User closes chat and keeps current workspace state unchanged.

## UI Sections
1. Chat rail header (title + close button)
2. Thread transcript list
3. Context chips (active item/candidate/filter)
4. Suggested actions panel
5. Composer (prompt input + send)

## State Behavior
- Loading: assistant thinking indicator.
- Empty: starter prompts and guidance examples.
- Error: provider/permission error with retry + manual fallback links.
- Success: threaded responses with structured action suggestions.

## Acceptance Criteria
- [ ] Chat can open and close without changing current route/screen.
- [ ] Composer is focusable by keyboard shortcut.
- [ ] Mutating actions require explicit preview and confirm.
- [ ] Thread history is restored per profile.
- [ ] Context chips reflect current workspace target when available.

## Test Cases
- `CHAT-001` open/close chat preserves active screen.
- `CHAT-002` send message and receive response.
- `CHAT-003` preview and confirm add-item action.
- `CHAT-004` ensure no mutation when suggestion is not confirmed.
- `CHAT-005` restore previous thread after app reload.
