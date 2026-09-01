import { createHash } from 'node:crypto'

const sourceCommitPattern = /^[0-9a-f]{40}$/
const shaHexLengths = new Map([
  ['SHA-256', 64],
  ['SHA-384', 96],
  ['SHA-512', 128],
])

function fail(code) {
  throw new Error(code)
}

function requireString(value, code) {
  if (typeof value !== 'string' || value.trim() === '') fail(code)
  return value
}

function validateIdentity({ version, sourceCommit, buildDate }) {
  requireString(version, 'cabinet_sbom_version_invalid')
  if (!sourceCommitPattern.test(sourceCommit ?? '')) fail('cabinet_sbom_source_commit_invalid')
  if (typeof buildDate !== 'string' || Number.isNaN(Date.parse(buildDate)) || !/T.*(?:Z|[+-]\d{2}:\d{2})$/.test(buildDate)) {
    fail('cabinet_sbom_build_date_invalid')
  }
}

function hashFromIntegrity(integrity, context) {
  if (typeof integrity !== 'string') fail(`cabinet_sbom_component_hash_missing:${context}`)
  const match = integrity.match(/^(sha256|sha384|sha512)-([A-Za-z0-9+/]+={0,2})(?:\s|$)/i)
  if (!match) fail(`cabinet_sbom_component_hash_invalid:${context}`)
  const algorithm = match[1].toUpperCase().replace('SHA', 'SHA-')
  let bytes
  try {
    bytes = Buffer.from(match[2], 'base64')
  } catch {
    fail(`cabinet_sbom_component_hash_invalid:${context}`)
  }
  const expectedLength = shaHexLengths.get(algorithm)
  const content = bytes.toString('hex')
  if (!expectedLength || content.length !== expectedLength) fail(`cabinet_sbom_component_hash_invalid:${context}`)
  return { alg: algorithm, content }
}

function hashFromGoSum(sum, context) {
  if (typeof sum !== 'string' || !sum.startsWith('h1:')) fail(`cabinet_sbom_component_hash_missing:${context}`)
  let bytes
  try {
    bytes = Buffer.from(sum.slice(3), 'base64')
  } catch {
    fail(`cabinet_sbom_component_hash_invalid:${context}`)
  }
  if (bytes.length !== 32) fail(`cabinet_sbom_component_hash_invalid:${context}`)
  return { alg: 'SHA-256', content: bytes.toString('hex') }
}

function encodePurlSegment(value) {
  return String(value).split('/').map((part) => encodeURIComponent(part)).join('/')
}

function goComponent(module) {
  const effective = module.Replace ?? module
  const path = requireString(effective.Path, 'cabinet_sbom_go_module_path_invalid')
  const version = requireString(effective.Version, `cabinet_sbom_go_module_version_missing:${path}`)
  const purl = `pkg:golang/${encodePurlSegment(path)}@${encodeURIComponent(version)}`
  return {
    type: 'library',
    'bom-ref': purl,
    name: path,
    version,
    purl,
    hashes: [hashFromGoSum(effective.Sum ?? module.Sum, purl)],
    properties: [{ name: 'cabinet:dependency_ecosystem', value: 'go' }],
  }
}

function npmPackageName(packagePath) {
  const marker = 'node_modules/'
  const index = packagePath.lastIndexOf(marker)
  if (index < 0) return null
  return packagePath.slice(index + marker.length)
}

function dependencyPackagePath(packages, parentPath, dependencyName) {
  let current = parentPath
  while (true) {
    const nested = current ? `${current}/node_modules/${dependencyName}` : `node_modules/${dependencyName}`
    if (packages[nested]) return nested
    const parentMarker = current.lastIndexOf('/node_modules/')
    if (parentMarker < 0) break
    current = current.slice(0, parentMarker)
  }
  const rootPath = `node_modules/${dependencyName}`
  return packages[rootPath] ? rootPath : null
}

