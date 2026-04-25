# TASK_LIST

Last synced from GitHub issues: 2026-04-25 22:55 +10:00

## Status Legend
- `OPEN`: issue is open and not yet complete.
- `DONE`: issue is closed and merged or otherwise resolved.
- Work is executed one focused issue branch at a time, validated, merged into `develop`, then the demo is rebuilt and restarted.

## Current Execution Queue
- [ ] #665 [OPEN] Workspace collection deletes should remove stale options from compact filters
- [ ] #643 [OPEN] Collections: implement two-table selected collection layout
- [ ] #644 [OPEN] Collections: remove duplicate Selected collection section
- [ ] #645 [OPEN] Collections: convert Collection members section into items table
- [ ] #637 [OPEN] Epic: Redesign Inventory as a table-first collection browser
- [x] #640 [DONE] Move Inventory photos into item-scoped modal with row quick action
- [x] #641 [DONE] Move Inventory barcodes into item-scoped modal with row quick action
- [x] #642 [DONE] Add responsive Cypress coverage for Inventory table-first redesign
- [x] #646 [DONE] Inventory: add assign-to-collection action on each item row
- [x] #636 [DONE] Make inventory collection-browser context menu actions visibly effective
- [x] #635 [DONE] Implement real mobile camera photo capture for inventory items
- [x] #634 [DONE] Unify page header actions and reclaim workspace height

## CI, Runtime, And Repository Hygiene
- [ ] #553 [OPEN] Consolidate duplicated main gate GitHub Actions workflows
- [ ] #555 [OPEN] Stabilize main gate CI failures on release PR
- [ ] #551 [OPEN] chore(repo): archive stale non-archive branches
- [ ] #547 [OPEN] [Blocker] UI journeys regression runner path is stale
- [ ] #546 [OPEN] [Blocker] Runtime pre-canceled Run fast-exit test flake
- [ ] #545 [OPEN] [Blocker] Onboarding sample-data category coverage regression
- [ ] #544 [OPEN] [Blocker] Commerce lifecycle tests red in app suite
- [ ] #543 [OPEN] [Blocker] Chat thread schema mismatch: chat_threads missing metadata_json
- [ ] #542 [OPEN] [Blocker] OpenAPI gate fails: missing required 4XX response contracts
- [ ] #521 [OPEN] chore(ui): reduce lint and format baseline blocking focused PR validation
- [ ] #518 [OPEN] bug(runtime): parallel app startup hits migration timeouts and canceled runs miss fast-exit path
- [ ] #480 [OPEN] test(runtime): /api/test/reset fails on shared managed DB when schema table is missing

## Product Follow-Up
- [ ] #537 [OPEN] feat(collections): convert collections screen into shared-table management surface
- [ ] #526 [OPEN] feat(shell): set document title to Cabinet - <Page Title>
- [ ] #513 [OPEN] bug(inbox): inbox workspace empty state exposes no actionable affordances
- [ ] #487 [OPEN] bug(inbox): top-level /inbox route resolves to 404
- [ ] #484 [OPEN] bug(collections): page shows duplicate create affordances in same view
- [ ] #479 [OPEN] OpenAI integration: browser-login/provider-runtime follow-up beyond stub UX

## Integrations Follow-Up
- [ ] #512 [OPEN] bug(integrations): OpenAI validate reports success even when token is missing
- [ ] #511 [OPEN] bug(integrations): empty-token save validation is not clearly bound to the token field
- [ ] #510 [OPEN] bug(integrations): needs-config OpenAI connect dialog lacks a clear labeled save-first setup flow
- [ ] #509 [OPEN] bug(integrations): needs-config connect dialog exposes Validate and Sync before setup is completed
- [ ] #508 [OPEN] bug(integrations): provider dialog exposes inert Sync action without visible result
- [ ] #507 [OPEN] bug(integrations): validate success feedback does not update visible provider health state
- [ ] #506 [OPEN] bug(integrations): provider edit dialog inputs have no visible or programmatic labels
- [ ] #505 [OPEN] bug(integrations): route leaks raw integrations.title and integrations.description keys

## Exploration And Traceability
- [ ] #502 [OPEN] [Exploration] Traceability / backlog closure pass
- [ ] #501 [OPEN] [Exploration] Cross-cutting UX audit and traceability pass
- [ ] #500 [OPEN] [Exploration] Runtime / operational states audit and traceability pass
- [ ] #499 [OPEN] [Exploration] Settings / preferences audit and traceability pass
- [ ] #498 [OPEN] [Exploration] Inbox / communications audit and traceability pass
- [ ] #497 [OPEN] [Exploration] Assistant / AI surfaces audit and traceability pass
- [ ] #496 [OPEN] [Exploration] Integrations audit and traceability pass
- [ ] #495 [OPEN] [Exploration] Tasks / operational surfaces audit and traceability pass
- [ ] #493 [OPEN] [Exploration] Collections audit and traceability pass
- [ ] #492 [OPEN] [Exploration] Item detail / workflow audit and traceability pass
- [ ] #489 [OPEN] [Exploration] App shell / navigation audit and traceability pass
- [ ] #488 [OPEN] [Exploration] Public / entry audit and traceability pass
- [ ] #189 [OPEN] Cabinet Continuous Implementation Program - Wave Execution
