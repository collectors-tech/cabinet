import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'

const root = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1')
const read = (path) => readFileSync(join(root, path), 'utf8')

test('Cabinet 1.0 GA release controls have a canonical identity and guarded publisher', () => {
  const versionPath = join(root, 'release', 'cabinet-release-version.json')
  assert.equal(existsSync(versionPath), true)
  const version = JSON.parse(readFileSync(versionPath, 'utf8'))
  assert.deepEqual(version, { version: '1.0.0', channel: 'ga' })

  const packager = read('scripts/package-installers.ps1')
  assert.match(packager, /\$ReleaseChannel/)
  assert.match(packager, /cabinet-release-version\.json/)
  assert.match(packager, /ga_candidate_not_published/)

  const companionConfig = JSON.parse(readFileSync(join(root, 'browser-extension', 'release-ga.json'), 'utf8'))
  assert.deepEqual(companionConfig.channel, 'ga')
  assert.deepEqual(companionConfig.version, '1.0.0')
  assert.match(read('scripts/package-browser-companion.mjs'), /--release-channel/)
  assert.match(read('scripts/lib/beta-candidate-bundle.mjs'), /ga_candidate_not_published/)
  assert.match(read('scripts/create-ga-candidate-bundle.mjs'), /ga_candidate_bundle_identity_invalid/)

  const recorder = read('scripts/lib/acceptance-evidence-recorder.mjs')
  assert.match(recorder, /ga_candidate_not_published/)
  assert.match(recorder, /Cabinet 1\.0 GA candidate/)

  const disclosure = JSON.parse(readFileSync(join(root, 'release', 'cabinet-ga-disclosure.json'), 'utf8'))
  assert.deepEqual(disclosure.release_channel, 'ga')
  const providerStatuses = new Map(disclosure.statements.map((statement) => [statement.id, statement.status]))
  assert.deepEqual(providerStatuses.get('provider-voglers'), 'supported')
  assert.deepEqual(providerStatuses.get('provider-hobbytech'), 'packaged_unproven')
  assert.deepEqual(providerStatuses.get('provider-frontline-bonza'), 'browser_assisted')
  assert.deepEqual(providerStatuses.get('assistant-agent-telegram'), 'preview')

  const publisher = read('.github/workflows/publish-ga-release.yml')
  assert.match(publisher, /APPROVE CABINET 1\.0 GA/)
  assert.match(publisher, /prerelease:\s*false/)
  assert.match(publisher, /ga_candidate_not_published/)
  assert.match(publisher, /Cabinet 1\.0 GA Candidate Gate/)
  assert.doesNotMatch(publisher, /private beta|private-beta/i)

  const candidate = read('.github/workflows/ga-release-candidate.yml')
  assert.match(candidate, /Cabinet 1\.0 GA Candidate Gate/)
  assert.match(candidate, /-ReleaseChannel ga/)
  assert.match(candidate, /--release-channel ga/)
  assert.match(candidate, /create-ga-candidate-bundle\.mjs/)
  assert.match(candidate, /npm install -g @fission-ai\/openspec@latest/)
  assert.match(candidate, /path:\s*package-source/)
  assert.match(candidate, /working-directory:\s*package-source/)
  assert.match(candidate, /\$env:GITHUB_WORKSPACE\/dist\/cabinet/)
  assert.match(candidate, /sbom-path:\s*dist\/cabinet\/cabinet-1\.0\.0-sbom\.cdx\.json/)
})
