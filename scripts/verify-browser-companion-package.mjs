import { resolve } from 'node:path'

import { verifyBrowserCompanionRelease } from './lib/browser-companion-verify.mjs'

const option = (name) => {
  const index = process.argv.indexOf(name)
  return index >= 0 ? process.argv[index + 1] : undefined
}

const manifest = option('--manifest')
if (!manifest) throw new Error('--manifest is required')
const release = await verifyBrowserCompanionRelease(resolve(manifest), {
  repositoryRoot: resolve('.'),
  expectedSourceCommit: option('--expected-source-commit'),
  previousManifestPath: option('--previous-manifest') ? resolve(option('--previous-manifest')) : undefined,
})
console.log(`Verified ${release.product} ${release.version_name} at ${release.source_commit}`)
