import assert from 'node:assert/strict'
import test from 'node:test'

import { normaliseCabinetURL } from '../runtime/config.mjs'

test('configuration accepts only credential-free loopback Cabinet addresses', () => {
  assert.equal(normaliseCabinetURL('http://127.0.0.1:17880/settings?x=1'), 'http://127.0.0.1:17880/')
  assert.equal(normaliseCabinetURL('http://localhost:8080/'), 'http://localhost:8080/')
  assert.equal(normaliseCabinetURL('http://[::1]:17880/'), 'http://[::1]:17880/')
  assert.throws(() => normaliseCabinetURL('https://cabinet.example/'), /loopback/)
  assert.throws(() => normaliseCabinetURL('http://user:pass@127.0.0.1:8080/'), /loopback/)
})
