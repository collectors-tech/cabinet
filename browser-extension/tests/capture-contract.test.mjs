import assert from 'node:assert/strict'
import test from 'node:test'

import { normaliseCapture } from '../runtime/capture-contract.mjs'

const module = {
  id: 'example-capture',
  browser: { url_patterns: ['https://shop.example.test/items/*'] },
  capture_schemas: [{ payload_type: 'item_observation', fields: ['title', 'price'], media_fields: ['image_url'] }],
}

test('capture contract keeps allow-listed fields and strips source URL secrets', () => {
  const result = normaliseCapture(module, {
    payload_type: 'item_observation',
    confidence_score: 0.8,
    data: { title: 'Example', price: 12.5, image_url: 'https://images.example.test/1.jpg' },
  }, 'https://shop.example.test/items/1?session=secret#private', 'profile-1')

  assert.equal(result.url, 'https://shop.example.test/items/1')
  assert.deepEqual(result.data, { title: 'Example', price: 12.5, image_url: 'https://images.example.test/1.jpg' })
  assert.equal(result.passive, true)
  assert.equal(result.attempted_write, false)
})

test('capture contract rejects raw page, credential and undeclared fields', () => {
  for (const data of [
    { title: 'Example', raw_html: '<html>' },
    { title: 'Example', cookie: 'secret' },
    { title: 'Example', access_token: 'secret' },
  ]) {
    assert.throws(() => normaliseCapture(module, {
      payload_type: 'item_observation', confidence_score: 1, data,
    }, 'https://shop.example.test/items/1', 'profile-1'), /field/)
  }
})
