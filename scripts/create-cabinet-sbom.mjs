import { execFile } from 'node:child_process'
import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { promisify } from 'node:util'

import { createCabinetSBOM, parseGoBuildModules, verifyCabinetSBOM } from './lib/cabinet-sbom.mjs'

const execFileAsync = promisify(execFile)

function option(name) {
  const index = process.argv.indexOf(name)
  if (index < 0 || index === process.argv.length - 1 || process.argv[index + 1].startsWith('--')) throw new Error(`cabinet_sbom_option_missing:${name}`)
  return process.argv[index + 1]
}

function options(name) {
  const values = []
  for (let index = 0; index < process.argv.length; index += 1) {
    if (process.argv[index] === name && process.argv[index + 1] && !process.argv[index + 1].startsWith('--')) values.push(process.argv[index + 1])
  }
  return values
}

const version = option('--version')
const sourceCommit = option('--source-commit')
const buildDate = option('--build-date')
const outputPath = resolve(option('--output'))
const repositoryRoot = resolve(import.meta.dirname, '..')
const goBinaries = options('--go-binary').map((path) => resolve(path))
if (goBinaries.length === 0) throw new Error('cabinet_sbom_go_binary_missing')

const [{ stdout: goBuildInfo }, packageLockText] = await Promise.all([
  execFileAsync('go', ['version', '-m', ...goBinaries], {
    cwd: repositoryRoot,
    encoding: 'utf8',
    maxBuffer: 16 * 1024 * 1024,
  }),
  readFile(resolve(repositoryRoot, 'ui.web', 'package-lock.json'), 'utf8'),
])

const sbom = createCabinetSBOM({
  version,
  sourceCommit,
  buildDate,
  goModules: parseGoBuildModules(goBuildInfo),
  npmLock: JSON.parse(packageLockText),
})
verifyCabinetSBOM(sbom, { version, sourceCommit, buildDate })
await writeFile(outputPath, `${JSON.stringify(sbom, null, 2)}\n`, { encoding: 'utf8', flag: 'wx' })
console.log(`cabinet_sbom_created:${outputPath}`)
