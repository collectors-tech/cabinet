# Assistant OpenAI Workflow Plan

Status: planning contract for issue #847
Related Cabinet issues: #834, #845, #846, #825
Source review: softlystudio/scheduling-assistant#740, softlystudio/scheduling-assistant#751, softlystudio/scheduling-assistant#770, softlystudio/scheduling-assistant#771

## Review Summary

Cabinet should reuse the product lessons from the Scheduling Assistant OpenAI review without copying its intermediate mistakes. The durable contract is:

- OpenAI configuration must show truthful method state for Browser Auth and API key paths.
- Browser Auth must never be marked connected from navigation alone.
- The UI must not show duplicate method narration such as "OpenAI is using: Browser Auth" when status pills already convey the active method.
- Provider-backed workflows must return setup-needed or unavailable states until credentials, provider health, and the required capability are proven.
- Mutating app actions must flow through the assistant capability registry with preview and explicit confirmation before apply.

The #834 OpenAI config slice covers the first UI/config parity layer. The #845 capability registry slice gives Cabinet a deterministic discovery API. This plan binds the next layer: OpenAI-backed workflow capabilities, DB-backed runs, Telegram intake alignment, and child implementation slices.

## Capability Contracts

Every assistant capability that can call OpenAI or mutate Cabinet state must declare:

- stable capability id
- provider requirements and fallback/setup-needed behavior
- input schema and validation rules
- output schema and preview shape
- execution mode: read-only, preview-only, confirm-required, or unavailable
- confirmation policy
- audit/event output
- result destination and deep link

Initial capability ids:

| Capability | Purpose | Mode |
| --- | --- | --- |
| catalog_add_from_photo | Use item photos to propose a catalog draft with provenance. | confirm-required |
| catalog_add_from_barcode | Use barcode image or typed code lookup to propose a catalog draft. | confirm-required |
| catalog_add_from_text | Extract likely item fields from typed notes and ask follow-up questions when ambiguous. | confirm-required |
| catalog_item_enrich | Enrich an existing item from approved context and source evidence. | confirm-required |
| image_analyze | Read image contents and metadata without mutating records. | preview-only |
| image_process | Produce processed media variants while preserving original evidence. | confirm-required |
| content_generate | Draft catalog/listing copy from approved item context. | preview-only |
| listing_draft_generate | Create a marketplace listing draft from item data and provider constraints. | confirm-required |
| purchase_reconcile | Suggest purchase-to-item matches with confidence and explanations. | preview-only |
| package_reconcile | Suggest package/shipment-to-purchase matches with confidence and explanations. | preview-only |
| provider_test | Test the active provider/method and return truthful readiness evidence. | read-only |
| app_action_preview | Render a structured mutation preview without applying it. | preview-only |
| app_action_apply_after_confirm | Apply a previously confirmed preview and persist audit/result links. | confirm-required |

## Workflow Run Model

OpenAI-backed and agent-backed work should create durable workflow run records rather than opaque chat messages. Each run needs:

- run id, workflow id, capability id, profile id, source channel, source thread/message ids
- status: queued, running, needs_input, completed, failed, cancelled
- progress events and timestamps
- normalized input payload and validated schema version
- preview/result payload and result links
- provider trace: provider id, method, model, request id where available, and setup-needed/unavailable reason when blocked
- error code/message with retry guidance
- confirmation state for mutating actions

The run model must support single-item and bulk runs. Bulk runs must expose per-item progress and failure state so one bad item cannot silently poison the whole batch.

## Telegram Intake Alignment

Telegram is a channel into the same governed capability system, not a separate mutation path.

Telegram-originated photo, barcode, text, or mixed messages must:

- map sender/chat to an authorized Cabinet user/profile before any mutation is possible
- preserve original media/source metadata and Cabinet profile scope
- group mixed photo/barcode/text messages into one draft session when they arrive together or the user indicates they belong together
- create draft previews first and require explicit confirmation before inventory/catalog creation or update
- ask a follow-up question when recognition or lookup is ambiguous
- write an audit trail linking Telegram source, media ids, proposed fields, confirmation decision, and applied item links

## Backlog Breakdown

Recommended implementation slices after #847:

1. #846 Telegram catalog intake shell: authorized text-only draft preview through the #845 capability registry.
2. Catalog barcode/photo draft capabilities: lookup and image analysis providers return setup-needed until credentials and lookup support are proven.
3. Workflow run persistence: queued/running/completed/failed records for image analysis, content generation, provider tests, and app-action previews.
4. Content/listing draft generation: OpenAI-backed copy generation with preview-only output and marketplace/provider constraints.
5. Purchase/package reconciliation skills: confidence-scored suggestions with manual confirmation and audit trail.

## Avoided SCHA Gaps

- No static ChatGPT/OpenAI login URL is treated as proof of Browser Auth readiness.
- No connected state from outbound navigation alone.
- No duplicate active-method copy where pills/labels already show method state.
- No provider test passes merely because a credential string exists.
- No direct mutation path from chat or Telegram without preview plus explicit confirmation.
