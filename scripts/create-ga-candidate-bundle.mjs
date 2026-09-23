import { resolve } from 'node:path'

import { createCandidateBundle } from './lib/beta-candidate-bundle.mjs'

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
const bundle = await createCandidateBundle({
  cabinetManifestPath: resolve(cabinetManifestPath),
  companionManifestPath: resolve(companionManifestPath),
  outputPath: resolve(outputPath),
  expectedSourceCommit,
})
if (bundle.channel !== 'ga' || bundle.publication_state !== 'ga_candidate_not_published') {
  throw new Error('ga_candidate_bundle_identity_invalid')
}
console.log(`Created exact GA candidate bundle for ${bundle.source_commit}`)
