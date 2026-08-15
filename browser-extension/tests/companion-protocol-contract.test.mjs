import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  CompanionClient,
  CompanionProtocolError,
  companionStorageKeys,
} from '../src/companion-client.mjs'

const contract = JSON.parse(
  await readFile(
    new URL('../contracts/companion-protocol-v1.json', import.meta.url),
    'utf8'
  )
)

test('v1 supports the complete pairing and reconnect lifecycle', () => {
  assert.equal(contract.protocol_version, '1')
  assert.match(contract.base_url, /^http:\/\/(127\.0\.0\.1|localhost):\d+$/)

  for (const operation of [
    'pair',
    'approve',
    'exchange',
    'session',
    'rotate',
    'revoke',
    'modules',
    'captures',
    'media',
  ]) {
    assert.ok(contract.operations[operation], `missing ${operation} operation`)
  }
})

test('credentials are header-only and never use the legacy predictable scheme', () => {
  assert.equal(contract.authentication.transport, 'Authorization header')
  assert.deepEqual(contract.authentication.forbidden_transports, [
    'query',
    'fragment',
    'path',
  ])
  assert.equal(contract.authentication.credential_prefix, 'cabcmp_')
  assert.doesNotMatch(JSON.stringify(contract), /companion:<profile|access_token=/)

  for (const operation of Object.values(contract.operations)) {
    assert.doesNotMatch(operation.path, /token|credential|secret/i)
  }
})

test('authenticated discovery and sync use bounded versioned capabilities', () => {
  assert.deepEqual(contract.capabilities, [
    'modules:read',
    'captures:submit',
    'media:submit',
    'session:manage',
  ])
  assert.equal(contract.operations.modules.auth, true)
  assert.equal(contract.operations.captures.auth, true)
  assert.equal(contract.operations.media.auth, true)
  assert.ok(contract.limits.pairing_json_bytes <= 16 * 1024)
  assert.ok(contract.limits.capture_json_bytes <= 1024 * 1024)
  assert.ok(contract.limits.media_bytes <= 8 * 1024 * 1024)
})

test('client pairs, reconnects after restart, rotates and revokes without URL secrets', async () => {
  const stored = new Map()
  const storage = {
    get: async (key) => stored.get(key),
    set: async (key, value) => stored.set(key, value),
    delete: async (key) => stored.delete(key),
  }
  const server = { approved: false, credential: '', revoked: false }
  const requests = []
  const fetchImpl = async (url, init) => {
    requests.push({ url: url.toString(), ...init })
    const path = url.pathname
    const authorization = init.headers.Authorization
    if (path === '/api/companion/pairing/requests') {
      return Response.json(
        {
          request_id: 'pairing-1',
          exchange_secret: 'exchange-secret',
          pairing_code: '472981',
          status: 'pending',
          expires_at: '2026-08-06T12:10:00Z',
          protocol_version: '1',
          capabilities: ['modules:read'],
        },
        { status: 201 }
      )
    }
    if (path === '/api/companion/pairing/exchanges') {
      if (!server.approved) {
        return Response.json(
          { error: 'companion_pairing_exchange_invalid' },
          { status: 400 }
        )
      }
      server.credential = 'cabcmp_first'
      return Response.json({
        credential: server.credential,
        session: { id: 'session-1', protocol_version: '1' },
      })
    }
    if (
      authorization !== `Bearer ${server.credential}` ||
      server.revoked
    ) {
      return Response.json(
        { error: 'companion_auth_required' },
        { status: 401 }
      )
    }
    if (path === '/api/companion/session/rotate') {
      server.credential = 'cabcmp_second'
      return Response.json({
        credential: server.credential,
        session: { id: 'session-2', rotated_from_id: 'session-1' },
      })
    }
    if (path === '/api/companion/session' && init.method === 'DELETE') {
      server.revoked = true
      return Response.json({ revoked: true })
    }
    return Response.json({ id: 'session-1', protocol_version: '1' })
  }

  const createClient = () =>
    new CompanionClient({
      baseURL: contract.base_url,
      deviceID: 'chrome-windows-1',
      fetchImpl,
      storage,
    })
  const firstRuntime = createClient()
  await firstRuntime.startPairing('Chrome on Windows', ['modules:read'])
  await assert.rejects(
    firstRuntime.exchangePairing(),
    (error) =>
      error instanceof CompanionProtocolError &&
      error.code === 'companion_pairing_exchange_invalid'
  )
  assert.ok(await storage.get(companionStorageKeys.pendingPairing))

  server.approved = true
  await firstRuntime.exchangePairing()
  assert.equal(await storage.get(companionStorageKeys.credential), 'cabcmp_first')
  assert.equal(await storage.get(companionStorageKeys.pendingPairing), undefined)

  const restartedRuntime = createClient()
  assert.equal((await restartedRuntime.reconnect()).protocol_version, '1')
  assert.equal((await restartedRuntime.rotate()).rotated_from_id, 'session-1')
  assert.equal(await storage.get(companionStorageKeys.credential), 'cabcmp_second')
  await restartedRuntime.revoke()
  assert.equal(await storage.get(companionStorageKeys.credential), undefined)
  await assert.rejects(
    restartedRuntime.reconnect(),
    (error) =>
      error instanceof CompanionProtocolError &&
      error.code === 'companion_auth_required'
  )

  for (const request of requests) {
    assert.doesNotMatch(request.url, /cabcmp_|exchange-secret|credential=/)
    assert.equal(
      request.headers['X-Cabinet-Companion-Device'],
      'chrome-windows-1'
    )
  }
})