function productionNpmPackagePaths(npmLock) {
  const packages = npmLock?.packages
  if (!packages || typeof packages !== 'object' || !packages['']) fail('cabinet_sbom_npm_lock_invalid')
  const pending = Object.keys(packages[''].dependencies ?? {}).sort().map((name) => ({ parentPath: '', name }))
  const selected = new Set()
  while (pending.length > 0) {
    const { parentPath, name } = pending.shift()
    const packagePath = dependencyPackagePath(packages, parentPath, name)
    if (!packagePath) fail(`cabinet_sbom_npm_dependency_missing:${name}`)
    if (selected.has(packagePath)) continue
    const entry = packages[packagePath]
    if (entry.dev === true) fail(`cabinet_sbom_npm_production_dependency_marked_dev:${name}`)
    selected.add(packagePath)
    for (const child of Object.keys(entry.dependencies ?? {}).sort()) pending.push({ parentPath: packagePath, name: child })
  }
  return [...selected]
}

function npmComponent(packagePath, entry) {
  const name = npmPackageName(packagePath)
  if (!name) fail(`cabinet_sbom_npm_package_path_invalid:${packagePath}`)
  const version = requireString(entry.version, `cabinet_sbom_npm_package_version_missing:${name}`)
  const purl = `pkg:npm/${encodePurlSegment(name)}@${encodeURIComponent(version)}`
  return {
    type: 'library',
    'bom-ref': purl,
    name,
    version,
    purl,
    hashes: [hashFromIntegrity(entry.integrity, purl)],
    properties: [{ name: 'cabinet:dependency_ecosystem', value: 'npm-production' }],
  }
}

function deduplicateAndSort(components) {
  const byPurl = new Map()
  for (const component of components) {
    const existing = byPurl.get(component.purl)
    if (existing && JSON.stringify(existing) !== JSON.stringify(component)) fail(`cabinet_sbom_component_conflict:${component.purl}`)
    byPurl.set(component.purl, component)
  }
  return [...byPurl.values()].sort((left, right) => left.purl.localeCompare(right.purl, 'en'))
}

export function createCabinetSBOM({ version, sourceCommit, buildDate, goModules, npmLock }) {
  validateIdentity({ version, sourceCommit, buildDate })
  if (!Array.isArray(goModules)) fail('cabinet_sbom_go_modules_invalid')
  const components = deduplicateAndSort([
    ...goModules.filter((module) => module && module.Main !== true).map(goComponent),
    ...productionNpmPackagePaths(npmLock).map((packagePath) => npmComponent(packagePath, npmLock.packages[packagePath])),
  ])
  if (components.length === 0) fail('cabinet_sbom_components_empty')

  const rootPurl = `pkg:github/collectors-tech/cabinet@${encodeURIComponent(version)}`
  const rootComponent = {
    type: 'application',
    'bom-ref': rootPurl,
    name: 'Cabinet',
    version,
    purl: rootPurl,
    properties: [
      { name: 'cabinet:source_commit', value: sourceCommit },
      { name: 'cabinet:target', value: 'windows-amd64' },
    ],
    externalReferences: [{
      type: 'vcs',
      url: `https://github.com/collectors-tech/cabinet/tree/${sourceCommit}`,
    }],
  }
  const dependencies = [
    { ref: rootPurl, dependsOn: components.map((component) => component['bom-ref']) },
    ...components.map((component) => ({ ref: component['bom-ref'], dependsOn: [] })),
  ]
  const sbom = {
    $schema: 'https://cyclonedx.org/schema/bom-1.7.schema.json',
    bomFormat: 'CycloneDX',
    specVersion: '1.7',
    version: 1,
    metadata: { timestamp: buildDate, component: rootComponent },
    components,
    dependencies,
  }
  return verifyCabinetSBOM(sbom, { version, sourceCommit, buildDate })
}

