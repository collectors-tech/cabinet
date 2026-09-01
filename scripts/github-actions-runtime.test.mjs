import assert from 'node:assert/strict'
import { readdir, readFile } from 'node:fs/promises'
import test from 'node:test'

const workflowsDirectory = new URL('../.github/workflows/', import.meta.url)
const governedActions = new Map([
  ['checkout', { commit: '3d3c42e5aac5ba805825da76410c181273ba90b1', tag: 'v7.0.1' }],
  ['setup-node', { commit: '820762786026740c76f36085b0efc47a31fe5020', tag: 'v7.0.0' }],
  ['setup-go', { commit: 'b7ad1dad31e06c5925ef5d2fc7ad053ef454303e', tag: 'v7.0.0' }],
  ['upload-artifact', { commit: '043fb46d1a93c77aae656e7c1c64a875d1fc6a0a', tag: 'v7.0.1' }],
  ['download-artifact', { commit: '3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c', tag: 'v8.0.1' }],
])

test('governed first-party workflow actions use audited immutable Node 24 releases', async () => {
  const files = (await readdir(workflowsDirectory)).filter((file) => file.endsWith('.yml')).sort()
  const seen = new Set()
  const uses = /uses:\s*actions\/(checkout|setup-node|setup-go|upload-artifact|download-artifact)@([^\s#]+)(?:\s+#\s*(v[^\s]+))?/g

  for (const file of files) {
    const workflow = await readFile(new URL(file, workflowsDirectory), 'utf8')
    for (const match of workflow.matchAll(uses)) {
      const [, action, reference, releaseTag] = match
      const expected = governedActions.get(action)
      seen.add(action)
      assert.equal(reference, expected.commit, `${file}: actions/${action} must use the audited immutable commit`)
      assert.equal(releaseTag, expected.tag, `${file}: actions/${action} must retain the audited release tag comment`)
    }
  }

  assert.deepEqual([...seen].sort(), [...governedActions.keys()].sort())
})
