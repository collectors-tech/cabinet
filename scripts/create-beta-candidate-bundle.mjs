import { resolve } from 'node:path'

import { createBetaCandidateBundle } from './lib/beta-candidate-bundle.mjs'

const option = (name) => {
  const index = process.argv.indexOf(name)
  return index >= 0 ? process.argv[index + 1] : undefined
}

const cabinetManifestPath = option('--cabinet-manifest')
const companionManifestPath = option('--companion-manifest')
const outputPath = option('--output')
const expectedSourceCommit = option('--expected-source-commit')
if (!cabinetManifestPath || !companionManifestPath || !outputPath || !expectedSourceCommit) {
  throw new Error('--cabinet-manifest, --companion-manifest, --output and --expected-source-commit are required')
}
const bundle = await createBetaCandidateBundle({
  cabinetManifestPath: resolve(cabinetManifestPath),
  companionManifestPath: resolve(companionManifestPath),
  outputPath: resolve(outputPath),
  expectedSourceCommit,
})
console.log(`Created exact private candidate bundle for ${bundle.source_commit}`)