export function verifyCabinetSBOM(sbom, { version, sourceCommit, buildDate }) {
  validateIdentity({ version, sourceCommit, buildDate })
  if (!sbom || typeof sbom !== 'object' || Array.isArray(sbom)) fail('cabinet_sbom_invalid')
  if (sbom.$schema !== 'https://cyclonedx.org/schema/bom-1.7.schema.json' || sbom.bomFormat !== 'CycloneDX' || sbom.specVersion !== '1.7' || sbom.version !== 1) {
    fail('cabinet_sbom_format_mismatch')
  }
  if (sbom.metadata?.timestamp !== buildDate) fail('cabinet_sbom_build_date_mismatch')
  const root = sbom.metadata?.component
  if (root?.type !== 'application' || root?.name !== 'Cabinet' || root?.version !== version) fail('cabinet_sbom_root_identity_mismatch')
  const sourceProperty = root.properties?.find((item) => item?.name === 'cabinet:source_commit')
  if (sourceProperty?.value !== sourceCommit) fail('cabinet_sbom_source_commit_mismatch')
  const expectedRootPurl = `pkg:github/collectors-tech/cabinet@${encodeURIComponent(version)}`
  if (root['bom-ref'] !== expectedRootPurl || root.purl !== expectedRootPurl) fail('cabinet_sbom_root_purl_mismatch')

  if (!Array.isArray(sbom.components) || sbom.components.length === 0) fail('cabinet_sbom_components_empty')
  const seen = new Set()
  for (const component of sbom.components) {
    if (typeof component?.purl !== 'string' || component['bom-ref'] !== component.purl) fail('cabinet_sbom_component_identity_invalid')
    if (seen.has(component.purl)) fail(`cabinet_sbom_component_duplicate:${component.purl}`)
    seen.add(component.purl)
    if (!Array.isArray(component.hashes) || component.hashes.length !== 1) fail(`cabinet_sbom_component_hash_invalid:${component.purl}`)
    const hash = component.hashes[0]
    if (!shaHexLengths.has(hash?.alg) || !new RegExp(`^[0-9a-f]{${shaHexLengths.get(hash?.alg)}}$`).test(hash?.content ?? '')) {
      fail(`cabinet_sbom_component_hash_invalid:${component.purl}`)
    }
  }
  const sorted = [...seen].sort((left, right) => left.localeCompare(right, 'en'))
  if (JSON.stringify([...seen]) !== JSON.stringify(sorted)) fail('cabinet_sbom_components_not_sorted')
  if (!sbom.components.some((component) => component.purl.startsWith('pkg:golang/'))) fail('cabinet_sbom_go_components_missing')
  if (!sbom.components.some((component) => component.purl.startsWith('pkg:npm/'))) fail('cabinet_sbom_npm_components_missing')

  if (!Array.isArray(sbom.dependencies) || sbom.dependencies.length !== sbom.components.length + 1) fail('cabinet_sbom_dependencies_invalid')
  const rootDependency = sbom.dependencies[0]
  const componentRefs = sbom.components.map((component) => component['bom-ref'])
  if (rootDependency?.ref !== expectedRootPurl || JSON.stringify(rootDependency.dependsOn) !== JSON.stringify(componentRefs)) {
    fail('cabinet_sbom_root_dependencies_mismatch')
  }
  const declaredRefs = sbom.dependencies.slice(1).map((dependency) => dependency?.ref)
  if (JSON.stringify(declaredRefs) !== JSON.stringify(componentRefs) || sbom.dependencies.slice(1).some((dependency) => !Array.isArray(dependency.dependsOn) || dependency.dependsOn.length !== 0)) {
    fail('cabinet_sbom_component_dependencies_mismatch')
  }
  return sbom
}

export function parseGoBuildModules(text) {
  if (typeof text !== 'string' || text.trim() === '') fail('cabinet_sbom_go_build_info_invalid')
  const modules = []
  let previousDependency
  for (const line of text.split(/\r?\n/)) {
    const fields = line.replace(/^\s+/, '').split('\t')
    if (fields[0] === 'dep') {
      if (fields.length !== 4 || !fields[1] || !fields[2] || !fields[3]?.startsWith('h1:')) fail('cabinet_sbom_go_build_info_invalid')
      previousDependency = { Path: fields[1], Version: fields[2], Sum: fields[3] }
      modules.push(previousDependency)
    } else if (fields[0] === '=>') {
      if (!previousDependency || fields.length < 3 || fields.length > 4 || !fields[1] || !fields[2]) fail('cabinet_sbom_go_build_info_invalid')
      previousDependency.Replace = { Path: fields[1], Version: fields[2] }
      if (fields[3]) previousDependency.Replace.Sum = fields[3]
    }
  }
  if (modules.length === 0) fail('cabinet_sbom_go_build_info_invalid')
  return modules
}
