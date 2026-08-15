import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { mkdtemp, readFile, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import test from 'node:test'

import { writeStoreZip } from './lib/browser-companion-package.mjs'
import { createBetaCandidateBundle } from './lib/beta-candidate-bundle.mjs'
import { verifyCabinetReleasePackage } from './lib/cabinet-release-verify.mjs'
import { loadBetaDisclosure, renderBetaDisclosureMarkdown } from './render-beta-disclosure.mjs'

const sourceCommit = 'c'.repeat(40)
const sha256 = (value) => createHash('sha256').update(value).digest('hex')

const cabinetFixture = async () => {
  const directory = await mkdtemp(join(tmpdir(), 'cabinet-release-package-'))
  const filename = 'cabinet-0.1.0-beta.3-windows-amd64-portable.zip'
  const checksumFilename = `${filename}.sha256`
  const notesFilename = 'cabinet-0.1.0-beta.3-release-notes.md'
  const packageEntries = new Map([
    ['README.md', Buffer.from('r')],
    ['WINDOWS-PORTABLE-BETA.md', Buffer.from('w')],
    ['cabinet-mcp.exe', Buffer.from('m')],
    ['cabinet.exe', Buffer.from('c')],
  ])
  const archive = writeStoreZip(packageEntries, 1_786_000_000)
  const checksum = sha256(archive)
  const disclosure = await loadBetaDisclosure(resolve('release/cabinet-beta-disclosure.json'))
  const disclosureNotes = renderBetaDisclosureMarkdown(disclosure, { format: 'release-notes' })
  await writeFile(join(directory, filename), archive)
  await writeFile(join(directory, checksumFilename), `${checksum}  ${filename}\n`)
  await writeFile(join(directory, notesFilename), `# Cabinet 0.1.0-beta.3\n\nCommit: ${sourceCommit}\n\nWindows portable package; not an installer.\n\n${disclosureNotes}`)
  const manifest = {
    schema_version: 1,
    product: 'Cabinet',
    channel: 'private-beta',
    version: '0.1.0-beta.3',
    source_commit: sourceCommit,
    build_date: '2026-08-06T00:00:00Z',
    publication_state: 'private_candidate_not_published',
    artifact: { target: 'windows-amd64', kind: 'portable_zip', filename, sha256_filename: checksumFilename, sha256: checksum, size_bytes: archive.length },
    release_notes_filename: notesFilename,
    package_files: [...packageEntries].map(([path, data]) => ({ path, size_bytes: data.length, sha256: sha256(data) })),
  }
  const manifestPath = join(directory, 'cabinet-release-manifest.json')
  await writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`)
  return { directory, manifest, manifestPath, archivePath: join(directory, filename) }
}

test('verifies exact Cabinet package identity, checksum, notes and required file inventory', async () => {
  const fixture = await cabinetFixture()
  const verified = await verifyCabinetReleasePackage(fixture.manifestPath, {
    repositoryRoot: resolve('.'),
    expectedSourceCommit: sourceCommit,
  })
  assert.equal(verified.source_commit, sourceCommit)

  await writeFile(fixture.archivePath, 'tampered')
  await assert.rejects(
    () => verifyCabinetReleasePackage(fixture.manifestPath, { repositoryRoot: resolve('.'), expectedSourceCommit: sourceCommit }),
    /cabinet_artifact_checksum_mismatch/,
  )
})

test('creates one source-bound non-publishing Cabinet and companion bundle', async () => {
  const fixture = await cabinetFixture()
  const companionPath = join(fixture.directory, 'browser-companion-release-manifest.json')
  await writeFile(companionPath, `${JSON.stringify({
    schema_version: 1,
    product: 'Cabinet Browser Companion',
    channel: 'private-beta',
    version_name: '0.1.0-beta.3.gcccccccccccc',
    source_commit: sourceCommit,
    publication_state: 'private_candidate_not_published',
    protocol_compatibility: { minimum: '1', maximum: '1' },
    release_notes_filename: 'cabinet-browser-companion-notes.md',
    artifacts: [
      { target: 'chrome', filename: 'chrome.zip', sha256_filename: 'chrome.zip.sha256', sha256: '1'.repeat(64) },
      { target: 'edge', filename: 'edge.zip', sha256_filename: 'edge.zip.sha256', sha256: '2'.repeat(64) },
    ],
  }, null, 2)}\n`)
  const outputPath = join(fixture.directory, 'beta-candidate-bundle-manifest.json')
  const bundle = await createBetaCandidateBundle({
    cabinetManifestPath: fixture.manifestPath,
    companionManifestPath: companionPath,
    outputPath,
    expectedSourceCommit: sourceCommit,
  })
  assert.equal(bundle.source_commit, sourceCommit)
  assert.equal(bundle.publication_state, 'private_candidate_not_published')
  assert.deepEqual(bundle.components.map((component) => component.product), ['Cabinet', 'Cabinet Browser Companion'])
  assert.deepEqual(JSON.parse(await readFile(outputPath, 'utf8')), bundle)
})
