import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'

const root = new URL('..', import.meta.url).pathname.replace(
  /^\/([A-Za-z]:)/,
  '$1'
)
const readRepoFile = (path) => readFileSync(join(root, path), 'utf8')

test('README describes the shipped Windows portable beta and executable-local data boundary', () => {
  const readme = readRepoFile('README.md')
  const portableGuide = readRepoFile(
    'openspec/migration/windows-portable-beta.md'
  )

  assert.match(readme, /Cabinet 0\.1.+private beta/is)
  assert.match(readme, /Windows portable ZIP/i)
  assert.match(readme, /not an installer/i)
  assert.match(readme, /unsigned/i)
  assert.match(readme, /beside `cabinet\.exe`.+`data`/is)
  assert.match(readme, /CABINET_DATA_DIR.+override/is)
  assert.match(
    readme,
    /\/api\/runtime.+authoritative|authoritative.+\/api\/runtime/is
  )
  assert.match(readme, /Browser Companion.+Chrome.+Edge/is)
  assert.match(readme, /local.+ZITADEL/is)
  assert.match(readme, /diagnostics.+disabled by default/is)
  assert.match(
    readme,
    /\[Windows portable install, upgrade, rollback, and removal\]\(WINDOWS-PORTABLE-BETA\.md\)/i
  )
  assert.match(readme, /Help Center.+Integrations/is)
  assert.doesNotMatch(readme, /\]\(openspec\/migration\//i)
  assert.doesNotMatch(readme, /\]\(docs\/help-center\//i)
  assert.doesNotMatch(readme, /runtime scaffold implemented/i)
  assert.doesNotMatch(
    readme,
    /installer packaging workflow for Windows and macOS/i
  )
  assert.doesNotMatch(readme, /Windows:\s*`?%LOCALAPPDATA%\\+Cabinet/i)
  assert.doesNotMatch(readme, /Help Center section drafts/i)

  assert.match(portableGuide, /`data`.+beside `cabinet\.exe`/is)
  assert.match(portableGuide, /CABINET_DATA_DIR.+override/is)
  assert.match(portableGuide, /\/api\/runtime.+authoritative/is)
  assert.match(portableGuide, /backup.+outside.+extracted.+folder/is)
  assert.match(
    portableGuide,
    /deleting.+whole extracted folder.+deletes.+default.+data/is
  )
})

test('public privacy notice covers actual beta processing, retention, controls and support boundary', () => {
  const privacy = readRepoFile(
    'ui.web/src/features/auth/privacy-policy/index.tsx'
  )

  for (const pattern of [
    /local storage and data paths/i,
    /Browser Companion/i,
    /loopback/i,
    /supported page data/i,
    /cookies.+passwords.+tokens/is,
    /provider processing/i,
    /ZITADEL/i,
    /remote diagnostics/i,
    /disabled by default/i,
    /configured remote endpoint/i,
    /redact/i,
    /retention and deletion/i,
    /no fixed automatic retention/i,
    /JSON.+CSV.+backup/is,
    /beta coordinator/i,
    /do not include.+credentials|never include.+credentials/is,
  ]) {
    assert.match(privacy, pattern)
  }

  assert.doesNotMatch(
    privacy,
    /guarantee|certified|fully secure|compliant with all|delete within \d+/i
  )
})

test('public terms describe beta, provider, companion, auth, diagnostics and support limits', () => {
  const terms = readRepoFile(
    'ui.web/src/features/auth/terms-of-service/index.tsx'
  )

  for (const pattern of [
    /private beta/i,
    /Windows\s+portable/i,
    /third-party providers/i,
    /provider terms/i,
    /Browser Companion/i,
    /user-present/i,
    /no unattended crawling|must not use.+unattended/is,
    /ZITADEL/i,
    /diagnostics.+opt in/is,
    /no support service-level commitment/i,
    /beta coordinator/i,
  ]) {
    assert.match(terms, pattern)
  }

  assert.doesNotMatch(
    terms,
    /guarantee|certified|fully secure|compliant with all|governed by the laws/i
  )
})

test('Help Center publishes beta guidance without draft or plan labels', () => {
  const index = readRepoFile('docs/help-center/README.md')
  const articles = readRepoFile('ui.web/src/features/help-center/articles.ts')
  const gettingStarted = readRepoFile(
    'docs/help-center/getting-started/login-and-database-setup.md'
  )
  const settings = readRepoFile('docs/help-center/sections/settings.md')

  assert.match(index, /Cabinet Help Center/i)
  assert.match(index, /published private-beta guidance/i)
  assert.doesNotMatch(index, /\bdrafts?\b|Docs Plan/i)
  assert.doesNotMatch(articles, /Help Center Docs Plan|help-center-docs-plan/i)
  assert.match(articles, /About the Help Center/i)

  assert.match(gettingStarted, /local mode/i)
  assert.match(gettingStarted, /ZITADEL/i)
  assert.match(gettingStarted, /extracted folder.+`data`/is)
  assert.match(gettingStarted, /\/api\/runtime/)

  assert.match(settings, /remote diagnostics.+disabled by default/is)
  assert.match(settings, /no fixed automatic retention/i)
  assert.match(settings, /remote endpoint.+retention/is)
  assert.match(settings, /JSON.+CSV.+backup/is)
})

test('exact candidate release notes link source-bound user guidance', () => {
  const packageScript = readRepoFile('scripts/package-installers.ps1')

  assert.match(
    packageScript,
    /github\.com\/collectors-tech\/cabinet\/blob\/\$buildRevision/
  )
  assert.match(packageScript, /\$guidanceBaseURL\/README\.md/)
  assert.match(
    packageScript,
    /\$guidanceBaseURL\/openspec\/migration\/windows-portable-beta\.md/
  )
  assert.match(
    packageScript,
    /\$guidanceBaseURL\/docs\/help-center\/sections\/integrations\.md/
  )
  assert.match(packageScript, /Guidance supplied with this candidate/i)
  assert.match(
    packageScript,
    /README\.md.+WINDOWS-PORTABLE-BETA\.md.+Help Center.+Integrations/is
  )
})

test('OpenSpec and traceability bind the beta documentation alignment contract', () => {
  const spec = readRepoFile(
    'openspec/specs/general/documentation-governance/spec.md'
  )
  const traceability = readRepoFile('openspec/traceability.md')

  assert.match(spec, /DOCUMENTATION-GOVERNANCE-006/)
  assert.match(spec, /README.+privacy.+terms.+Help Center/is)
  for (const pattern of [
    /portable/i,
    /data path/i,
    /Browser Companion/i,
    /ZITADEL/i,
    /diagnostics/i,
    /provider/i,
    /retention/i,
    /deletion/i,
    /export/i,
    /support/i,
  ]) {
    assert.match(spec, pattern)
  }
  assert.match(
    traceability,
    /DOCUMENTATION-GOVERNANCE-006.+#2057.+beta-product-docs-contract\.test\.mjs/is
  )
})

test('GA roadmap keeps the minimum supported contract and release sequence explicit', () => {
  const roadmap = readRepoFile(
    'openspec/migration/cabinet-1.0-ga-roadmap.md'
  )

  assert.match(roadmap, /## Recommended minimum 1\.0 contract/)
  assert.match(roadmap, /proposed for owner approval.+#2546/is)
  assert.match(roadmap, /Voglers.+Hobbytech.+supported providers/is)
  assert.match(roadmap, /Frontline.+Bonza.+Preview/is)
  assert.match(roadmap, /Chat.+Agent.+Preview/is)
  assert.match(roadmap, /Telegram.+post-1\.0/is)
  assert.match(roadmap, /portable-only.+signed installer/is)

  assert.match(roadmap, /## Execution order after scope approval/)
  for (const issue of [
    '#2546',
    '#2057',
    '#1868',
    '#2034',
    '#1946',
    '#1869',
    '#1867',
    '#2488',
    '#1864',
  ]) {
    assert.match(roadmap, new RegExp(issue))
  }
  assert.match(roadmap, /two consecutive exact candidates/is)
  assert.match(roadmap, /Every code change invalidates the candidate/is)
  assert.match(roadmap, /separate explicit approval.+`main`/is)
})
