# Cabinet OpenSpec Hierarchy

This directory is organized by UI section hierarchy.

## Top-level Sections
- `general`
- `dashboard`
- `inventory`
- `wishlist`
- `integrations`
- `chats`
- `users`
- `settings`
- `helpcenter`

Each section includes:
- `README.md` with purpose and sub-spec listing
- sub-spec folders containing `spec.md`
- section-root `spec.md` only where already existing (`integrations`, `settings`)

## Notes
- Requirement IDs remain unchanged (append-only discipline).
- Legacy flat spec paths were moved into section subfolders and all OpenSpec references were updated.
- Coverage remains intact across all previously existing capabilities.
- Feature modeling rule: one capability spec per feature; avoid multi-feature bundled specs.
- Documentation governance capability: `openspec/specs/general/documentation-governance/spec.md`.
