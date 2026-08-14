
# Agent acquisition workflow evidence

Issue: #2083
Parent: #1701
Status: incremental backend evidence; the affected coverage-matrix rows remain `partial` until the unified Chat UI and live-provider gates are complete.

## Natural-language routing

Main and contextual Chat acquisition requests are eligible for the configured assistant-provider planner when they combine a supported action with an Integrations, provider, Market Watch, Discoveries/listing, or Purchases/order concept. Generic background handoff phrases do not take precedence over a recognized acquisition request.

Evidence:

- `internal/app/chat_agent_planner_test.go` — `TestChatAgentPlannerRecognisesDiscoveryAndAcquisitionLanguage`
- `ui.web/cypress/e2e/chats/assistant-acquisition-workflows/spec.cy.ts` — `AGENT-ACQUISITION-001/#2083`

## Provider truth

Planner failures preserve a redacted provider error class and setup-next-action. The public planner error uses `assistant_provider_<class>` for known setup/runtime categories, while raw provider error detail and credentials remain absent.

The acquisition preview contract returned in `agent_planner.preview_result` is:

```json
{
  "preview_kind": "agent_skill",
  "skill_id": "cabinet.discoveries.send_to_wishlist",
  "status": "previewed",
  "preview_only": true,
  "mutation_applied": false,
  "confirmation_required": true,
  "source_surface": "discoveries.result.card",
  "source_channel": "in-app",
  "target": {
    "provider_id": "ebay",
    "result_id": "disc-2083",
    "source_url": "https://example.test/listing/2083"
  },
  "parameters": {},
  "apply_endpoint": "/api/agent/skills/apply",
  "apply_request": {
    "confirm": true
  }
}
```

The actual payload contains the profile, thread/message provenance, and non-secret planner parameters. Secret-like keys are removed. The UI MUST present the preview and obtain explicit user confirmation before sending `apply_request`; it MUST NOT dispatch the request merely because it is present in the response. An `authority_blocker` such as `agent_authority_external_write_not_approved` makes apply unavailable until the profile authority policy permits it.

Evidence:

- `internal/app/chat_agent_planner_test.go` — `TestChatAgentPlannerPreservesProviderSetupFailureTaxonomy`
- `internal/app/chat_agent_planner_test.go` — `TestChatAgentPlannerPreviewsDiscoveryHandoffWithoutMutation`
- `internal/app/chat_agent_planner_test.go` — `TestChatAgentPlannerPreviewsExternalWriteSelectionWithoutApplyAuthority`

## Remaining gates

- #2081 must consume `agent_planner.preview_result` from both main and contextual Chat and render the shared confirmation surface.
- Successful live-provider search, timeout, login-needed, partial-result, and retry evidence still require safe provider credentials or controlled provider fixtures.
- Cancel, duplicate-delivery, confirmed acquisition apply, and packaged-runtime coverage remain open under #2083.
- Integrations, Market Watch, Discoveries, and Purchases remain `partial` in `agent-skill-coverage.md` until those gates are proven.
