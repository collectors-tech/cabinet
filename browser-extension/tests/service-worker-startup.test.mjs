import assert from 'node:assert/strict'
import test from 'node:test'

const never = new Promise(() => {})

test('MV3 worker handles popup state while Cabinet reconnect is still pending', async () => {
  let messageListener
  const stored = new Map()

  globalThis.chrome = {
    action: {
      setBadgeBackgroundColor: async () => {},
      setBadgeText: async () => {},
    },
    alarms: {
      create: async () => {},
      onAlarm: { addListener: () => {} },
    },
    permissions: {
      contains: async () => false,
      remove: async () => false,
      request: async () => false,
    },
    runtime: {
      getManifest: () => ({ version: '0.1.0-test' }),
      getURL: () => 'chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/',
      onInstalled: { addListener: () => {} },
      onMessage: { addListener: (listener) => { messageListener = listener } },
      onStartup: { addListener: () => {} },
    },
    scripting: { executeScript: async () => {} },
    storage: {
      local: {
        get: async () => never,
        remove: async (key) => { stored.delete(key) },
        set: async (values) => {
          for (const [key, value] of Object.entries(values)) stored.set(key, value)
        },
        setAccessLevel: async () => {},
      },
    },
    tabs: {
      create: async () => ({}),
      query: async () => [],
      sendMessage: async () => ({}),
      update: async () => ({}),
    },
  }
  globalThis.fetch = async () => never

  const workerStarted = import(new URL(
    `../background/service-worker.mjs?startup=${Date.now()}`,
    import.meta.url,
  ))
  const startedBeforeReconnect = await Promise.race([
    workerStarted.then(() => true),
    new Promise((resolve) => setTimeout(() => resolve(false), 100)),
  ])

  assert.equal(
    startedBeforeReconnect,
    true,
    'service-worker module activation must not await a pending Cabinet reconnect',
  )
  assert.equal(typeof messageListener, 'function')

  const response = await new Promise((resolve, reject) => {
    const keepPortOpen = messageListener({ type: 'host:get-state' }, {}, resolve)
    if (keepPortOpen !== true) reject(new Error('message_port_not_kept_open'))
  })
  assert.equal(response.ok, true)
  assert.equal(response.result.connection, 'disconnected')
})

test('MV3 restart preserves an outstanding pairing receipt', async () => {
  const stored = new Map([
    ['cabinet.companion.pending-pairing.v1', {
      request_id: 'pairing-restart',
      exchange_secret: 'exchange-restart',
      pairing_code: '654321',
    }],
  ])
  let messageListener
  globalThis.chrome.storage.local.get = async (key) => ({ [key]: stored.get(key) })
  globalThis.chrome.storage.local.set = async (values) => {
    for (const [key, value] of Object.entries(values)) stored.set(key, value)
  }
  globalThis.chrome.runtime.onMessage.addListener = (listener) => { messageListener = listener }

  await import(new URL(
    `../background/service-worker.mjs?pairing=${Date.now()}`,
    import.meta.url,
  ))
  await new Promise((resolve) => setTimeout(resolve, 20))

  const response = await new Promise((resolve) => {
    messageListener({ type: 'host:get-state' }, {}, resolve)
  })
  assert.equal(response.ok, true)
  assert.equal(response.result.connection, 'approval_required')
  assert.equal(response.result.pairing_code, '654321')
})
