import { createHash } from 'node:crypto'
import { access, mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'

const zipLocalSignature = 0x04034b50
const zipCentralSignature = 0x02014b50
const zipEndSignature = 0x06054b50
const utf8Flag = 0x0800
const storedMethod = 0

const crcTable = (() => {
  const table = new Uint32Array(256)
  for (let index = 0; index < table.length; index += 1) {
    let value = index
    for (let bit = 0; bit < 8; bit += 1) value = (value & 1) ? (0xedb88320 ^ (value >>> 1)) : (value >>> 1)
    table[index] = value >>> 0
  }
  return table
})()

const crc32 = (buffer) => {
  let value = 0xffffffff
  for (const byte of buffer) value = crcTable[(value ^ byte) & 0xff] ^ (value >>> 8)
  return (value ^ 0xffffffff) >>> 0
}

export const sha256Buffer = (buffer) => createHash('sha256').update(buffer).digest('hex')

const stableJSON = (value) => Buffer.from(`${JSON.stringify(value, null, 2)}\n`, 'utf8')

const dosTimestamp = (epoch) => {
  const date = new Date(Number(epoch) * 1000)
  const year = Math.max(1980, Math.min(2107, date.getUTCFullYear()))
  const time = ((date.getUTCHours() & 0x1f) << 11) | ((date.getUTCMinutes() & 0x3f) << 5) | ((Math.floor(date.getUTCSeconds() / 2)) & 0x1f)
  const day = ((year - 1980) << 9) | (((date.getUTCMonth() + 1) & 0x0f) << 5) | (date.getUTCDate() & 0x1f)
  return { time, day }
}

export const writeStoreZip = (inputEntries, epoch) => {
  const entries = [...inputEntries.entries()]
    .map(([path, raw]) => ({ path, name: Buffer.from(path, 'utf8'), data: Buffer.from(raw) }))
    .sort((left, right) => left.path.localeCompare(right.path))
  const localParts = []
  const centralParts = []
  const { time, day } = dosTimestamp(epoch)
  let offset = 0

  for (const entry of entries) {
    const checksum = crc32(entry.data)
    const local = Buffer.alloc(30)
    local.writeUInt32LE(zipLocalSignature, 0)
    local.writeUInt16LE(20, 4)
    local.writeUInt16LE(utf8Flag, 6)
    local.writeUInt16LE(storedMethod, 8)
    local.writeUInt16LE(time, 10)
    local.writeUInt16LE(day, 12)
    local.writeUInt32LE(checksum, 14)
    local.writeUInt32LE(entry.data.length, 18)
    local.writeUInt32LE(entry.data.length, 22)
    local.writeUInt16LE(entry.name.length, 26)
    local.writeUInt16LE(0, 28)
    localParts.push(local, entry.name, entry.data)

    const central = Buffer.alloc(46)
    central.writeUInt32LE(zipCentralSignature, 0)
    central.writeUInt16LE(0x0314, 4)
    central.writeUInt16LE(20, 6)
    central.writeUInt16LE(utf8Flag, 8)
    central.writeUInt16LE(storedMethod, 10)
    central.writeUInt16LE(time, 12)
    central.writeUInt16LE(day, 14)
    central.writeUInt32LE(checksum, 16)
    central.writeUInt32LE(entry.data.length, 20)
    central.writeUInt32LE(entry.data.length, 24)
    central.writeUInt16LE(entry.name.length, 28)
    central.writeUInt16LE(0, 30)
    central.writeUInt16LE(0, 32)
    central.writeUInt16LE(0, 34)
    central.writeUInt16LE(0, 36)
    central.writeUInt32LE(0x81a40000, 38)
    central.writeUInt32LE(offset, 42)
    centralParts.push(central, entry.name)
    offset += local.length + entry.name.length + entry.data.length
  }

  const centralDirectory = Buffer.concat(centralParts)
  const end = Buffer.alloc(22)
  end.writeUInt32LE(zipEndSignature, 0)
  end.writeUInt16LE(0, 4)
  end.writeUInt16LE(0, 6)
  end.writeUInt16LE(entries.length, 8)
  end.writeUInt16LE(entries.length, 10)
  end.writeUInt32LE(centralDirectory.length, 12)
  end.writeUInt32LE(offset, 16)
  end.writeUInt16LE(0, 20)
  return Buffer.concat([...localParts, centralDirectory, end])
}

export const readStoreZip = async (archivePath) => {
  const archive = await readFile(archivePath)
  let endOffset = -1
  for (let offset = archive.length - 22; offset >= Math.max(0, archive.length - 65_557); offset -= 1) {
    if (archive.readUInt32LE(offset) === zipEndSignature) {
      endOffset = offset
      break
    }
  }
  if (endOffset < 0) throw new Error('zip_end_record_missing')
  const entryCount = archive.readUInt16LE(endOffset + 10)
  let centralOffset = archive.readUInt32LE(endOffset + 16)
  const result = new Map()
  for (let index = 0; index < entryCount; index += 1) {
    if (archive.readUInt32LE(centralOffset) !== zipCentralSignature) throw new Error('zip_central_record_invalid')
    const method = archive.readUInt16LE(centralOffset + 10)
    const checksum = archive.readUInt32LE(centralOffset + 16)
    const size = archive.readUInt32LE(centralOffset + 24)
    const nameLength = archive.readUInt16LE(centralOffset + 28)
    const extraLength = archive.readUInt16LE(centralOffset + 30)
    const commentLength = archive.readUInt16LE(centralOffset + 32)
    const localOffset = archive.readUInt32LE(centralOffset + 42)
    const name = archive.subarray(centralOffset + 46, centralOffset + 46 + nameLength).toString('utf8')
    if (method !== storedMethod || archive.readUInt32LE(localOffset) !== zipLocalSignature) throw new Error(`zip_entry_method_invalid:${name}`)
    const localNameLength = archive.readUInt16LE(localOffset + 26)
    const localExtraLength = archive.readUInt16LE(localOffset + 28)
    const dataOffset = localOffset + 30 + localNameLength + localExtraLength
    const data = Buffer.from(archive.subarray(dataOffset, dataOffset + size))
    if (data.length !== size || crc32(data) !== checksum) throw new Error(`zip_entry_checksum_invalid:${name}`)
    if (result.has(name)) throw new Error(`zip_entry_duplicate:${name}`)
    result.set(name, data)
    centralOffset += 46 + nameLength + extraLength + commentLength
  }
  return result
}

const ensureAbsent = async (path) => {
  try {
    await access(path)
    throw new Error(`output_already_exists:${path}`)
  } catch (error) {
    if (error?.code !== 'ENOENT') throw error
  }
}

const writeStaging = async (root, entries) => {
  await ensureAbsent(root)
  for (const [path, data] of entries) {
    const target = join(root, path)
    await mkdir(dirname(target), { recursive: true })
    await writeFile(target, data, { flag: 'wx' })
  }
}

const validateReleaseConfig = (config) => {
  if (config?.schema_version !== 1 || !/^\d+\.\d+\.\d+$/.test(config.version) ||
      !/^\d+\.\d+\.\d+-[a-z0-9.-]+$/.test(config.version_name_prefix) || config.channel !== 'private-beta' ||
      !Array.isArray(config.targets) || config.targets.join(',') !== 'chrome,edge' ||
      !Array.isArray(config.package_files) || new Set(config.package_files).size !== config.package_files.length ||
      config.package_files.some((path) => !path || path.startsWith('/') || path.includes('\\') ||
        path.split('/').includes('..') || path === 'manifest.json' || path === 'release-channel.json')) {
    throw new Error('browser_companion_release_config_invalid')
  }
}

export const packageBrowserCompanion = async ({
  repositoryRoot = resolve('.'),
  outputDirectory = resolve('dist/browser-companion'),
  sourceCommit,
  sourceDateEpoch,
  keepStaging = false,
} = {}) => {
  if (!/^[a-f0-9]{40}$/i.test(String(sourceCommit ?? ''))) throw new Error('source_commit_must_be_full_sha')
  if (!Number.isInteger(Number(sourceDateEpoch)) || Number(sourceDateEpoch) <= 315_532_800) throw new Error('source_date_epoch_invalid')
  const extensionRoot = join(repositoryRoot, 'browser-extension')
  const releaseConfig = JSON.parse(await readFile(join(extensionRoot, 'release.json'), 'utf8'))
  const developmentManifest = JSON.parse(await readFile(join(extensionRoot, 'manifest.json'), 'utf8'))
  validateReleaseConfig(releaseConfig)
  const candidateVersion = `${releaseConfig.version_name_prefix}.g${sourceCommit.toLowerCase().slice(0, 12)}`
  if (!/Development/.test(developmentManifest.name) || developmentManifest.version_name === candidateVersion) {
    throw new Error('development_channel_not_distinct')
  }

  const releaseManifestPath = join(outputDirectory, 'browser-companion-release-manifest.json')
  await ensureAbsent(releaseManifestPath)
  await mkdir(outputDirectory, { recursive: true })
  const sourceFiles = new Map()
  for (const path of releaseConfig.package_files) sourceFiles.set(path, await readFile(join(extensionRoot, path)))

  const artifacts = []
  for (const target of releaseConfig.targets) {
    const filename = `cabinet-browser-companion-${candidateVersion}-${target}.zip`
    const archivePath = join(outputDirectory, filename)
    const checksumPath = `${archivePath}.sha256`
    await ensureAbsent(archivePath)
    await ensureAbsent(checksumPath)
    const {
      key: _key,
      update_url: _updateURL,
      version_name: _developmentVersionName,
      ...manifestBase
    } = developmentManifest
    const manifest = {
      ...manifestBase,
      name: 'Cabinet Browser Companion',
      version: releaseConfig.version,
      version_name: candidateVersion,
      permissions: releaseConfig.permissions,
      host_permissions: releaseConfig.host_permissions,
      optional_host_permissions: releaseConfig.optional_host_permissions,
      content_security_policy: { extension_pages: "script-src 'self'; object-src 'self'" },
    }
    const channel = {
      schema_version: 1,
      product: 'Cabinet Browser Companion',
      target,
      channel: releaseConfig.channel,
      version: releaseConfig.version,
      version_name: candidateVersion,
      source_commit: sourceCommit.toLowerCase(),
      protocol_compatibility: releaseConfig.protocol_compatibility,
      distribution: releaseConfig.distribution,
      automatic_updates: releaseConfig.automatic_updates,
    }
    const entries = new Map(sourceFiles)
    entries.set('manifest.json', stableJSON(manifest))
    entries.set('release-channel.json', stableJSON(channel))
    const archive = writeStoreZip(entries, Number(sourceDateEpoch))
    const archiveHash = sha256Buffer(archive)
    await writeFile(archivePath, archive, { flag: 'wx' })
    await writeFile(checksumPath, `${archiveHash}  ${filename}\n`, { flag: 'wx' })
    if (keepStaging) await writeStaging(join(outputDirectory, '.staging', target), entries)
    artifacts.push({
      target,
      filename,
      sha256_filename: `${filename}.sha256`,
      sha256: archiveHash,
      size_bytes: archive.length,
      permissions: [...manifest.permissions],
      host_permissions: [...manifest.host_permissions],
      optional_host_permissions: [...manifest.optional_host_permissions],
      files: [...entries.entries()].sort(([left], [right]) => left.localeCompare(right)).map(([path, data]) => ({
        path,
        size_bytes: data.length,
        sha256: sha256Buffer(data),
      })),
    })
  }

  const releaseNotesFilename = `cabinet-browser-companion-${candidateVersion}-release-notes.md`
  const releaseNotesTemplate = await readFile(join(extensionRoot, 'RELEASE_NOTES.md'), 'utf8')
  const releaseNotes = `Source commit: \`${sourceCommit.toLowerCase()}\`  \nChannel: \`${releaseConfig.channel}\`  \nProtocol: \`${releaseConfig.protocol_compatibility.minimum}\` through \`${releaseConfig.protocol_compatibility.maximum}\`\n\n${releaseNotesTemplate}`
  await writeFile(join(outputDirectory, releaseNotesFilename), releaseNotes, { flag: 'wx' })
  const releaseManifest = {
    schema_version: 1,
    product: 'Cabinet Browser Companion',
    channel: releaseConfig.channel,
    version: releaseConfig.version,
    version_name: candidateVersion,
    immutable_tag: `${releaseConfig.immutable_tag_prefix}${candidateVersion}`,
    source_commit: sourceCommit.toLowerCase(),
    source_date_epoch: Number(sourceDateEpoch),
    created_at: new Date(Number(sourceDateEpoch) * 1000).toISOString(),
    protocol_compatibility: releaseConfig.protocol_compatibility,
    distribution: releaseConfig.distribution,
    automatic_updates: releaseConfig.automatic_updates,
    publication_state: 'private_candidate_not_published',
    release_notes_filename: releaseNotesFilename,
    artifacts,
  }
  await writeFile(releaseManifestPath, stableJSON(releaseManifest), { flag: 'wx' })
  const summary = [
    '# Cabinet Browser Companion candidate', '',
    `Source commit: ${releaseManifest.source_commit}`,
    `Version: ${releaseManifest.version_name}`,
    `Channel: ${releaseManifest.channel}`,
    'Release manifest: browser-companion-release-manifest.json', '',
    ...artifacts.map((artifact) => `- ${artifact.target}: ${artifact.filename} — ${artifact.sha256}`), '',
    'This build is a private candidate. It does not publish a store listing, release or mutable latest URL.', '',
  ].join('\n')
  await writeFile(join(outputDirectory, 'browser-companion-candidate-summary.md'), summary, { flag: 'wx' })
  return { releaseManifest, releaseManifestPath, outputDirectory }
}
