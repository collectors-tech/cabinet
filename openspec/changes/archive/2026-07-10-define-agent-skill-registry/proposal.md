## Why

Cabinet now has governed assistant capabilities, guided workflows, UI targets, integration workflows, shell command bus operations, preview/apply boundaries, Action Timeline records, and audit evidence. Those pieces are still low-level machinery from the user's point of view.

Agent Skills need to become the first-class user-facing model that explains what Cabinet Agent can do, what safety boundary applies, whether setup is required, and how built-in and imported skills are managed without allowing archive imports to bypass existing governance.

## What Changes

- Define an Agent Skill Registry OpenSpec capability for skill identity, versioning, source type, status lifecycle, safety levels, permissions, context requirements, and lower-level binding metadata.
- Define local skill archive structure and validation rules for future `.cabinet-skill.zip` imports.
- Define installed skill enable/disable state, safe failure states, and Action Timeline/audit requirements.
- Define Skills page list, detail, import, and marketplace-deferred behavior.
- Link the specification to the implementation and documentation child issues.

## Capabilities

### New Capabilities

- `agent-skills-registry`: first-class Agent Skill Registry, local archive import contract, installed skill lifecycle, and Skills page behavior.

## Impact

- Affected specs: `openspec/changes/define-agent-skill-registry/specs/agent-skills-registry/spec.md`
- Affected traceability: `openspec/traceability.md`
- Related issues: #1666, #1667, #1668, #1669, #1670, #1671, #1672
