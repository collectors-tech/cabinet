import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { mkdtemp } from 'node:fs/promises'
import test from 'node:test'

import {
  packageBrowserCompanion,
  readStoreZip,
} from './lib/browser-companion-package.mjs'
import {
  scanPackagedFiles,
  verifyBrowserCompanionRelease,
} from './lib/browser-companion-verify.mjs'

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const sourceCommit = 'a'.repeat(40)
const sourceDateEpoch = 1_786_000_000

const build = async (name, commit = sourceCommit, epoch = sourceDateEpoch) => {
  const outputDirectory = await mkdtemp(join(tmpdir(), `cabinet-companion-${name}-`))
  const result = await packageBrowserCompanion({
    repositoryRoot,
    outputDirectory,
    sourceCommit: commit,
    sourceDateEpoch: epoch,
    keepStaging: true,
  })
  return { ...result, outputDirectory }
}

test('packages reproducible Chrome and Edge private-beta archives from one exact source', async () => {
  const first = await build('first')
  const second = await build('second')

  assert.equal(first.releaseManifest.source_commit, sourceCommit)
  assert.equal(first.releaseManifest.source_date_epoch, sourceDateEpoch)
  assert.equal(first.releaseManifest.version_name, '0.1.0-beta.7.gaaaaaaaaaaaa')
  assert.equal(first.releaseManifest.channel, 'private-beta')
  assert.deepEqual(first.releaseManifest.protocol_compatibility, { minimum: '1', maximum: '1' })
  assert.deepEqual(first.releaseManifest.artifacts.map((item) => item.target), ['chrome', 'edge'])
  assert.deepEqual(
    first.releaseManifest.artifacts.map((item) => item.sha256),
    second.releaseManifest.artifacts.map((item) => item.sha256),
  )

  for (const artifact of first.releaseManifest.artifacts) {
    assert.match(artifact.filename, /^cabinet-browser-companion-0\.1\.0-beta\.7\.ga{12}-(chrome|edge)\.zip$/)
    assert.match(artifact.sha256, /^[a-f0-9]{64}$/)
    assert.equal(artifact.sha256_filename, `${artifact.filename}.sha256`)
    assert.ok(artifact.files.every((file) => /^[a-f0-9]{64}$/.test(file.sha256)))
  }

  await verifyBrowserCompanionRelease(first.releaseManifestPath, { expectedSourceCommit: sourceCommit })
})

test('production archives are explicit, minimal, CSP-bound and distinct from development', async () => {
  const built = await build('manifest')
  const developmentManifest = JSON.parse(await readFile(join(repositoryRoot, 'browser-extension', 'manifest.json'), 'utf8'))
  assert.match(developmentManifest.name, /Development/)

  for (const artifact of built.releaseManifest.artifacts) {
    const archive = await readStoreZip(join(built.outputDirectory, artifact.filename))
    const manifest = JSON.parse(archive.get('manifest.json').toString('utf8'))
    const channel = JSON.parse(archive.get('release-channel.json').toString('utf8'))
    assert.equal(manifest.name, 'Cabinet Browser Companion')
    assert.equal(manifest.version, '0.1.0')
    assert.equal(manifest.version_name, '0.1.0-beta.7.gaaaaaaaaaaaa')
    assert.deepEqual(manifest.content_security_policy, { extension_pages: "script-src 'self'; object-src 'self'" })
    assert.deepEqual(manifest.host_permissions, ['http://127.0.0.1/*', 'http://localhost/*', 'http://[::1]/*'])
    assert.deepEqual(manifest.optional_host_permissions, ['https://*/*'])
    assert.equal(channel.target, artifact.target)
    assert.equal(channel.channel, 'private-beta')
    assert.equal(channel.source_commit, sourceCommit)
    assert.equal(channel.distribution, 'verified_zip_load_unpacked')
    assert.equal(channel.automatic_updates, false)
    assert.equal([...archive.keys()].some((path) => /(^|\/)(tests?|fixtures?)\/|\.map$|README/i.test(path)), false)
  }
})

test('verifier rejects secrets, challenge bypass, source maps and reused versions', async () => {
  assert.throws(
    () => scanPackagedFiles(new Map([['runtime/private.txt', Buffer.from('-----BEGIN PRIVATE KEY-----')]])),
    /private_key_material_rejected/,
  )
  assert.throws(
    () => scanPackagedFiles(new Map([['modules/provider.js', Buffer.from('document.cookie = atob(value)')]])),
    /provider_challenge_or_session_bypass_rejected/,
  )
  assert.throws(
    () => scanPackagedFiles(new Map([['background/service-worker.mjs.map', Buffer.from('{}')]])),
    /source_map_rejected/,
  )

  const previous = await build('previous', sourceCommit)
  const reused = await build('reused', sourceCommit, sourceDateEpoch + 2)
  await assert.rejects(
    () => verifyBrowserCompanionRelease(reused.releaseManifestPath, {
      expectedSourceCommit: sourceCommit,
      previousManifestPath: previous.releaseManifestPath,
    }),
    /immutable_version_reused/,
  )
})

test('candidate workflow validates and uploads exact companion evidence without publishing', async () => {
  const workflow = await readFile(join(repositoryRoot, '.github', 'workflows', 'beta-release-candidate.yml'), 'utf8')
  assert.match(workflow, /package-browser-companion\.mjs/)
  assert.match(workflow, /verify-browser-companion-package\.mjs/)
  assert.match(workflow, /browser-companion-release-manifest\.json/)
  assert.match(workflow, /cabinet-browser-companion-\*\.zip\.sha256/)
  assert.doesNotMatch(workflow, /create-release|upload-release-asset|webstore|edge.*publish/i)
})

test('packaged browser loader reads target roots without shadowing the Node process', async () => {
  const loader = await readFile(join(repositoryRoot, 'scripts', 'verify-browser-extension-load.mjs'), 'utf8')
  assert.match(loader, /process\.env\[browser\.rootVariable\]/)
  assert.match(loader, /const browserProcess = spawn\(/)
  assert.match(loader, /startupTimeoutMilliseconds/)
  assert.match(loader, /attempts/)
  assert.doesNotMatch(loader, /const process = spawn\(/)
})

test('packaged browser loader resolves installed Chrome and Edge on Windows', async () => {
  const loader = await readFile(join(repositoryRoot, 'scripts', 'verify-browser-extension-load.mjs'), 'utf8')
  assert.match(loader, /process\.platform === ["']win32["']/)
  assert.match(loader, /CABINET_CHROME_BIN/)
  assert.match(loader, /CABINET_EDGE_BIN/)
  assert.match(loader, /PROGRAMFILES/)
  assert.match(loader, /PROGRAMFILES\(X86\)/)
  assert.match(loader, /LOCALAPPDATA/)
  assert.match(loader, /existsSync/)
  assert.match(loader, /--edge-skip-compat-layer-relaunch/)
})

test('private-beta operator guide is truthful about install, permissions, rollback and removal', async () => {
  const guide = (await readFile(join(repositoryRoot, 'openspec', 'migration', 'browser-companion-private-beta-package-guide.md'), 'utf8')).toLowerCase()
  for (const required of [
    'not an installer',
    'sha-256',
    'load unpacked',
    'exact provider origin',
    'no automatic updates',
    'protocol compatibility',
    'rollback',
    'revoke',
    'uninstall',
    '#1864',
    '#1869',
  ]) assert.match(guide, new RegExp(required.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  assert.doesNotMatch(guide, /available (?:in|from) (?:the )?(?:chrome|edge) (?:web )?store/)
})
