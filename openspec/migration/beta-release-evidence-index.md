# Cabinet 0.1 beta release evidence index

Issue: #1864 `epic(beta): ship Cabinet 0.1 private beta`
Reconciled: 2026-08-06 AEST under #2036

This is the single repository-side release sequence and evidence hand-off for
#1864. It does not approve, publish or promote a release.

## Acyclic gate model

1. **Source/live ready for packaging.** Merge and validate canonical media,
   all required direct/provider modules, secure Browser Companion transport,
   host, persistence and package controls. A provider's source/live-ready state
   does not require #1869 closure or packaged evidence.
2. **Temporary release freeze.** Nominate one exact `develop` commit after all
   source-ready prerequisites are merged. During the temporary release freeze, accept only a P0 release blocker,
   directly required test/evidence repair or release-document
   correction through a focused issue and green pull request. Any accepted fix
   creates a new exact candidate commit and invalidates earlier package evidence.
3. **Private/internal candidate.** #1868 runs the exact-commit beta gate and
   creates Cabinet plus #2034 Browser Companion candidate artefacts, release
   notes, manifests and separate SHA-256 checksums.
   Internal candidate creation does not require final #1864 approval.
4. **#1869 packaged acceptance.** Test the exact Cabinet and companion files on
   Windows, including collector journeys, Voglers, Hobbytech, Frontline, Bonza,
   canonical/migrated media and recovery states.
5. **#1867 packaged data-safety.** Attach upgrade, backup, export, restore,
   relocation, media-manifest and zero-data-loss evidence from the same candidate.
6. **Decision and publication.** Link all exact-candidate evidence to #1864 and obtain explicit #1864 approval.
   Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.
   Rejection or any release-blocking fix returns
   the programme to the freeze/candidate gate with a new commit.

## Source-ready checklist

