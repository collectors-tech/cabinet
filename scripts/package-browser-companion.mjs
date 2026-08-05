import { resolve } from 'node:path'

import { packageBrowserCompanion } from './lib/browser-companion-package.mjs'

const option = (name) => {
  const index = process.argv.indexOf(name)
  return index >= 0 ? process.argv[index + 1] : undefined
}

const sourceCommit = option('--source-commit')
const sourceDateEpoch = Number(option('--source-date-epoch'))
const outputDirectory = resolve(option('--output') ?? 'dist/browser-companion')
const keepStaging = process.argv.includes('--keep-staging')

const result = await packageBrowserCompanion({
  repositoryRoot: resolve('.'),
  outputDirectory,
  sourceCommit,
  sourceDateEpoch,
  keepStaging,
})

console.log(`Browser Companion release manifest: ${result.releaseManifestPath}`)
for (const artifact of result.releaseManifest.artifacts) console.log(`${artifact.target}: ${artifact.filename} ${artifact.sha256}`)
