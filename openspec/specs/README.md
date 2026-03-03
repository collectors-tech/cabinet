# Cabinet OpenSpec Hierarchy

This directory is organized by UI section hierarchy.

## Top-level Sections
- `general`
- `dashboard`
- `inventory`
- `wishlist`
- `integrations`
- `media`
- `chats`
- `users`
- `settings`
- `helpcenter`
- `collections`

Each section includes:
- `README.md` with purpose and sub-spec listing
- sub-spec folders containing `spec.md`
- screen-level specs split per route/screen (one spec per screen)
- section-root `spec.md` only for cross-screen/domain contracts

## Notes
- Requirement IDs remain unchanged (append-only discipline).
- Legacy flat spec paths were moved into section subfolders and all OpenSpec references were updated.
- Coverage remains intact across all previously existing capabilities.
- Feature modeling rule: one capability spec per feature; avoid multi-feature bundled specs.
- Documentation governance capability: `openspec/specs/general/documentation-governance/spec.md`.
