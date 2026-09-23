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

  const publisher = read('.github/workflows/publish-ga-release.yml')
  assert.match(publisher, /APPROVE CABINET 1\.0 GA/)
  assert.match(publisher, /prerelease:\s*false/)
  assert.match(publisher, /ga_candidate_not_published/)
  assert.match(publisher, /Beta Release Candidate Gate/)
  assert.doesNotMatch(publisher, /private beta|private-beta/i)
})
