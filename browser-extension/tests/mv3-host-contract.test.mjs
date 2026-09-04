import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const extensionRoot = new URL('../', import.meta.url)

const readJSON = async (path) =>
  JSON.parse(await readFile(new URL(path, extensionRoot), 'utf8'))

test('ships one installable Chromium MV3 host for Chrome and Edge', async () => {
  const manifest = await readJSON('manifest.json')
  const targets = await readJSON('targets.json')

  assert.equal(manifest.manifest_version, 3)
  assert.equal(manifest.background.type, 'module')
  assert.ok(manifest.background.service_worker)
  assert.equal(manifest.action.default_popup, 'popup/popup.html')
  assert.equal(manifest.side_panel.default_path, 'popup/popup.html')
  assert.equal(targets.source_manifest, 'manifest.json')
  assert.deepEqual(targets.browsers, ['chrome', 'edge'])
  assert.ok(manifest.optional_host_permissions.includes('https://*/*'))

  for (const path of [
    manifest.background.service_worker,
    manifest.action.default_popup,
    'platform/browser-api.mjs',
    'runtime/module-contract.mjs',
  ]) {
    const source = await readFile(new URL(path, extensionRoot), 'utf8')
    assert.ok(source.length > 0, `${path} is empty`)
  }

  const worker = await readFile(new URL(manifest.background.service_worker, extensionRoot), 'utf8')
  assert.doesNotMatch(worker, /globalThis\.(chrome|browser)/)
  assert.doesNotMatch(worker, /frontline|bonza|ebay|hobbytech/i)
  assert.match(worker, /browserPlatform/)
	assert.match(worker, /accepted\?\.committed !== true/)
	assert.match(worker, /\['completed', 'partial', 'review'\]/)
	assert.match(worker, /await normaliseCapture/)
})

test('popup provider projection is Cabinet-driven rather than hardcoded', async () => {
  const popup = await readFile(
    new URL('popup/popup-controller.mjs', extensionRoot),
    'utf8'
  )
  assert.doesNotMatch(popup, /frontline|bonza|ebay|hobbytech/i)
  assert.match(popup, /modules/i)
  assert.match(popup, /integration_instance_id/)
})
