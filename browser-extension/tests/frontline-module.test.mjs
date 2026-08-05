import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import vm from 'node:vm'

const fixture = JSON.parse(await readFile(new URL('../fixtures/frontline-search-v1.json', import.meta.url), 'utf8'))
const source = await readFile(new URL('../modules/frontlinehobbies.js', import.meta.url), 'utf8')

const fakeNode = (value = {}) => ({
  textContent: value.text ?? '',
  href: value.href ?? '',
  src: value.src ?? '',
  currentSrc: value.src ?? '',
  getAttribute: (name) => value[name] ?? value.attributes?.[name] ?? null,
  getClientRects: () => [{}],
})

const fakeCard = (value) => ({
  getAttribute: (name) => value.attributes?.[name] ?? null,
  querySelector: (selector) => value.nodes?.[selector] ? fakeNode(value.nodes[selector]) : null,
  getClientRects: () => [{}],
})

const fakeDocument = (example) => ({
  querySelector: (selector) => example.markers.includes(selector) ? fakeNode() : null,
  querySelectorAll: (selector) => selector === 'li.product, article.product, .product-item, [data-product-id], [data-product-sku], [itemtype="https://schema.org/Product"], [itemtype="http://schema.org/Product"]'
    ? example.cards.map(fakeCard)
    : [],
})

const loadCapture = () => {
  let hooks
  const listeners = []
  const context = vm.createContext({
    URL,
    console,
    chrome: { runtime: { onMessage: { addListener: (listener) => listeners.push(listener) } } },
    __cabinetFrontlineCaptureTestHooks: (value) => { hooks = value },
  })
  vm.runInContext(source, context, { filename: 'modules/frontlinehobbies.js' })
  return { hooks, listeners }
}

test('Frontline fixture v1 classifies ready, partial, challenge, login and selector drift fail closed', () => {
  assert.equal(fixture.fixture_version, '1')
  const { hooks, listeners } = loadCapture()
  assert.equal(typeof hooks?.capture, 'function')
  assert.equal(listeners.length, 1)

  for (const example of fixture.cases) {
    for (const card of example.cards) {
      assert.ok(hooks.productItem(fakeCard(card), new URL(example.url)), `${example.name}: card did not parse ${JSON.stringify(card)}`)
    }
    const result = hooks.capture(fakeDocument(example), new URL(example.url), example.query, example.pagination)
    assert.equal(result.state, example.expected.state, example.name)
    if (example.expected.error) {
      assert.equal(result.error, example.expected.error, example.name)
      assert.equal(result.data, undefined, example.name)
      continue
    }
    assert.equal(result.payload_type, 'search_results', example.name)
    assert.equal(result.page_complete, example.expected.complete, example.name)
    assert.equal(result.data.complete, example.expected.complete, example.name)
    assert.equal(result.data.items.length, example.expected.items, `${example.name}: ${JSON.stringify(result)}`)
    assert.equal(result.data.items.every((item) => item.currency === 'AUD' && item.url.startsWith('https://www.frontlinehobbies.com.au/')), true, example.name)
    assert.equal(result.data.items.every((item) => !item.url.includes('?') && !item.image_url.includes('?')), true, example.name)
  }
})

test('Frontline capture is passive, bounded and contains no session or remote-write mechanisms', () => {
  assert.doesNotMatch(source, /document\.(cookie|body)|localStorage|sessionStorage|fetch\s*\(|XMLHttpRequest|\.click\s*\(|submit\s*\(/)
  assert.match(source, /passive/)
  assert.match(source, /frontline_selector_drift/)
  assert.match(source, /frontline_challenge_action_required/)
  assert.match(source, /frontline_login_required/)
})