| Work | Source/live state | Packaged state |
| --- | --- | --- |
| #1936/#1937 canonical media and migration | Complete and merged | Revalidate layout, links, hashes, relocation and restore in #1869/#1867 |
| Voglers (#1871) | Complete and merged | Run exact packaged journey in #1869 |
| Hobbytech (#1943) | Direct live proof merged; source/live ready | Keep #1943 open only for exact packaged evidence |
| #2033 secure companion pairing | Complete and merged | Pair/reconnect/rotate/revoke in #1869 |
| #2035 shared MV3 host/config/readiness | Complete and merged | Install and verify browser/Cabinet/provider states in #1869 |
| #2032 item/image observation persistence | Source implementation complete | Commit-before-ack, fixtures, replay, media, restart, safe export and relocated restore are automated; exact packaged proof remains in #1869/#1867 |
| Frontline (#1944) | Exact-origin module, fail-closed fixtures, persistence and confirmed hand-off implemented; #2054 bounded-timeout source contract required; external user-present live proof pending | Exact packaged journey and stalled/unavailable timeout proof in #1869 |
| Bonza (#1945) | Decoder/retry removed; exact-origin module, fail-closed fixtures and canonical persistence implemented; #2054 bounded-timeout source contract required; external user-present live proof pending | Exact packaged journey and stalled/unavailable timeout proof in #1869 |
| #2034 Chrome/Edge packaging | Deterministic target packages, release manifest, checksums, verifier, immutable-version preflight and manual recovery guidance implemented; exact candidate files pending #1868 | Nominate and install/pair the same files/checksums in #1868/#1869 |

#1929 may record a provider as source/live ready for packaging while its focused
issue remains open for #1869 packaged evidence. #1929 closure is not a prerequisite
for #1868; completion of every source-ready checklist row is.

## Child issue state

| Issue | Current evidence | Release state |
| --- | --- | --- |
| #1865 Develop/release gates | PR #1879 merged | Done |
| #1866 truthful local authentication | Merged evidence linked on the issue | Done |
| #1867 data safety | `beta-data-safety-evidence-matrix.md` records source proof | Source complete; exact packaged proof pending |
| #1870 entitlement alignment | Merged evidence linked on the issue | Done |
| #1871 Voglers live path | PR #1908 merged | Source/live ready; packaged rerun pending |
| #1936/#1937 canonical media | Merged implementation and migration evidence | Source complete; packaged rerun pending |
| #1943 Hobbytech | PR #1955 plus live provider-run proof | Source/live ready; packaged rerun pending |
| #2037 truthful OpenAPI gate | PR #2038; Develop Quality Gate `31018831485` green | Done |
| #2054 bounded provider HTTP behavior | Shared commerce-provider timeout, cancellation, partial-response and same-client recovery contract tests | Source gate required before candidate; Frontline/Bonza packaged timeout proof remains in #1869 |
| #2033/#2035/#2032/#1944/#1945/#2034 | Browser Companion release programme | Packaging source implemented; external Frontline/Bonza and exact-candidate evidence pending |
| #1868 exact internal candidate | Exact-clean Cabinet manifest/verifier, combined Cabinet/companion manifest, validate-before-package workflow and approval-evidenced prerelease source implemented | Run only after Frontline/Bonza source/live gate and freeze; retain exact candidate run/artifacts |
| #1869 packaged acceptance | Stable checklist exists | Execute after #1868 internal candidate |

## Evidence commands and artefacts

### Source gate

```text
go test ./internal/... ./cmd/... -count=1
npm --prefix ui.web run build
npx --yes @fission-ai/openspec@latest validate --all --strict --no-interactive
go run ./cmd/openapi-parity-gate
```

Each focused issue also runs its named Go/Cypress/extension tests and records the
exact command in its pull request. #1868 produces the final Cabinet and extension
files only after the exact validation job passes and records them in
`beta-candidate-bundle-manifest.json`.

### Freeze and candidate

```text
git fetch origin develop
git rev-parse origin/develop
git status --short
gh workflow run beta-release-candidate.yml --ref develop -f commit_sha=<40-character-origin-develop-sha>
gh run watch <release-candidate-run-id> --exit-status
pwsh -NoLogo -NoProfile -File .\scripts\package-installers.ps1
```

The private/internal candidate is the only candidate accepted by the later gates.
Record the exact commit, workflow run, Cabinet filename/version/checksum, companion
filename/version/checksum, both release manifests, combined candidate manifest and release notes. Store these as
private/internal acceptance artefacts; do not create or update an external release.

After #1869 and #1867 pass, final #1864 approval must be a trusted comment containing
`APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>`. The explicit publication workflow
accepts that comment ID and successful candidate run ID, downloads and reverifies the
same artefacts, rejects version reuse and creates the prerelease. Main CI does not
publish automatically.

### Packaged acceptance and decision

Run `beta-packaged-core-workflow-acceptance.md`,
`windows-portable-artifact-validation.md` and
`windows-portable-upgrade-validation.md` against the nominated files. Link evidence
to #1869, #1867 and then #1864. Approval occurs only after those results exist.

## Stranded provider branch disposition

The branches are archived as read-only evidence; their commits remain reachable,
but they must not be merged wholesale because both branch tips predate hundreds of
`develop` commits and implement superseded direct-transport assumptions.

| Branch | Preserved evidence | Disposition |
| --- | --- | --- |
| `issue-1944-frontline-live-provider` | Tip `e04ca27e`; five commits cover Frontline UI projection, saved-watch routing, fail-closed warnings and Cypress provenance/handoff tests | Reuse requirements/fixtures selectively under #1944 after #2033/#2035/#2032; no PR from the stale branch |
| `issue-1929-bonza-live-path-probe` | Tip `2c0e51b2`; bounded probe test records the Sucuri challenge response | Preserve as historical blocker evidence under #1945; do not reuse its transport as challenge solving is prohibited |

## Invariants

- #1864 remains P0 and is the single final evidence/approval parent.
- An issue may stay open for packaged evidence without blocking internal candidate creation once its source-ready row passes.
- A screenshot, fixture or registry entry alone is not live provider evidence.
- No governance change publishes a release or merges `develop` into `main`.
