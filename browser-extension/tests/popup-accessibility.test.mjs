import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const root = new URL('../', import.meta.url)

test('popup exposes accessible zero, one and many module controls', async () => {
  const html = await readFile(new URL('popup/popup.html', root), 'utf8')
  const css = await readFile(new URL('popup/popup.css', root), 'utf8')
  assert.match(html, /<main/)
  assert.match(html, /aria-live="polite"/)
  assert.match(html, /id="module-list"/)
  assert.match(html, /<template[^>]+id="module-row"/)
  assert.match(html, /data-module-icon/)
  for (const action of ['setup', 'sync', 'pause', 'review']) {
    assert.match(html, new RegExp(`data-action="${action}"`))
  }
  assert.match(html, /Open Cabinet/)
  assert.match(html, /id="cabinet-url"/)
  assert.match(html, /id="sync-toggle"/)
  assert.match(css, /:focus-visible/)
  assert.match(css, /prefers-reduced-motion/)
})

test('privacy disclosure names optional origins and prohibited data', async () => {
  const privacy = await readFile(new URL('PRIVACY.md', root), 'utf8')
  assert.match(privacy, /optional site access/i)
  assert.match(privacy, /cookies/i)
  assert.match(privacy, /challenge/i)
  assert.match(privacy, /remove.*permission/i)
})
