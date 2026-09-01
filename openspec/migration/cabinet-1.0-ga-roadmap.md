# Cabinet 1.0 General Availability Roadmap

Status: active OpenSpec program plan  
Canonical tracking issue: [#2546](https://github.com/collectors-tech/cabinet/issues/2546)  
Project: [collectors-tech Project #2](https://github.com/orgs/collectors-tech/projects/2)  
Baseline date: 2026-09-01  
Baseline `develop`: `46f4640882bc3d3ba0b2f6225d26ff12bc2c3b10`

## Purpose

This roadmap defines the smallest truthful path from Cabinet's current private
beta to 1.0 General Availability (GA). GA does not require completing the whole
backlog. It requires a deliberately bounded product contract, repeatable
packaged evidence, a supportable distribution, and explicit owner approval for
the exact release.

GitHub issues are the implementation source of truth. This document describes
the release strategy and exit gates; [#2546](https://github.com/collectors-tech/cabinet/issues/2546)
and its linked issues record current execution state.

## GA product contract

### Supported in 1.0

- Windows desktop collector workspace.
- Dashboard, Inventory, Wishlist, Collections, Media, search, import/export,
  backup and recovery.
- Clean onboarding and profile setup without developer assistance.
- A deliberately small, named set of provider workflows that pass both live
  and packaged acceptance.
- Safe upgrades, rollback/recovery, actionable diagnostics, documentation,
  privacy and support processes.
- Chat and Agent workflows only where the packaged journey is independently
  proven. Other assistant capabilities must be labelled Preview.

### Preview or post-1.0 by default

- Frontline and Bonza if Cabinet cannot accept the support obligations of their
  browser-assisted integrations.
- Broader Agent and Telegram capabilities.
- Metadata Studio.
- eBay seller, Shopify, crawler and parts-finder expansion.
- Trust and peer-to-peer features.
- Companion side-panel polish unless beta evidence shows it is necessary for a
  supported provider journey.

Moving an item into or out of the GA contract requires an owner decision and an
update to #2546 and the affected release issue. Scope must not change silently
to make a gate appear green.

## Starting position

| Gate | 2026-09-01 position | Required next evidence |
| --- | --- | --- |
| Source qualification | Exact-head hosted validation passed 146/146 UI specifications, 742 tests, and zero failures in [run 33178615689](https://github.com/collectors-tech/cabinet/actions/runs/33178615689). | Re-run at the frozen candidate SHA. A scheduled `no-change` result is not a substitute. |
| First run | A clean production-mode source run returned `setup_complete_failed_401` when choosing **Use Defaults**. | Reproduce against the exact packaged candidate under [#1946](https://github.com/collectors-tech/cabinet/issues/1946), then remediate test-first if confirmed. |
| Live providers | Frontline [#1944](https://github.com/collectors-tech/cabinet/issues/1944) and Bonza [#1945](https://github.com/collectors-tech/cabinet/issues/1945) lack user-present browser proof. | Complete the lawful live journeys or explicitly classify the providers as Preview/post-1.0 and update the release contract. |
| Candidate artifacts | The latest published package is [v0.1.0-beta.9](https://github.com/collectors-tech/cabinet/releases/tag/v0.1.0-beta.9), behind the baseline `develop` SHA. | Freeze and package an exact-current candidate under [#1868](https://github.com/collectors-tech/cabinet/issues/1868). |
| SBOM and provenance | [#2550](https://github.com/collectors-tech/cabinet/issues/2550) is complete. Private package [run 33508229009](https://github.com/collectors-tech/cabinet/actions/runs/33508229009) retained and independently verified a manifest-bound CycloneDX 1.7 SBOM plus build and SBOM attestations for exact `develop` commit `4a41a415b0ee0f30d4052947cc395029062f6d11`. | Reuse this gate for each frozen candidate. This evidence does not replace packaged product, provider, recovery, legal, approval or publication gates. |
| Packaged acceptance | Exact-current Companion, collector workflow and recovery acceptance are incomplete. | Complete [#2034](https://github.com/collectors-tech/cabinet/issues/2034), [#1869](https://github.com/collectors-tech/cabinet/issues/1869) and [#1867](https://github.com/collectors-tech/cabinet/issues/1867) against one candidate. |
| Dependency security | The local production gate reports critical/high/moderate `0/0/0` and low `1`; GitHub still reports open medium/low advisories. | Reconcile every advisory against the exact lock, installed production tree and packaged contents. Record evidence for fixes or dismissals. |
| Legal and support | Entity, privacy, retention, licensing, consumer and support decisions remain open in [#2057](https://github.com/collectors-tech/cabinet/issues/2057). | Record owner decisions and verify the exact candidate documentation. |
| Publication | No exact-current approved release exists. | Obtain exact-commit approval, publish, independently redownload and replay, then re-read release, issue and Project state. |

## Delivery roadmap

The durations below are planning ranges, not commitments. Calendar dates depend
on team capacity, user-present provider sessions, signing arrangements and owner
availability. With one focused release lane, the earliest credible GA window is
12–16 weeks after this baseline.

### Phase 0 — Define and govern GA (week 1)

Outcomes:

- Owner-approved platform, provider, Chat/Agent and distribution contracts.
- `Beta Exit`, `Validated Beta`, `0.9 RC`, `1.0 GA` and `Post-1.0`
  milestones in GitHub.
- Every open issue assigned to GA-critical, beta-evidence, Preview or post-1.0.
- Project #2 reconciled with repository issues, including status and priority.

Exit gate:

- #2546 records the approved product contract and milestone membership.
- No P0/P1 GA item is unowned, unmilestoned or missing an acceptance gate.

### Phase 1 — Publish an exact-current beta (weeks 2–4)

Execute this dependency order:

1. Complete or explicitly rescope Frontline #1944 and Bonza #1945.
2. Resolve owner/legal decisions in #2057.
3. Reproduce and remediate clean-start onboarding under #1946.
4. Re-fetch `origin/develop` and freeze an exact candidate SHA.
5. Produce Cabinet and Companion artifacts under #1868.
6. Install, pair and revoke the Chrome and Edge packages under #2034.
7. Complete packaged collector and provider acceptance under #1869.
8. Complete same-candidate media, relocation and recovery evidence under #1867.
9. Obtain exact approval under #1864/#2488 and publish.
10. Independently redownload, checksum and replay the published assets.

Exit gate:

- A public beta from one exact SHA passes clean-start, core collector,
  supported-provider, browser package, restart, backup and recovery acceptance.
- The release and retained evidence can be downloaded and replayed independently.

### Phase 2 — Validated beta (weeks 5–9)

Run an invited-user cohort against published packages, not development servers.
Capture only consented evidence and avoid treating fixture success as user proof.

Measure:

- onboarding completion and time to first useful collector action;
- create/edit/find flows across Inventory, Wishlist and Collections;
- supported-provider success and truthful fail-closed behavior;
- restart, upgrade, backup and restore success;
- actionable diagnostic and support outcomes;
- accessibility, information density and large-library responsiveness.

Exit gate:

- Two consecutive exact candidates pass required source and package gates.
- No open P0/P1 defect affects the GA contract.
- Beta evidence, rather than backlog age, determines remaining GA work.

### Phase 3 — Release candidate hardening (weeks 10–13)

Required work:

- Decide the distribution contract: a signed installer is preferred; a
  portable-only release requires explicit approval and truthful documentation.
- Provision and validate code-signing and update-signing keys.
- Test install, uninstall, upgrade, rollback and supported data migrations.
- Reproduce the completed SBOM and provenance gate from #2550 for the immutable RC.
- Reconcile all security advisories against the exact packaged contents.
- Complete threat-model review, corrupted-input handling and recovery exercises.
- Finalize privacy, legal, licensing, support and incident-response material.
- Qualify performance and accessibility on supported Windows hardware.

Exit gate:

- One immutable RC has complete source, artifact, package, recovery, security,
  legal and support evidence.
- Zero open P0/P1 GA defects.

### Phase 4 — GA publication (weeks 14–16)

Required work:

- Soak the immutable RC for 7–14 days without replacing its artifacts.
- Re-fetch all live issue, Project, CI, artifact and approval state.
- Obtain explicit owner approval for the exact commit and assets.
- Publish 1.0 GA and independently redownload, checksum, install and replay it.
- Re-read the release, canonical issues and archived Project items.

Exit gate:

- The published artifacts, not just source or CI, satisfy the GA contract.
- Issue and Project state match the evidence.
- Any `develop` to `main` promotion remains a separate explicit approval.

## Backlog disposition

### Current GA-critical release spine

| Area | Issues |
| --- | --- |
| Coordination and final approval | [#2546](https://github.com/collectors-tech/cabinet/issues/2546), [#2488](https://github.com/collectors-tech/cabinet/issues/2488), [#1864](https://github.com/collectors-tech/cabinet/issues/1864) |
| First run and packaged Chat setup | [#1946](https://github.com/collectors-tech/cabinet/issues/1946), [#2185](https://github.com/collectors-tech/cabinet/issues/2185), [#2190](https://github.com/collectors-tech/cabinet/issues/2190), [#2332](https://github.com/collectors-tech/cabinet/issues/2332) |
| Provider contract and proof | [#1944](https://github.com/collectors-tech/cabinet/issues/1944), [#1945](https://github.com/collectors-tech/cabinet/issues/1945) |
| Candidate and browser artifacts | [#1868](https://github.com/collectors-tech/cabinet/issues/1868), [#2034](https://github.com/collectors-tech/cabinet/issues/2034) |
| Packaged product and recovery | [#1869](https://github.com/collectors-tech/cabinet/issues/1869), [#1867](https://github.com/collectors-tech/cabinet/issues/1867) |
| Legal, privacy and support | [#2057](https://github.com/collectors-tech/cabinet/issues/2057) |

### Completed GA evidence enablers

| Area | Evidence |
| --- | --- |
| SBOM and provenance | [#2550](https://github.com/collectors-tech/cabinet/issues/2550); retained private package [run 33508229009](https://github.com/collectors-tech/cabinet/actions/runs/33508229009); build attestation [#44409596](https://github.com/collectors-tech/cabinet/attestations/44409596); CycloneDX attestation [#44409608](https://github.com/collectors-tech/cabinet/attestations/44409608) |

### Beta-evidence candidates

Onboarding, inventory and dashboard improvements such as #1946, #1947 and #1948
join the GA critical path only when a supported package or beta cohort shows a
contract failure. Priority changes must cite that evidence.

### Post-1.0 by default

- Metadata Studio issues.
- eBay seller, Shopify, crawler and parts-finder expansion.
- Trust and peer-to-peer features.
- Broad Telegram/Agent expansion beyond the approved 1.0 contract.
- Presentation-only Companion enhancements that do not block a supported journey.

## GA evidence rules

- Source checks, live-provider proof, packaged acceptance, recovery, legal
  decisions, approval and publication are separate gates.
- A green `no-change` workflow is not a fresh validation run.
- A merge waiver cannot convert failed checks or missing artifacts into evidence.
- The candidate SHA and runtime `build_revision` must match in full.
- Provider fixtures do not replace lawful user-present browser evidence.
- Package validation must not use test-only hooks or development servers.
- Release claims require public or retained-artifact redownload and replay.
- `develop` to `main` promotion requires separate explicit approval.

## Operating metrics

Before GA, #2546 must record targets and evidence for:

- clean onboarding completion rate;
- time to first useful collector action;
- core collector workflow completion rate;
- supported-provider success and fail-closed rate;
- upgrade, restart, backup and restore success;
- unresolved P0/P1 defects;
- crash-free sessions and actionable diagnostic capture;
- accessibility and large-library performance;
- support response and incident ownership.

## Decision record

Record decisions in #2546 and summarize them here when they change the release
contract. At minimum, Cabinet needs explicit decisions for:

1. Supported Windows versions and hardware baseline.
2. Signed installer versus approved portable-only distribution.
3. Supported versus Preview providers.
4. Supported versus Preview Chat/Agent capabilities.
5. Legal entity, privacy controller, retention and support commitments.
6. Exact GA approval and whether a separate `main` promotion is authorized.
