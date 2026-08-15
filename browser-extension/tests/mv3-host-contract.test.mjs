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

test('service worker answers host state while startup reconnect is pending', async () => {
  const previousChrome = globalThis.chrome
  const previousFetch = globalThis.fetch
  const listeners = { onMessage: [], onStartup: [], onInstalled: [], onAlarm: [] }
  const storage = {
    'cabinet.companion.credential.v1': 'stored-credential',
    'cabinet.companion.host-state.v1': {
      cabinet_url: 'http://127.0.0.1:19010/',
      connection: 'connected',
      modules: [],
    },
  }
  const addListener = (name) => (listener) => listeners[name].push(listener)

  globalThis.chrome = {
    action: {
      setBadgeBackgroundColor: (_details) => {},
      setBadgeText: (_details) => {},
    },
    alarms: {
      create: (_name, _options) => {},
      onAlarm: { addListener: addListener('onAlarm') },
    },
    permissions: {
      contains: () => false,
      remove: () => true,
      request: () => true,
    },
    runtime: {
      getManifest: () => ({ version: '0.0.0-test' }),
      onInstalled: { addListener: addListener('onInstalled') },
      onMessage: { addListener: addListener('onMessage') },
      onStartup: { addListener: addListener('onStartup') },
    },
    scripting: { executeScript: () => {} },
    storage: {
      local: {
        get: (key) => ({ [key]: storage[key] }),
        remove: (key) => { delete storage[key] },
        set: (items) => Object.assign(storage, items),
        setAccessLevel: (_details) => {},
      },
    },
    tabs: {
      create: (_details) => ({}),
      query: (_query) => [],
      sendMessage: (_tabID, _message) => ({}),
      update: (_tabID, _update) => ({}),
    },
  }
  globalThis.fetch = () => new Promise(() => {})

  try {
    await Promise.race([
      import(`../background/service-worker.mjs?mv3-startup-${Date.now()}`),
      new Promise((_, reject) => setTimeout(() => reject(new Error('service_worker_startup_blocked_by_reconnect')), 150)),
    ])

    assert.equal(listeners.onMessage.length, 1)
    const response = await new Promise((resolve) => {
      const keepChannelOpen = listeners.onMessage[0]({ type: 'host:get-state' }, {}, resolve)
      assert.equal(keepChannelOpen, true)
    })
    assert.equal(response.ok, true)
    assert.equal(response.result.cabinet_url, 'http://127.0.0.1:19010/')
    assert.equal(response.result.connection, 'connected')
  } finally {
    if (previousChrome === undefined) {
      delete globalThis.chrome
    } else {
      globalThis.chrome = previousChrome
    }
    globalThis.fetch = previousFetch
  }
})
