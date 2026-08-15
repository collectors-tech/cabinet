import assert from 'node:assert/strict'
import test from 'node:test'

class FakeElement {
  constructor() {
    this.disabled = false
    this.hidden = false
    this.listeners = new Map()
    this.textContent = ''
    this.value = ''
  }

  addEventListener(type, listener) { this.listeners.set(type, listener) }
  append() {}
  replaceChildren() {}
}

test('Connect retains its button across the asynchronous state lookup', async () => {
  const elements = new Map([
    ['#announcement', new FakeElement()],
    ['#cabinet-detail', new FakeElement()],
    ['#cabinet-url', new FakeElement()],
    ['#connect', new FakeElement()],
    ['#connection', new FakeElement()],
    ['#empty-state', new FakeElement()],
    ['#error', new FakeElement()],
    ['#module-list', new FakeElement()],
    ['#open-cabinet', new FakeElement()],
    ['#refresh', new FakeElement()],
    ['#settings-form', new FakeElement()],
    ['#sync-toggle', new FakeElement()],
    ['#module-row', new FakeElement()],
  ])
  globalThis.document = { querySelector: (selector) => elements.get(selector) }

  const calls = []
  const state = {
    cabinet_url: 'http://127.0.0.1:19010/',
    connection: 'pairing_required',
    extension_version: '0.1.0-test',
    modules: [],
    profile_id: '',
  }
  globalThis.chrome = {
    runtime: {
      onInstalled: { addListener: () => {} },
      onMessage: { addListener: () => {} },
      onStartup: { addListener: () => {} },
      sendMessage: async (message) => {
        calls.push(message.type)
        if (message.type === 'host:start-pairing') {
          return { ok: true, result: { ...state, connection: 'approval_required', pairing_code: '123456' } }
        }
        return { ok: true, result: state }
      },
    },
    storage: { local: {} },
  }

  await import(new URL(`../popup/popup-controller.mjs?connect=${Date.now()}`, import.meta.url))
  const connect = elements.get('#connect')
  const event = { currentTarget: connect }
  const click = connect.listeners.get('click')(event)
  event.currentTarget = null
  await click

  assert.deepEqual(calls, ['host:get-state', 'host:get-state', 'host:start-pairing'])
  assert.equal(connect.disabled, false)
  assert.equal(elements.get('#connection').textContent, 'approval required')
  assert.equal(connect.textContent, 'Finish pairing')
})
