import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'

const root = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1')

const readRepoFile = (path) => readFileSync(join(root, path), 'utf8')

test('Cabinet 0.1 beta disclosure has one governed source and required release boundaries', () => {
  const disclosurePath = join(root, 'release', 'cabinet-beta-disclosure.json')
  assert.equal(existsSync(disclosurePath), true)
  const disclosure = JSON.parse(readFileSync(disclosurePath, 'utf8'))

  assert.equal(disclosure.schema_version, 1)
  assert.equal(disclosure.product, 'Cabinet')
  assert.equal(disclosure.release_channel, 'private-beta')
  assert.equal(disclosure.release_version, '0.1.0-beta.2')
  assert.match(disclosure.generated_heading, /Cabinet 0\.1 private beta capabilities and limitations/)
  assert.ok(Array.isArray(disclosure.statements))

  const required = new Map([
    ['windows-portable-package', /portable ZIP.+not an installer/i],
    ['code-signing-installer-limit', /unsigned.+signed installer/i],
    ['browser-companion-targets', /Chrome.+Edge.+developer-mode|unpacked/i],
    ['browser-companion-updates', /no silent automatic browser-store updates/i],
    ['provider-voglers', /direct public storefront/i],
    ['provider-hobbytech', /packaged-unproven/i],
    ['provider-frontline', /browser-assisted.+action-required/i],
    ['provider-bonza', /browser-assisted.+action-required/i],
    ['auth-local-zitadel', /local account.+ZITADEL/i],
    ['assistant-agent-preview', /preview.+optional/i],
    ['telegram-live-limitation', /Telegram live acceptance.+not complete/i],
    ['post-beta-exclusions', /Metadata Studio.+public identity.+provider expansion.+eBay seller/i],
    ['data-ownership-export-recovery', /export.+backup.+restore.+not trapped behind a paid gate/i],
  ])

  const byID = new Map(disclosure.statements.map((statement) => [statement.id, statement]))
  for (const [id, pattern] of required) {
    const statement = byID.get(id)
    assert.ok(statement, `missing disclosure statement ${id}`)
    assert.equal(statement.channel, 'private-beta')
    assert.match(statement.status, /^(supported|preview|browser_assisted|action_required|unavailable|packaged_unproven|excluded|limited)$/)
    assert.match(statement.user_facing, pattern)
    assert.doesNotMatch(statement.user_facing, /C:\\|secret value|fixture|latest/i)
  }
})

test('release notes and Help Center projection are generated from the governed disclosure source', () => {
  const packageScript = readRepoFile('scripts/package-installers.ps1')
  const articles = readRepoFile('ui.web/src/features/help-center/articles.ts')
  const generator = readRepoFile('scripts/render-beta-disclosure.mjs')
  const releaseVerifier = readRepoFile('scripts/lib/cabinet-release-verify.mjs')

  assert.match(packageScript, /render-beta-disclosure\.mjs.+--format release-notes/s)
  assert.match(packageScript, /cabinet-beta-disclosure\.json/)
  assert.match(articles, /cabinet-private-beta-disclosure\.md\?raw/)
  assert.match(generator, /cabinet-beta-disclosure\.json/)
  assert.match(releaseVerifier, /verifyCabinetReleaseDisclosure/)
})
