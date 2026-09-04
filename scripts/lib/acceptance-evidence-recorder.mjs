import { createHash } from 'node:crypto'
import { access, readFile, rename, rm, writeFile } from 'node:fs/promises'
import { basename, dirname, join } from 'node:path'

import { verifyCabinetSBOM } from './cabinet-sbom.mjs'

const rows = (section, prefix, titles, requiresHumanConfirmation = false) => titles.map((title, index) => ({
  id: `${prefix}-${String(index + 1).padStart(2, '0')}`,
  section,
  title,
  requires_human_confirmation: requiresHumanConfirmation,
}))

export const acceptanceRows = Object.freeze([
  ...rows('Candidate Identity', 'IDENTITY', [
    'OS version and host profile are recorded.',
    'Cabinet package filename is recorded.',
    'Cabinet package SHA-256 is recorded and matches its `.sha256` file.',
    'Cabinet source commit SHA is recorded.',
    'Successful Beta Release Candidate Gate run ID and exact artifact name are recorded.',
    'Cabinet, Browser Companion and combined candidate manifest paths all name the same source commit.',
    'Browser name/version and Browser Companion package filename are recorded.',
    'Browser Companion package SHA-256, source commit, extension version and release-manifest path are recorded.',
    'Browser Companion production identity, target, protocol compatibility and immutable candidate version match the release manifest; the Development source build is not used.',
    '`/api/runtime.app_version` and full `/api/runtime.build_revision`, build date, runtime port, and pid are recorded; build revision equals the Cabinet manifest `source_commit`.',
    'Cabinet and Browser Companion release notes paths are recorded.',
  ]),
  ...rows('Required Collector Journey', 'COLLECTOR', [
    'Fresh start and onboarding/profile setup complete from a clean Windows data directory.',
    'Inventory item can be created, edited, searched, filtered, reloaded, and verified after restart.',
    'Media can be attached, marked primary, and verified after restart.',
    '#1937 media migration evidence records discovered, migrated, already-migrated, duplicate, skipped, failed, and orphan counts from the packaged or explicit maintenance smoke run.',
    'Wishlist item can be created, reprioritised, status-updated, and marked purchased into Inventory.',
    'Collection can be created/edited, receive/move an item, soft-delete safely, and protect All Items.',
    'Data export and backup both complete with non-secret artefacts.',
    'Backup restore into an isolated target preserves core record counts and relationships.',
    'Discovery review can hand an item to Wishlist or Inventory without ownership confusion.',
    'One failed provider and one invalid import/restore input show useful recovery/error behavior.',
  ], true),
  ...rows('Required Provider and Companion Journey', 'PROVIDER', [
    'Install the exact Chrome and Edge packages through the documented beta path without developer source tools.',
    'Upgrade and rollback use only verified versioned packages, preserve visible pending jobs, revoke stale origins and fail closed on checksum or protocol mismatch.',
    'Pair to Cabinet through #2033 and verify reconnect, credential rotation, revoke-one and revoke-all.',
    'Enabled browser-capable integration changes propagate from Cabinet without rebuilding the extension.',
    'Cabinet/provider open-focus and ready, login-required, action-required, partial, selector-drift and disconnected states are truthful.',
    'A real saved Market Watch and Discovery hand-off pass independently for Voglers.',
    'A real saved Market Watch and Discovery hand-off pass independently for Hobbytech.',
    'A user-present real Frontline search persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.',
    'A user-present real Bonza search after normal browser interaction persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.',
    'A stalled or unavailable Frontline request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.',
    'A stalled or unavailable Bonza request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.',
    "Failure of one provider does not prevent, mutate or corrupt another provider's watches or observations.",
    'Replaying one capture proves item and media idempotency with transport/module/schema provenance.',
    'One durable protected-provider image uses the canonical asset manifest/layout and survives restart, backup, relocation and restore.',
    'Browser-closed, Cabinet-restart and extension-service-worker recovery resume without duplicate observations.',
  ], true),
  ...rows('Cross-Cutting Proof', 'CROSS', [
    'Persistence is verified after reload and application restart.',
    'Active-profile isolation is verified for at least one created record, companion session and export/restore path.',
    'No raw translation keys, placeholder security claims, or unsigned-installer claims appear in release UI/docs.',
    'No cookie/token export, challenge solving, hidden crawling or silent inventory mutation occurs.',
    'Empty and error states are useful enough for a beta user to recover or report the issue.',
    'Exact Cabinet and extension versions/commits are visible in recorded evidence.',
  ], true),
  ...rows('Failure Handling', 'FAILURE', [
    'Every failure creates or links a focused GitHub issue with route/surface, expected behavior, actual behavior, repro steps, evidence, requirement link, and planned validation target.',
    'Release-blocking failures are linked back to #1864 and #1869 before rerun.',
    'The acceptance pack is rerun after release-blocking fixes against a new exact candidate commit.',
    'Final evidence explicitly states pass, fail with blockers, or not run, without using visual toasts or redirects as persistence proof.',
    'If all gates pass, the proposed #1864 approval comment records `APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>`; publication is not invoked by this checklist.',
  ], true),
  ...rows('Prohibited Shortcuts', 'SHORTCUT', [
    'Final packaged acceptance uses the packaged binary, not a dev server.',
    'Final packaged acceptance does not require test-only hooks.',
    'Final packaged acceptance does not rely on a dirty worktree or unpublished local changes.',
    'Final packaged acceptance does not merge `develop` into `main` or publish a release without #1864 approval.',
  ], true),
])

