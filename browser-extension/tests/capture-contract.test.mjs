import assert from 'node:assert/strict'
import test from 'node:test'

import { normaliseCapture } from '../runtime/capture-contract.mjs'

const module = {
  id: 'example-capture',
	module_version: '1.0.0',
	fixture_version: '1',
	provider_id: 'example',
	integration_instance_id: 'instance-1',
	redaction_rules: ['no_cookies', 'no_tokens', 'no_raw_page'],
  browser: { url_patterns: ['https://shop.example.test/items/*'] },
  capture_schemas: [{ payload_type: 'item_observation', fields: ['title', 'price'], media_fields: ['image_url'] }],
}

test('capture contract keeps typed allow-listed fields and creates a versioned digest', async () => {
  const result = await normaliseCapture(module, {
    payload_type: 'item_observation',
    confidence_score: 0.8,
	page_complete: true,
	data: { title: 'Example', price: 12.5, image_url: 'https://images.example.test/1.jpg' },
	}, 'https://shop.example.test/items/1?session=secret#private', 'profile-1', 'capture-1')

  assert.equal(result.url, 'https://shop.example.test/items/1')
  assert.deepEqual(result.data, { title: 'Example', price: 12.5, image_url: 'https://images.example.test/1.jpg' })
  assert.equal(result.passive, true)
  assert.equal(result.attempted_write, false)
	assert.equal(result.protocol_version, '1')
	assert.equal(result.integration_instance_id, 'instance-1')
	assert.equal(result.idempotency_key, 'capture-1')
	assert.match(result.payload_hash, /^sha256:[a-f0-9]{64}$/)
	assert.equal(result.page_complete, true)
})

test('capture contract rejects raw page, credential and undeclared fields recursively', async () => {
  for (const data of [
    { title: 'Example', raw_html: '<html>' },
    { title: 'Example', cookie: 'secret' },
    { title: 'Example', access_token: 'secret' },
	{ title: 'Example', price: { nested: { cookie: 'secret' } } },
  ]) {
	await assert.rejects(() => normaliseCapture(module, {
      payload_type: 'item_observation', confidence_score: 1, data,
	}, 'https://shop.example.test/items/1', 'profile-1', 'capture-2'), /field/)
  }
})

test('capture contract supports bounded arrays of typed item objects', async () => {
	const batchModule = { ...module, capture_schemas: [{ payload_type: 'search_results', fields: ['items'], media_fields: [] }] }
	const result = await normaliseCapture(batchModule, {
	  payload_type: 'search_results', confidence_score: 0.92, page_complete: false,
	  data: { items: [{ listing_id: 'A-1', title: 'First', price: 10.5, in_stock: true }] },
	}, 'https://shop.example.test/items/search', 'profile-1', 'capture-batch')
	assert.equal(result.data.items[0].listing_id, 'A-1')
	assert.equal(result.page_complete, false)
})
