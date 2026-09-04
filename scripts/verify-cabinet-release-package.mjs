import { resolve } from 'node:path'

import { verifyCabinetReleasePackage } from './lib/cabinet-release-verify.mjs'

const option = (name) => {
  const index = process.argv.indexOf(name)
  return index >= 0 ? process.argv[index + 1] : undefined
}

const manifest = option('--manifest')
if (!manifest) throw new Error('--manifest is required')
const release = await verifyCabinetReleasePackage(resolve(manifest), {
  repositoryRoot: resolve('.'),
  expectedSourceCommit: option('--expected-source-commit'),
})
console.log(`Verified Cabinet ${release.version} at ${release.source_commit}`)