test('client invokes browser fetch without an illegal CompanionClient receiver', async () => {
  let receiver = 'not-called'
  const fetchImpl = async function () {
    receiver = this
    return Response.json({
      request_id: 'pairing-browser',
      exchange_secret: 'exchange-browser',
      pairing_code: '123456',
      status: 'pending',
      expires_at: '2026-08-15T12:10:00Z',
      protocol_version: '1',
      capabilities: ['modules:read'],
    }, { status: 201 })
  }
  const storage = {
    get: async () => undefined,
    set: async () => {},
    delete: async () => {},
  }
  const client = new CompanionClient({
    baseURL: contract.base_url,
    deviceID: 'brave-windows-1',
    fetchImpl,
    storage,
  })

  await client.startPairing('Brave on Windows', ['modules:read'])
  assert.equal(receiver, undefined)
})

test('client projects its runtime extension origin for authenticated Chromium GETs', async () => {
  const stored = new Map([[companionStorageKeys.credential, 'cabcmp_browser']])
  let request
  const client = new CompanionClient({
    baseURL: contract.base_url,
    deviceID: 'brave-windows-1',
    origin: 'chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    fetchImpl: async (url, init) => {
      request = { url, init }
      return Response.json({ id: 'session-browser', protocol_version: '1' })
    },
    storage: {
      get: async (key) => stored.get(key),
      set: async (key, value) => stored.set(key, value),
      delete: async (key) => stored.delete(key),
    },
  })

  await client.reconnect()
  assert.equal(
    request.init.headers['X-Cabinet-Companion-Origin'],
    'chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
  )
})

test('capture and media writes carry idempotent typed metadata and require committed acknowledgements', async () => {
  const stored = new Map([[companionStorageKeys.credential, 'cabcmp_test']])
  const storage = {
    get: async (key) => stored.get(key),
    set: async (key, value) => stored.set(key, value),
    delete: async (key) => stored.delete(key),
  }
  const requests = []
  const fetchImpl = async (url, init) => {
    requests.push({ url, init })
    if (url.pathname.endsWith('/payloads')) {
      return Response.json({ committed: true, state: 'completed', capture_id: 'capture-1' }, { status: 202 })
    }
    return Response.json({ committed: true, asset_id: 'asset-1', deduplicated: false }, { status: 201 })
  }
  const client = new CompanionClient({ baseURL: contract.base_url, deviceID: 'device-1', fetchImpl, storage })
  await client.submitCapture({ idempotency_key: 'capture-key', data: { cards: [] } })
  await client.submitMedia({
    bytes: new Uint8Array([1, 2, 3]), profileID: 'profile-1', captureID: 'capture-1', fieldName: 'image_url',
    filename: 'item.png', mimeType: 'image/png', sha256: 'a'.repeat(64), idempotencyKey: 'media-key',
  })

  assert.equal(requests[0].init.headers['X-Cabinet-Idempotency-Key'], 'capture-key')
  assert.equal(requests[1].init.headers['X-Cabinet-Capture-ID'], 'capture-1')
  assert.equal(requests[1].init.headers['X-Cabinet-Media-Field'], 'image_url')
  assert.equal(requests[1].init.headers['X-Cabinet-Media-Filename'], 'item.png')
  assert.equal(requests[1].init.headers['X-Cabinet-Media-SHA256'], 'a'.repeat(64))
  assert.equal(requests[1].init.body instanceof Uint8Array, true)
})
