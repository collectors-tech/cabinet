import assert from 'node:assert/strict'
import { execFile } from 'node:child_process'
import { createHash } from 'node:crypto'
import { mkdtemp, readFile, rename, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import test from 'node:test'
import { promisify } from 'node:util'

import {
  acceptanceRows,
  createOrResumeAcceptanceRun,
  redactAcceptanceText,
  recordAcceptanceResult,
  renderAcceptanceMarkdown,
  validateAcceptanceState,
} from './lib/acceptance-evidence-recorder.mjs'

const commit = 'a'.repeat(40)
const sha256 = (value) => createHash('sha256').update(value).digest('hex')
const execFileAsync = promisify(execFile)

const writeArtifact = async (directory, filename, contents) => {
  const buffer = Buffer.from(contents)
  const digest = sha256(buffer)
  await writeFile(join(directory, filename), buffer)
  await writeFile(join(directory, `${filename}.sha256`), `${digest}  ${filename}\n`)
  return { target: filename.includes('edge') ? 'edge' : filename.includes('chrome') ? 'chrome' : 'windows-amd64', filename, sha256_filename: `${filename}.sha256`, sha256: digest, size_bytes: buffer.length }
}

const candidateFixture = async ({ sourceCommit = commit, suffix = 'one' } = {}) => {
  const directory = await mkdtemp(join(tmpdir(), 'cabinet-acceptance-recorder-'))
  const candidateVersion = `0.1.0-beta.7.g${sourceCommit.slice(0, 12)}`
  const cabinetArtifact = await writeArtifact(directory, 'cabinet-0.1.0-beta.7-windows-amd64-portable.zip', `cabinet-${suffix}`)
  cabinetArtifact.kind = 'portable_zip'
  const chromeArtifact = await writeArtifact(directory, `cabinet-browser-companion-${candidateVersion}-chrome.zip`, `chrome-${suffix}`)
  const edgeArtifact = await writeArtifact(directory, `cabinet-browser-companion-${candidateVersion}-edge.zip`, `edge-${suffix}`)
  const cabinet = {
    schema_version: 1,
    product: 'Cabinet',
    channel: 'private-beta',
    version: '0.1.0-beta.7',
    source_commit: sourceCommit,
    build_date: '2026-08-11T00:00:00Z',
    publication_state: 'private_candidate_not_published',
    artifact: cabinetArtifact,
    release_notes_filename: 'cabinet-0.1.0-beta.7-release-notes.md',
  }
  const companion = {
    schema_version: 1,
    product: 'Cabinet Browser Companion',
    channel: 'private-beta',
    version: '0.1.0',
    version_name: candidateVersion,
    immutable_tag: `browser-companion-v${candidateVersion}`,
    source_commit: sourceCommit,
    publication_state: 'private_candidate_not_published',
    protocol_compatibility: { minimum: '1', maximum: '1' },
    release_notes_filename: `cabinet-browser-companion-${candidateVersion}-release-notes.md`,
    artifacts: [chromeArtifact, edgeArtifact],
  }
  const bundle = {
    schema_version: 1,
    product: 'Cabinet 0.1 private beta candidate',
    channel: 'private-beta',
    source_commit: sourceCommit,
    publication_state: 'private_candidate_not_published',
    components: [
      { product: cabinet.product, version: cabinet.version, manifest_filename: 'cabinet-release-manifest.json', release_notes_filename: cabinet.release_notes_filename, artifacts: [cabinet.artifact] },
      { product: companion.product, version: companion.version_name, manifest_filename: 'browser-companion-release-manifest.json', release_notes_filename: companion.release_notes_filename, protocol_compatibility: companion.protocol_compatibility, artifacts: companion.artifacts.map(({ target, filename, sha256_filename, sha256 }) => ({ target, filename, sha256_filename, sha256 })) },
    ],
  }
  const cabinetManifestPath = join(directory, 'cabinet-release-manifest.json')
  const companionManifestPath = join(directory, 'browser-companion-release-manifest.json')
  const bundleManifestPath = join(directory, 'beta-candidate-bundle-manifest.json')
  await writeFile(cabinetManifestPath, `${JSON.stringify(cabinet, null, 2)}\n`)
  await writeFile(companionManifestPath, `${JSON.stringify(companion, null, 2)}\n`)
  await writeFile(bundleManifestPath, `${JSON.stringify(bundle, null, 2)}\n`)
  await writeFile(join(directory, cabinet.release_notes_filename), 'Cabinet release notes')
  await writeFile(join(directory, companion.release_notes_filename), 'Companion release notes')
  return { directory, cabinetManifestPath, companionManifestPath, bundleManifestPath }
}

const environment = {
  os_version: 'Windows 11 24H2',
  host_profile: 'isolated acceptance host',
  browser_name: 'Chrome',
  browser_version: '140.0.0.0',
  isolated_profile: 'cabinet-beta-acceptance',
  isolated_data_directory: 'C:\\Users\\operator\\AppData\\Local\\Cabinet\\acceptance-secret',
  runtime: { app_version: '0.1.0-beta.7', build_revision: commit, build_date: '2026-08-11T00:00:00Z', port: 17900, pid: 4242 },
}

const start = async (fixture, outputPath = join(fixture.directory, 'acceptance.json'), runtimeEnvironment = environment) => createOrResumeAcceptanceRun({
  cabinetManifestPath: fixture.cabinetManifestPath,
  companionManifestPath: fixture.companionManifestPath,
  bundleManifestPath: fixture.bundleManifestPath,
  releaseCandidateRunId: '31123456789',
  releaseCandidateArtifactName: 'cabinet-beta-candidate-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
  environment: runtimeEnvironment,
  outputPath,
})

test('catalog covers every current #1869 checklist row with unique stable identifiers', async () => {
  const checklist = await readFile(resolve('openspec/migration/beta-packaged-core-workflow-acceptance.md'), 'utf8')
  const checklistRows = [...checklist.matchAll(/^- \[ \] (.+)$/gm)].map((match) => match[1])
  assert.equal(acceptanceRows.length, 51)
  assert.deepEqual(acceptanceRows.map((row) => row.title), checklistRows)
  assert.equal(new Set(acceptanceRows.map((row) => row.id)).size, acceptanceRows.length)
  assert.ok(acceptanceRows.filter((row) => /Frontline|Bonza|Install the exact|recovery/i.test(row.title)).every((row) => row.requires_human_confirmation))
})

test('candidate identity, three manifests, and independently verified package checksums are mandatory', async () => {
  const fixture = await candidateFixture()
  await assert.rejects(
    () => createOrResumeAcceptanceRun({ outputPath: join(fixture.directory, 'missing.json'), environment }),
    /acceptance_candidate_identity_required/,
  )

  const state = await start(fixture)
  assert.equal(state.candidate.source_commit, commit)
  assert.equal(state.environment.runtime.build_revision, commit)
  assert.equal(state.candidate.cabinet.package.filename, 'cabinet-0.1.0-beta.7-windows-amd64-portable.zip')
  assert.deepEqual(state.candidate.companion.packages.map((item) => item.target), ['chrome', 'edge'])
  assert.equal(state.rows.length, acceptanceRows.length)

  const cabinet = JSON.parse(await readFile(fixture.cabinetManifestPath, 'utf8'))
  await writeFile(join(fixture.directory, cabinet.artifact.filename), 'tampered')
  await assert.rejects(() => start(fixture, join(fixture.directory, 'tampered.json')), /acceptance_package_checksum_mismatch/)
})

test('running package revision must be a full lowercase SHA matching the candidate manifest', async () => {
  const fixture = await candidateFixture()
  for (const buildRevision of [undefined, 'a'.repeat(39), 'A'.repeat(40), 'not-a-commit']) {
    await assert.rejects(
      () => start(fixture, join(fixture.directory, `invalid-${String(buildRevision)}.json`), {
        ...environment,
        runtime: { ...environment.runtime, build_revision: buildRevision },
      }),
      /acceptance_environment_identity_required/,
    )
  }
  await assert.rejects(
    () => start(fixture, join(fixture.directory, 'wrong-runtime.json'), {
      ...environment,
      runtime: { ...environment.runtime, build_revision: 'b'.repeat(40) },
    }),
    /acceptance_runtime_source_commit_mismatch/,
  )
})

test('resume preserves completed rows once and archives stale candidate evidence', async () => {
  const fixture = await candidateFixture()
  const outputPath = join(fixture.directory, 'acceptance.json')
  let state = await start(fixture, outputPath)
  state = await recordAcceptanceResult({
    state,
    rowId: 'COLLECTOR-01',
    status: 'pass',
    evidenceReferences: ['evidence/onboarding-screenshot.png'],
    operatorNotes: 'Completed in the packaged application.',
    operatorConfirmed: true,
  })
  await writeFile(outputPath, `${JSON.stringify(state, null, 2)}\n`)
  await rename(outputPath, `${outputPath}.previous`)
  const resumed = await start(fixture, outputPath)
  assert.equal(resumed.rows.length, acceptanceRows.length)
  assert.equal(resumed.rows.find((row) => row.id === 'COLLECTOR-01').status, 'pass')

  const replacement = await candidateFixture({ sourceCommit: 'b'.repeat(40), suffix: 'two' })
  const reset = await createOrResumeAcceptanceRun({
    cabinetManifestPath: replacement.cabinetManifestPath,
    companionManifestPath: replacement.companionManifestPath,
    bundleManifestPath: replacement.bundleManifestPath,
    releaseCandidateRunId: '31123456790',
    releaseCandidateArtifactName: 'cabinet-beta-candidate-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
    environment: {
      ...environment,
      runtime: { ...environment.runtime, build_revision: 'b'.repeat(40) },
    },
    outputPath,
  })
  assert.ok(reset.rows.every((row) => row.status === 'not_run'))
  assert.equal(reset.archived_prior_evidence.length, 1)
  assert.match(reset.archived_prior_evidence[0], new RegExp(`^acceptance\\.stale-${resumed.candidate.fingerprint.slice(0, 12)}-[a-f0-9]{12}\\.json$`))
  assert.deepEqual(JSON.parse(await readFile(join(fixture.directory, reset.archived_prior_evidence[0]), 'utf8')), resumed)
})

test('invalid transitions and evidence-free terminal states fail closed', async () => {
  const fixture = await candidateFixture()
  let state = await start(fixture)
  await assert.rejects(() => recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'pass', operatorConfirmed: true }), /acceptance_evidence_reference_required/)
  await assert.rejects(() => recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'pass', evidenceReferences: ['evidence/onboarding.png'], operatorConfirmed: true }), /acceptance_operator_notes_required/)
  await assert.rejects(() => recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'blocked' }), /acceptance_unblock_condition_required/)
  await assert.rejects(() => recordAcceptanceResult({ state, rowId: 'PROVIDER-08', status: 'pass', evidenceReferences: ['evidence/frontline.png'], operatorNotes: 'Observed user-present flow.' }), /acceptance_human_confirmation_required/)
  state = await recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'pass', evidenceReferences: ['evidence/onboarding.png'], operatorNotes: 'Packaged flow completed.', operatorConfirmed: true })
  await assert.rejects(() => recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'blocked', unblockCondition: 'rerun later' }), /acceptance_status_transition_invalid/)
  const idempotent = await recordAcceptanceResult({ state, rowId: 'COLLECTOR-01', status: 'pass', evidenceReferences: ['evidence/onboarding.png'], operatorNotes: 'Packaged flow completed.', operatorConfirmed: true })
  assert.deepEqual(idempotent, state)
  const blocked = await recordAcceptanceResult({ state, rowId: 'PROVIDER-08', status: 'blocked', unblockCondition: 'User-present Frontline session is available.' })
  assert.equal(blocked.overall_result, 'fail_with_blockers')
})

