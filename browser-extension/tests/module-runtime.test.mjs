import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

import {
  normaliseRegistry,
  permissionOrigins,
} from '../runtime/module-contract.mjs'
import { classifyReadiness } from '../runtime/readiness.mjs'

const instance = (overrides = {}) => ({
  id: 'example-observation',
  module_version: '1.0.0',
  site: 'Example Store',
  provider_id: 'example',
  integration_instance_id: 'instance-1',
  actions: ['capture_item'],
  passive_only: true,
  capture_schemas: [{ payload_type: 'item_observation', fields: ['title', 'price'], media_fields: ['image_url'] }],
  workflows: ['manual_item_capture'],
  redaction_rules: ['no_cookies', 'no_tokens', 'no_raw_page'],
  fixture_version: '1',
  display: { name: 'Example Store', icon: 'icons/provider.svg' },
  browser: {
    start_url: 'https://shop.example.test/search',
    origins: ['https://shop.example.test/*'],
    url_patterns: ['https://shop.example.test/search*'],
    readiness: {
      ready: ['[data-account-menu]'],
      logged_out: ['a[href*="login"]'],
      challenge: ['[data-challenge]'],
    },
  },
  configuration: {
    capture_mode: 'manual_user_present',
    item_fields: ['title', 'price'],
    media_policy: 'review_before_canonical_persistence',
    review_destination: 'discoveries',
    rate_limit_per_minute: 6,
    help_url: '/help-center/integrations',
    setup_required: false,
    sync_available: false,
  },
  safe_config: { region: 'AU' },
  ...overrides,
})

test('normalises a Cabinet registry without provider-specific extension code', () => {
  const result = normaliseRegistry({
    protocol_version: '1',
    profile_id: 'profile-1',
    modules: [instance()],
  })

  assert.equal(result.modules.length, 1)
  assert.equal(result.modules[0].integration_instance_id, 'instance-1')
  assert.equal(result.modules[0].status, 'permission_required')
  assert.equal(result.modules[0].fixture_version, '1')
  assert.equal(result.modules[0].configuration.capture_mode, 'manual_user_present')
  assert.deepEqual(permissionOrigins(result.modules[0]), [
    'https://shop.example.test/*',
  ])
  assert.equal(Object.keys(result.modules[0].safe_config).some((key) => /cookie|token|secret/i.test(key)), false)
})

test('rejects unsafe or non-passive module definitions', () => {
  assert.throws(
    () => normaliseRegistry({ protocol_version: '1', modules: [instance({ passive_only: false })] }),
    /passive/
  )
  assert.throws(
    () => normaliseRegistry({
      protocol_version: '1',
      modules: [instance({ browser: { origins: ['http://shop.example.test/*'] } })],
    }),
    /https/
  )
  assert.throws(
    () => normaliseRegistry({
      protocol_version: '1',
      modules: [instance({ browser: { start_url: 'https://shop.example.test/search', origins: ['https://*.example.test/*'] } })],
    }),
    /exact/
  )
  assert.throws(
    () => normaliseRegistry({
      protocol_version: '1',
      modules: [instance({ safe_config: { access_token: 'do-not-project' } })],
    }),
    /secret-like/
  )
})

test('readiness is bounded and distinguishes ready, logged out and challenge pages', () => {
  const definition = instance().browser.readiness
  assert.equal(classifyReadiness(definition, ['[data-account-menu]']).state, 'ready')
  assert.equal(classifyReadiness(definition, ['a[href*="login"]']).state, 'logged_out')
  assert.equal(classifyReadiness(definition, ['[data-challenge]']).state, 'action_required')
  assert.equal(classifyReadiness(definition, []).state, 'unsupported')
  assert.equal(
    classifyReadiness(definition, ['[data-account-menu]', '[data-challenge]']).state,
    'action_required'
  )
})

test('versioned readiness fixtures cover every truthful page state', async () => {
  const fixture = JSON.parse(await readFile(new URL('../fixtures/readiness-v1.json', import.meta.url), 'utf8'))
  assert.equal(fixture.fixture_version, '1')
  for (const example of fixture.cases) {
    assert.equal(classifyReadiness(fixture.definition, example.evidence).state, example.expected, example.name)
  }
})

test('readiness bridge returns evidence identifiers without page or session data', async () => {
  const bridge = await readFile(new URL('../content/provider-bridge.js', import.meta.url), 'utf8')
  assert.doesNotMatch(bridge, /document\.(cookie|body)|innerHTML|outerHTML|localStorage|sessionStorage|location\.href/)
  assert.match(bridge, /matched/)
})
