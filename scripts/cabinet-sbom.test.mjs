import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import test from 'node:test'

import { createCabinetSBOM, parseGoBuildModules, verifyCabinetSBOM } from './lib/cabinet-sbom.mjs'

const sourceCommit = 'a'.repeat(40)
const buildDate = '2026-09-01T00:00:00Z'
const version = '1.0.0'
const integrity = (algorithm, byte, size) => `${algorithm}-${Buffer.alloc(size, byte).toString('base64')}`
const goSum = (byte) => `h1:${Buffer.alloc(32, byte).toString('base64')}`

const goModules = [
  { Path: 'github.com/collectors-tech/cabinet', Main: true, Dir: 'C:/repo' },
  {
    Path: 'github.com/example/runtime',
    Version: 'v1.2.3',
    Sum: goSum(1),
  },
]

const npmLock = {
  name: 'cabinet-ui-web',
  version: '2.2.1',
  lockfileVersion: 3,
  packages: {
    '': {
      name: 'cabinet-ui-web',
      version: '2.2.1',
      dependencies: { react: '19.2.3' },
      devDependencies: { vite: '7.3.6' },
    },
    'node_modules/react': {
      version: '19.2.3',
      integrity: integrity('sha512', 2, 64),
    },
    'node_modules/vite': {
      version: '7.3.6',
      dev: true,
      integrity: integrity('sha512', 3, 64),
    },
  },
}

const create = () => createCabinetSBOM({ version, sourceCommit, buildDate, goModules, npmLock })

test('creates deterministic CycloneDX 1.7 from Go and production npm dependencies', () => {
  const first = create()
  const second = create()
  assert.deepEqual(first, second)
  assert.equal(`${JSON.stringify(first, null, 2)}\n`, `${JSON.stringify(second, null, 2)}\n`)
  assert.equal(first.bomFormat, 'CycloneDX')
  assert.equal(first.specVersion, '1.7')
  assert.match(first.serialNumber, /^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
  const differentSource = createCabinetSBOM({ version, sourceCommit: 'b'.repeat(40), buildDate, goModules, npmLock })
  assert.notEqual(differentSource.serialNumber, first.serialNumber)
  assert.equal(first.metadata.timestamp, buildDate)
  assert.equal(first.metadata.component.version, version)
  assert.equal(first.metadata.component.properties.find((item) => item.name === 'cabinet:source_commit')?.value, sourceCommit)
  assert.deepEqual(first.components.map((item) => item.purl), [
    'pkg:golang/github.com/example/runtime@v1.2.3',
    'pkg:npm/react@19.2.3',
  ])
  assert.equal(first.components.some((item) => item.purl.includes('/vite@')), false)
  assert.equal(first.components[0].hashes[0].alg, 'SHA-256')
  assert.equal(first.components[1].hashes[0].alg, 'SHA-512')
  assert.equal(first.dependencies[0].ref, first.metadata.component['bom-ref'])
  assert.deepEqual(first.dependencies[0].dependsOn, first.components.map((item) => item['bom-ref']))
  assert.deepEqual(verifyCabinetSBOM(first, { version, sourceCommit, buildDate }), first)
})

test('deduplicates repeated package identities and rejects malformed or drifted evidence', () => {
  const duplicateLock = structuredClone(npmLock)
  duplicateLock.packages['node_modules/example/node_modules/react'] = structuredClone(duplicateLock.packages['node_modules/react'])
  const deduplicated = createCabinetSBOM({ version, sourceCommit, buildDate, goModules, npmLock: duplicateLock })
  assert.equal(deduplicated.components.filter((item) => item.purl === 'pkg:npm/react@19.2.3').length, 1)

  const wrongSource = structuredClone(create())
  wrongSource.metadata.component.properties.find((item) => item.name === 'cabinet:source_commit').value = 'b'.repeat(40)
  assert.throws(() => verifyCabinetSBOM(wrongSource, { version, sourceCommit, buildDate }), /cabinet_sbom_source_commit_mismatch/)

  const duplicate = structuredClone(create())
  duplicate.components.push(structuredClone(duplicate.components[0]))
  assert.throws(() => verifyCabinetSBOM(duplicate, { version, sourceCommit, buildDate }), /cabinet_sbom_component_duplicate/)

  const unsorted = structuredClone(create())
  unsorted.components.reverse()
  assert.throws(() => verifyCabinetSBOM(unsorted, { version, sourceCommit, buildDate }), /cabinet_sbom_components_not_sorted/)

  const malformedHash = structuredClone(create())
  malformedHash.components[0].hashes[0].content = createHash('sha256').update('short').digest('hex').slice(1)
  assert.throws(() => verifyCabinetSBOM(malformedHash, { version, sourceCommit, buildDate }), /cabinet_sbom_component_hash_invalid/)

  const missingSerial = structuredClone(create())
  delete missingSerial.serialNumber
  assert.throws(() => verifyCabinetSBOM(missingSerial, { version, sourceCommit, buildDate }), /cabinet_sbom_serial_number_mismatch/)
})

test('rejects invalid generator inputs instead of emitting partial inventory', () => {
  assert.throws(
    () => createCabinetSBOM({ version, sourceCommit: 'short', buildDate, goModules, npmLock }),
    /cabinet_sbom_source_commit_invalid/,
  )
  const missingProductionGraph = structuredClone(npmLock)
  delete missingProductionGraph.packages['node_modules/react']
  assert.throws(
    () => createCabinetSBOM({ version, sourceCommit, buildDate, goModules, npmLock: missingProductionGraph }),
    /cabinet_sbom_npm_dependency_missing:react/,
  )
})

test('parses the stable go version -m format used by the repository toolchain', () => {
  const buildInfo = [
    'cabinet.exe: go1.24.0',
    '\tpath\tgithub.com/collectors-tech/cabinet/cmd/cabinet',
    '\tmod\tgithub.com/collectors-tech/cabinet\t(devel)\t',
    `\tdep\tgithub.com/example/runtime\tv1.2.3\t${goSum(1)}`,
    '\tbuild\tGOOS=windows',
    'cabinet-mcp.exe: go1.24.0',
    '\tpath\tgithub.com/collectors-tech/cabinet/cmd/cabinet-mcp',
    `\tdep\tgithub.com/example/runtime\tv1.2.3\t${goSum(1)}`,
    `\tdep\tgithub.com/example/old\tv1.0.0\t${goSum(2)}`,
    `\t=>\tgithub.com/example/new\tv1.0.1\t${goSum(3)}`,
  ].join('\n')
  assert.deepEqual(parseGoBuildModules(buildInfo), [
    { Path: 'github.com/example/runtime', Version: 'v1.2.3', Sum: goSum(1) },
    { Path: 'github.com/example/runtime', Version: 'v1.2.3', Sum: goSum(1) },
    {
      Path: 'github.com/example/old',
      Version: 'v1.0.0',
      Sum: goSum(2),
      Replace: { Path: 'github.com/example/new', Version: 'v1.0.1', Sum: goSum(3) },
    },
  ])
  assert.throws(() => parseGoBuildModules('cabinet.exe: go1.24.0\n\tdep\tmissing-version'), /cabinet_sbom_go_build_info_invalid/)
})