test('saved environment identity and evidence references fail closed when tampered', async () => {
  const fixture = await candidateFixture()
  const state = await start(fixture)

  const changedEnvironment = structuredClone(state)
  changedEnvironment.environment.browser_version = 'tampered-version'
  assert.throws(
    () => validateAcceptanceState(changedEnvironment),
    /acceptance_state_environment_fingerprint_invalid/,
  )

  const unsafeReference = structuredClone(state)
  unsafeReference.rows[0].evidence_references = ['..\\private\\proof.png']
  assert.throws(
    () => validateAcceptanceState(unsafeReference),
    /acceptance_evidence_reference_sensitive:IDENTITY-01/,
  )
})

test('secret material and sensitive paths never leak to JSON or Markdown', async () => {
  assert.doesNotMatch(redactAcceptanceText('password="hunter two"'), /hunter|two/)
  assert.doesNotMatch(redactAcceptanceText('password=hunter two'), /hunter|two/)
  assert.doesNotMatch(redactAcceptanceText('proof at C:\\Users\\Max Barrass\\Desktop\\proof.png'), /Max|Barrass|Desktop/)
  const fixture = await candidateFixture()
  let state = await start(fixture)
  await assert.rejects(() => recordAcceptanceResult({
    state,
    rowId: 'COLLECTOR-01',
    status: 'fail',
    evidenceReferences: ['Bearer super-secret-token'],
    operatorNotes: 'failed',
  }), /acceptance_evidence_reference_sensitive/)
  state = await recordAcceptanceResult({
    state,
    rowId: 'COLLECTOR-01',
    status: 'fail',
    evidenceReferences: ['evidence/issues/2050.md'],
    operatorNotes: 'password="hunter two"; Cookie: cabinet_session=abc [private]customer order 123[/private] at C:\\Users\\Max Barrass\\Desktop\\proof.png',
  })
  const json = `${JSON.stringify(validateAcceptanceState(state), null, 2)}\n`
  const markdown = renderAcceptanceMarkdown(state)
  for (const output of [json, markdown]) {
    assert.doesNotMatch(output, /hunter two|cabinet_session|customer order 123|C:\\Users\\Max|Barrass\\Desktop/i)
    assert.match(output, /<redacted-/)
  }
})