const sha256 = (value) => createHash('sha256').update(value).digest('hex')
const isObject = (value) => value !== null && typeof value === 'object' && !Array.isArray(value)
const stableValue = (value) => Array.isArray(value)
  ? value.map(stableValue)
  : isObject(value)
    ? Object.fromEntries(Object.keys(value).sort().map((key) => [key, stableValue(value[key])]))
    : value
const stableStringify = (value) => JSON.stringify(stableValue(value))
const fullCommit = (value) => typeof value === 'string' && /^[a-f0-9]{40}$/.test(value)
const checksum = (value) => typeof value === 'string' && /^[a-f0-9]{64}$/.test(value)
const safeFilename = (value) => typeof value === 'string' && value.length > 0 && value === basename(value)
const requiredText = (value) => typeof value === 'string' && value.trim().length > 0

export const redactAcceptanceText = (input) => String(input ?? '')
  .replace(/\[private\][\s\S]*?\[\/private\]/gi, '<redacted-private-content>')
  .replace(/\bBearer\s+[A-Za-z0-9._~+/=-]+/gi, 'Bearer <redacted-token>')
  .replace(/\b(?:Cookie|Set-Cookie)\s*:[^\r\n]*/gi, 'Cookie: <redacted-cookie>')
  .replace(/\b(password|passwd|client_secret|access_token|refresh_token|id_token|api_key|authorization)\s*[:=]\s*(?:"[^"]*"|'[^']*'|[^,;\r\n]+)/gi, '$1=<redacted-credential>')
  .replace(/\\\\[^,;|\r\n]+|\b[A-Za-z]:\\[^,;|\r\n]+|\/(?:Users|home)\/[^,;|\r\n]+/g, '<redacted-local-path>')

const parseJSON = async (path, code) => {
  try {
    return { raw: await readFile(path), path }
  } catch (error) {
    throw new Error(`${code}:${error.code ?? 'read_failed'}`)
  }
}

const verifyArtifact = async (directory, artifact, target) => {
  if (!isObject(artifact) || artifact.target !== target || !safeFilename(artifact.filename) ||
      artifact.sha256_filename !== `${artifact.filename}.sha256` || !checksum(artifact.sha256)) {
    throw new Error(`acceptance_package_identity_invalid:${target}`)
  }
  const packagePath = join(directory, artifact.filename)
  const data = await readFile(packagePath).catch(() => { throw new Error(`acceptance_package_missing:${target}`) })
  const actual = sha256(data)
  if (actual !== artifact.sha256 || (Number.isInteger(artifact.size_bytes) && artifact.size_bytes !== data.length)) {
    throw new Error(`acceptance_package_checksum_mismatch:${target}`)
  }
  const checksumText = await readFile(join(directory, artifact.sha256_filename), 'utf8')
    .catch(() => { throw new Error(`acceptance_checksum_file_missing:${target}`) })
  if (checksumText !== `${actual}  ${artifact.filename}\n`) throw new Error(`acceptance_checksum_file_mismatch:${target}`)
  return {
    target,
    filename: artifact.filename,
    sha256_filename: artifact.sha256_filename,
    sha256: actual,
    size_bytes: data.length,
  }
}

const requireReleaseNotes = async (manifestPath, filename, target) => {
  if (!safeFilename(filename)) throw new Error(`acceptance_release_notes_identity_invalid:${target}`)
  await access(join(dirname(manifestPath), filename)).catch(() => { throw new Error(`acceptance_release_notes_missing:${target}`) })
  return filename
}

const verifySBOM = async (manifestPath, cabinet, cabinetPackage) => {
  const sbom = cabinet.sbom
  const expectedFilename = `cabinet-${cabinet.version}-sbom.cdx.json`
  if (!isObject(sbom) || sbom.filename !== expectedFilename || !safeFilename(sbom.filename) ||
      sbom.embedded_path !== 'CABINET-SBOM.cdx.json' || sbom.format !== 'cyclonedx-json' || sbom.spec_version !== '1.7' ||
      sbom.predicate_type !== 'https://cyclonedx.org/bom' || !checksum(sbom.sha256) || !Number.isInteger(sbom.size_bytes) || sbom.size_bytes < 1 ||
      sbom.source_commit !== cabinet.source_commit || sbom.subject_artifact_sha256 !== cabinetPackage.sha256) {
    throw new Error('acceptance_sbom_identity_invalid')
  }
  const data = await readFile(join(dirname(manifestPath), sbom.filename))
    .catch(() => { throw new Error('acceptance_sbom_missing') })
  if (data.length !== sbom.size_bytes || sha256(data) !== sbom.sha256) throw new Error('acceptance_sbom_checksum_mismatch')
  let document
  try {
    document = JSON.parse(data)
  } catch {
    throw new Error('acceptance_sbom_json_invalid')
  }
  verifyCabinetSBOM(document, {
    version: cabinet.version,
    sourceCommit: cabinet.source_commit,
    buildDate: cabinet.build_date,
  })
  return {
    filename: sbom.filename,
    embedded_path: sbom.embedded_path,
    format: sbom.format,
    spec_version: sbom.spec_version,
    predicate_type: sbom.predicate_type,
    sha256: sbom.sha256,
    size_bytes: sbom.size_bytes,
    source_commit: sbom.source_commit,
    subject_artifact_sha256: sbom.subject_artifact_sha256,
  }
}

const verifyCandidate = async ({
  cabinetManifestPath,
  companionManifestPath,
  bundleManifestPath,
  releaseCandidateRunId,
  releaseCandidateArtifactName,
}) => {
  if (![cabinetManifestPath, companionManifestPath, bundleManifestPath].every(requiredText) ||
      !/^\d+$/.test(String(releaseCandidateRunId ?? '')) || !safeFilename(releaseCandidateArtifactName)) {
    throw new Error('acceptance_candidate_identity_required')
  }
  const cabinetFile = await parseJSON(cabinetManifestPath, 'acceptance_cabinet_manifest_missing')
  const companionFile = await parseJSON(companionManifestPath, 'acceptance_companion_manifest_missing')
  const bundleFile = await parseJSON(bundleManifestPath, 'acceptance_bundle_manifest_missing')
  let cabinet
  let companion
  let bundle
  try {
    cabinet = JSON.parse(cabinetFile.raw)
    companion = JSON.parse(companionFile.raw)
    bundle = JSON.parse(bundleFile.raw)
  } catch {
    throw new Error('acceptance_candidate_manifest_json_invalid')
  }
  if (!fullCommit(cabinet.source_commit) || companion.source_commit !== cabinet.source_commit || bundle.source_commit !== cabinet.source_commit) {
    throw new Error('acceptance_candidate_source_commit_mismatch')
  }
  for (const manifest of [cabinet, companion, bundle]) {
    if (manifest.channel !== 'private-beta' || manifest.publication_state !== 'private_candidate_not_published') {
      throw new Error('acceptance_candidate_publication_boundary_invalid')
    }
  }
  if (cabinet.product !== 'Cabinet' || !requiredText(cabinet.version) || companion.product !== 'Cabinet Browser Companion' ||
      !requiredText(companion.version_name) || !companion.version_name.endsWith(`.g${cabinet.source_commit.slice(0, 12)}`) ||
      !requiredText(companion.immutable_tag) || !isObject(companion.protocol_compatibility)) {
    throw new Error('acceptance_candidate_product_identity_invalid')
  }
  if (basename(cabinetManifestPath) !== 'cabinet-release-manifest.json' ||
      basename(companionManifestPath) !== 'browser-companion-release-manifest.json' ||
      basename(bundleManifestPath) !== 'beta-candidate-bundle-manifest.json') {
    throw new Error('acceptance_candidate_manifest_filename_invalid')
  }
  const cabinetPackage = await verifyArtifact(dirname(cabinetManifestPath), cabinet.artifact, 'windows-amd64')
  const cabinetSBOM = await verifySBOM(cabinetManifestPath, cabinet, cabinetPackage)
  const companionArtifacts = new Map((companion.artifacts ?? []).map((artifact) => [artifact.target, artifact]))
  if (companionArtifacts.size !== 2 || !companionArtifacts.has('chrome') || !companionArtifacts.has('edge')) {
    throw new Error('acceptance_companion_targets_invalid')
  }
  const companionPackages = []
  for (const target of ['chrome', 'edge']) companionPackages.push(await verifyArtifact(dirname(companionManifestPath), companionArtifacts.get(target), target))

  const expectedComponents = [
    { product: cabinet.product, version: cabinet.version, manifest_filename: basename(cabinetManifestPath), release_notes_filename: cabinet.release_notes_filename, artifacts: [cabinet.artifact], sbom: cabinet.sbom },
    { product: companion.product, version: companion.version_name, manifest_filename: basename(companionManifestPath), release_notes_filename: companion.release_notes_filename, protocol_compatibility: companion.protocol_compatibility, artifacts: companion.artifacts.map(({ target, filename, sha256_filename, sha256 }) => ({ target, filename, sha256_filename, sha256 })) },
  ]
  if (bundle.schema_version !== 1 || bundle.product !== 'Cabinet 0.1 private beta candidate' || stableStringify(bundle.components) !== stableStringify(expectedComponents)) {
    throw new Error('acceptance_combined_manifest_identity_mismatch')
  }
  const identity = {
    source_commit: cabinet.source_commit,
    release_candidate: { run_id: String(releaseCandidateRunId), artifact_name: releaseCandidateArtifactName },
    cabinet: {
      version: cabinet.version,
      build_date: cabinet.build_date,
      manifest_filename: basename(cabinetManifestPath),
      manifest_sha256: sha256(cabinetFile.raw),
      package: cabinetPackage,
      sbom: cabinetSBOM,
      release_notes_filename: await requireReleaseNotes(cabinetManifestPath, cabinet.release_notes_filename, 'cabinet'),
    },
    companion: {
      version: companion.version,
      version_name: companion.version_name,
      immutable_tag: companion.immutable_tag,
      protocol_compatibility: companion.protocol_compatibility,
      manifest_filename: basename(companionManifestPath),
      manifest_sha256: sha256(companionFile.raw),
      packages: companionPackages,
      release_notes_filename: await requireReleaseNotes(companionManifestPath, companion.release_notes_filename, 'companion'),
    },
    combined_manifest_filename: basename(bundleManifestPath),
    combined_manifest_sha256: sha256(bundleFile.raw),
    publication_state: 'private_candidate_not_published',
  }
  return { ...identity, fingerprint: sha256(stableStringify(identity)) }
}

const sanitizeEnvironment = (environment) => {
  const runtime = environment?.runtime
  if (![environment?.os_version, environment?.host_profile, environment?.browser_name, environment?.browser_version,
    environment?.isolated_profile, environment?.isolated_data_directory, runtime?.app_version, runtime?.build_revision, runtime?.build_date].every(requiredText) ||
    !fullCommit(runtime?.build_revision) ||
    !Number.isInteger(runtime?.port) || runtime.port < 1 || runtime.port > 65535 || !Number.isInteger(runtime?.pid) || runtime.pid < 1) {
    throw new Error('acceptance_environment_identity_required')
  }
  const identity = {
    os_version: redactAcceptanceText(environment.os_version),
    host_profile: redactAcceptanceText(environment.host_profile),
    browser_name: redactAcceptanceText(environment.browser_name),
    browser_version: redactAcceptanceText(environment.browser_version),
    isolated_profile: redactAcceptanceText(environment.isolated_profile),
    isolated_data_directory: {
      display: '<redacted-local-path>',
      sha256: sha256(environment.isolated_data_directory),
    },
    runtime: {
      app_version: redactAcceptanceText(runtime.app_version),
      build_revision: runtime.build_revision,
      build_date: redactAcceptanceText(runtime.build_date),
      port: runtime.port,
      pid: runtime.pid,
    },
  }
  return { ...identity, fingerprint: sha256(stableStringify(identity)) }
}

const calculateOverall = (rowsToCheck) => {
  if (rowsToCheck.some((row) => row.status === 'fail' || row.status === 'blocked')) return 'fail_with_blockers'
  if (rowsToCheck.every((row) => row.status === 'pass')) return 'pass'
  return 'not_run'
}

const newRows = () => acceptanceRows.map((row) => ({
  ...row,
  status: 'not_run',
  evidence_references: [],
  operator_notes: '',
  unblock_condition: '',
  operator_confirmed: false,
}))

const writeState = async (path, state) => {
  const pending = `${path}.pending`
  const previous = `${path}.previous`
  await writeFile(pending, `${JSON.stringify(state, null, 2)}\n`, { flag: 'w' })
  let hadCurrent = true
  try {
    await rename(path, previous)
  } catch (error) {
    if (error.code !== 'ENOENT') throw error
    hadCurrent = false
  }
  try {
    await rename(pending, path)
    if (hadCurrent) await rm(previous, { force: true })
  } catch (error) {
    if (hadCurrent) await rename(previous, path).catch(() => {})
    throw error
  }
}

const loadRecoverableState = async (path) => {
  try {
    return JSON.parse(await readFile(path, 'utf8'))
  } catch (error) {
    if (error.code !== 'ENOENT') throw new Error('acceptance_state_json_invalid')
  }
  try {
    await rename(`${path}.previous`, path)
    return JSON.parse(await readFile(path, 'utf8'))
  } catch (error) {
    if (error.code !== 'ENOENT') throw new Error('acceptance_state_recovery_failed')
    return undefined
  }
}

export const validateAcceptanceState = (state) => {
  if (!isObject(state) || state.schema_version !== 1 || state.recorder !== 'Cabinet packaged acceptance evidence' || !isObject(state.candidate) || !isObject(state.environment) || !Array.isArray(state.rows)) {
    throw new Error('acceptance_state_identity_invalid')
  }
  const { fingerprint, ...identity } = state.candidate
  if (!checksum(fingerprint) || sha256(stableStringify(identity)) !== fingerprint) throw new Error('acceptance_state_candidate_fingerprint_invalid')
  const { fingerprint: environmentFingerprint, ...environmentIdentity } = state.environment
  if (!checksum(environmentFingerprint) || sha256(stableStringify(environmentIdentity)) !== environmentFingerprint) {
    throw new Error('acceptance_state_environment_fingerprint_invalid')
  }
  if (!Array.isArray(state.archived_prior_evidence) ||
      state.archived_prior_evidence.some((filename) => !safeFilename(filename)) ||
      new Set(state.archived_prior_evidence).size !== state.archived_prior_evidence.length) {
    throw new Error('acceptance_state_archive_history_invalid')
  }
  if (state.rows.length !== acceptanceRows.length) throw new Error('acceptance_state_rows_incomplete')
  for (let index = 0; index < acceptanceRows.length; index += 1) {
    const expected = acceptanceRows[index]
    const row = state.rows[index]
    if (!isObject(row) || row.id !== expected.id || row.section !== expected.section || row.title !== expected.title || row.requires_human_confirmation !== expected.requires_human_confirmation ||
        !['not_run', 'blocked', 'pass', 'fail'].includes(row.status) || !Array.isArray(row.evidence_references) ||
        typeof row.operator_notes !== 'string' || typeof row.unblock_condition !== 'string' || typeof row.operator_confirmed !== 'boolean') {
      throw new Error(`acceptance_state_row_invalid:${expected.id}`)
    }
    for (const reference of row.evidence_references) {
      if (!requiredText(reference) || redactAcceptanceText(reference) !== reference || /(^|[\\/])\.\.([\\/]|$)/.test(reference)) {
        throw new Error(`acceptance_evidence_reference_sensitive:${row.id}`)
      }
    }
    if (['pass', 'fail'].includes(row.status) && row.evidence_references.length === 0) throw new Error(`acceptance_evidence_reference_required:${row.id}`)
    if (['pass', 'fail'].includes(row.status) && !requiredText(row.operator_notes)) throw new Error(`acceptance_operator_notes_required:${row.id}`)
    if (row.status === 'blocked' && !requiredText(row.unblock_condition)) throw new Error(`acceptance_unblock_condition_required:${row.id}`)
    if (row.status === 'pass' && row.requires_human_confirmation && row.operator_confirmed !== true) throw new Error(`acceptance_human_confirmation_required:${row.id}`)
    for (const value of [...row.evidence_references, row.operator_notes, row.unblock_condition]) {
      if (redactAcceptanceText(value) !== value) throw new Error(`acceptance_state_secret_leak:${row.id}`)
    }
  }
  if (state.overall_result !== calculateOverall(state.rows)) throw new Error('acceptance_overall_result_invalid')
  return state
}

export const createOrResumeAcceptanceRun = async (options) => {
  const candidate = await verifyCandidate(options ?? {})
  const environment = sanitizeEnvironment(options?.environment)
  if (environment.runtime.build_revision !== candidate.source_commit) {
    throw new Error('acceptance_runtime_source_commit_mismatch')
  }
  if (!requiredText(options?.outputPath)) throw new Error('acceptance_output_path_required')
  const existing = await loadRecoverableState(options.outputPath)
  if (existing) {
    validateAcceptanceState(existing)
    if (existing.candidate.fingerprint === candidate.fingerprint) {
      if (stableStringify(existing.environment) !== stableStringify(environment)) throw new Error('acceptance_resume_environment_mismatch')
      return existing
    }
    const evidenceFingerprint = sha256(stableStringify(existing)).slice(0, 12)
    const archiveFilename = `${basename(options.outputPath, '.json')}.stale-${existing.candidate.fingerprint.slice(0, 12)}-${evidenceFingerprint}.json`
    const archivePath = join(dirname(options.outputPath), archiveFilename)
    const archiveContents = `${JSON.stringify(existing, null, 2)}\n`
    try {
      await writeFile(archivePath, archiveContents, { flag: 'wx' })
    } catch (error) {
      if (error.code !== 'EEXIST' || await readFile(archivePath, 'utf8') !== archiveContents) throw error
    }
    const state = {
      schema_version: 1,
      recorder: 'Cabinet packaged acceptance evidence',
      candidate,
      environment,
      rows: newRows(),
      overall_result: 'not_run',
      archived_prior_evidence: [...(existing.archived_prior_evidence ?? []), archiveFilename],
    }
    await writeState(options.outputPath, state)
    return state
  }
  const state = {
    schema_version: 1,
    recorder: 'Cabinet packaged acceptance evidence',
    candidate,
    environment,
    rows: newRows(),
    overall_result: 'not_run',
    archived_prior_evidence: [],
  }
  await writeState(options.outputPath, state)
  return state
}

const allowedTransitions = {
  not_run: new Set(['blocked', 'pass', 'fail']),
  blocked: new Set(['blocked', 'pass', 'fail']),
  fail: new Set(['fail', 'pass']),
  pass: new Set(['pass']),
}

export const recordAcceptanceResult = async ({
  state,
  rowId,
  status,
  evidenceReferences = [],
  operatorNotes = '',
  unblockCondition = '',
  operatorConfirmed = false,
}) => {
  validateAcceptanceState(state)
  const index = state.rows.findIndex((row) => row.id === rowId)
  if (index < 0) throw new Error('acceptance_row_unknown')
  if (!['blocked', 'pass', 'fail'].includes(status)) throw new Error('acceptance_status_invalid')
  const current = state.rows[index]
  if (!allowedTransitions[current.status].has(status)) throw new Error(`acceptance_status_transition_invalid:${current.status}:${status}`)
  if (!Array.isArray(evidenceReferences)) throw new Error('acceptance_evidence_reference_invalid')
  for (const reference of evidenceReferences) {
    if (!requiredText(reference) || redactAcceptanceText(reference) !== reference || /(^|[\\/])\.\.([\\/]|$)/.test(reference)) {
      throw new Error('acceptance_evidence_reference_sensitive')
    }
  }
  if (['pass', 'fail'].includes(status) && evidenceReferences.length === 0) throw new Error('acceptance_evidence_reference_required')
  if (['pass', 'fail'].includes(status) && !requiredText(operatorNotes)) throw new Error('acceptance_operator_notes_required')
  if (status === 'blocked' && !requiredText(unblockCondition)) throw new Error('acceptance_unblock_condition_required')
  if (status === 'pass' && current.requires_human_confirmation && operatorConfirmed !== true) throw new Error('acceptance_human_confirmation_required')
  const updated = {
    ...current,
    status,
    evidence_references: [...new Set(evidenceReferences)].sort(),
    operator_notes: redactAcceptanceText(operatorNotes),
    unblock_condition: status === 'blocked' ? redactAcceptanceText(unblockCondition) : '',
    operator_confirmed: operatorConfirmed === true,
  }
  if (stableStringify(updated) === stableStringify(current)) return state
  const next = { ...state, rows: state.rows.map((row, rowIndex) => rowIndex === index ? updated : row) }
  next.overall_result = calculateOverall(next.rows)
  return validateAcceptanceState(next)
}

const markdownCell = (value) => String(value).replaceAll('|', '\\|').replaceAll('\n', '<br>')

export const renderAcceptanceMarkdown = (state) => {
  validateAcceptanceState(state)
  const candidate = state.candidate
  const lines = [
    '# Cabinet packaged acceptance evidence',
    '',
    `Overall result: \`${state.overall_result}\``,
    '',
    '## Exact candidate',
    '',
    `- Source commit: \`${candidate.source_commit}\``,
    `- Candidate fingerprint: \`${candidate.fingerprint}\``,
    `- Candidate gate run/artifact: \`${candidate.release_candidate.run_id}\` / \`${candidate.release_candidate.artifact_name}\``,
    `- Cabinet: \`${candidate.cabinet.version}\`, \`${candidate.cabinet.package.filename}\`, SHA-256 \`${candidate.cabinet.package.sha256}\``,
    `- Browser Companion: \`${candidate.companion.version_name}\`, immutable tag \`${candidate.companion.immutable_tag}\``,
    ...candidate.companion.packages.map((item) => `  - ${item.target}: \`${item.filename}\`, SHA-256 \`${item.sha256}\``),
    `- Manifests: \`${candidate.cabinet.manifest_filename}\`, \`${candidate.companion.manifest_filename}\`, \`${candidate.combined_manifest_filename}\``,
    `- Release notes: \`${candidate.cabinet.release_notes_filename}\`, \`${candidate.companion.release_notes_filename}\``,
    '',
    '## Isolated environment',
    '',
    `- OS/host: ${state.environment.os_version} / ${state.environment.host_profile}`,
    `- Browser: ${state.environment.browser_name} ${state.environment.browser_version}`,
    `- Profile/data directory: ${state.environment.isolated_profile} / ${state.environment.isolated_data_directory.display} (SHA-256 \`${state.environment.isolated_data_directory.sha256}\`)`,
    `- Runtime: ${state.environment.runtime.app_version}, revision \`${state.environment.runtime.build_revision}\`, build ${state.environment.runtime.build_date}, port ${state.environment.runtime.port}, pid ${state.environment.runtime.pid}`,
    '',
  ]
  for (const section of [...new Set(state.rows.map((row) => row.section))]) {
    lines.push(`## ${section}`, '', '| ID | Status | Check | Evidence | Notes / exact unblock condition |', '| --- | --- | --- | --- | --- |')
    for (const row of state.rows.filter((item) => item.section === section)) {
      const notes = row.status === 'blocked' ? row.unblock_condition : row.operator_notes
      lines.push(`| ${row.id} | ${row.status} | ${markdownCell(row.title)} | ${markdownCell(row.evidence_references.join(', '))} | ${markdownCell(notes)} |`)
    }
    lines.push('')
  }
  lines.push('This recorder only captures evidence. It has no release, branch-promotion, provider-interaction, or browser-automation operation.', '')
  return lines.join('\n')
}

export const writeAcceptanceOutputs = async ({ state, jsonPath, markdownPath }) => {
  validateAcceptanceState(state)
  if (!requiredText(jsonPath) || !requiredText(markdownPath) || jsonPath === markdownPath) throw new Error('acceptance_output_paths_invalid')
  await writeState(jsonPath, state)
  await writeFile(markdownPath, renderAcceptanceMarkdown(state), { flag: 'w' })
}