test('dry-run outputs are deterministic and never auto-pass human steps', async () => {
  const fixture = await candidateFixture()
  const state = await start(fixture)
  const first = renderAcceptanceMarkdown(state)
  const second = renderAcceptanceMarkdown(validateAcceptanceState(JSON.parse(JSON.stringify(state))))
  assert.equal(first, second)
  assert.equal(state.overall_result, 'not_run')
  assert.ok(state.rows.every((row) => row.status === 'not_run'))
  assert.match(first, /Overall result: `not_run`/)
  assert.match(first, /PROVIDER-08.*not_run/)
  assert.match(first, /PROVIDER-09.*not_run/)
  assert.match(first, /has no release, branch-promotion, provider-interaction, or browser-automation operation/)
})

test('operator CLI exposes evidence-only commands', async () => {
  const { stdout } = await execFileAsync(process.execPath, ['scripts/record-beta-acceptance.mjs', '--help'], { cwd: resolve('.') })
  for (const command of ['init', 'record', 'validate', 'render']) assert.match(stdout, new RegExp(`\\b${command}\\b`))
  assert.doesNotMatch(stdout, /^\s+(?:publish|promote|merge)\b/im)
})

test('operator CLI initializes and resumes a real JSON and Markdown dry run', async () => {
  const fixture = await candidateFixture()
  const jsonPath = join(fixture.directory, 'operator-acceptance.json')
  const markdownPath = join(fixture.directory, 'operator-acceptance.md')
  const common = { cwd: resolve('.') }
  await execFileAsync(process.execPath, [
    'scripts/record-beta-acceptance.mjs', 'init',
    '--cabinet-manifest', fixture.cabinetManifestPath,
    '--companion-manifest', fixture.companionManifestPath,
    '--bundle-manifest', fixture.bundleManifestPath,
    '--candidate-run-id', '31123456789',
    '--candidate-artifact', 'cabinet-beta-candidate-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    '--os-version', environment.os_version,
    '--host-profile', environment.host_profile,
    '--browser-name', environment.browser_name,
    '--browser-version', environment.browser_version,
    '--isolated-profile', environment.isolated_profile,
    '--data-directory', environment.isolated_data_directory,
    '--app-version', environment.runtime.app_version,
    '--build-revision', environment.runtime.build_revision,
    '--build-date', environment.runtime.build_date,
    '--runtime-port', String(environment.runtime.port),
    '--runtime-pid', String(environment.runtime.pid),
    '--json', jsonPath,
    '--markdown', markdownPath,
  ], common)
  await execFileAsync(process.execPath, [
    'scripts/record-beta-acceptance.mjs', 'record',
    '--json', jsonPath,
    '--markdown', markdownPath,
    '--row', 'PROVIDER-08',
    '--status', 'blocked',
    '--unblock', 'User-present Frontline session is available.',
  ], common)
  const state = JSON.parse(await readFile(jsonPath, 'utf8'))
  assert.equal(state.rows.find((row) => row.id === 'PROVIDER-08').status, 'blocked')
  assert.equal(state.overall_result, 'fail_with_blockers')
  assert.match(await readFile(markdownPath, 'utf8'), /PROVIDER-08 \| blocked/)
})
